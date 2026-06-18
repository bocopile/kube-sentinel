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
- OSquery inventory collection and Trivy-based delivery image vulnerability
  assessment.
- An OpenTelemetry pipeline that routes signals to Grafana LGTM.
- Delivery artifact security assessment for source, secrets, container images,
  SBOM/integrity, Kubernetes YAML, RBAC, Dockerfile, and deployment scripts.
- Applied Biz Cluster configuration assessment with read-only access.
- A Final Check Dashboard for review, scan health, findings, and exception
  tracking.

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
5. [docs/FRONTEND_ARCHITECTURE.md](docs/FRONTEND_ARCHITECTURE.md)
6. [docs/ROADMAP.md](docs/ROADMAP.md)
7. [docs/ORCHESTRATOR.md](docs/ORCHESTRATOR.md)
8. [docs/PROMPTS.md](docs/PROMPTS.md)

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
