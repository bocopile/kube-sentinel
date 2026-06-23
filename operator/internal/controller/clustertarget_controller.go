package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/bocopile/kube-sentinel/operator/api/v1alpha1"
	"github.com/bocopile/kube-sentinel/operator/internal/target"
)

const defaultKubeconfigNamespace = "kube-sentinel-system"

// Requeue intervals by phase (PLAN.md: PoC periodic re-validation default 60s).
const (
	requeueReady   = 60 * time.Second
	requeueDegrade = 30 * time.Second
	requeueFailed  = 5 * time.Minute
)

// ClusterTargetReconciler validates Biz Cluster credentials and discovers
// version/capability/namespace state into ClusterTarget.status (M0 readiness).
type ClusterTargetReconciler struct {
	client.Client
	// APIReader is the uncached reader used to fetch the kubeconfig Secret
	// (get-only; the cached client is not used for Secrets).
	APIReader client.Reader
	Scheme    *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=clustertargets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=clustertargets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=clustertargets/finalizers,verbs=update
// Secret access is deliberately get-only (narrow access to the sensitive
// kubeconfig Secret, per docs/ARCHITECTURE.md). Reads MUST use the uncached
// APIReader (mgr.GetAPIReader); the cached client is not used for Secrets so no
// list/watch is required.
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

// Reconcile runs the ClusterTarget readiness preflight and records the result.
func (r *ClusterTargetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var ct securityv1alpha1.ClusterTarget
	if err := r.Get(ctx, req.NamespacedName, &ct); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	res := r.preflight(ctx, &ct)
	log.Info("ClusterTarget preflight",
		"name", ct.Name,
		"phase", res.Phase,
		"kubernetesVersion", res.KubernetesVersion,
		"visibleNamespaces", len(res.Namespaces),
	)

	r.applyStatus(&ct, res)
	if err := r.Status().Update(ctx, &ct); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: requeueFor(res.Phase)}, nil
}

// preflight loads the kubeconfig Secret and runs the target readiness checks.
func (r *ClusterTargetReconciler) preflight(ctx context.Context, ct *securityv1alpha1.ClusterTarget) target.PreflightResult {
	ref := ct.Spec.KubeconfigRef
	ns := ref.Namespace
	if ns == "" {
		ns = defaultKubeconfigNamespace
	}
	key := ref.Key
	if key == "" {
		key = "kubeconfig"
	}

	var secret corev1.Secret
	if err := r.APIReader.Get(ctx, types.NamespacedName{Namespace: ns, Name: ref.Name}, &secret); err != nil {
		// Do not include the error verbatim (may reference Secret contents indirectly); name the ref only.
		return target.FailResult(target.PhaseAuthFailed, target.CondKubeconfigValid, target.ReasonInvalidSecret,
			fmt.Sprintf("kubeconfig Secret %s/%s unavailable", ns, ref.Name))
	}

	cfg, err := target.RestConfigFromKubeconfig(secret.Data[key])
	if err != nil {
		return target.FailResult(target.PhaseAuthFailed, target.CondKubeconfigValid, target.ReasonInvalidKubeconfig,
			fmt.Sprintf("kubeconfig key %q is missing or malformed", key))
	}
	cs, err := target.ClientSet(cfg)
	if err != nil {
		return target.FailResult(target.PhaseAuthFailed, target.CondKubeconfigValid, target.ReasonInvalidKubeconfig,
			"failed to build target client from kubeconfig")
	}

	res := target.Preflight(ctx, cs, target.PreflightInput{
		TargetNamespace:          ct.Spec.TargetNamespace,
		NamespaceAllowlist:       ct.Spec.NamespaceAllowlist,
		WantTrivyOperatorReports: ct.Spec.Capabilities.TrivyOperatorReports,
	})
	// The kubeconfig was valid enough to build a client; record it.
	res.Conditions = append([]target.ConditionResult{
		{Type: target.CondKubeconfigValid, OK: true, Reason: target.ReasonReady},
	}, res.Conditions...)
	return res
}

// applyStatus projects a PreflightResult onto the ClusterTarget status.
func (r *ClusterTargetReconciler) applyStatus(ct *securityv1alpha1.ClusterTarget, res target.PreflightResult) {
	ct.Status.Phase = res.Phase
	ct.Status.ObservedGeneration = ct.Generation
	ct.Status.KubernetesVersion = res.KubernetesVersion
	ct.Status.Namespaces = res.Namespaces
	ct.Status.Capabilities = securityv1alpha1.TargetCapabilityStatus{
		ScannerJobs:          res.Capabilities.ScannerJobs,
		ReadOnlyInspection:   res.Capabilities.ReadOnlyInspection,
		TrivyOperatorReports: res.Capabilities.TrivyOperatorReports,
		// ImageAccess/ReportUpload probes land in M1; HostPath stays false (reserved).
	}

	// lastValidatedAt advances only when a full validation pass completed
	// (target reachable + read RBAC ok): Ready or Degraded.
	if res.Phase == target.PhaseReady || res.Phase == target.PhaseDegraded {
		ct.Status.LastValidatedAt = metav1.Now()
	}

	for _, c := range res.Conditions {
		status := metav1.ConditionFalse
		if c.OK {
			status = metav1.ConditionTrue
		}
		reason := c.Reason
		if reason == "" {
			reason = target.ReasonReady
		}
		meta.SetStatusCondition(&ct.Status.Conditions, metav1.Condition{
			Type:               c.Type,
			Status:             status,
			Reason:             reason,
			Message:            c.Message,
			ObservedGeneration: ct.Generation,
		})
	}
}

func requeueFor(phase string) time.Duration {
	switch phase {
	case target.PhaseReady:
		return requeueReady
	case target.PhaseDegraded:
		return requeueDegrade
	default:
		return requeueFailed
	}
}

// SetupWithManager registers the controller with the manager.
func (r *ClusterTargetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.ClusterTarget{}).
		Named("clustertarget").
		Complete(r)
}
