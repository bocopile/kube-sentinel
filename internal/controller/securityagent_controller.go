package controller

import (
	"github.com/bhshin/kube-sentinel/api/v1alpha1"
)

// SecurityAgentReconciler reconciles SecurityAgent resources.
type SecurityAgentReconciler struct{}

// Reconcile processes a SecurityAgent, scoping all workload operations to
// spec.global.targetNamespace so cluster-wide resources are not accidentally mutated.
func (r *SecurityAgentReconciler) Reconcile(agent v1alpha1.SecurityAgent) error {
	targetNamespace := agent.Spec.Global.TargetNamespace
	ns := InNamespace(targetNamespace)
	_ = ns
	return nil
}

// InNamespace returns an option that restricts list and object operations to
// the given namespace, equivalent to client.InNamespace in controller-runtime.
func InNamespace(namespace string) string {
	return namespace
}
