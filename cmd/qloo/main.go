package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	glooflags "github.com/solo-io/gloo/pkg/bootstrap/flags"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/gloo/pkg/signals"
	"github.com/solo-io/qloo/pkg/bootstrap"
	"github.com/solo-io/qloo/pkg/bootstrap/flags"
	"github.com/solo-io/qloo/pkg/core"
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
	Use:   "qloo",
	Short: "runs qloo",
	RunE: func(cmd *cobra.Command, args []string) error {
		eventLoop, err := core.Setup(opts)
		if err != nil {
			return errors.Wrap(err, "setting up event loop")
		}

		stop := signals.SetupSignalHandler()
		eventLoop.Run(stop)

		log.Printf("shutting down QLoo")
		return nil
	},
}

func init() {
	glooflags.AddConfigStorageOptionFlags(rootCmd, &opts.Options)
	glooflags.AddFileFlags(rootCmd, &opts.Options)
	glooflags.AddKubernetesFlags(rootCmd, &opts.Options)
	glooflags.AddConsulFlags(rootCmd, &opts.Options)
	flags.AddQLooFlags(rootCmd, &opts)
}
