# Orchestrator prompts

Use these prompts with `orchestrator plan` first, then `orchestrator run` only
after the plan looks acceptable.

All prompts assume the project root is the current checkout of:

```text
github.com/bocopile/kube-sentinel
```

## Command pattern

Dry run:

```bash
orchestrator plan --project . --request "<prompt>"
```

Implementation:

```bash
orchestrator run --project . --request "<prompt>" --auto-approve
```

## Milestone mapping

| Prompt | Roadmap stage | Roadmap milestone | Purpose |
| --- | --- | --- | --- |
| P0 | Foundation | First implementation block | Go management controller skeleton and core API contracts. |
| P1 | S0 | M0 | Assessment readiness checks. |
| P2 | S0.5 | M0.5 | Delivery artifact security assessment baseline. |
| P3 | S1 | M1 | Report store, finding schema, evidence bundle, and dashboard backend. |
| P4 | S2 | M2 | Management controller core and security assessment scaffold. |
| P5 | S2 | M3 | Security Assessment feature. |
| P6 | S3 | M4 | Applied cluster configuration scan. |
| P7 | S3 | M5 | Trivy delivery image scan, image integrity, and optional VulnerabilityReport ingestion. |
| P8 | S5 | M6 | Phase 2 optional inventory/telemetry extension. |
| P9 | S4 | M7 | Final Check Dashboard. |
| P10 | S4 | M8 | Final-check validation, reports, exceptions, and garbage collection. |

## Global instruction block

Add this block to milestone prompts when the request is complex:

```text
Use docs/PLAN.md as the source plan, and use docs/REQUIREMENTS.md,
docs/ARCHITECTURE.md, docs/SECURITY_ASSESSMENT.md,
docs/ASSESSMENT_SUPPORT_FEATURES.md, docs/FRONTEND_ARCHITECTURE.md,
docs/ROADMAP.md, and docs/ORCHESTRATOR.md as the implementation contract.
Keep changes scoped to the requested milestone. Do not implement Phase 2
inventory, telemetry, runtime sensors, or automatic remediation unless they are
explicitly part of the milestone. Preserve buildability after the change. Add
focused tests for new logic. Verification must include go test ./... and
go build ./....
```

## P0 - Create project skeleton

```text
Use docs/PLAN.md, docs/REQUIREMENTS.md, docs/ARCHITECTURE.md, and
docs/ROADMAP.md as the project contract.

Implement the first kube-sentinel code block from docs/ROADMAP.md only:

- Ensure this is a Go Kubernetes management controller project using module
  github.com/bocopile/kube-sentinel.
- Add or complete ClusterTarget, SecurityAssessment, and ScanRun API types under
  api/v1alpha1.
- Add an empty but buildable controller reconciler.
- Add assessment registry interfaces and deterministic priority ordering.
- Add tests for registry ordering and unknown feature validation.
- Do not implement optional inventory, OTel manifests, LGTM integration,
  runtime sensors, security assessment jobs, Trivy, or dashboards yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- ClusterTarget contains target kubeconfigRef, targetNamespace,
  namespaceAllowlist, output, capabilities, and status fields.
- SecurityAssessment contains selected targets and scan profiles.
- ScanRun contains scan execution status and per-target results.
- Unknown feature names can be detected and reported by pure Go unit tests.
- Registry ordering is deterministic by priority and feature ID.
```

## P1 - M0 assessment readiness checks

```text
Use docs/ROADMAP.md S0/M0 and docs/ASSESSMENT_SUPPORT_FEATURES.md as the target.

Implement assessment readiness assets for kube-sentinel:

- Namespace manifest for kube-sentinel-system.
- Target preflight check for kubeconfig presence, API reachability, namespace
  existence, read-only RBAC, image pull access, and report store write access.
- Guard that detects accidental Secret read permission in target credentials and
  reports it as a preflight risk.
- Documentation for how to run and interpret the checks.

Do not implement runtime sensors, OTel/LGTM telemetry, privileged DaemonSets, or
automatic remediation.

Acceptance criteria:

- go test ./... passes if Go packages exist.
- go build ./... passes if Go packages exist.
- Kubernetes YAML can be rendered or applied with documented commands.
- Preflight distinguishes target environment failures from scanner findings.
- Secret raw values are not read.
```

## P2 - M0.5 delivery artifact security assessment baseline

```text
Use docs/SECURITY_ASSESSMENT.md, docs/ASSESSMENT_SUPPORT_FEATURES.md, and
docs/ROADMAP.md S0.5/M0.5 as the target.

Implement the first security assessment baseline:

- Scanner configuration placeholders for Semgrep/gosec, Gitleaks, Trivy/Grype,
  Syft, Cosign/Notation, Crane, kube-linter, conftest, Hadolint, and ShellCheck.
- artifact-input.example.yaml for source paths, image list, digest list,
  Helm/YAML, RBAC, Dockerfile, and scripts.
- Scanner version and vulnerability DB/rule baseline capture.
- scripts/run-security-assessment.sh orchestration skeleton.
- scripts/verify-image-digest.sh for approved digest comparison.
- scripts/normalize-findings.sh placeholder for scanner result normalization.
- Scan health output for missing artifacts, unsupported targets, scanner
  errors, stale baselines, and registry pull failures.

Do not implement runtime event correlation, OSQuery, OTel/LGTM, or automatic
remediation.
M0.5 creates scanner configuration, input validation, baseline capture, and
scan-health skeletons only. Actual delivery image vulnerability scanning is
implemented in M5.

Acceptance criteria:

- go test ./... passes if Go packages exist.
- go build ./... passes if Go packages exist.
- Running the assessment script without required inputs reports scan health
  failures rather than a false pass.
- Required artifact inputs are documented.
- Scanner baseline data is written with the report.
- No Secret raw values are written to reports.
```

## P3 - M1 report store, schema, evidence, and dashboard backend

```text
Implement M1 from docs/ROADMAP.md.

Scope:

- Report Store interfaces for raw scanner reports, normalized findings, scan
  health, final decision records, and evidence bundles.
- Security Finding Schema and schema validator.
- Stable finding ID and deduplication helpers.
- Secret redaction guard for reports, logs, dashboard records, and artifacts.
- Evidence bundle export structure.
- Base dashboard/read-model records for Overview, Targets, Assessments,
  Findings, Reports, and Governance.

Do not implement OTel/LGTM telemetry or Grafana-specific dashboards in this
milestone.

Acceptance criteria:

- go test ./... passes if Go packages exist.
- go build ./... passes if Go packages exist.
- Duplicate fixture findings produce the same stable finding ID.
- Invalid normalized findings fail schema validation.
- Evidence bundle references raw report, normalized findings, scan health,
  final decision, and exception candidates.
- Secret-like fixture values are redacted or rejected before persistence.
```

## P4 - M2 management controller core and assessment scaffold

```text
Implement M2 from docs/ROADMAP.md.

Scope:

- ClusterTarget, SecurityAssessment, and ScanRun reconciler core.
- Finalizer handling.
- Assessment registry integration.
- Desired state store.
- Remote apply client skeleton using ClusterTarget kubeconfigRef.
- Server-side apply skeleton with managed labels and annotations from
  docs/ARCHITECTURE.md.
- Status patching with observedGeneration and workflow conditions.
- security_assessment feature scaffold that can create assessment Job/CronJob
  resources without implementing all scanner logic.
- Report writer skeleton for ScanRun results.

Do not implement optional inventory, OTel/LGTM, runtime sensors, automatic
remediation, or Trivy feature logic yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Unit tests cover finalizer behavior, unknown feature status, registry
  ordering, desired state labels, remote apply label generation, and status
  phase calculation.
- Sample ClusterTarget, SecurityAssessment, and ScanRun YAML exists for a
  minimal assessment deployment.
```

## P5 - M3 Security Assessment feature

```text
Implement M3 from docs/ROADMAP.md: the Security Assessment feature.

Scope:

- security_assessment feature config defaults and validation.
- Assessment Job/CronJob resources for delivery artifact scans.
- Scanner config mount points and report output conventions.
- Finding normalization invocation.
- Scan health reporting for scanner failures and missing artifacts.
- Artifact input manifest validation.
- Scanner baseline capture.

Do not implement optional inventory, Trivy delivery image scan, or applied
cluster configuration scan yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Generated assessment resources contain kube-sentinel ownership labels.
- Disabling the security_assessment feature removes or marks stale run-scoped
  resources for GC.
- Scanner failures are represented as scan health findings.
- Evidence bundle output includes raw report and normalized finding references.
```

## P6 - M4 Applied cluster configuration scan

```text
Implement M4 from docs/ROADMAP.md: applied cluster configuration scan.

Scope:

- Read-only Kubernetes client access for approved namespaces.
- Workload spec inspection for securityContext, volume, image, and
  ServiceAccount settings.
- RBAC inspection for Role, RoleBinding, ClusterRole, and ClusterRoleBinding
  risks.
- Secret reference inspection without reading raw Secret values.
- Service/Ingress exposure inspection as an optional warning category.
- Namespace allowlist validator.
- Normalized findings for applied configuration risks.

Do not implement optional inventory, runtime sensors, or automatic remediation.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Applied cluster inspection uses read-only permissions.
- Secret raw values are not read or persisted.
- Sample SecurityAssessment can enable security_assessment with applied cluster
  scan settings.
- Documentation includes validation commands and expected report fields.
```

## P7 - M5 Trivy delivery image scan and integrity

```text
Implement M5 from docs/ROADMAP.md: Trivy delivery image scan plus image
integrity.

Scope:

- trivy feature config defaults and validation for delivery image scanning.
- Registry digest or image tar scan flow.
- SBOM generation using Syft or Trivy SBOM output.
- Digest verification using Crane and approved digest lists.
- Optional signature verification hook for Cosign or Notation.
- Optional read-only Trivy Operator VulnerabilityReport ingestion when the CRD
  exists and the ClusterTarget has get/list/watch permission.
- Deterministic finding ID:
  <imageRepository>/<imageDigest>/<vulnerabilityID>/<packageName>
- Tests for duplicate-safe finding generation across direct Trivy scan and
  optional VulnerabilityReport input.

Do not install or operate Trivy Operator as part of this milestone. Do not fail
the whole assessment when VulnerabilityReport is unavailable; record optional
input unavailable in scan health.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Duplicate Trivy fixture ingestion produces the same finding ID.
- Optional VulnerabilityReport fixture ingestion normalizes to the same finding
  schema.
- Vulnerability findings are written to Report Store records and evidence
  bundles.
- Documentation includes install-independent verification commands.
```

## P8 - M6 Phase 2 optional inventory/telemetry extension

```text
Implement M6 from docs/ROADMAP.md only if Phase 2 inventory or telemetry is
approved after a separate design review.

Scope candidates:

- OSQuery or equivalent inventory sensor.
- OTel/LGTM export path from normalized findings and report events.
- Runtime event or drift assessment.
- Long-running sensor DaemonSet model.

Do not implement this during the first final-check PoC unless the product scope
explicitly requires it.

Acceptance criteria must be defined in a separate design document before work
starts.
```

## P9 - M7 Final Check Dashboard

```text
Implement M7 from docs/ROADMAP.md and docs/FRONTEND_ARCHITECTURE.md.

Scope:

- Final Check Dashboard assets for Overview, Targets, Assessments, Findings,
  Reports, and Governance.
- Findings table or documented panel query conventions.
- Report menu for final-check reports, evidence bundles, raw reports,
  normalized findings, and scan health summaries.
- Dashboard variables for environment, target version/build, scan run ID,
  namespace, image, severity, category, scanner, scan status, and exception
  status.

Acceptance criteria:

- go test ./... passes if Go packages exist.
- go build ./... passes if Go packages exist.
- Dashboard assets or setup instructions are deterministic.
- Screenshots or documented queries cover each menu.
- Reports menu exposes evidence bundle and final decision data.
```

## P10 - M8 final-check validation

```text
Implement M8 from docs/ROADMAP.md.

Scope:

- End-to-end validation assets for Code / Artifact Scan, Biz Cluster Scan, and
  Full Final Check.
- Garbage collection verification for disabled profiles and stale ScanRuns.
- Delivery artifact assessment validation.
- Applied cluster configuration assessment validation.
- Secret redaction validation.
- Evidence bundle and exception review validation.
- No-auto-remediation guardrail validation.
- Documentation of expected kubectl diff/get output and final-check report
  output.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Validation scripts cover security_assessment and trivy.
- Stale resource cleanup behavior is documented.
- Final-check report output includes scan health, evidence bundle references,
  exception status, and no automatic Biz Cluster infrastructure mutation.
```

## Prompt quality checklist

Before running `orchestrator run`, verify the prompt has:

- A single milestone target.
- Explicit files or modules in scope.
- Explicit out-of-scope items.
- At least three acceptance criteria.
- Required verification commands.
- References to docs rather than restating the whole plan.
