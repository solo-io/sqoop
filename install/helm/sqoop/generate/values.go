package generate

import "github.com/solo-io/gloo/install/helm/gloo/generate"

type Config struct {
	Gloo  *generate.Config `json:"gloo,omitempty"`
	Sqoop *Sqoop           `json:"sqoop,omitempty"`
}

// Common

type OAuth struct {
	Server string `json:"server"`
	Client string `json:"client"`
}

type Rbac struct {
	Create bool `json:"create"`
}

// Sqoop

type Sqoop struct {
	Deployment *SqoopDeployment `json:"deployment,omitempty"`
	Service    *SqoopService    `json:"service,omitempty"`
	ConfigMap  *SqoopConfigMap  `json:"configMap,omitempty"`
}

type SqoopDeployment struct {
	Image *generate.Image `json:"image,omitempty"`
	Proxy *generate.Image `json:"proxy,omitempty"`
	*generate.DeploymentSpec
}

type SqoopService struct {
	Port string `json:"port"`
	Name string `json:"name"`
}

type SqoopConfigMap struct {
	Name string `json:"name"`
}
