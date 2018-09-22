package install

import (
	"fmt"
	"os"

	"os/exec"

	"github.com/spf13/cobra"
)

const sqoopYamlURI = "https://raw.githubusercontent.com/solo-io/sqoop/master/install/kube/install.yaml"

var installKubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "install Sqoop on Kubernetes",
	Long: `
	Installs latest Sqoop into a Kubernetes cluster. It downloads the latest installation YAML
	file and installs to the current kubectl context.`,
	Run: func(c *cobra.Command, a []string) {
		err := kubeInstall(dryRun, sqoopYamlURI)
		if err != nil {
			fmt.Printf("Unable to isntall Sqoop to Kubernetes %q\n", err)
			os.Exit(1)
		}
		if !dryRun {
			fmt.Println("Sqoop successfully installed.")
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
