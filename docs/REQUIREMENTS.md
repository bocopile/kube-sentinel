# Requirements

## Goal

kube-sentinel is a Kubernetes security telemetry PoC. A single
`SecurityAgent` custom resource should deploy and manage security sensors,
an OpenTelemetry pipeline, and Elasticsearch/Kibana outputs for CTEM-oriented
security visibility.

## Success criteria

| ID | Requirement | Verification |
| --- | --- | --- |
| G1 | Applying one `SecurityAgent` CR creates the enabled sensors and OTel pipeline. | `kubectl apply` followed by `kubectl get ds,deploy,cm,secret` in the target namespace. |
| G2 | Feature toggles create or remove each feature's managed resources. | Patch `spec.features[].enabled` and verify resource creation/deletion. |
| G3 | Overrides can change resources and scheduling fields for selected features. | Patch `spec.override`, then inspect generated workload specs. |
| G4 | Falco, Tetragon, OSquery, and Trivy data lands in CTEM-specific Elasticsearch indices. | Query `security-events`, `security-inventory`, and `security-vuln`. |
| G5 | Kibana dashboards expose event, inventory, and vulnerability views. | Capture dashboard screenshots. |
| G6 | MITRE ATT&CK validation scenarios produce detectable events. | Run scenario script and query Elasticsearch for expected fields. |
| G7 | CTEM mapping checklist passes for Scope, Discovery, Priority, and Validation. | Fill and review `docs/ctem-mapping-results.md`. |

## Non-goals

- Multi-cluster management.
- Inline blocking or policy enforcement.
- Kafka or streaming middleware.
- Complete OCSF normalization.
- Production-grade high availability.

## Required project capabilities

- A Go Kubernetes operator built with controller-runtime.
- `SecurityAgent` CRD under `security.kube-sentinel.io/v1alpha1`.
- Feature-as-plugin architecture using a registry and feature priorities.
- Server-side apply for managed Kubernetes resources.
- Status reporting for feature readiness, config errors, apply errors, and degraded runtime state.
- OTel Node Collector and Gateway configuration generation.
- Elasticsearch index routing for events, inventory, and vulnerabilities.
- MITRE scenario test assets.

## Runtime assumptions

- Kubernetes cluster allows the required privileged DaemonSets and hostPath mounts.
- Worker nodes expose `/sys/kernel/btf/vmlinux` for eBPF-based tools.
- Elasticsearch and Kibana are available in or reachable from the cluster.
- ECK, Falco, Tetragon, OSquery, Trivy Operator, OTel Collector, and required images can be installed.
