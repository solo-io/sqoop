package core

import (
	"github.com/solo-io/qloo/pkg/bootstrap"
	"github.com/solo-io/gloo/pkg/bootstrap/configstorage"
	"github.com/pkg/errors"
	qloobootstrap "github.com/solo-io/qloo/pkg/bootstrap"
	"github.com/solo-io/qloo/pkg/configwatcher"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/qloo/pkg/api/types/v1"
	"github.com/solo-io/qloo/pkg/operator"
	"github.com/solo-io/qloo/pkg/graphql"
	"github.com/solo-io/qloo/pkg/storage"
	"github.com/solo-io/qloo/pkg/reporter"
	"github.com/solo-io/qloo/pkg/resolvers"
	"github.com/solo-io/qloo/pkg/exec"
	"github.com/vektah/gqlgen/neelance/schema"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/qloo/pkg/util"
	"fmt"
	"net/http"
)

type EventLoop struct {
	cfgWatcher configwatcher.Interface
	operator   *operator.GlooOperator
	router     *graphql.Router
	qloo       storage.Interface
	reporter   reporter.Interface
	proxyAddr  string
	bindAddr   string
}

func Setup(opts bootstrap.Options) (*EventLoop, error) {
	gloo, err := configstorage.Bootstrap(opts.Options)
	if err != nil {
		return nil, errors.Wrap(err, "creating gloo client")
	}
	qloo, err := qloobootstrap.Bootstrap(opts.Options)
	if err != nil {
		return nil, errors.Wrap(err, "creating qloo client")
	}
	if err := gloo.V1().Register(); err != nil {
		return nil, errors.Wrap(err, "registering gloo client")
	}
	if err := qloo.V1().Register(); err != nil {
		return nil, errors.Wrap(err, "registering qloo storage client")
	}
	cfgWatcher, err := configwatcher.NewConfigWatcher(qloo)
	if err != nil {
		return nil, errors.Wrap(err, "starting watch for QLoo config")
	}
	op := operator.NewGlooOperator(gloo, opts.VirtualServiceName, opts.RoleName)
	router := graphql.NewRouter()
	rep := reporter.NewReporter(qloo)
	return &EventLoop{
		cfgWatcher: cfgWatcher,
		operator:   op,
		router:     router,
		qloo:       qloo,
		reporter:   rep,
		proxyAddr:  opts.ProxyAddr,
		bindAddr:   opts.BindAddr,
	}, nil
}

func (el *EventLoop) Run(stop <-chan struct{}) {
	go el.cfgWatcher.Run(stop)
	go func(){
		log.Fatalf("failed to start server: %v", http.ListenAndServe(el.bindAddr, el.router))
	}()
	errs := make(chan error)
	for {
		select {
		case cfg := <-el.cfgWatcher.Config():
			if err := el.update(cfg); err != nil {
				errs <- errors.Wrap(err, "update failed")
			}
		case err := <-el.cfgWatcher.Error():
			errs <- errors.Wrap(err, "config watcher error")
		case err := <-errs:
			log.Warnf("error in event loop: %v", err)
		}
	}
}

func (el *EventLoop) update(cfg *v1.Config) error {
	endpoints, reports := el.createGraphqlEndpoints(cfg)
	el.router.UpdateEndpoints(endpoints...)
	return el.reporter.WriteReports(reports)
}

func (el *EventLoop) createGraphqlEndpoints(cfg *v1.Config) ([]*graphql.Endpoint, []reporter.ConfigObjectReport) {
	var (
		endpoints          []*graphql.Endpoint
		schemaReports      []reporter.ConfigObjectReport
		resolverMapReports []reporter.ConfigObjectReport
	)
	resolverMapErrs := make(map[*v1.ResolverMap]error)

	for _, schema := range cfg.Schemas {
		schemaReport := reporter.ConfigObjectReport{
			CfgObject: schema,
		}
		// empty map means we should generate a skeleton and update the schema to point to it
		ep, schemaErr, resolverMapErr := el.handleSchema(schema, cfg.ResolverMaps)
		if schemaErr != nil {
			resolverMapErr.err = multierror.Append(resolverMapErr.err, errors.Wrap(schemaErr, "schema was not accepted"))
		}
		if resolverMapErr.resolverMap != nil {
			err := resolverMapErrs[resolverMapErr.resolverMap]
			if resolverMapErr.err != nil {
				err = multierror.Append(resolverMapErrs[resolverMapErr.resolverMap], resolverMapErr.err)
			}
			resolverMapErrs[resolverMapErr.resolverMap] = err
		}
		schemaReport.Err = schemaErr
		schemaReports = append(schemaReports, schemaReport)
		if ep == nil {
			continue
		}
		endpoints = append(endpoints, ep)
	}
	for resolverMap, err := range resolverMapErrs {
		resolverMapReports = append(resolverMapReports, reporter.ConfigObjectReport{
			CfgObject: resolverMap,
			Err:       err,
		})
	}
	return endpoints, append(schemaReports, resolverMapReports...)
}

type resolverMapError struct {
	resolverMap *v1.ResolverMap
	err         error
}

func (el *EventLoop) handleSchema(schema *v1.Schema, resolvers []*v1.ResolverMap) (*graphql.Endpoint, error, resolverMapError) {
	if schema.ResolverMap == "" {
		return nil, el.createEmptyResolverMap(schema), resolverMapError{}
	}
	for _, resolverMap := range resolvers {
		if resolverMap.Name == schema.ResolverMap {
			ep, schemaErr, resolverErr := el.createGraphqlEndpoint(schema, resolverMap)
			return ep, schemaErr, resolverMapError{resolverMap: resolverMap, err: resolverErr}
		}
	}
	return nil, errors.Errorf("resolver map %v for schema %v not found", schema.ResolverMap, schema.Name), resolverMapError{}
}

// create an empty resolver map and
func (el *EventLoop) createEmptyResolverMap(schema *v1.Schema) error {
	resolverName := resolverMapName(schema)
	parsedSchema, err := parseSchemaString(schema)
	if err != nil {
		return errors.Wrap(err, "failed to parse schema")
	}
	generatedResolvers := util.GenerateResolverMapSkeleton(resolverName, parsedSchema)

	// update existing schema with the new schema name
	// important to do this first or we may retry creating the resolver map in a race
	schemaToUpdate, err := el.qloo.V1().Schemas().Get(schema.Name)
	if err != nil {
		return errors.Wrapf(err, "retrieving schema %v from storage", schema.Name)
	}
	schemaToUpdate.ResolverMap = resolverName
	if _, err := el.qloo.V1().Schemas().Update(schemaToUpdate); err != nil {
		return errors.Wrapf(err, "updating schema %v in storage", schema.Name)
	}

	if _, err := el.qloo.V1().ResolverMaps().Create(generatedResolvers); err != nil {
		return errors.Wrapf(err, "writing resolver map %v to storage", resolverName)
	}
	return nil
}

func (el *EventLoop) createGraphqlEndpoint(schema *v1.Schema, resolverMap *v1.ResolverMap) (*graphql.Endpoint, error, error) {
	resolverFactory := resolvers.NewResolverFactory(el.proxyAddr, resolverMap)
	parsedSchema, err := parseSchemaString(schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse schema"), nil
	}
	executableResolvers, err := exec.NewExecutableResolvers(parsedSchema, resolverFactory.CreateResolver)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate resolvers from map")
	}
	executableSchema := exec.NewExecutableSchema(parsedSchema, executableResolvers)
	return &graphql.Endpoint{
		SchemaName: schema.Name,
		RootPath:   "/" + schema.Name,
		QueryPath:  "/" + schema.Name + "/query",
		ExecSchema: executableSchema,
	}, nil, nil
}

func parseSchemaString(sch *v1.Schema) (*schema.Schema, error) {
	parsedSchema := schema.New()
	return parsedSchema, parsedSchema.Parse(sch.InlineSchema)
}

func resolverMapName(schema *v1.Schema) string {
	return fmt.Sprintf("%v-resolvers", schema.Name)
}
