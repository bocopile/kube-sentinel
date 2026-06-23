// Package feature defines the Feature-as-Plugin abstraction. Each assessment
// capability (SAST, secret scan, image vuln, manifest policy, ...) implements
// the Feature interface and self-registers via Register() in its init(). The
// reconciler only orchestrates workflow, status, GC, and remote apply; it never
// hardcodes a feature. Adding a new tool is: implement Feature, call Register,
// add one import line in cmd/main.go.
package feature

import (
	"context"

	securityv1alpha1 "github.com/bocopile/kube-sentinel/operator/api/v1alpha1"
)

// Feature is one assessment capability plugin. The lifecycle methods are driven
// by the ScanRun reconciler in priority order.
type Feature interface {
	// ID is the stable registry feature ID (e.g. "source_security").
	ID() string
	// Priority orders feature execution ascending (see docs/PLAN.md registry).
	Priority() int
	// Validate checks config/preconditions and returns CRD-surfaced conditions.
	Validate(ctx FeatureContext) []Condition
	// Preflight verifies target/environment readiness before any scan.
	Preflight(ctx FeatureContext) []CheckResult
	// Build returns the desired Mgmt-local and remote resources to apply.
	Build(ctx FeatureContext) DesiredState
	// Collect gathers raw scanner outputs as artifact references.
	Collect(ctx FeatureContext) []ArtifactRef
	// Normalize converts raw outputs into normalized findings.
	Normalize(ctx FeatureContext) []Finding
}

// FeatureContext carries everything a feature needs for one ScanRun execution.
// Fields are intentionally minimal for the skeleton and will grow with target
// clients, artifact-store handles, and the desired-state store.
type FeatureContext struct {
	Ctx        context.Context
	Run        *securityv1alpha1.ScanRun
	Assessment *securityv1alpha1.SecurityAssessment
	// Config is the opaque per-feature config from SecurityAssessment.spec.features[].config.
	Config []byte
}

// Condition is a validation result surfaced to ScanRun.status.features[].
type Condition struct {
	Type    string
	Status  bool
	Reason  string // Disabled, Ready, ConfigError, ApplyError, NotReady
	Message string
}

// CheckResult is a single preflight outcome.
type CheckResult struct {
	Name    string
	Passed  bool
	Message string
}

// DesiredState is the set of resources a feature wants applied. Mgmt-local
// objects can carry ownerReferences; remote (Biz Cluster) objects are tracked
// by label + spec-hash for GC.
type DesiredState struct {
	// Placeholder. Will hold []client.Object for Mgmt and per-target remote
	// unstructured resources once remote apply lands (M2).
	MgmtObjects   []RuntimeObject
	RemoteObjects []RemoteObject
}

// RuntimeObject is a placeholder for a Mgmt-local apply target.
type RuntimeObject struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

// RemoteObject is a placeholder for a Biz Cluster apply target.
type RemoteObject struct {
	Target   string
	Object   RuntimeObject
	SpecHash string
}

// ArtifactRef references a stored scanner artifact (raw report, SBOM, evidence).
type ArtifactRef struct {
	Type     string // raw_report, sbom, evidence_bundle, human_report, ...
	Path     string
	Checksum string
}

// Finding is a normalized finding (security.finding/v1). Minimal for the
// skeleton; the canonical schema lives in docs/DATABASE.md and the normalizer.
type Finding struct {
	FindingID       string
	Scanner         string
	Category        string // sast, secret, image_vulnerability, sbom, integrity, kubernetes, rbac, secret_ref, network, dockerfile, script, scan_health
	Severity        string // Critical, High, Medium, Low, Info
	ScanStatus      string // Pass, Fail, Error, Skipped, Unsupported
	ExceptionStatus string // None, Required, Requested, Approved, Expired, Rejected
	TargetCluster   string // empty = Code/Artifact manifest; set = Biz applied
	Message         string
	Remediation     string
}
