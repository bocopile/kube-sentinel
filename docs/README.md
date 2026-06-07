# kube-sentinel docs

This directory is the project contract for kube-sentinel. Implementation should
start only after these documents make the target behavior, architecture, and
verification gates explicit.

## Reading order

1. [PLAN.md](./PLAN.md) - original PoC plan and source of truth.
2. [REQUIREMENTS.md](./REQUIREMENTS.md) - user-visible goals and acceptance criteria.
3. [ARCHITECTURE.md](./ARCHITECTURE.md) - operator, feature registry, and data pipeline design.
4. [ROADMAP.md](./ROADMAP.md) - implementation stages and exit criteria.
5. [ORCHESTRATOR.md](./ORCHESTRATOR.md) - how to use `~/IdeaProjects/orchestrator` for this repo.
6. [PROMPTS.md](./PROMPTS.md) - milestone prompts for orchestrator `plan` and `run`.

## Documentation policy

- Keep `PLAN.md` as the high-level planning document.
- Put implementation contracts in focused documents so they can be reviewed and
  used as orchestrator prompts.
- Prefer English for new focused contract documents. `PLAN.md` may remain in its
  original mixed Korean/English form until a dedicated translation pass is made.
- Every implementation milestone must have an exit criterion that can be tested
  by command, Kubernetes object inspection, Elasticsearch query, or screenshot.
- Any deviation from the architecture should be recorded in the relevant focused
  document before code changes are made.
