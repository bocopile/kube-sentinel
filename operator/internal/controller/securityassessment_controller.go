package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	securityv1alpha1 "github.com/bocopile/kube-sentinel/operator/api/v1alpha1"
)

// SecurityAssessmentReconciler maintains the last-run aggregate summary used by
// the dashboard Overview/Assessments views.
type SecurityAssessmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=securityassessments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=securityassessments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=securityassessments/finalizers,verbs=update
// +kubebuilder:rbac:groups=security.kube-sentinel.io,resources=scanruns,verbs=get;list;watch;create

// Reconcile is the SecurityAssessment reconcile loop. Skeleton: fetch + log only.
// ScanRun creation and summary aggregation land in M2.
func (r *SecurityAssessmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var assessment securityv1alpha1.SecurityAssessment
	if err := r.Get(ctx, req.NamespacedName, &assessment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("reconciling SecurityAssessment", "name", assessment.Name, "targets", assessment.Spec.Targets)

	// TODO(M2): resolve profiles→features, create/track ScanRun, aggregate
	// last-run summary into status.summary.
	return ctrl.Result{}, nil
}

// SetupWithManager registers the controller with the manager.
func (r *SecurityAssessmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1alpha1.SecurityAssessment{}).
		Named("securityassessment").
		Complete(r)
}
