# Requirements

## Goal

kube-sentinel is a Mgmt Cluster based Kubernetes final-check security
assessment PoC. Mgmt Cluster CRDs should register Biz Clusters, run delivery
artifact security assessment, inspect applied Biz Cluster configuration through
remote apply/read-only access, and publish report/evidence results to the
dashboard.

## Success criteria

| ID | Requirement | Verification |
| --- | --- | --- |
| G1 | Creating `ClusterTarget` and `SecurityAssessment` CRs lets the management controller create enabled assessment workloads on selected Biz Clusters. | `kubectl get clustertarget,securityassessment,scanrun` in the Mgmt Cluster and `kubectl get job,cronjob,cm,sa,role,rolebinding` in the Biz Cluster target namespace. |
| G2 | Feature toggles create or remove each feature's managed resources. | Patch `spec.features[].enabled` and verify resource creation/deletion. |
| G3 | Allowlisted scan resource config can change resources and scheduling fields for selected scan Jobs. | Patch `spec.scanResources`, then inspect generated workload specs and rejected forbidden fields. |
| G4 | Trivy and security assessment data is normalized into report artifacts and dashboard records. | Review normalized findings, scan health, and report artifacts. |
| G5 | Dashboard views expose finding, vulnerability, scan health, and final-check security assessment results. | Capture dashboard screenshots. |
| G6 | Final Check Dashboard exposes the assessment result by decision-oriented menus. | Capture dashboard screenshots. |
| G7 | Evidence and decision mapping passes for scope, discovery, priority, validation, and exception review. | Review evidence bundle, final decision summary, and exception review file. |
| G8 | Source code static analysis identifies risky code and security anti-patterns. | Review Semgrep/gosec reports. |
| G9 | Hardcoded secrets, tokens, credentials, and account information are detected. | Review Gitleaks reports. |
| G10 | Container image Critical vulnerabilities and risky base images are detected. | Review Trivy/Grype image scan reports. |
| G11 | Image digest, SBOM, and signature/integrity verification results are generated. | Review Syft/Cosign/Crane outputs. |
| G12 | Helm/YAML, RBAC, Dockerfile, and deployment script high-risk settings are detected. | Review kube-linter/conftest/hadolint/shellcheck reports. |
| G13 | Applied development cluster configuration risks are detected for Kubernetes YAML, RBAC, and Secret references. | Inspect rendered/applied workload specs, RBAC, ServiceAccount, Service/Ingress, and Secret reference paths. |
| G14 | Scanner failures, unsupported scans, and missing required artifacts are surfaced as failed scan health findings. | Review security assessment summary. |
| G15 | First-scope assessment support features provide reproducible inputs, scanner baselines, stable finding IDs, Secret redaction, evidence bundles, exception review, and scan health summaries. | Review `artifact-input.yaml`, scanner baseline report, normalized findings, evidence bundle, exception review file, and scan health summary. |
| G16 | Trivy Operator `VulnerabilityReport` is supported as an optional read-only input when present, without making it a mandatory dependency. | Verify optional `VulnerabilityReport` ingestion or `optional input unavailable` scan health status. |

## Non-goals

- Per-Biz-Cluster operator installation.
- Automatic infrastructure remediation or mutation of customer application
  workloads.
- Inline blocking or policy enforcement.
- Kafka or streaming middleware.
- Complete OCSF normalization.
- Production-grade high availability.

## Required project capabilities

- A Go Kubernetes operator built with controller-runtime.
- `ClusterTarget`, `SecurityAssessment`, and `ScanRun` CRDs under `security.kube-sentinel.io/v1alpha1`.
- Assessment workflow architecture for Code / Artifact Scan, Biz Cluster Scan,
  and Final Decision.
- Server-side apply for Mgmt-local resources and Biz-remote scan resources only.
- Status reporting for feature readiness, config errors, apply errors, and degraded runtime state.
- Report Store and Evidence Bundle generation for raw scanner outputs,
  normalized findings, scan health, final decision, and exception candidates.
- Result persistence and retrieval using immutable artifacts, normalized
  JSONL/JSON documents, metadata indexes, stable artifact references, and
  dashboard/API read models.
- Delivery artifact security assessment for source, secret, image, SBOM, integrity, Kubernetes manifest, RBAC, Dockerfile, and script risks.
- Applied cluster configuration assessment for Pod security settings, RBAC, Secret references, ServiceAccount token behavior, and Service/Ingress exposure.
- Separate scan phases for Code / Artifact Scan and Biz Cluster Scan so artifact failures, cluster connectivity failures, RBAC denied errors, and skipped cluster scans are not conflated.
- Scan health reporting for scanner errors, unsupported targets, missing required artifacts, and stale vulnerability databases or policy rules.
- Exception and remediation tracking for findings that require approval before delivery.
- Report generation with findings, evidence, remediation recommendations, scan
  health, and exception candidates. The PoC does not automatically fix Biz
  Cluster infrastructure.
- Assessment support features defined in
  [ASSESSMENT_SUPPORT_FEATURES.md](./ASSESSMENT_SUPPORT_FEATURES.md), with
  first-scope required functions implemented before optional telemetry or
  inventory extensions.
- Final Check Dashboard assets.

## Environment assumptions

- Mgmt Cluster stores Biz Cluster kubeconfig Secrets and runs the kube-sentinel management controller.
- Biz Clusters allow the required assessment jobs when remote scanner Jobs are enabled.
- Report Store and Dashboard storage are available in or reachable from the Mgmt Cluster.
- Report metadata storage is available for dashboard/API filtering and can be
  rebuilt from report artifacts when needed.
- Image scanners and required scanner images can be installed or executed by the selected runner.
- Biz Clusters can be queried with read-only credentials scoped to approved namespaces and cluster-level RBAC resources required for applied configuration assessment.
- Target kubeconfig Secret data is never exposed through status, dashboards, logs, or reports.
- Private registry access, approved image digest lists, and optional offline image tar artifacts are available for image vulnerability and integrity checks.
- Vulnerability databases and scanner rule sets are updated or pinned to an approved baseline date before final-check execution.
- Secret values must not be collected or written to reports; only Secret references, mounts, environment references, and ServiceAccount token settings are assessed.
- Trivy Operator `VulnerabilityReport` can be read as an optional Biz Cluster
  input only when the CRD and read-only permissions already exist. Installing
  or operating Trivy Operator is not a first-scope requirement.
