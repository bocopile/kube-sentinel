package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterTargetSpec defines a Biz Cluster target registered in the Mgmt Cluster.
type ClusterTargetSpec struct {
	DisplayName        string                       `json:"displayName,omitempty"`
	Environment        string                       `json:"environment,omitempty"` // dev | final-check | prod
	KubeconfigRef      SecretKeyRef                 `json:"kubeconfigRef"`
	TargetNamespace    string                       `json:"targetNamespace,omitempty"`
	NamespaceAllowlist []string                     `json:"namespaceAllowlist,omitempty"`
	Output             TargetOutputSpec             `json:"output,omitempty"`
	Capabilities       TargetCapabilitySpec         `json:"capabilities,omitempty"`
	BootstrapPolicy    ClusterTargetBootstrapPolicy `json:"bootstrapPolicy,omitempty"`
}

// TargetOutputSpec carries report routing for a target.
type TargetOutputSpec struct {
	ReportTenantID string `json:"reportTenantID,omitempty"`
}

// TargetCapabilitySpec declares what kube-sentinel may do against the Biz Cluster.
type TargetCapabilitySpec struct {
	ScannerJobs          bool `json:"scannerJobs,omitempty"`
	ReadOnlyInspection   bool `json:"readOnlyInspection,omitempty"`
	TrivyOperatorReports bool `json:"trivyOperatorReports,omitempty"`
	HostPath             bool `json:"hostPath,omitempty"`
}

// ClusterTargetBootstrapPolicy declares which preflight-detected missing items
// the Mgmt operator may auto-install. Items not listed are reported as
// "install required" but never installed.
type ClusterTargetBootstrapPolicy struct {
	InstallMissingNamespace  bool            `json:"installMissingNamespace,omitempty"`
	InstallManagedRBAC       bool            `json:"installManagedRBAC,omitempty"`
	InstallScannerResources  bool            `json:"installScannerResources,omitempty"`
	AttachImagePullSecretRef *LocalObjectRef `json:"attachImagePullSecretRef,omitempty"`
}

// ClusterTargetStatus records connection, discovery, and credential state.
type ClusterTargetStatus struct {
	ObservedGeneration       int64                  `json:"observedGeneration,omitempty"`
	Phase                    string                 `json:"phase,omitempty"` // Pending, Ready, Degraded, AuthFailed, Unreachable, PermissionDenied
	LastValidatedAt          metav1.Time            `json:"lastValidatedAt,omitempty"`
	LastCredentialRotationAt metav1.Time            `json:"lastCredentialRotationAt,omitempty"`
	KubernetesVersion        string                 `json:"kubernetesVersion,omitempty"`
	Capabilities             TargetCapabilityStatus `json:"capabilities,omitempty"`
	Namespaces               []string               `json:"namespaces,omitempty"`
	Conditions               []metav1.Condition     `json:"conditions,omitempty"`
}

// TargetCapabilityStatus is the observed capability projection after discovery.
type TargetCapabilityStatus struct {
	ScannerJobs          bool `json:"scannerJobs,omitempty"`
	ReadOnlyInspection   bool `json:"readOnlyInspection,omitempty"`
	TrivyOperatorReports bool `json:"trivyOperatorReports,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ct
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.spec.environment`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="K8sVersion",type=string,JSONPath=`.status.kubernetesVersion`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ClusterTarget registers a Biz Cluster as a scan target in the Mgmt Cluster.
type ClusterTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterTargetSpec   `json:"spec,omitempty"`
	Status ClusterTargetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterTargetList contains a list of ClusterTarget.
type ClusterTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterTarget `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterTarget{}, &ClusterTargetList{})
}
