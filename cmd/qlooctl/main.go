package main

import (
	"fmt"
	"os"

	_ "github.com/solo-io/qloo/pkg/qlooctl/install"
	_ "github.com/solo-io/qloo/pkg/qlooctl/resolvermap"
	_ "github.com/solo-io/qloo/pkg/qlooctl/schema"
	"github.com/solo-io/qloo/pkg/qlooctl"
)

func main() {
	if err := qlooctl.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
