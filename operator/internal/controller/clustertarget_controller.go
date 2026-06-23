package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/bocopile/kube-sentinel/operator/api/v1alpha1"
)

// ClusterTargetReconciler validates Biz Cluster credentials and discovers
// version/capability/namespace state into ClusterTarget.status.
type ClusterTargetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=clustertargets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=clustertargets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=clustertargets/finalizers,verbs=update
// Secret access is deliberately get-only (narrow access to the sensitive
// kubeconfig Secret, per docs/ARCHITECTURE.md). Reads MUST use the uncached
// APIReader (mgr.GetAPIReader); the cached client is not used for Secrets so no
// list/watch is required.
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

// Reconcile is the ClusterTarget reconcile loop. Skeleton: fetch + log only.
// Preflight/discovery (kubeconfig load, RBAC check, version discovery) lands in M0/M2.
func (r *ClusterTargetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var target securityv1alpha1.ClusterTarget
	if err := r.Get(ctx, req.NamespacedName, &target); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("reconciling ClusterTarget", "name", target.Name, "environment", target.Spec.Environment)

	// TODO(M0/M2): load kubeconfig Secret, validate connection/RBAC, discover
	// Kubernetes version + capabilities, and write status.phase.
	return ctrl.Result{}, nil
}

// SetupWithManager registers the controller with the manager.
func (r *ClusterTargetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.ClusterTarget{}).
		Named("clustertarget").
		Complete(r)
}
