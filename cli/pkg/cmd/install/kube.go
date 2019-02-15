package install

import (
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/sqoop/cli/pkg/flagutils"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
)

func KubeCmd(opts *options.Options, optionsFunc... cliutils.OptionsFunc) *cobra.Command {
	const glooGatewayUrlTemplate = "https://github.com/solo-io/sqoop/releases/download/v%s/sqoop.yaml"
	cmd := &cobra.Command{
		Use:   "kube",
		Short: "install sqoop on kubernetes",
		Long:  "requires kubectl to be installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := preInstall(); err != nil {
				return errors.Wrapf(err, "pre-install failed")
			}
			if err := installFromUri(opts, opts.Install.ManifestOverride, glooGatewayUrlTemplate); err != nil {
				return errors.Wrapf(err, "installing ingress from manifest")
			}
			return nil
		},
	}
	pflags := cmd.PersistentFlags()
	flagutils.AddInstallFlags(pflags, &opts.Install)

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
