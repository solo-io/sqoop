package v1

import (
	"github.com/solo-io/gloo/pkg/api/types/v1"
)

// implement v1.ConfigObject

var _ v1.ConfigObject = &Schema{}
var _ v1.ConfigObject = &ResolverMap{}

// because proto refuses to do setters

func (item *Schema) SetName(name string) {
	item.Name = name
}

func (item *Schema) SetStatus(status *v1.Status) {
	item.Status = status
}

func (item *Schema) SetMetadata(meta *v1.Metadata) {
	item.Metadata = meta
}

func (item *ResolverMap) SetName(name string) {
	item.Name = name
}

func (item *ResolverMap) SetStatus(status *v1.Status) {
	item.Status = status
}

func (item *ResolverMap) SetMetadata(meta *v1.Metadata) {
	item.Metadata = meta
}
