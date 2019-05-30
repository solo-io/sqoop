package syncer

import (
	"context"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	v1 "github.com/solo-io/sqoop/pkg/api/v1"
	"github.com/solo-io/sqoop/pkg/engine"
	"github.com/solo-io/sqoop/pkg/engine/router"
	TODO "github.com/solo-io/sqoop/pkg/todo"
	"github.com/solo-io/sqoop/pkg/translator"
	"github.com/vektah/gqlgen/neelance/schema"
)

type GraphQLSyncer struct {
	writeNamespace    string
	reporter          reporter.Reporter
	writeErrs         chan error
	proxyReconciler   gloov1.ProxyReconciler
	resolverMapClient v1.ResolverMapClient
	engine            *engine.Engine
	router            *router.Router
}

func NewGraphQLSyncer(writeNamespace string,
	reporter reporter.Reporter,
	writeErrs chan error,
	proxyReconciler gloov1.ProxyReconciler,
	resolverMapClient v1.ResolverMapClient,
	engine *engine.Engine,
	router *router.Router) v1.ApiSyncer {
	s := &GraphQLSyncer{
		writeNamespace:    writeNamespace,
		reporter:          reporter,
		writeErrs:         writeErrs,
		proxyReconciler:   proxyReconciler,
		resolverMapClient: resolverMapClient,
		engine:            engine,
		router:            router,
	}
	return s
}

func (s *GraphQLSyncer) Sync(ctx context.Context, snap *v1.ApiSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "syncer")

	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v (%v schemas %v resolverMaps)",
		snap.Hash(),
		len(snap.Schemas),
		len(snap.ResolverMaps),
	)
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("%v", snap)

	resourceErrs := make(reporter.ResourceErrors)

	proxy := translator.Translate(s.writeNamespace, snap, resourceErrs)
	logger.Infof("creating proxy %v", proxy.Metadata.Ref())
	if err := s.proxyReconciler.Reconcile(s.writeNamespace, gloov1.ProxyList{proxy}, TODO.TransitionFunction, clients.ListOpts{
		Ctx: ctx,
		Selector: map[string]string{
			"created_by": "sqoop",
		},
	}); err != nil {
		return err
	}

	var endpoints []*router.Endpoint
	var resolverMapsToGenerate v1.ResolverMapList
	var schemasToUpdate v1.SchemaList
	for _, schema := range snap.Schemas.List() {
		resolverMap, err := snap.ResolverMaps.List().Find(schema.Metadata.Ref().Strings())
		if err != nil {
			newMeta := core.Metadata{
				Name:        schema.Metadata.Name,
				Namespace:   schema.Metadata.Namespace,
				Annotations: map[string]string{"created_for": schema.Metadata.Name},
			}
			parsedSchema, err := parseSchemaString(schema)
			if err != nil {
				resourceErrs.AddError(schema, errors.Wrapf(err, "failed to parse schema"))
				continue
			}

			rm := translator.GenerateResolverMapSkeleton(newMeta, parsedSchema)

			resolverMapsToGenerate = append(resolverMapsToGenerate, rm)
			schemasToUpdate = append(schemasToUpdate, schema)

			// nothing to do for this schema yet, need to receive some resolvers
			continue
		}
		resourceErrs.Accept(schema)

		// this time should succeed
		resolverMap, err = snap.ResolverMaps.List().Find(schema.Metadata.Ref().Strings())
		if err != nil {
			resourceErrs.AddError(schema, errors.Wrapf(err, "finding resolvermap for schema"))
			continue
		}

		resourceErrs.Accept(resolverMap)

		endpoint, schemaErr, resolverErr := s.engine.CreateGraphqlEndpoint(schema, resolverMap)
		if schemaErr != nil {
			resourceErrs.AddError(schema, schemaErr)
		}
		if resolverErr != nil {
			resourceErrs.AddError(resolverMap, resolverErr)
		}
		if schemaErr != nil || resolverErr != nil {
			continue
		}
		endpoints = append(endpoints, endpoint)
	}
	if err := s.reporter.WriteReports(ctx, resourceErrs, nil); err != nil {
		return errors.Wrapf(err, "writing reports")
	}
	if err := resourceErrs.Validate(); err != nil {
		logger.Errorf("snapshot %v was rejected due to invalid config: %v", err)
		return nil
	}

	// final results
	s.router.UpdateEndpoints(endpoints)

	// TODO(ilackarms): use reconciler, and allow resolvermaps to transition between snapshots
	// then we can always generate ratehr than only generating for nonexisting!
	for _, rm := range resolverMapsToGenerate {
		if _, err := s.resolverMapClient.Write(rm, clients.WriteOpts{}); err != nil && !errors.IsExist(err) {
			return errors.Wrapf(err, "writing generated resolver maps to storage")
		}
	}

	// start propagating for new set of resources
	// TODO(ilackarms): reinstate propagator
	return nil // s.propagator.PropagateStatuses(snap, proxy, clients.WatchOpts{Ctx: ctx})

}

func parseSchemaString(sch *v1.Schema) (*schema.Schema, error) {
	parsedSchema := schema.New()
	return parsedSchema, parsedSchema.Parse(sch.InlineSchema)
}
