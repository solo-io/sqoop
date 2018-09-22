package kube_e2e

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/gloo/test/helpers"
)

const (
	// gloo labels
	testrunner        = "testrunner"
	controlPlane      = "control-plane"
	sqoopContainer    = "sqoop"
	upstreamDiscovery = "upstream-discovery"
	funcitonDiscovery = "function-discovery"
	starWars          = "starwars"
)

func SetupKubeForE2eTest(namespace string, buildImages, push, debug bool) error {
	if err := helpers.SetupKubeForTest(namespace); err != nil {
		return err
	}
	if buildImages {
		if err := BuildPushContainers(push, debug); err != nil {
			return err
		}
	}
	kubeResourcesDir := filepath.Join(KubeE2eDirectory(), "kube_resources")

	envoyImageTag := os.Getenv("ENVOY_IMAGE_TAG")
	if envoyImageTag == "" {
		log.Warnf("no ENVOY_IMAGE_TAG specified, defaulting to latest")
		envoyImageTag = "latest"
	}

	pullPolicy := "IfNotPresent"

	if push {
		pullPolicy = "Always"
	}

	data := templateData{Namespace: namespace, ImageTag: helpers.ImageTag(), ImagePullPolicy: pullPolicy, Debug: ""}
	if debug {
		data.Debug = "-debug"
	}

	testingResources := "testing-resources.yaml"
	installResources := "test-install.yaml"

	if err := GenerateKubeYaml(kubeResourcesDir, "testing-resources.tmpl.yaml", testingResources, data); err != nil {
		return err
	}

	if err := GenerateKubeYaml(kubeResourcesDir, "install.tmpl.yaml", installResources, data); err != nil {
		return err
	}

	if err := helpers.Kubectl("apply", "-f", filepath.Join(kubeResourcesDir, installResources)); err != nil {
		return errors.Wrapf(err, "creating kube resource from install.yml")
	}
	if err := helpers.Kubectl("apply", "-f", filepath.Join(kubeResourcesDir, testingResources)); err != nil {
		return errors.Wrapf(err, "creating kube resource from testing-resources.yml")
	}
	if err := helpers.WaitPodsRunning(
		testrunner,
		starWars,
	); err != nil {
		return errors.Wrap(err, "waiting for pods to start")
	}

	if err := helpers.WaitPodsRunning(
		controlPlane,
		sqoopContainer,
		upstreamDiscovery,
		funcitonDiscovery,
	); err != nil {
		return errors.Wrap(err, "waiting for pods to start")
	}
	return nil
}

func SqoopSDirectory() string {
	return filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "solo-io", "sqoop")
}

func KubeE2eDirectory() string {
	return filepath.Join(SqoopSDirectory(), "test", "kube_e2e")
}

// builds and pushes all docker containers needed for test
func BuildPushContainers(push, debug bool) error {
	if os.Getenv("SKIP_BUILD") == "1" {
		return nil
	}
	imageTag := helpers.ImageTag()
	os.Setenv("IMAGE_TAG", imageTag)

	// make the gloo containers
	for _, component := range []string{"sqoop"} {
		target := component
		target += "-docker"
		if push {
			target += "-push"
		}

		if debug {
			target += "-debug"
		}

		cmd := exec.Command("make", target)
		cmd.Dir = SqoopSDirectory()
		cmd.Stdout = ginkgo.GinkgoWriter
		cmd.Stderr = ginkgo.GinkgoWriter
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

type templateData struct {
	Namespace       string
	ImageTag        string
	ImagePullPolicy string
	Debug           string
}

func GenerateKubeYaml(kubeResourcesDir string, templateFile, outFile string, data templateData) error {
	testingResourcesTmpl, err := template.New("Test_Resources").ParseFiles(filepath.Join(kubeResourcesDir, templateFile))
	if err != nil {
		return errors.Wrapf(err, "parsing template from %s", templateFile)
	}

	buf := &bytes.Buffer{}
	if err := testingResourcesTmpl.ExecuteTemplate(buf, templateFile, data); err != nil {
		return errors.Wrapf(err, "executing template")
	}

	err = ioutil.WriteFile(filepath.Join(kubeResourcesDir, outFile), buf.Bytes(), 0644)
	if err != nil {
		return errors.Wrapf(err, "writing generated test resources template")
	}
	return nil
}
