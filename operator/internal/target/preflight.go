package target

import (
	"context"
	"fmt"
	"sort"
	"strings"

	authzv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterTarget.status.phase values.
const (
	PhasePending          = "Pending"
	PhaseReady            = "Ready"
	PhaseDegraded         = "Degraded"
	PhaseAuthFailed       = "AuthFailed"
	PhaseUnreachable      = "Unreachable"
	PhasePermissionDenied = "PermissionDenied"
)

// Condition types written to status.conditions[].
const (
	CondKubeconfigValid     = "KubeconfigValid"
	CondAPIServerReachable  = "APIServerReachable"
	CondAuthenticationValid = "AuthenticationValid"
	CondRBACValid           = "RBACValid"
	CondScannerRBAC         = "ScannerRBACValid"
	CondNamespacesVisible   = "NamespacesVisible"
	CondTrivyOperatorReady  = "TrivyOperatorReady"
)

// Condition reasons.
const (
	ReasonReady             = "Ready"
	ReasonInvalidSecret     = "InvalidSecret"
	ReasonInvalidKubeconfig = "InvalidKubeconfig"
	ReasonUnreachable       = "Unreachable"
	ReasonAuthFailed        = "AuthFailed"
	ReasonPermissionDenied  = "PermissionDenied"
	ReasonNotFound          = "NotFound"
	ReasonCRDNotFound       = "CRDNotFound"
)

// trivyGroupVersion is the Trivy Operator VulnerabilityReport API group/version.
const trivyGroupVersion = "aquasecurity.github.io/v1alpha1"

// ConditionResult is a single preflight check outcome (mapped to metav1.Condition
// by the controller).
type ConditionResult struct {
	Type    string
	OK      bool
	Reason  string
	Message string
}

// CapabilityResult is the discovered capability set.
type CapabilityResult struct {
	ScannerJobs          bool
	ReadOnlyInspection   bool
	TrivyOperatorReports bool
}

// PreflightResult is the full readiness outcome for one ClusterTarget.
type PreflightResult struct {
	Phase             string
	KubernetesVersion string
	Capabilities      CapabilityResult
	Namespaces        []string
	Conditions        []ConditionResult
}

func (r *PreflightResult) add(t string, ok bool, reason, msg string) {
	r.Conditions = append(r.Conditions, ConditionResult{Type: t, OK: ok, Reason: reason, Message: msg})
}

// FailResult builds a terminal PreflightResult for a check that failed before
// any target API call (e.g. missing/invalid kubeconfig Secret).
func FailResult(phase, condType, reason, msg string) PreflightResult {
	return PreflightResult{
		Phase:      phase,
		Conditions: []ConditionResult{{Type: condType, OK: false, Reason: reason, Message: msg}},
	}
}

// PreflightInput is the spec-derived input to a preflight run.
type PreflightInput struct {
	TargetNamespace          string
	NamespaceAllowlist       []string
	WantTrivyOperatorReports bool
}

// Preflight runs read-only readiness checks against the target cluster and
// returns the resulting phase, discovered state, and per-check conditions. It
// performs no writes to the target cluster.
func Preflight(ctx context.Context, cs kubernetes.Interface, in PreflightInput) PreflightResult {
	res := PreflightResult{Phase: PhaseReady}

	// 1. Connectivity + authentication via version discovery.
	ver, err := cs.Discovery().ServerVersion()
	if err != nil {
		phase, reason := ClassifyAPIError(err)
		res.Phase = phase
		if phase == PhaseAuthFailed {
			res.add(CondAuthenticationValid, false, reason, err.Error())
		} else {
			res.add(CondAPIServerReachable, false, reason, err.Error())
		}
		return res
	}
	res.KubernetesVersion = ver.GitVersion
	res.add(CondAPIServerReachable, true, ReasonReady, "")
	res.add(CondAuthenticationValid, true, ReasonReady, "")

	// 2. Read-only RBAC probe (list pods in the target namespace, or cluster-wide).
	readOK, err := canI(ctx, cs, "", "list", "pods", in.TargetNamespace)
	if err != nil {
		phase, reason := ClassifyAPIError(err)
		res.Phase = phase
		res.add(CondRBACValid, false, reason, err.Error())
		return res
	}
	if !readOK {
		res.Phase = PhasePermissionDenied
		res.add(CondRBACValid, false, ReasonPermissionDenied, "read-only RBAC denied: list pods")
		return res
	}
	res.add(CondRBACValid, true, ReasonReady, "")
	res.Capabilities.ReadOnlyInspection = true

	// 3. Scanner-Job capability (create jobs) — non-fatal.
	jobsOK, _ := canI(ctx, cs, "batch", "create", "jobs", in.TargetNamespace)
	res.Capabilities.ScannerJobs = jobsOK
	if jobsOK {
		res.add(CondScannerRBAC, true, ReasonReady, "")
	} else {
		res.add(CondScannerRBAC, false, ReasonPermissionDenied, "cannot create scanner Jobs in target namespace")
	}

	// 4. Optional Trivy Operator VulnerabilityReport CRD — non-blocking.
	if in.WantTrivyOperatorReports {
		exists := groupVersionExists(cs, trivyGroupVersion)
		res.Capabilities.TrivyOperatorReports = exists
		if exists {
			res.add(CondTrivyOperatorReady, true, ReasonReady, "")
		} else {
			res.add(CondTrivyOperatorReady, false, ReasonCRDNotFound, "VulnerabilityReport CRD not installed (optional input)")
		}
	}

	// 5. Namespace discovery within the allowlist.
	nsList, err := cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		visible, missing := NamespaceVisibility(in.NamespaceAllowlist, namespaceNames(nsList.Items))
		res.Namespaces = visible
		if len(missing) > 0 {
			res.add(CondNamespacesVisible, false, ReasonNotFound, fmt.Sprintf("namespaces in allowlist not found: %s", strings.Join(missing, ", ")))
			if res.Phase == PhaseReady {
				res.Phase = PhaseDegraded
			}
		} else {
			res.add(CondNamespacesVisible, true, ReasonReady, "")
		}
	}

	return res
}

// ClassifyAPIError maps a target API error to a phase and reason. Unauthorized
// (401) → AuthFailed; Forbidden (403) → PermissionDenied; anything else
// (network/timeout/DNS) → Unreachable.
func ClassifyAPIError(err error) (phase, reason string) {
	switch {
	case apierrors.IsUnauthorized(err):
		return PhaseAuthFailed, ReasonAuthFailed
	case apierrors.IsForbidden(err):
		return PhasePermissionDenied, ReasonPermissionDenied
	default:
		return PhaseUnreachable, ReasonUnreachable
	}
}

// NamespaceVisibility intersects the allowlist with existing namespaces. An
// empty allowlist returns all non-system namespaces. Results are sorted.
func NamespaceVisibility(allowlist, existing []string) (visible, missing []string) {
	set := make(map[string]bool, len(existing))
	for _, n := range existing {
		set[n] = true
	}
	if len(allowlist) == 0 {
		for _, n := range existing {
			if !isSystemNamespace(n) {
				visible = append(visible, n)
			}
		}
		sort.Strings(visible)
		return visible, nil
	}
	for _, a := range allowlist {
		if set[a] {
			visible = append(visible, a)
		} else {
			missing = append(missing, a)
		}
	}
	sort.Strings(visible)
	sort.Strings(missing)
	return visible, missing
}

// isSystemNamespace reports whether a namespace is a Kubernetes system namespace
// excluded from the default (empty-allowlist) visible set.
func isSystemNamespace(name string) bool {
	return strings.HasPrefix(name, "kube-")
}

func namespaceNames(items []corev1.Namespace) []string {
	out := make([]string, 0, len(items))
	for _, ns := range items {
		out = append(out, ns.Name)
	}
	return out
}

// canI runs a SelfSubjectAccessReview on the target cluster (allowed for any
// authenticated user) to probe a verb/resource without needing extra RBAC.
func canI(ctx context.Context, cs kubernetes.Interface, group, verb, resource, namespace string) (bool, error) {
	ssar := &authzv1.SelfSubjectAccessReview{
		Spec: authzv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authzv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     group,
				Resource:  resource,
			},
		},
	}
	resp, err := cs.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return resp.Status.Allowed, nil
}

// groupVersionExists reports whether the target cluster serves the given
// group/version (used to detect optional CRDs like Trivy Operator).
func groupVersionExists(cs kubernetes.Interface, groupVersion string) bool {
	_, err := cs.Discovery().ServerResourcesForGroupVersion(groupVersion)
	return err == nil
}
