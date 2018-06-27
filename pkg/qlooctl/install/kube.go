package install

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"os/exec"
)

const qlooYamlURI = "https://raw.githubusercontent.com/solo-io/qloo/master/install/kube/install.yaml"

var installKubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "install QLoo on Kubernetes",
	Long: `
	Installs latest QLoo into a Kubernetes cluster. It downloads the latest installation YAML
	file and installs to the current kubectl context.`,
	Run: func(c *cobra.Command, a []string) {
		err := kubeInstall(dryRun, qlooYamlURI)
		if err != nil {
			fmt.Printf("Unable to isntall QLoo to Kubernetes %q\n", err)
			os.Exit(1)
		}
		if !dryRun {
			fmt.Println("QLoo successfully installed.")
		}
	},
}

var dryRun bool

func init() {
	installKubeCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "If true, only "+
		"print the objects that will be installed")
	installCmd.AddCommand(installKubeCmd)
}

// Install setups Gloo on Kubernetes using kubectl and current context
func kubeInstall(dryRun bool, uri string) error {
	// using kubectl with latest install.yaml
	args := []string{"apply", "--filename",
		uri}
	if dryRun {
		args = append(args, "--dry-run=true")
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
