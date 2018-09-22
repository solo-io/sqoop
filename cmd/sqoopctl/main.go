package main

import (
	"fmt"
	"os"

	_ "github.com/solo-io/sqoop/pkg/sqoopctl/install"
	_ "github.com/solo-io/sqoop/pkg/sqoopctl/resolvermap"
	_ "github.com/solo-io/sqoop/pkg/sqoopctl/schema"
	"github.com/solo-io/sqoop/pkg/sqoopctl"
)

func main() {
	if err := sqoopctl.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
