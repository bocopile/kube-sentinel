package v1alpha1

// ScanProfile is a high-level scan profile that expands to a base set of
// registry feature IDs via the canonical profile→feature mapping table
// (see docs/ARCHITECTURE.md).
// +kubebuilder:validation:Enum=SourceSecurity;ImageSupplyChain;KubernetesConfig;RBACAndSecretReference;BuildAndDeploy
type ScanProfile string

const (
	ProfileSourceSecurity         ScanProfile = "SourceSecurity"
	ProfileImageSupplyChain       ScanProfile = "ImageSupplyChain"
	ProfileKubernetesConfig       ScanProfile = "KubernetesConfig"
	ProfileRBACAndSecretReference ScanProfile = "RBACAndSecretReference"
	ProfileBuildAndDeploy         ScanProfile = "BuildAndDeploy"
)

// SecretKeyRef references a single key inside a Secret in the Mgmt Cluster.
// The referenced value (e.g. kubeconfig) is never exposed through API, logs,
// status, or reports.
type SecretKeyRef struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"`
}

// LocalObjectRef references another object by name (and optional namespace).
type LocalObjectRef struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}
