# Using orchestrator

`~/IdeaProjects/orchestrator` should be used as a workflow runner for
kube-sentinel, not as a source template. Its built-in scaffold command currently
targets Node/TypeScript projects, while kube-sentinel is a Go Kubernetes
operator.

## Current bootstrap state

This repository is in a pre-skeleton Go state. It has `go.mod` and planning
docs, but it does not yet have Kubebuilder-generated `cmd/`, `api/`,
`internal/`, `config/`, `PROJECT`, or `.orchestrator/config.yaml` files.

That means there are two valid execution modes:

- If local orchestrator `plan` and `run` are available, use
  [PROMPTS.md](./PROMPTS.md) through orchestrator.
- If orchestrator is still in a foundation/skeleton phase or fails to run
  against this repo, use the P0-P3 prompts directly in Claude Code first. Return
  to orchestrator after it can initialize and execute Go projects reliably.

## Local prerequisites

The current machine already has:

- Node 24
- `orchestrator`
- Go
- `claude`
- `codex`
- `cursor-agent`
- `kubectl`
- `helm`

Still needed for the operator workflow:

- `kubebuilder`
- `controller-gen`
- `kustomize`
- `kind` or `minikube`

## Recommended sequence

From the repository root:

```bash
kubebuilder init --domain kube-sentinel.io --repo github.com/bocopile/kube-sentinel
kubebuilder create api --group security --version v1alpha1 --kind ClusterTarget --namespaced=false
kubebuilder create api --group security --version v1alpha1 --kind SecurityAssessment --namespaced=false
kubebuilder create api --group security --version v1alpha1 --kind ScanRun --namespaced=false
go test ./...
go build ./...
orchestrator init --project . --yes
```

Do not run the Kubebuilder commands twice against an already initialized repo
without checking the generated files first. If `go.mod`, `PROJECT`, `api/`, or
`config/` already exist, inspect them and continue from the matching prompt
instead of reinitializing.

Then run orchestrator per milestone:

```bash
orchestrator plan --project . --request "Implement the first kube-sentinel code block from docs/ROADMAP.md"
orchestrator run --project . --request "Implement the first kube-sentinel code block from docs/ROADMAP.md" --auto-approve
```

Use `plan` before `run` for every stage. The plan command is useful because it
stops before code changes and reports whether the request has enough acceptance
criteria and project context.

## Suggested request style

Use narrow milestone requests instead of broad requests.

Good:

```text
Implement M2 management controller core from docs/ROADMAP.md: CRD type,
assessment registry, desired state store, target kubeconfig loader, remote
apply skeleton, SSA apply skeleton, report writer skeleton, and tests. Do not
implement runtime sensors yet.
```

Avoid:

```text
Build kube-sentinel.
```

## Verification commands

The orchestrator config for this repo should use Go commands:

```bash
go test ./...
go build ./...
```

Cluster stages should add explicit manual checks in the milestone request, for
example:

```bash
kubectl --context mgmt get clustertarget,securityassessment,scanrun -A
kubectl --context mgmt logs -n kube-sentinel-system deploy/kube-sentinel-controller-manager

kubectl --context biz-a get namespace kube-sentinel-system
kubectl --context biz-a get ds,deploy,job,cronjob,cm -n kube-sentinel-system
```

Use explicit kubeconfig contexts in every milestone request. Mgmt Cluster
commands inspect kube-sentinel CRDs, controller logs, and status. Biz Cluster
commands inspect only remotely applied resources and read-only scan targets.

Report stages should include concrete checks against raw report artifacts,
normalized finding records, scan health summaries, final decision records,
evidence bundles, and dashboard screenshots.
