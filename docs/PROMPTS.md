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

For larger stages, omit `--auto-approve` if you want to manually approve the
task graph.

## Milestone mapping

| Prompt | Roadmap stage | Roadmap milestone | Purpose |
| --- | --- | --- | --- |
| P0 | Foundation | First implementation block | Go management controller skeleton and core API contracts. |
| P1 | S0 | M0 | Cluster prerequisite checks and LGTM connectivity. |
| P2 | S0.5 | M0.5 | Delivery artifact security assessment baseline. |
| P3 | S1 | M1 | Grafana LGTM backend and base dashboard checks. |
| P4 | S2 | M2 | Operator core, OTel feature, and security assessment scaffold. |
| P5 | S2 | M3 | Security Assessment feature. |
| P6 | S3 | M4 | Applied cluster configuration scan. |
| P7 | S3 | M5 | OSquery feature. |
| P8 | S3 | M6 | Trivy delivery image scan and image integrity feature. |
| P9 | S4 | M7 | Final Check Dashboard. |
| P10 | S4 | M8 | Toggle, override, final-check validation, and garbage collection. |

## Global instruction block

Add this block to milestone prompts when the request is complex:

```text
Use docs/PLAN.md as the source plan, and use docs/REQUIREMENTS.md,
docs/ARCHITECTURE.md, docs/SECURITY_ASSESSMENT.md,
docs/FRONTEND_ARCHITECTURE.md, docs/ROADMAP.md, and docs/ORCHESTRATOR.md as
the implementation contract. Keep changes scoped to the requested milestone.
Do not implement later sensors or dashboards unless they are explicitly part of
the milestone. Preserve buildability after the change. Add focused tests for
new logic. Verification must include go test ./... and go build ./....
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
- Add feature registry interfaces and deterministic priority ordering.
- Add tests for registry ordering and unknown feature validation.
- Do not implement OSquery, Trivy, OTel manifests, LGTM
  integration, security assessment jobs, or dashboards yet.

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

## P1 - M0 infrastructure readiness checks

```text
Use docs/ROADMAP.md S0/M0 as the target.

Implement cluster prerequisite assets for kube-sentinel:

- Namespace manifest for kube-sentinel-system.
- A privileged preflight DaemonSet or Job that verifies required host access.
- A script that checks /sys/kernel/btf/vmlinux from a node-level pod.
- A script or manifest for sending test telemetry to Loki, Mimir, and Tempo.
- Documentation for how to run and interpret the checks.

Do not implement sensor deployment yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Kubernetes YAML can be rendered or applied with documented commands.
- The preflight check reports privileged workload status, BTF availability, and
  LGTM write connectivity.
```

## P2 - M0.5 delivery artifact security assessment baseline

```text
Use docs/SECURITY_ASSESSMENT.md and docs/ROADMAP.md S0.5/M0.5 as the target.

Implement the first security assessment baseline:

- Scanner configuration placeholders for Semgrep/gosec, Gitleaks, Trivy/Grype,
  Syft, Cosign/Notation, Crane, kube-linter, conftest, Hadolint, and ShellCheck.
- scripts/run-security-assessment.sh orchestration skeleton.
- scripts/verify-image-digest.sh for approved digest comparison.
- scripts/normalize-findings.sh placeholder for scanner result normalization.
- Report directory conventions for raw reports, normalized findings, and scan
  health.

Do not implement runtime event correlation or Trivy Operator
VulnerabilityReport ingestion.

Acceptance criteria:

- go test ./... passes if Go packages exist.
- go build ./... passes if Go packages exist.
- Running the assessment script without required inputs reports scan health
  failures rather than a false pass.
- Required artifact inputs are documented.
- No Secret raw values are written to reports.
```

## P3 - M1 Grafana LGTM backend

```text
Implement M1 from docs/ROADMAP.md: Grafana LGTM backend for the kube-sentinel
PoC.

Scope:

- Kubernetes manifests or Helm values for Loki, Mimir, Tempo, and Grafana.
- Datasource provisioning for Loki, Mimir, and Tempo.
- Basic dashboard assets for event, inventory, vulnerability, and security
  finding signals.
- Test telemetry scripts for Loki logs, Mimir metrics, and Tempo traces.
- Documentation for install, readiness checks, and teardown.

Do not implement the kube-sentinel operator, OTel pipeline, or sensors in this
milestone.

Acceptance criteria:

- go test ./... passes if Go packages exist.
- go build ./... passes if Go packages exist.
- Manifests or values are deterministic and documented.
- Grafana datasources are provisioned.
- Documentation includes kubectl checks and sample LogQL/PromQL queries.
```

## P4 - M2 management controller core, OTel feature, and assessment scaffold

```text
Implement M2 from docs/ROADMAP.md.

Scope:

- ClusterTarget, SecurityAssessment, and ScanRun reconciler core.
- Finalizer handling.
- Feature registry integration.
- Desired state store.
- Remote apply client skeleton using ClusterTarget kubeconfigRef.
- Override hook structure.
- Server-side apply skeleton with managed labels and annotations from
  docs/ARCHITECTURE.md.
- Status patching with observedGeneration and feature conditions.
- otel_pipeline feature that contributes buildable Kubernetes objects.
- security_assessment feature scaffold that can create assessment Job/CronJob
  resources without implementing all scanner logic.

Do not implement OSquery or Trivy feature logic yet.

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
- Sample SecurityAssessment enabling otel_pipeline and security_assessment.

Do not implement OSquery, Trivy delivery image scan, or applied cluster
configuration scan yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Generated assessment resources contain kube-sentinel ownership labels.
- Disabling the security_assessment feature removes or marks stale resources
  for GC.
- Scanner failures are represented as scan health findings.
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
- Normalized findings for applied configuration risks.

Do not implement OSquery or Trivy.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Applied cluster inspection uses read-only permissions.
- Secret raw values are not read or persisted.
- Sample SecurityAssessment can enable otel_pipeline and security_assessment
  with applied cluster scan settings.
- Documentation includes validation commands and expected Loki/Grafana fields.
```

## P7 - M5 OSquery feature

```text
Implement M5 from docs/ROADMAP.md: the OSquery feature.

Scope:

- osquery feature config defaults and validation.
- OSquery DaemonSet and config for CTEM Scope inventory.
- Minimal query pack for system, kernel, port, and container inventory.
- OTel receiver fragment for OSquery result logs.
- Readiness assessment and status updates.

Do not implement Trivy.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- OSquery inventory documents route to Loki inventory streams and Mimir
  inventory counters.
- Sample SecurityAssessment can enable otel_pipeline and osquery for selected
  ClusterTargets.
- Documentation includes query and LGTM verification commands.
```

## P8 - M6 Trivy delivery image scan and integrity

```text
Implement M6 from docs/ROADMAP.md: Trivy delivery image scan plus image
integrity.

Scope:

- trivy feature config defaults and validation for delivery image scanning.
- Registry digest or image tar scan flow.
- SBOM generation using Syft or Trivy SBOM output.
- Digest verification using Crane and approved digest lists.
- Optional signature verification hook for Cosign or Notation.
- Deterministic finding ID:
  <imageRepository>/<imageDigest>/<vulnerabilityID>/<packageName>
- Tests for duplicate-safe finding generation.

Do not implement Trivy Operator VulnerabilityReport ingestion in this milestone.
That is a Next Version extension.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Duplicate Trivy fixture ingestion produces the same finding ID.
- Vulnerability findings route to Loki vulnerability/security_finding streams
  and Mimir counters.
- Documentation includes install and verification commands.
```

## P9 - M7 Final Check Dashboard

```text
Implement M7 from docs/ROADMAP.md and docs/FRONTEND_ARCHITECTURE.md.

Scope:

- Final Check Dashboard assets for Overview, Source & Secrets, Images &
  Integrity, Kubernetes Config & RBAC, Dockerfile & Scripts, Scan Health, and
  Exceptions & Remediation.
- Findings table or documented panel query conventions.
- Dashboard variables for environment, target version/build, scan run ID,
  namespace, image, severity, category, scanner, scan status, and exception
  status.
- docs/ctem-mapping-results.md template.

Acceptance criteria:

- go test ./... passes if Go packages exist.
- go build ./... passes if Go packages exist.
- Dashboard assets or setup instructions are deterministic.
- Screenshots or documented queries cover each menu.
- CTEM results template maps Scope, Discovery, Priority, and Validation.
```

## P10 - M8 toggle, override, and final-check validation

```text
Implement M8 from docs/ROADMAP.md.

Scope:

- End-to-end validation assets for feature enable/disable.
- Override validation for nodeAgent and feature-specific resource overrides.
- Garbage collection verification for disabled features.
- Delivery artifact assessment validation.
- Applied cluster configuration assessment validation.
- Documentation of expected kubectl diff/get output and final-check report
  output.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Feature toggle tests or scripts cover security_assessment and one additional
  feature.
- Override tests or scripts verify resources and tolerations are reflected in
  generated workload specs.
- Stale resource cleanup behavior is documented.
- Final-check report output includes scan health and exception status.
```

## Prompt quality checklist

Before running `orchestrator run`, verify the prompt has:

- A single milestone target.
- Explicit files or modules in scope.
- Explicit out-of-scope items.
- At least three acceptance criteria.
- Required verification commands.
- References to docs rather than restating the whole plan.
