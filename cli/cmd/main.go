package main

import (
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"os"
	"time"

	"github.com/solo-io/sqoop/version"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/sqoop/cli/pkg/cmd"
)

func main() {
	start := time.Now()
	defer check.CallCheck("sqoopctl", version.Version, start)

	app := cmd.App(version.Version)
	if err := app.Execute(); err != nil {
		if err != nil {
			log.Fatalf("Error running CLI: %v", err)
		}
		os.Exit(0)
	}
}
