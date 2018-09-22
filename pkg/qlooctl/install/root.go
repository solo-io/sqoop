package install

import (
	"github.com/solo-io/sqoop/pkg/sqoopctl"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Sqoop and dependencies to supported environments",
	Long:  `sqoopctl currently suppports installations using docker-compose and Kubernetes`,
}

func init() {
	sqoopctl.RootCmd.AddCommand(installCmd)
}
