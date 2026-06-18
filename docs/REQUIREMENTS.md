# Requirements

## Goal

kube-sentinel is a Kubernetes security telemetry PoC. A single
`SecurityAgent` custom resource should deploy and manage security sensors,
an OpenTelemetry pipeline, delivery artifact security assessment, applied
cluster configuration assessment, and Grafana LGTM outputs for CTEM-oriented
security visibility.

## Success criteria

| ID | Requirement | Verification |
| --- | --- | --- |
| G1 | Applying one `SecurityAgent` CR creates the enabled sensors and OTel pipeline. | `kubectl apply` followed by `kubectl get ds,deploy,cm,secret` in the target namespace. |
| G2 | Feature toggles create or remove each feature's managed resources. | Patch `spec.features[].enabled` and verify resource creation/deletion. |
| G3 | Overrides can change resources and scheduling fields for selected features. | Patch `spec.override`, then inspect generated workload specs. |
| G4 | Falco, Tetragon, OSquery, Trivy, and security assessment data lands in CTEM-specific LGTM streams and metrics. | Query Loki streams and Mimir counters for events, inventory, vulnerabilities, and findings. |
| G5 | Grafana dashboards expose event, inventory, vulnerability, and final-check security assessment views. | Capture dashboard screenshots. |
| G6 | MITRE ATT&CK validation scenarios produce detectable events. | Run scenario script and query Loki/Mimir for expected fields. |
| G7 | CTEM mapping checklist passes for Scope, Discovery, Priority, and Validation. | Fill and review `docs/ctem-mapping-results.md`. |
| G8 | Source code static analysis identifies risky code and security anti-patterns. | Review Semgrep/gosec reports. |
| G9 | Hardcoded secrets, tokens, credentials, and account information are detected. | Review Gitleaks reports. |
| G10 | Container image Critical vulnerabilities and risky base images are detected. | Review Trivy/Grype image scan reports. |
| G11 | Image digest, SBOM, and signature/integrity verification results are generated. | Review Syft/Cosign/Crane outputs. |
| G12 | Helm/YAML, RBAC, Dockerfile, and deployment script high-risk settings are detected. | Review kube-linter/conftest/hadolint/shellcheck reports. |
| G13 | Applied development cluster configuration risks are detected for Kubernetes YAML, RBAC, and Secret references. | Inspect rendered/applied workload specs, RBAC, ServiceAccount, Service/Ingress, and Secret reference paths. |
| G14 | Scanner failures, unsupported scans, and missing required artifacts are surfaced as failed scan health findings. | Review security assessment summary. |

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
- Grafana LGTM routing for events, inventory, vulnerabilities, and normalized security findings.
- Delivery artifact security assessment for source, secret, image, SBOM, integrity, Kubernetes manifest, RBAC, Dockerfile, and script risks.
- Applied cluster configuration assessment for Pod security settings, RBAC, Secret references, ServiceAccount token behavior, and Service/Ingress exposure.
- Scan health reporting for scanner errors, unsupported targets, missing required artifacts, and stale vulnerability databases or policy rules.
- Exception and remediation tracking for findings that require approval before delivery.
- MITRE scenario test assets.

## Environment assumptions

- Kubernetes cluster allows the required privileged DaemonSets and hostPath mounts.
- Worker nodes expose `/sys/kernel/btf/vmlinux` for eBPF-based tools.
- Loki, Mimir, Tempo, and Grafana are available in or reachable from the cluster.
- Falco, Tetragon, OSquery, Trivy Operator, OTel Collector, and required images can be installed.
- The development cluster can be queried with read-only credentials scoped to approved namespaces and cluster-level RBAC resources required for applied configuration assessment.
- Private registry access, approved image digest lists, and optional offline image tar artifacts are available for image vulnerability and integrity checks.
- Vulnerability databases and scanner rule sets are updated or pinned to an approved baseline date before final-check execution.
- Secret values must not be collected or written to reports; only Secret references, mounts, environment references, and ServiceAccount token settings are assessed.
