# Architecture

## Overview

kube-sentinel is centered on one custom resource:

```yaml
apiVersion: security.kube-sentinel.io/v1alpha1
kind: SecurityAgent
```

The operator reconciles this CR into sensor workloads, OpenTelemetry pipeline
resources, Elasticsearch ingestion helpers, and status conditions.

## API scope decision

`SecurityAgent` is cluster-scoped.

Kubebuilder must create the API with:

```bash
kubebuilder create api --group security --version v1alpha1 --kind SecurityAgent --namespaced=false
```

Rationale:

- The managed sensors are cluster-wide DaemonSets.
- The operator needs cluster-scoped RBAC for nodes, ClusterRoles,
  ClusterRoleBindings, and third-party CRDs such as Tetragon policies.
- A namespaced owner cannot own cluster-scoped resources cleanly.
- Changing CRD scope later requires a CRD migration.

Namespaced workloads should be placed in `spec.global.targetNamespace`, defaulting
to `kube-sentinel-system`. Samples should not set `metadata.namespace` on the
`SecurityAgent` object.

## Main components

| Component | Responsibility |
| --- | --- |
| `SecurityAgent` CRD | Single user-facing entry point for features, output, overrides, and tests. |
| Controller | Reconciles desired state, applies resources, performs garbage collection, and updates status. |
| Feature registry | Builds enabled features in deterministic priority order. |
| Desired state store | Collects Kubernetes objects contributed by features before apply. |
| Override layer | Applies global node-agent overrides and feature-specific overrides. |
| OTel config builder | Merges receiver/exporter fragments into Node Collector and Gateway configs. |
| Feature packages | Own tool-specific defaults, config validation, resources, OTel fragments, and readiness checks. |

## Managed infrastructure boundary

The kube-sentinel operator does not create Elasticsearch, Kibana, or ECK
resources.

The `otel_pipeline` feature manages only kube-sentinel collection components:

- OTel Node Collector DaemonSet and ConfigMap.
- OTel Gateway Deployment, Service, ConfigMap, and related RBAC.
- Pipeline wiring from enabled feature receiver fragments to the configured
  Elasticsearch endpoint.

Elasticsearch and Kibana are prerequisites. PoC installation assets may live
under `config/elasticsearch/`, but they are applied manually or by a separate
platform workflow during M1. The operator reads `spec.output.elasticsearch` and
reports connection/runtime failures through status; it must not reconcile ECK
`Elasticsearch` or `Kibana` custom resources.

## Feature priorities

| Priority | Feature | Reason |
| --- | --- | --- |
| 10 | `otel_pipeline` | Collection infrastructure must exist before sensors emit data. |
| 100 | `falco` | Runtime event sensor. |
| 100 | `tetragon` | Runtime event sensor. |
| 100 | `osquery` | Inventory sensor. |
| 200 | `trivy` | Vulnerability ingestion depends on Trivy reports and direct Elasticsearch upsert. |

## Reconcile flow

1. Add finalizer.
2. Load `SecurityAgent` spec.
3. Validate feature names and feature config.
4. Build active features in priority order.
5. Ask each feature to contribute Kubernetes resources.
6. Collect OTel receiver fragments and generate OTel configs.
7. Apply overrides.
8. Apply desired resources using server-side apply.
9. Garbage collect disabled or stale feature resources.
10. Assess feature readiness and patch status.

## Override policy

Overrides are allowlisted, not arbitrary patches.

Allowed override fields:

| Path | Allowed fields |
| --- | --- |
| `override.nodeAgent` | `resources`, `nodeSelector`, `affinity`, `tolerations` |
| `override.otelGateway` | `resources`, `replicas`, `nodeSelector`, `affinity`, `tolerations` |
| `override.falco` | `resources`, `nodeSelector`, `affinity`, `tolerations` |
| `override.tetragon` | `resources`, `nodeSelector`, `affinity`, `tolerations` |
| `override.osquery` | `resources`, `nodeSelector`, `affinity`, `tolerations` |
| `override.trivy` | `resources`, `scanSchedule`, `severityThreshold` |

Forbidden override behavior:

- Adding arbitrary containers, init containers, volumes, hostPath mounts, service
  account names, image names, image pull policies, security contexts, commands,
  or arguments.
- Adding `tolerations: [{ operator: Exists }]`.
- Tolerating control-plane taints unless the operator is configured with an
  explicit installation-time allow-control-plane setting.
- Raising privileges beyond each feature's built-in security context.

Toleration validation must be implemented before applying overrides. Invalid
overrides set the relevant feature to `ConfigError` and must not be applied.

## HostPath policy

HostPath mounts are feature-owned and fixed by code. Overrides cannot add or
change them.

Minimum intended hostPath set:

| Feature | Path | Access | Purpose |
| --- | --- | --- | --- |
| `otel_pipeline` | `/var/log/pods` | read-only | Collect Kubernetes pod logs for Tetragon and other stdout sources. |
| `otel_pipeline` | `/var/log/containers` | read-only | Runtime log symlink compatibility. |
| `otel_pipeline` | `/var/log/kube-sentinel` | read-write | Shared sensor file-log directory. |
| `falco` | `/var/log/kube-sentinel/falco` | read-write | Falco JSON event output. |
| `falco` | `/sys/kernel/btf` | read-only | eBPF BTF discovery. |
| `tetragon` | `/sys/kernel/btf` | read-only | eBPF BTF discovery. |
| `osquery` | `/var/log/kube-sentinel/osquery` | read-write | OSquery result logs. |

Additional host paths require an architecture update and a security review.

## Ownership model

Every managed object should include:

```yaml
metadata:
  labels:
    app.kubernetes.io/managed-by: kube-sentinel
    security.kube-sentinel.io/instance: <security-agent-name>
    security.kube-sentinel.io/feature: <feature-id>
  annotations:
    security.kube-sentinel.io/spec-hash: <sha256>
```

Server-side apply field managers should use:

```text
kube-sentinel/<feature-id>
```

## Data routing

| Source | Collection path | Destination index | CTEM phase |
| --- | --- | --- | --- |
| Falco | File log through OTel Node Collector | `security-events` | Validation |
| Tetragon | Pod log through OTel Node Collector | `security-events` | Validation |
| OSquery | File log through OTel Node Collector | `security-inventory` | Scope |
| Trivy | VulnerabilityReport read by ingestor job | `security-vuln` | Discovery / Priority |

## OTel resiliency policy

The OTel config builder must generate bounded failure behavior. Elasticsearch
outages must not cause unbounded memory growth.

Required defaults:

- `memory_limiter` processor enabled in Node Collector and Gateway.
- `batch` processor enabled with bounded batch sizes.
- Elasticsearch exporter timeout set explicitly.
- Exporter sending queue enabled with a bounded queue size.
- Retry enabled with finite backoff and finite max elapsed time.
- Data is dropped after retry exhaustion and reflected in collector metrics.
- Persistent disk queue is out of scope for the PoC unless explicitly enabled in
  a later production profile.

The config builder types should represent this explicitly, for example:

```go
type OTelExporterConfig struct {
    Endpoint          string
    Timeout          metav1.Duration
    QueueSize         int
    NumConsumers      int
    RetryInitial      metav1.Duration
    RetryMax          metav1.Duration
    RetryMaxElapsed   metav1.Duration
    MemoryLimitMiB    int
    MemorySpikeMiB    int
}
```

Operational status should surface export failures through OTel metrics and the
`otel_pipeline` feature condition.

## Controller RBAC

The controller needs explicit RBAC for the resources it reconciles or observes.
Kubebuilder markers must be derived from this list rather than added ad hoc
during feature implementation.

Core resources:

- `securityagents`: get, list, watch, create, update, patch, delete
- `securityagents/status`: get, update, patch
- `securityagents/finalizers`: update
- `namespaces`: get, list, watch
- `nodes`: get, list, watch
- `pods`: get, list, watch
- `configmaps`: get, list, watch, create, update, patch, delete
- `secrets`: get, list, watch
- `services`: get, list, watch, create, update, patch, delete
- `serviceaccounts`: get, list, watch, create, update, patch, delete
- `events`: create, patch

Workload resources:

- `apps/daemonsets`: get, list, watch, create, update, patch, delete
- `apps/deployments`: get, list, watch, create, update, patch, delete
- `batch/jobs`: get, list, watch, create, update, patch, delete
- `batch/cronjobs`: get, list, watch, create, update, patch, delete

RBAC resources:

- `rbac.authorization.k8s.io/roles`: get, list, watch, create, update, patch, delete
- `rbac.authorization.k8s.io/rolebindings`: get, list, watch, create, update, patch, delete
- `rbac.authorization.k8s.io/clusterroles`: get, list, watch, create, update, patch, delete
- `rbac.authorization.k8s.io/clusterrolebindings`: get, list, watch, create, update, patch, delete

Third-party resources:

- Tetragon `TracingPolicy` resources: get, list, watch, create, update, patch, delete
- Trivy `VulnerabilityReport` resources: get, list, watch

Secrets are read-only. The operator must not create or mutate Elasticsearch
credentials.

## Status model

The operator should expose:

- `status.observedGeneration`
- `status.phase`: `Ready`, `Progressing`, or `Degraded`
- `status.features[]`
- `status.managedResources[]`

Feature status reasons should include:

- `Disabled`
- `Ready`
- `ConfigError`
- `ApplyError`
- `NotReady`

Unknown feature names are configuration errors and must not create resources.
