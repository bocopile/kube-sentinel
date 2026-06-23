// Package target loads Biz Cluster credentials and performs read-only preflight
// against a target cluster using its kubeconfig. The kubeconfig value is read
// from a Mgmt Cluster Secret via the uncached APIReader and is never written to
// status, logs, events, or reports.
package target

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// connectTimeout bounds target API calls so an unreachable server fails fast
// (and surfaces as Unreachable rather than hanging the reconcile).
const connectTimeout = 10 * time.Second

// RestConfigFromKubeconfig builds a *rest.Config from raw kubeconfig bytes.
func RestConfigFromKubeconfig(data []byte) (*rest.Config, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty kubeconfig")
	}
	cfg, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	cfg.Timeout = connectTimeout
	return cfg, nil
}

// ClientSet builds a typed clientset for the target cluster.
func ClientSet(cfg *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(cfg)
}
