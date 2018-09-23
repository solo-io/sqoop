package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	glooflags "github.com/solo-io/gloo/pkg/bootstrap/flags"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/gloo/pkg/signals"
	"github.com/solo-io/sqoop/pkg/bootstrap"
	"github.com/solo-io/sqoop/pkg/bootstrap/flags"
	"github.com/solo-io/sqoop/pkg/core"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var opts bootstrap.Options

var rootCmd = &cobra.Command{
	Use:   "sqoop",
	Short: "runs sqoop",
	RunE: func(cmd *cobra.Command, args []string) error {
		eventLoop, err := core.Setup(opts)
		if err != nil {
			return errors.Wrap(err, "setting up event loop")
		}

		stop := signals.SetupSignalHandler()
		eventLoop.Run(stop)

		log.Printf("shutting down Sqoop")
		return nil
	},
}

func init() {
	glooflags.AddConfigStorageOptionFlags(rootCmd, &opts.Options)
	glooflags.AddFileFlags(rootCmd, &opts.Options)
	glooflags.AddKubernetesFlags(rootCmd, &opts.Options)
	glooflags.AddConsulFlags(rootCmd, &opts.Options)
	flags.AddSqoopFlags(rootCmd, &opts)
}
