package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/bocopile/kube-sentinel/operator/api/v1alpha1"
	"github.com/bocopile/kube-sentinel/operator/internal/feature"
)

// ScanRunReconciler orchestrates one assessment execution: resolve the enabled
// feature set, run each feature in priority order (preflight → build → apply →
// collect → normalize), persist findings, and compute the final decision.
type ScanRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=scanruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=scanruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=scanruns/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the ScanRun reconcile loop. Skeleton: fetch + log the resolved
// feature ordering only. Feature execution, remote apply, normalization, and
// final-decision computation land in M2/M3.
func (r *ScanRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var run securityv1alpha1.ScanRun
	if err := r.Get(ctx, req.NamespacedName, &run); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Demonstrate the registry: list registered features in execution order.
	ordered := feature.All()
	ids := make([]string, 0, len(ordered))
	for _, f := range ordered {
		ids = append(ids, f.ID())
	}
	log.Info("reconciling ScanRun",
		"name", run.Name,
		"assessment", run.Spec.AssessmentRef.Name,
		"retryScope", run.Annotations[securityv1alpha1.RetryScopeAnnotation],
		"registeredFeatures", ids,
	)

	// TODO(M2/M3): merge profiles+features, run feature lifecycle in priority
	// order, remote-apply scanner Jobs, normalize findings, persist to
	// PostgreSQL, and write status.finalDecision.
	return ctrl.Result{}, nil
}

// SetupWithManager registers the controller with the manager.
func (r *ScanRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.ScanRun{}).
		Named("scanrun").
		Complete(r)
}
