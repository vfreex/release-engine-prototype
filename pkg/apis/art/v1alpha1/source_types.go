package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SourceSpec defines the desired state of Source
// +k8s:openapi-gen=true
type SourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Source SourceDefinition `json:"source"`
	// +listType=set
	Relationships []SourceRelationship `json:"relationships"`
}

// +k8s:openapi-gen=true
type SourceDefinition struct {
	Git GitSource `json:"git"`
}

// +k8s:openapi-gen=true
type GitSource struct {
	// URI points to the source that will be built. The structure of the source
	// will depend on the type of build to run
	URI string `json:"uri"`

	// Ref is the branch/tag/ref to build.
	Ref string `json:"ref"`
}

// +k8s:openapi-gen=true
type SourceRelationship struct {
	Type    string              `json:"type"`
	// +kubebuilder:validation:Optional
	DistGit DistGitRelationship `json:"distGit"`
	// +kubebuilder:validation:Optional
	Koji    KojiRelationship    `json:"koji"`
}

// +k8s:openapi-gen=true
type DistGitRelationship struct {
	Key      string `json:"key"`
	Branch   string `json:"branch"`
	Instance string `json:"instance"`
}

// +k8s:openapi-gen=true
type KojiRelationship struct {
	ComponentName string `json:"componentName"`
	Instance      string `json:"instance"`
}

// SourceStatus defines the observed state of Source
// +k8s:openapi-gen=true
type SourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// +listType=set
	Conditions []string `json:"conditions"`
	Phase      string   `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Source is the Schema for the sources API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sources,scope=Namespaced
type Source struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SourceSpec   `json:"spec,omitempty"`
	Status SourceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SourceList contains a list of Source
type SourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Source `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Source{}, &SourceList{})
}
