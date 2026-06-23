package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// RetryScopeAnnotation requests a partial re-run. Values: Full,
	// ArtifactOnly, ClusterOnly, FinalDecisionOnly. Carried as a metadata
	// annotation (not a spec field) so retry intent is auditable but does not
	// mutate the desired spec.
	RetryScopeAnnotation       = "security.kube-sentinel.io/retry-scope"
	RetryRequestIDAnnotation   = "security.kube-sentinel.io/retry-request-id"
	RetryRequestedAtAnnotation = "security.kube-sentinel.io/retry-requested-at"
)

// ScanRunSpec is an immutable execution of a SecurityAssessment.
type ScanRunSpec struct {
	AssessmentRef LocalObjectRef `json:"assessmentRef"`
	Targets       []string       `json:"targets,omitempty"`  // override; else SecurityAssessment.spec.targets
	Profiles      []ScanProfile  `json:"profiles,omitempty"` // override; else SecurityAssessment.spec.profiles
}

// ScanRunStatus tracks workflow phase and the final decision.
type ScanRunStatus struct {
	ObservedGeneration int64               `json:"observedGeneration,omitempty"`
	Phase              string              `json:"phase,omitempty"` // Pending, Running, Completed, Failed, Canceled
	ArtifactScan       ScanPhaseStatus     `json:"artifactScan,omitempty"`
	ClusterScan        ScanPhaseStatus     `json:"clusterScan,omitempty"`
	Features           []FeatureCondition  `json:"features,omitempty"`
	Targets            []TargetRunStatus   `json:"targets,omitempty"`
	RemoteResources    []RemoteResourceRef `json:"remoteResources,omitempty"`
	FinalDecision      *FinalDecision      `json:"finalDecision,omitempty"`
}

// FinalDecision expresses the verdict as status + reasons, not a single exit
// code. The REST API flattens status to a snake_case final_decision string and
// preserves the full object as evidence/summary.
type FinalDecision struct {
	Status    string                `json:"status,omitempty"` // Pass, Fail, Warning
	Reasons   []FinalDecisionReason `json:"reasons,omitempty"`
	DecidedAt metav1.Time           `json:"decidedAt,omitempty"`
}

// FinalDecisionReason is a single failure/warning reason.
type FinalDecisionReason struct {
	Code      string `json:"code"` // critical_finding, secret_exposure, digest_mismatch, missing_artifact, unapproved_exception, ...
	Message   string `json:"message,omitempty"`
	Severity  string `json:"severity,omitempty"`
	Category  string `json:"category,omitempty"`
	Count     int32  `json:"count,omitempty"`
	FindingID string `json:"findingID,omitempty"`
}

// RemoteResourceRef tracks a resource remote-applied to a Biz Cluster. Remote
// objects cannot carry an ownerReference to the Mgmt CR, so GC is driven by
// label (target/feature/scope) and spec-hash annotation.
type RemoteResourceRef struct {
	Target     string `json:"target"`
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name"`
	Feature    string `json:"feature,omitempty"`
	Scope      string `json:"scope,omitempty"` // target | run
	SpecHash   string `json:"specHash,omitempty"`
}

// ScanPhaseStatus tracks one workflow phase (artifact scan or cluster scan).
type ScanPhaseStatus struct {
	Phase      string             `json:"phase,omitempty"` // Pending, Running, Completed, Failed, Skipped
	StartedAt  metav1.Time        `json:"startedAt,omitempty"`
	FinishedAt metav1.Time        `json:"finishedAt,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// FeatureCondition is the per-feature execution status.
type FeatureCondition struct {
	Name               string      `json:"name"`
	Enabled            bool        `json:"enabled"`
	Ready              bool        `json:"ready"`
	Reason             string      `json:"reason,omitempty"` // Disabled, Ready, ConfigError, ApplyError, NotReady
	Message            string      `json:"message,omitempty"`
	ObservedGeneration int64       `json:"observedGeneration,omitempty"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// TargetRunStatus is the per-target execution status within a ScanRun.
type TargetRunStatus struct {
	Target     string             `json:"target"`
	Phase      string             `json:"phase,omitempty"` // Pending, Running, Completed, Failed, Skipped
	Message    string             `json:"message,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=sr
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Decision",type=string,JSONPath=`.status.finalDecision.status`
// +kubebuilder:printcolumn:name="Assessment",type=string,JSONPath=`.spec.assessmentRef.name`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ScanRun is one immutable execution of a SecurityAssessment.
type ScanRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScanRunSpec   `json:"spec,omitempty"`
	Status ScanRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScanRunList contains a list of ScanRun.
type ScanRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScanRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScanRun{}, &ScanRunList{})
}
