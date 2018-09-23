package reporter

import (
	"github.com/pkg/errors"
	"github.com/solo-io/sqoop/pkg/storage"

	gloov1 "github.com/solo-io/gloo/pkg/api/types/v1"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
)

type reporter struct {
	store storage.Interface
}

func NewReporter(store storage.Interface) *reporter {
	return &reporter{store: store}
}

func (r *reporter) WriteReports(reports []ConfigObjectReport) error {
	for _, report := range reports {
		if err := r.writeReport(report); err != nil {
			return errors.Wrapf(err, "failed to write report for config object %v", report.CfgObject)
		}
		log.Debugf("wrote report for %v", report.CfgObject.GetName())
	}
	return nil
}

func (r *reporter) writeReport(report ConfigObjectReport) error {
	status := &gloov1.Status{
		State: gloov1.Status_Accepted,
	}
	if report.Err != nil {
		status.State = gloov1.Status_Rejected
		status.Reason = report.Err.Error()
	}
	name := report.CfgObject.GetName()
	switch report.CfgObject.(type) {
	case *v1.Schema:
		schema, err := r.store.V1().Schemas().Get(report.CfgObject.GetName())
		if err != nil {
			return errors.Wrapf(err, "failed to find schema %v", name)
		}
		// only update if status doesn't match
		if schema.Status.Equal(status) {
			return nil
		}
		schema.Status = status
		if _, err := r.store.V1().Schemas().Update(schema); err != nil {
			return errors.Wrapf(err, "failed to update schema with status report")
		}
	case *v1.ResolverMap:
		resolverMap, err := r.store.V1().ResolverMaps().Get(name)
		if err != nil {
			return errors.Wrapf(err, "failed to find resolverMap %v", name)
		}
		// only update if status doesn't match
		if resolverMap.Status.Equal(status) {
			return nil
		}
		resolverMap.Status = status
		if _, err := r.store.V1().ResolverMaps().Update(resolverMap); err != nil {
			return errors.Wrapf(err, "failed to update resolverMap store with status report")
		}
	}
	return nil
}
