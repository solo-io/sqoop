package helpers

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/solo-io/sqoop/pkg/api/v1"

	glooV1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var memoryResourceClient *factory.MemoryResourceClientFactory

func UseMemoryClients() {
	memoryResourceClient = &factory.MemoryResourceClientFactory{
		Cache: memory.NewInMemoryResourceCache(),
	}
}

func MustSchemaClient() v1.SchemaClient {
	client, err := SchemaClient()
	if err != nil {
		log.Fatalf("failed to create schema client: %v", err)
	}
	return client
}

func SchemaClient() (v1.SchemaClient, error) {
	if memoryResourceClient != nil {
		return v1.NewSchemaClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	cache := kube.NewKubeCache(context.TODO())
	schemaClient, err := v1.NewSchemaClient(&factory.KubeResourceClientFactory{
		Crd:         v1.SchemaCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating resolver map client")
	}
	if err := schemaClient.Register(); err != nil {
		return nil, err
	}
	return schemaClient, nil
}

func MustResolverMapClient() v1.ResolverMapClient {
	client, err := ResolverMapClient()
	if err != nil {
		log.Fatalf("failed to create resolver map client: %v", err)
	}
	return client
}

func ResolverMapClient() (v1.ResolverMapClient, error) {
	if memoryResourceClient != nil {
		return v1.NewResolverMapClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	cache := kube.NewKubeCache(context.TODO())
	resolverMapClient, err := v1.NewResolverMapClient(&factory.KubeResourceClientFactory{
		Crd:         v1.ResolverMapCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating resolver map client")
	}
	if err := resolverMapClient.Register(); err != nil {
		return nil, err
	}
	return resolverMapClient, nil
}

func MustSettingsClient() glooV1.SettingsClient {
	client, err := SettingsClient()
	if err != nil {
		log.Fatalf("failed to create settings client: %v", err)
	}
	return client
}

func SettingsClient() (glooV1.SettingsClient, error) {
	if memoryResourceClient != nil {
		return glooV1.NewSettingsClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	cache := kube.NewKubeCache(context.TODO())
	settingsClient, err := glooV1.NewSettingsClient(&factory.KubeResourceClientFactory{
		Crd:         glooV1.SettingsCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating settings client")
	}
	if err := settingsClient.Register(); err != nil {
		return nil, err
	}
	return settingsClient, nil
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := getKubernetesConfig(0)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubeClient, nil
}

func getKubernetesConfig(timeout time.Duration) (*rest.Config, error) {
	config, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, fmt.Errorf("Error retrieving Kubernetes configuration: %v \n", err)
	}
	config.Timeout = timeout
	return config, nil
}
