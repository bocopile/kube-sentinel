# Using orchestrator

`~/IdeaProjects/orchestrator` should be used as a workflow runner for
kube-sentinel, not as a source template. Its built-in scaffold command currently
targets Node/TypeScript projects, while kube-sentinel is a Go Kubernetes
operator.

## Current bootstrap state

This repository is not initialized as a Go/Kubebuilder project yet. At the time
of this document, the repo contains planning docs but no `go.mod`, `cmd/`,
`api/`, `internal/`, `config/`, or `.orchestrator/config.yaml`.

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

From `/Users/bhshin/projects/kube-sentinel`:

```bash
kubebuilder init --domain kube-sentinel.io --repo github.com/bhshin/kube-sentinel
kubebuilder create api --group security --version v1alpha1 --kind SecurityAgent --namespaced=false
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
Implement M2 operator core from docs/ROADMAP.md: CRD type, registry, desired
state store, override hook, SSA apply skeleton, and tests. Do not implement
Falco yet.
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
kubectl get securityagent -A
kubectl get ds,deploy,cm -n kube-sentinel-system
kubectl logs -n kube-sentinel-system deploy/kube-sentinel-controller-manager
```

Elasticsearch stages should include concrete query checks against:

- `security-events`
- `security-inventory`
- `security-vuln`
