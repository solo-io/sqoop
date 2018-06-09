package install

import (
	"fmt"
	"os"

	"github.com/solo-io/glooctl/pkg/install/kube"
	"github.com/spf13/cobra"
)

func kubeCmd() *cobra.Command {
	dryRun := false
	cmd := &cobra.Command{
		Use:   "kube",
		Short: "install gloo on Kubernetes",
		Long: `
	Installs latest gloo on Kubernetes. It downloads the latest installation YAML
	file and installs to the current Kubectl context.`,
		Run: func(c *cobra.Command, a []string) {
			err := kube.Install(dryRun)
			if err != nil {
				fmt.Printf("Unable to isntall gloo to Kubernetes %q\n", err)
				os.Exit(1)
			}
			if !dryRun {
				fmt.Println("Gloo successfully installed.")
			}
		},
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false,
		"If true, only print the objects that will be setup, without sending it")
	return cmd
}
