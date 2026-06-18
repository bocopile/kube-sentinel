# Requirements

## Goal

kube-sentinel is a Mgmt Cluster based Kubernetes final-check security
assessment PoC. Mgmt Cluster CRDs should register Biz Clusters, run delivery
artifact security assessment, inspect applied Biz Cluster configuration through
remote apply/read-only access, and publish results to Grafana LGTM.

## Success criteria

| ID | Requirement | Verification |
| --- | --- | --- |
| G1 | Creating `ClusterTarget` and `SecurityAssessment` CRs lets the management controller create enabled assessment workloads on selected Biz Clusters. | `kubectl get clustertarget,securityassessment,scanrun` in the Mgmt Cluster and `kubectl get job,cronjob,deploy,cm` in the Biz Cluster target namespace. |
| G2 | Feature toggles create or remove each feature's managed resources. | Patch `spec.features[].enabled` and verify resource creation/deletion. |
| G3 | Overrides can change resources and scheduling fields for selected features. | Patch `spec.override`, then inspect generated workload specs. |
| G4 | OSquery, Trivy, and security assessment data lands in CTEM-specific LGTM streams and metrics. | Query Loki streams and Mimir counters for inventory, vulnerabilities, and findings. |
| G5 | Grafana dashboards expose event, inventory, vulnerability, and final-check security assessment views. | Capture dashboard screenshots. |
| G6 | Final Check Dashboard exposes the assessment result by decision-oriented menus. | Capture dashboard screenshots. |
| G7 | CTEM mapping checklist passes for Scope, Discovery, Priority, and Validation. | Fill and review `docs/ctem-mapping-results.md`. |
| G8 | Source code static analysis identifies risky code and security anti-patterns. | Review Semgrep/gosec reports. |
| G9 | Hardcoded secrets, tokens, credentials, and account information are detected. | Review Gitleaks reports. |
| G10 | Container image Critical vulnerabilities and risky base images are detected. | Review Trivy/Grype image scan reports. |
| G11 | Image digest, SBOM, and signature/integrity verification results are generated. | Review Syft/Cosign/Crane outputs. |
| G12 | Helm/YAML, RBAC, Dockerfile, and deployment script high-risk settings are detected. | Review kube-linter/conftest/hadolint/shellcheck reports. |
| G13 | Applied development cluster configuration risks are detected for Kubernetes YAML, RBAC, and Secret references. | Inspect rendered/applied workload specs, RBAC, ServiceAccount, Service/Ingress, and Secret reference paths. |
| G14 | Scanner failures, unsupported scans, and missing required artifacts are surfaced as failed scan health findings. | Review security assessment summary. |

## Non-goals

- Per-Biz-Cluster operator installation.
- Inline blocking or policy enforcement.
- Kafka or streaming middleware.
- Complete OCSF normalization.
- Production-grade high availability.

## Required project capabilities

- A Go Kubernetes operator built with controller-runtime.
- `ClusterTarget`, `SecurityAssessment`, and `ScanRun` CRDs under `security.kube-sentinel.io/v1alpha1`.
- Feature-as-plugin architecture using a registry and feature priorities.
- Server-side apply for Mgmt-local and Biz-remote Kubernetes resources.
- Status reporting for feature readiness, config errors, apply errors, and degraded runtime state.
- OTel Node Collector and Gateway configuration generation.
- Grafana LGTM routing for events, inventory, vulnerabilities, and normalized security findings.
- Delivery artifact security assessment for source, secret, image, SBOM, integrity, Kubernetes manifest, RBAC, Dockerfile, and script risks.
- Applied cluster configuration assessment for Pod security settings, RBAC, Secret references, ServiceAccount token behavior, and Service/Ingress exposure.
- Separate scan phases for Code / Artifact Scan and Biz Cluster Scan so artifact failures, cluster connectivity failures, RBAC denied errors, and skipped cluster scans are not conflated.
- Scan health reporting for scanner errors, unsupported targets, missing required artifacts, and stale vulnerability databases or policy rules.
- Exception and remediation tracking for findings that require approval before delivery.
- Final Check Dashboard assets.

## Environment assumptions

- Mgmt Cluster stores Biz Cluster kubeconfig Secrets and runs the kube-sentinel management controller.
- Biz Clusters allow the required assessment jobs and optional inventory DaemonSets.
- Loki, Mimir, Tempo, and Grafana are available in or reachable from the cluster.
- OSquery, OTel Collector, image scanners, and required images can be installed.
- Biz Clusters can be queried with read-only credentials scoped to approved namespaces and cluster-level RBAC resources required for applied configuration assessment.
- Target kubeconfig Secret data is never exposed through status, dashboards, logs, or reports.
- Private registry access, approved image digest lists, and optional offline image tar artifacts are available for image vulnerability and integrity checks.
- Vulnerability databases and scanner rule sets are updated or pinned to an approved baseline date before final-check execution.
- Secret values must not be collected or written to reports; only Secret references, mounts, environment references, and ServiceAccount token settings are assessed.
