package v1

import (
	"encoding/json"

	"github.com/solo-io/gloo/pkg/api/types/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Schema is the generic Kubernetes API object wrapper for Gloo Schemas
type Schema struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            *v1.Status `json:"status"`
	Spec              *Spec      `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SchemaList is the generic Kubernetes API object wrapper
type SchemaList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`
	Items           []Schema `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResolverMap is the generic Kubernetes API object wrapper for Gloo ResolverMaps
type ResolverMap struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            *v1.Status `json:"status"`
	Spec              *Spec      `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResolverMapList is the generic Kubernetes API object wrapper
type ResolverMapList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`
	metav1.Status   `json:"status,omitempty"`
	Items           []ResolverMap `json:"items"`
}

// spec implements deepcopy
type Spec map[string]interface{}

func (in *Spec) DeepCopyInto(out *Spec) {
	if in == nil {
		out = nil
		return
	}
	data, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &out)
	if err != nil {
		panic(err)
	}
}
