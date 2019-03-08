package options

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type Options struct {
	Metadata    core.Metadata
	Top         Top
	Install     Install
	Uninstall   Uninstall
	Schema      Schema
	ResolverMap ResolverMap
}

type Top struct {
	Interactive bool
	File        string
	Output      string
	Namespace   string
	Ctx         context.Context
}

type Install struct {
	DryRun            bool
	HelmChartOverride string
	Namespace         string
}

type Uninstall struct {
	Namespace string
}

type Schema struct {
	Name        string
	ResolverMap string
}

type ResolverMap struct {
	Upstream         string
	Function         string
	RequestTemplate  string
	ResponseTemplate string
	SchemaName       string
	TypeName         string
	FieldName        string
}
