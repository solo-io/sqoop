package sqoopctl

import (
	"fmt"

	"strings"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/bootstrap"
	glooflags "github.com/solo-io/gloo/pkg/bootstrap/flags"
	"github.com/solo-io/gloo/pkg/protoutil"
	"github.com/solo-io/glooctl/pkg/config"
	"github.com/solo-io/sqoop/pkg/api/types/v1"
	sqoopstorage "github.com/solo-io/sqoop/pkg/bootstrap"
	"github.com/solo-io/sqoop/pkg/storage"
	"github.com/spf13/cobra"
)

var Opts bootstrap.Options

var outputFormat string

var RootCmd = &cobra.Command{
	Use:   "sqoopctl",
	Short: "Interact with Sqoop's storage API from the command line",
	Long: "As Sqoop features a storage-based API, direct communication with " +
		"the Sqoop server is not necessary. sqoopctl simplifies the administration of " +
		"Sqoop by providing an easy way to create, read, update, and delete Sqoop storage objects.\n\n" +
		"" +
		"The primary concerns of sqoopctl are Schemas and ResolverMaps. Schemas contain your GraphQL schema;" +
		" ResolverMaps define how your schema fields are resolved.\n\n" +
		"" +
		"Start by creating a schema using sqoopctl schema create --from-file <path/to/your/graphql/schema>",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format for results")
	glooflags.AddConfigStorageOptionFlags(RootCmd, &Opts)
	glooflags.AddFileFlags(RootCmd, &Opts)
	glooflags.AddKubernetesFlags(RootCmd, &Opts)
	glooflags.AddConsulFlags(RootCmd, &Opts)
	config.LoadConfig(&Opts)
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
	fmt.Printf("%s\n", yam)
	return nil
}

func printAsJSON(msg proto.Message) error {
	jsn, err := protoutil.Marshal(msg)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", jsn)
	return nil
}

func printTable(msg proto.Message) error {
	switch obj := msg.(type) {
	case *v1.Schema:
		printSchema(obj)
	case *v1.ResolverMap:
		printResolverMap(obj)
	default:
		return errors.Errorf("unknown type %v", msg)
	}
	return nil
}

func printSchema(schema *v1.Schema) {
	fmt.Printf("%v", schema.Name)
}

func printResolverMap(resolverMap *v1.ResolverMap) {
	fmt.Printf("%v", resolverMap.Name)
}

func Print(msg proto.Message) error {
	switch strings.ToLower(outputFormat) {
	case "yaml":
		return printAsYaml(msg)
	case "json":
		return printAsJSON(msg)
	default:
		return printTable(msg)
	}
}

func MakeClient() (storage.Interface, error) {
	return sqoopstorage.Bootstrap(Opts)
}
