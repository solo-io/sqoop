package main

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/sqoop/version"
)

//go:generate go run generate.go

func main() {
	err := version.CheckVersions()
	if err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
	docsOpts := cmd.DocsOptions{
		Output: options.Hugo,
	}
	log.Printf("starting generate")
	if err := cmd.Run(".", true, &docsOpts, nil, nil); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
