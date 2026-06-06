package controller

import (
	"github.com/bhshin/kube-sentinel/api/v1alpha1"
)

// WorkloadClient performs namespaced workload operations scoped to a specific namespace.
type WorkloadClient interface {
	List(namespace string)
}

type noopClient struct{}

func (n *noopClient) List(namespace string) {}

// SecurityAgentReconciler reconciles SecurityAgent resources.
type SecurityAgentReconciler struct {
	Client WorkloadClient
}

// Reconcile processes a SecurityAgent, scoping all workload operations to
// spec.global.targetNamespace so cluster-wide resources are not accidentally mutated.
func (r *SecurityAgentReconciler) Reconcile(agent *v1alpha1.SecurityAgent) error {
	targetNamespace := agent.Spec.Global.TargetNamespace
	client := r.Client
	if client == nil {
		client = &noopClient{}
	}
	client.List(InNamespace(targetNamespace))
	return nil
}

// InNamespace returns a namespace-scoped option for list and object operations,
// equivalent to client.InNamespace in controller-runtime.
func InNamespace(namespace string) string {
	return namespace
}
