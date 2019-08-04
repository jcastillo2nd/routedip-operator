package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/api/core/v1"
	amt "k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RoutedIPSpec defines the desired state of RoutedIP
// +k8s:openapi-gen=true
type RoutedIPSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ClassName string `json:"className"`
	RoutedIP string `json:"routedIP"`
	ServiceRef amt.NamespacedName `json:"serviceRef"`
}

// RoutedIPStatus defines the observed state of RoutedIP
// +k8s:openapi-gen=true
type RoutedIPStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	AssignedNode []string `json:"assignedNode"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoutedIP is the Schema for the routedips API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type RoutedIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoutedIPSpec   `json:"spec,omitempty"`
	Status RoutedIPStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoutedIPList contains a list of RoutedIP
type RoutedIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoutedIP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoutedIP{}, &RoutedIPList{})
}
