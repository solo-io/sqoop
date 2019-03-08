package install

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/cliutil/install"
	"github.com/solo-io/gloo/projects/gloo/cli/pkg/constants"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/cobra"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func InstallCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.INSTALL_COMMAND.Use,
		Short: constants.INSTALL_COMMAND.Short,
		Long:  constants.INSTALL_COMMAND.Long,
	}

	cmd.AddCommand(
		KubeCmd(opts),
	)

	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}

func UninstallCmd(opts *options.Options, optionsFunc ...cliutils.OptionsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.UNINSTALL_COMMAND.Use,
		Short: constants.UNINSTALL_COMMAND.Short,
		Long:  constants.UNINSTALL_COMMAND.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Uninstalling Sqoop. This might take a while...")
			cfg, err := kubeutils.GetConfig("", "")
			if err != nil {
				return err
			}
			kubeClient, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return err
			}

			namespace, err := kubeClient.CoreV1().Namespaces().Get(opts.Uninstall.Namespace, v1.GetOptions{})
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					return errors.Errorf("namespace '%s' does not exist", opts.Uninstall.Namespace)
				}
				return errors.Wrapf(err, "failed to uninstall Sqoop")
			}

			if err := install.Kubectl(nil, "delete", "namespace", namespace.Name); err != nil {
				return errors.Wrapf(err, "failed to uninstall Sqoop")
			}

			fmt.Printf("Sqoop has been successfully uninstalled.\n")
			return nil
		},
	}
	cliutils.ApplyOptions(cmd, optionsFunc)
	return cmd
}
