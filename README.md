# kube-sentinel

kube-sentinel is a Mgmt Cluster based Kubernetes final-check assessment PoC.
The Mgmt Cluster is the cluster where this solution is installed; Biz Clusters
are business/application clusters inspected by the solution. The repository is
currently in a planning and pre-skeleton state: it has a Go module and
implementation contract documents, but the Kubebuilder controller skeleton has
not been generated yet.

## Current Scope

The PoC is designed around Mgmt Cluster CRDs that register Biz Clusters and run
final-check scans through remote apply:

- `ClusterTarget` for Biz Cluster kubeconfig references, capabilities, and
  status.
- `SecurityAssessment` for assessment templates and selected targets.
- `ScanRun` for one execution and its per-target results.
- Trivy-based delivery image vulnerability assessment. OSquery inventory is an
  optional later extension, not a required final-check control.
- Delivery artifact security assessment for source, secrets, container images,
  SBOM/integrity, Kubernetes YAML, RBAC, Dockerfile, and deployment scripts.
- Applied Biz Cluster configuration assessment with read-only access.
- Report Store and Evidence Bundle generation for scanner reports, normalized
  findings, scan health, final decision, and exception review candidates.
- A Final Check Dashboard for review, scan health, findings, and exception
  tracking.
- Assessment reliability support features such as target preflight, artifact
  input manifests, scanner baseline capture, stable finding IDs, Secret
  redaction, evidence bundle export, exception review artifacts, and scan health
  summaries.
- Middleware and scanner version baselines for Kubernetes, Kubebuilder,
  controller-runtime, Trivy, Semgrep, Gitleaks, SBOM/signing tools, and
  Kubernetes policy scanners.

Biz Clusters do not run a kube-sentinel operator and do not need kube-sentinel
CRDs installed. Runtime event correlation and runtime drift validation are
planned as later extensions.

## Repository State

This repository currently contains:

- `go.mod`
- planning and architecture documents under `docs/`

It does not yet contain:

- Kubebuilder `PROJECT`
- `cmd/`
- `api/`
- `internal/`
- `config/`

## Documentation

Start with the docs in this order:

1. [docs/PLAN.md](docs/PLAN.md)
2. [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md)
3. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
4. [docs/SECURITY_ASSESSMENT.md](docs/SECURITY_ASSESSMENT.md)
5. [docs/ASSESSMENT_SUPPORT_FEATURES.md](docs/ASSESSMENT_SUPPORT_FEATURES.md)
6. [docs/FRONTEND_ARCHITECTURE.md](docs/FRONTEND_ARCHITECTURE.md)
7. [docs/ROADMAP.md](docs/ROADMAP.md)
8. [docs/ORCHESTRATOR.md](docs/ORCHESTRATOR.md)
9. [docs/PROMPTS.md](docs/PROMPTS.md)

Documentation policy:

- Keep [docs/PLAN.md](docs/PLAN.md) as the high-level planning document.
- Put implementation contracts in focused documents so they can be reviewed and
  used as orchestrator prompts.
- Prefer English for implementation-facing contract documents. Korean is
  acceptable for user-facing review documents and scope decisions that are
  actively discussed in Korean.
- Every implementation milestone must have an exit criterion that can be tested
  by command, Kubernetes object inspection, report artifact, evidence bundle,
  dashboard screenshot, or status field.
- Record architecture deviations in the relevant focused document before code
  changes are made.

## Next Implementation Step

The first implementation PR should create the Kubebuilder skeleton and core API
contracts:

- initialize the module as `github.com/bocopile/kube-sentinel`
- create the `ClusterTarget`, `SecurityAssessment`, and `ScanRun` APIs
- add a buildable reconciler skeleton
- add feature registry interfaces
- add tests for registry ordering and unknown feature validation

Expected verification once code exists:

```bash
go test ./...
go build ./...
```
