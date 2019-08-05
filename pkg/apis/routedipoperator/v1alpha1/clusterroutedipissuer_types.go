package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	amt "k8s.io/apimachinery/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterRoutedIPIssuerSpec defines the desired state of ClusterRoutedIPIssuer
// +k8s:openapi-gen=true
// Will still need to patch CRD to Cluster. See: https://github.com/kubernetes/kubernetes/pull/80458
// +genclient:nonNamespaced
type ClusterRoutedIPIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ClassName          string        `json:"className"`
	FirewallPostAllow  bool          `json:"firewallPostAllow"`
	NodePortRange      string        `json:"nodePortRange"`
	PerNodeIPLimit     int32         `json:"perNodeIPLimit"`
	FromSecretsRef     string        `json:"fromSecretsRef"`
	UpdateRoutedIPSpec batchv1.JobSpec `json:"updateRoutedIPSpec"`
	UpdateFirewallSpec batchv1.JobSpec `json:"updateFirewallSpec"`
}

// ClusterRoutedIPIssuerStatus defines the observed state of ClusterRoutedIPIssuer
// +k8s:openapi-gen=true
type ClusterRoutedIPIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// RoutedIPs is a list of the RoutedIP names
	RoutedIPs []amt.NamespacedName `json:"routedIPs"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterRoutedIPIssuer is the Schema for the clusterroutedipissuers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ClusterRoutedIPIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterRoutedIPIssuerSpec   `json:"spec,omitempty"`
	Status ClusterRoutedIPIssuerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterRoutedIPIssuerList contains a list of ClusterRoutedIPIssuer
type ClusterRoutedIPIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRoutedIPIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterRoutedIPIssuer{}, &ClusterRoutedIPIssuerList{})
}
