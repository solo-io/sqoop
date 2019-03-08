package install

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/install/helm/gloo/generate"
	"github.com/solo-io/gloo/pkg/cliutil/install"
	"github.com/solo-io/go-utils/kubeutils"
	sqoopcliutil "github.com/solo-io/sqoop/pkg/cliutil"
	"github.com/solo-io/sqoop/pkg/defaults"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const (
	installNamespace      = defaults.GlooSystem
	PersistentVolumeClaim = "PersistentVolumeClaim"

	sqoopTemplateUrl = "https://storage.googleapis.com/sqoop-helm/charts/sqoop-%s.tgz"
)

func removeExistingPVCs(manifestBytes []byte, namespace string) ([]byte, error) {

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	var docs []string
	for _, doc := range strings.Split(string(manifestBytes), "---") {

		// We need to define this ourselves, because if we unmarshal into `apiextensions.CustomResourceDefinition`
		// we don't get the TypeMeta (in the yaml they are nested under `metadata`, but the k8s struct has
		// them as top level fields...)
		var resource struct {
			Metadata metav1.ObjectMeta
			metav1.TypeMeta
		}
		if err := yaml.Unmarshal([]byte(doc), &resource); err != nil {
			return nil, errors.Wrapf(err, "parsing resource: %s", doc)
		}

		// If this is a PVC, check if it already exists. If so, exclude this resource from the manifest.
		// We don't want to overwrite existing PVCs.
		if resource.TypeMeta.Kind == PersistentVolumeClaim {

			_, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(resource.Metadata.Name, metav1.GetOptions{})
			if err != nil {
				if !kubeerrors.IsNotFound(err) {
					return nil, errors.Wrapf(err, "retrieving %s: %s.%s", PersistentVolumeClaim, namespace, resource.Metadata.Name)
				}
			} else {
				// The PVC exists, exclude it from manifest
				continue
			}
		}

		docs = append(docs, doc)
	}
	return []byte(strings.Join(docs, install.YamlDocumentSeparator)), nil
}

func getValuesFromFile(helmChart *chart.Chart, fileName string) (*chart.Config, error) {
	rawAdditionalValues := "{}"
	if fileName != "" {
		var found bool
		for _, valueFile := range helmChart.Files {
			if valueFile.TypeUrl == fileName {
				rawAdditionalValues = string(valueFile.Value)
				found = true
			}
		}
		if !found {
			return nil, errors.Errorf("could not find value file [%s] in Helm chart archive", fileName)
		}
	}

	// Convert value file content to struct
	valueStruct := &generate.Config{}
	if err := yaml.Unmarshal([]byte(rawAdditionalValues), valueStruct); err != nil {
		return nil, errors.Errorf("invalid format for value file [%s] in Helm chart archive", fileName)
	}

	// Namespace creation is disabled by default, otherwise install with helm will fail
	// (`helm install --namespace=<namespace_name>` creates the given namespace)
	valueStruct.Namespace = &generate.Namespace{Create: true}

	valueBytes, err := yaml.Marshal(valueStruct)
	if err != nil {
		return nil, errors.Wrapf(err, "failed marshaling value file struct")
	}

	// Add license key. Ugly but it works
	valuesString := fmt.Sprintf("license_key: %s\n%s", string(valueBytes))

	// NOTE: config.Values is never used by helm
	return &chart.Config{Raw: valuesString}, nil
}

// TODO: copied over and modified for a quick fix, improve
//noinspection GoNameStartsWithPackageName
func installManifest(manifest []byte, isDryRun bool, namespace string) error {
	if isDryRun {
		fmt.Printf("%s", manifest)
		// For safety, print a YAML separator so multiple invocations of this function will produce valid output
		fmt.Println("\n---")
		return nil
	}

	// TODO(marco): this is hideous, but no time to wait on gloo build+release right now. I'll clean up soon.
	if namespace != "" {

		// Create namespace otherwise the next command might fail
		if _, err := createNamespaceIfNotExist(namespace); err != nil {
			return errors.Wrapf(err, "creating namespace [%s]", namespace)
		}

		if err := kubectlApplyWithNamespace(manifest, namespace); err != nil {
			return errors.Wrapf(err, "running kubectl apply on manifest")
		}
		return nil

	} else {

		if err := kubectlApply(manifest); err != nil {
			return errors.Wrapf(err, "running kubectl apply on manifest")
		}
		return nil
	}
}

func kubectlApplyWithNamespace(manifest []byte, namespace string) error {
	return kubectl(bytes.NewBuffer(manifest), "apply", "-n", namespace, "-f", "-")
}

func kubectlApply(manifest []byte) error {
	return kubectl(bytes.NewBuffer(manifest), "apply", "-f", "-")
}

func kubectl(stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = sqoopcliutil.Logger
	kubectl.Stderr = sqoopcliutil.Logger
	return kubectl.Run()
}

func createNamespaceIfNotExist(namespace string) (exists bool, err error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return false, err
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	installNamespace := &kubev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	if _, err := kubeClient.CoreV1().Namespaces().Create(installNamespace); err != nil {
		if kubeerrors.IsAlreadyExists(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// TODO: copied over and modified for a quick fix, improve
// Blocks until the given CRDs have been registered.
func waitForCrdsToBeRegistered(crds []string, timeout, interval time.Duration) error {
	if len(crds) == 0 {
		return nil
	}

	// TODO: think about improving
	// Just pick the last crd in the list an wait for it to be created. It is reasonable to assume that by the time we
	// get to applying the manifest the other ones will be ready as well.
	crdName := crds[len(crds)-1]

	elapsed := time.Duration(0)
	for {
		select {
		case <-time.After(interval):
			if err := kubectl(nil, "get", crdName); err == nil {
				return nil
			}
			elapsed += interval
			if elapsed > timeout {
				return errors.Errorf("failed to confirm knative crd registration after %v", timeout)
			}
		}
	}
}
