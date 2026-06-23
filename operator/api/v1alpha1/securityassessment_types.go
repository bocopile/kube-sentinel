package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// SecurityAssessmentSpec declares what to scan and which features/profiles apply.
type SecurityAssessmentSpec struct {
	Targets       []string           `json:"targets"`
	Profiles      []ScanProfile      `json:"profiles,omitempty"`
	ArtifactInput *ArtifactInputSpec `json:"artifactInput,omitempty"`
	AIRemediation *AIRemediationSpec `json:"aiRemediation,omitempty"`
	Features      []FeatureSpec      `json:"features,omitempty"`
	Output        OutputSpec         `json:"output,omitempty"`
	ScanResources *ScanResourceSpec  `json:"scanResources,omitempty"`
}

// ArtifactInputSpec links delivery artifact inputs (artifact-input manifest).
// The reproducible declaration is stored in the Artifact Store; the CRD holds
// only references to input locations/lists, never large inline payloads.
type ArtifactInputSpec struct {
	SourceRef   *ArtifactLocationRef `json:"sourceRef,omitempty"`
	ImageList   []ImageArtifactRef   `json:"imageList,omitempty"`
	DigestList  []ImageDigestRef     `json:"digestList,omitempty"`
	ManifestRef *ArtifactLocationRef `json:"manifestRef,omitempty"`
}

// ArtifactLocationRef points to source/manifest input in a path or Artifact Store.
type ArtifactLocationRef struct {
	Path              string `json:"path,omitempty"`
	ArtifactStorePath string `json:"artifactStorePath,omitempty"`
	Checksum          string `json:"checksum,omitempty"`
}

// ImageArtifactRef is a delivery image target.
type ImageArtifactRef struct {
	Image   string `json:"image"`
	Digest  string `json:"digest,omitempty"`
	TarPath string `json:"tarPath,omitempty"`
}

// ImageDigestRef is an approved image digest baseline entry.
type ImageDigestRef struct {
	Image  string `json:"image"`
	Digest string `json:"digest"`
}

// AIRemediationSpec is the opt-in (default OFF) AI remediation advisor config.
// Kept strongly-typed so egress/redaction/decision-safety fields are validated
// at the CRD schema. See docs/AI_REMEDIATION.md.
type AIRemediationSpec struct {
	Enabled               bool          `json:"enabled"`
	Provider              string        `json:"provider,omitempty"` // gemini | none
	APIKeySecretRef       *SecretKeyRef `json:"apiKeySecretRef,omitempty"`
	Model                 string        `json:"model,omitempty"`
	PromptTemplateID      string        `json:"promptTemplateID,omitempty"`
	SeverityFilter        []string      `json:"severityFilter,omitempty"`
	CategoryAllowlist     []string      `json:"categoryAllowlist,omitempty"`
	MaxFindingsPerScan    int32         `json:"maxFindingsPerScan,omitempty"`
	RequestTimeoutSeconds int32         `json:"requestTimeoutSeconds,omitempty"`
	MaxConcurrency        int32         `json:"maxConcurrency,omitempty"`
	RedactionProfile      string        `json:"redactionProfile,omitempty"`
}

// FeatureSpec enables/disables a registry feature (or umbrella name) and
// carries opaque per-feature config so adding a new tool needs no schema change.
type FeatureSpec struct {
	Name    string               `json:"name"`
	Enabled bool                 `json:"enabled"`
	Config  runtime.RawExtension `json:"config,omitempty"`
}

// OutputSpec configures report store routing for the assessment.
type OutputSpec struct {
	ReportStore ReportStoreSpec `json:"reportStore,omitempty"`
}

// ReportStoreSpec sets tenant/retention for stored results.
type ReportStoreSpec struct {
	TenantID      string            `json:"tenantID,omitempty"`
	RetentionDays int               `json:"retentionDays,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// ScanResourceSpec configures scan Job resource settings.
type ScanResourceSpec struct {
	SecurityAssessment ScanJobResourceSpec `json:"securityAssessment,omitempty"`
}

// ScanJobResourceSpec holds scan Job lifecycle settings.
type ScanJobResourceSpec struct {
	TTLSecondsAfterFinished int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// SecurityAssessmentStatus is the last-run aggregate, a read model for the
// dashboard Overview/Assessments views.
type SecurityAssessmentStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	LastRunRef         *LocalObjectRef    `json:"lastRunRef,omitempty"`
	Summary            AssessmentSummary  `json:"summary,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// AssessmentSummary mirrors DATABASE scan_runs.summary counters.
type AssessmentSummary struct {
	LastDecision           string      `json:"lastDecision,omitempty"` // Pass, Fail, Warning
	CriticalCount          int32       `json:"criticalCount,omitempty"`
	HighCount              int32       `json:"highCount,omitempty"`
	ExceptionRequiredCount int32       `json:"exceptionRequiredCount,omitempty"`
	ScanHealthFailCount    int32       `json:"scanHealthFailCount,omitempty"`
	ScannerBaselineDate    string      `json:"scannerBaselineDate,omitempty"`
	LastRunAt              metav1.Time `json:"lastRunAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=sa
// +kubebuilder:printcolumn:name="Decision",type=string,JSONPath=`.status.summary.lastDecision`
// +kubebuilder:printcolumn:name="Critical",type=integer,JSONPath=`.status.summary.criticalCount`
// +kubebuilder:printcolumn:name="High",type=integer,JSONPath=`.status.summary.highCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SecurityAssessment is a delivery final-check assessment definition.
type SecurityAssessment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityAssessmentSpec   `json:"spec,omitempty"`
	Status SecurityAssessmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecurityAssessmentList contains a list of SecurityAssessment.
type SecurityAssessmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityAssessment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityAssessment{}, &SecurityAssessmentList{})
}
