package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	glooGenerate "github.com/solo-io/gloo/install/helm/gloo/generate"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/sqoop/install/helm/sqoop/generate"
)

const (
	valuesTemplate       = "install/helm/sqoop/values-template.yaml"
	valuesOutput         = "install/helm/sqoop/values.yaml"
	chartTemplate        = "install/helm/sqoop/Chart-template.yaml"
	chartOutput          = "install/helm/sqoop/Chart.yaml"
	requirementsTemplate = "install/helm/sqoop/requirements-template.yaml"
	requirementsOutput   = "install/helm/sqoop/requirements.yaml"

	gopkgToml    = "Gopkg.toml"
	constraint   = "constraint"
	glooPkg      = "github.com/solo-io/gloo"
	nameConst    = "name"
	versionConst = "version"
	neverPull    = "Never"
	alwaysPull   = "Always"
	ifNotPresent = "IfNotPresent"
)

func main() {
	var version string
	if len(os.Args) < 2 {
		panic("Must provide version as argument")
	} else {
		version = os.Args[1]
	}
	log.Printf("Generating helm files.")
	if err := generateValuesYamls(version); err != nil {
		log.Fatalf("generating values.yaml failed!: %v", err)
	}
	if err := generateChartYaml(version); err != nil {
		log.Fatalf("generating Chart.yaml failed!: %v", err)
	}
	if err := generateRequirementsYaml(); err != nil {
		log.Fatalf("unable to parse Gopkg.toml for proper gloo version: %v", err)
	}
}

func readConfig(path string) (generate.Config, error) {
	var config generate.Config
	if err := readYaml(valuesTemplate, &config); err != nil {
		return config, err
	}
	return config, nil
}

func readYaml(path string, obj interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.Unmarshal(bytes, obj); err != nil {
		return errors.Wrap(err, "failed parsing configuration file")
	}

	return nil
}

func writeYaml(obj interface{}, path string) error {
	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "failed marshaling config struct")
	}

	err = ioutil.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failing writing config file")
	}
	return nil
}

func generateValuesYaml(version, pullPolicy, outputFile string) error {
	config, err := readConfig(valuesTemplate)
	if err != nil {
		return err
	}
	config.Sqoop.Deployment.Image.Tag = version

	config.Sqoop.Deployment.Image.PullPolicy = pullPolicy

	return writeYaml(&config, outputFile)
}

func generateValuesYamls(version string) error {
	// Generate values for standard manifest
	standardPullPolicy := alwaysPull
	if version == "dev" {
		standardPullPolicy = neverPull
	}
	if err := generateValuesYaml(version, standardPullPolicy, valuesOutput); err != nil {
		return err
	}

	return nil
}

func generateChartYaml(version string) error {
	var chart glooGenerate.Chart
	if err := readYaml(chartTemplate, &chart); err != nil {
		return err
	}

	chart.Version = version

	return writeYaml(&chart, chartOutput)
}

func generateRequirementsYaml() error {
	var dl generate.DependencyList
	if err := readYaml(requirementsTemplate, &dl); err != nil {
		return err
	}
	glooVersion, err := parseToml()
	if err != nil {
		return err
	}
	for i, v := range dl.Dependencies {
		if v.Name == "gloo" {
			dl.Dependencies[i].Version = glooVersion
		}
	}
	return writeYaml(dl, requirementsOutput)
}

func parseToml() (string, error) {
	config, err := toml.LoadFile(gopkgToml)
	if err != nil {
		return "", err
	}

	tomlTree := config.Get(constraint)
	var (
		tree    []*toml.Tree
		version string
	)

	switch typedTree := tomlTree.(type) {
	case []*toml.Tree:
		tree = typedTree
	default:
		return "", fmt.Errorf("unable to parse toml tree")
	}

	for _, v := range tree {
		if v.Get(nameConst) == glooPkg && v.Get(versionConst) != "" {
			version = v.Get(versionConst).(string)
		}
	}

	return version, nil
}
