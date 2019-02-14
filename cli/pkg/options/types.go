package options

import (
	"context"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type Options struct {
	Metadata core.Metadata
	Top      Top
	Install  Install
	Schema   Schema
}

type Top struct {
	Interactive bool
	File        string
	Output      string
	Namesapce   string
	Ctx         context.Context
}

type Install struct {
	DryRun           bool
	ReleaseVersion   string
	ManifestOverride string
}

type Schema struct {
	Name        string
	ResolverMap string
}
