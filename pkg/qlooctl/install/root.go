package install

import (
	"github.com/spf13/cobra"
	"github.com/solo-io/qloo/pkg/qlooctl"
)


var installCmd = &cobra.Command{
	Use: "install",
	Short: "Install QLoo and dependencies to supported environments",
	Long:  `qlooctl currently suppports installations using docker-compose and Kubernetes`,
}

func init() {
	qlooctl.RootCmd.AddCommand(installCmd)
}
