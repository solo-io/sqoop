package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/solo-io/gloo/pkg/bootstrap"
	glooflags "github.com/solo-io/gloo/pkg/bootstrap/flags"
	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/gloo/pkg/protoutil"
	"github.com/ghodss/yaml"
	"github.com/solo-io/qloo/pkg/storage"
	qloostorage "github.com/solo-io/qloo/pkg/bootstrap"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var opts bootstrap.Options

var rootCmd = &cobra.Command{
	Use:   "qlooctl",
	Short: "Interact with QLoo's storage API from the command line",
	Long: "As QLoo features a storage-based API, direct communication with " +
		"the QLoo server is not necessary. qlooctl simplifies the administration of " +
		"QLoo by providing an easy way to create, read, update, and delete QLoo storage objects.\n\n" +
		"" +
		"The primary concerns of qlooctl are Schemas and ResolverMaps. Schemas contain your GraphQL schema;" +
		" ResolverMaps define how your schema fields are resolved.\n\n" +
		"" +
		"Start by creating a schema using qlooctl schema create --from-file <path/to/your/graphql/schema>",
}

func init() {
	glooflags.AddConfigStorageOptionFlags(rootCmd, &opts)
	glooflags.AddFileFlags(rootCmd, &opts)
	glooflags.AddKubernetesFlags(rootCmd, &opts)
	glooflags.AddConsulFlags(rootCmd, &opts)
}

func printAsYaml(msg proto.Message) error {
	jsn, err := protoutil.Marshal(msg)
	if err != nil {
		return err
	}
	yam, err := yaml.JSONToYAML(jsn)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", yam)
}

func makeClient() (storage.Interface, error) {
	return qloostorage.Bootstrap(opts)
}
