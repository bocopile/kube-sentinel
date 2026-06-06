package v1alpha1

// SecurityAgentSpec defines the desired state of SecurityAgent.
type SecurityAgentSpec struct {
	Global   GlobalConfig   `json:"global"`
	Features FeaturesConfig `json:"features"`
	Output   OutputConfig   `json:"output"`
	Override OverrideConfig `json:"override"`
	Tests    TestsConfig    `json:"tests"`
}

// GlobalConfig holds global settings for the security agent.
type GlobalConfig struct {
	Enabled         bool   `json:"enabled,omitempty"`
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

// FeaturesConfig enables or disables individual features.
type FeaturesConfig struct {
	PodSecurity    bool `json:"podSecurity,omitempty"`
	NetworkPolicy  bool `json:"networkPolicy,omitempty"`
	SecretScanning bool `json:"secretScanning,omitempty"`
}

// OutputConfig controls how findings are reported.
type OutputConfig struct {
	Format    string `json:"format,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	ConfigMap string `json:"configMap,omitempty"`
}

// OverrideConfig allows per-namespace or per-workload policy overrides.
type OverrideConfig struct {
	Namespaces []string          `json:"namespaces,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// TestsConfig holds configuration for compliance test suites.
type TestsConfig struct {
	Enabled bool     `json:"enabled,omitempty"`
	Suites  []string `json:"suites,omitempty"`
}

// SecurityAgentStatus defines the observed state of SecurityAgent.
type SecurityAgentStatus struct {
	Phase string `json:"phase,omitempty"`
}

// ResourceMeta holds Kubernetes resource metadata (name, labels, annotations).
type ResourceMeta struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type (
	// +kubebuilder:object:root=true
	// +kubebuilder:subresource:status
	// +kubebuilder:resource:scope=Cluster
	// SecurityAgent is the schema for the securityagents API.
	SecurityAgent struct {
		Metadata ResourceMeta        `json:"metadata,omitempty"`
		Spec     SecurityAgentSpec   `json:"spec"`
		Status   SecurityAgentStatus `json:"status,omitempty"`
	}
)
