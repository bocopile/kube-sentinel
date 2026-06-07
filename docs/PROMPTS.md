# Orchestrator prompts

Use these prompts with `orchestrator plan` first, then `orchestrator run` only
after the plan looks acceptable. If the local orchestrator build does not yet
support `plan` and `run`, use the same prompt text directly in Claude Code for
P0 through P3, then return to orchestrator after its Phase 1 skeleton is ready.

All prompts assume the project root is:

```bash
/Users/bhshin/projects/kube-sentinel
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
| P0 | Foundation | First implementation block | Go operator skeleton and core API contracts. |
| P1 | S0 | M0 | Cluster prerequisite checks. |
| P2 | S1 | M0.5 | OTel/parser spike and fixture routing. |
| P3 | S2 | M1 | Elasticsearch, Kibana, and index templates. |
| P4 | S2 | M2 | Operator core and OTel pipeline feature. |
| P5 | S2 | M3 | Falco vertical slice. |
| P6 | S3 | M4 | Tetragon feature. |
| P7 | S3 | M5 | OSquery feature. |
| P8 | S3 | M6 | Trivy feature. |
| P9 | S4 | M7 | MITRE scenarios and Kibana dashboards. |
| P10 | S4 | M8 | Feature toggle, override, and garbage collection validation. |

## Global instruction block

Add this block to milestone prompts when the request is complex:

```text
Use docs/PLAN.md as the source plan, and use docs/REQUIREMENTS.md,
docs/ARCHITECTURE.md, docs/ROADMAP.md, and docs/ORCHESTRATOR.md as the
implementation contract. Keep changes scoped to the requested milestone. Do not
implement later sensors or dashboards unless they are explicitly part of the
milestone. Preserve buildability after the change. Add focused tests for new
logic. Verification must include go test ./... and go build ./....
```

## P0 - Create project skeleton

Use this after Kubebuilder has initialized the repository, or adapt it if the
operator skeleton is created manually.

```text
Use docs/PLAN.md, docs/REQUIREMENTS.md, docs/ARCHITECTURE.md, and
docs/ROADMAP.md as the project contract.

Implement the first kube-sentinel code block from docs/ROADMAP.md only:

- Ensure this is a Go Kubernetes operator project.
- Add or complete the cluster-scoped SecurityAgent API type under api/v1alpha1.
- Add an empty but buildable controller reconciler.
- Add feature registry interfaces and deterministic priority ordering.
- Add tests for registry ordering and unknown feature validation.
- Do not implement Falco, Tetragon, OSquery, Trivy, OTel manifests, or
  Elasticsearch integration yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- The SecurityAgent type contains spec fields for global, features, output,
  override, and tests.
- The SecurityAgent CRD is cluster-scoped and uses
  spec.global.targetNamespace for namespaced workloads.
- Unknown feature names can be detected and reported by pure Go unit tests.
- Registry ordering is deterministic by priority and feature ID.
```

## P1 - S0 cluster prerequisite checks

```text
Use docs/ROADMAP.md S0 as the target.

Implement cluster prerequisite assets for kube-sentinel:

- Namespace manifest for kube-sentinel-system.
- A privileged preflight DaemonSet or Job that verifies required host access.
- A script that checks /sys/kernel/btf/vmlinux from a node-level pod.
- A script or manifest for writing a test document to Elasticsearch.
- Documentation for how to run and interpret the checks.

Do not implement sensor deployment yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Kubernetes YAML can be rendered or applied with documented commands.
- The preflight check reports privileged workload status, BTF availability, and
  Elasticsearch write connectivity.
```

## P2 - S1 OTel and parser spike

```text
Use docs/ROADMAP.md S1 and docs/ARCHITECTURE.md data routing as the target.

Implement a local OTel/parser spike:

- Add sample log fixtures for Falco, Tetragon, and OSquery.
- Add a Trivy VulnerabilityReport fixture.
- Add OTel config fragments or generator code sufficient to route fixtures to
  security-events, security-inventory, and security-vuln according to
  docs/ARCHITECTURE.md.
- Add tests that validate routing decisions and stable Trivy document IDs.

Do not deploy real Falco, Tetragon, OSquery, or Trivy yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Fixtures cover one event from each source.
- Routing logic maps Falco and Tetragon to security-events, OSquery to
  security-inventory, and Trivy to security-vuln.
- Trivy upsert document ID is deterministic.
```

## P3 - M1 Elasticsearch and Kibana

```text
Implement M1 from docs/ROADMAP.md: Elasticsearch and Kibana infrastructure for
the kube-sentinel PoC.

Scope:

- Kubernetes manifests or scripts for ECK-based Elasticsearch and Kibana.
- Index template setup for security-events, security-inventory, and
  security-vuln.
- Secret and TLS assumptions documented without committing real credentials.
- A test-document write script for each index.
- Documentation for install, readiness checks, and teardown.

Do not implement the kube-sentinel operator, OTel pipeline, or sensors in this
milestone.

Acceptance criteria:

- go test ./... passes if Go code exists.
- go build ./... passes if Go code exists.
- Manifests or scripts are deterministic and documented.
- Index templates cover the three CTEM indices.
- Documentation includes kubectl checks for Elasticsearch/Kibana readiness and
  curl examples for test document insertion.
```

## P4 - M2 operator core and OTel feature

```text
Implement M2 from docs/ROADMAP.md.

Scope:

- SecurityAgent reconciler core.
- Finalizer handling.
- Feature registry integration.
- Desired state store.
- Override hook structure.
- Server-side apply skeleton with managed labels and annotations from
  docs/ARCHITECTURE.md.
- Status patching with observedGeneration and feature conditions.
- otel_pipeline feature that contributes buildable Kubernetes objects.

Do not implement Falco, Tetragon, OSquery, or Trivy features yet.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Unit tests cover finalizer behavior, unknown feature status, registry ordering,
  desired state labels, and status phase calculation.
- Sample SecurityAgent YAML exists for minimal OTel pipeline deployment.
```

## P5 - M3 Falco vertical slice

```text
Implement M3 from docs/ROADMAP.md: the Falco feature vertical slice.

Scope:

- falco feature config defaults and validation.
- Falco DaemonSet, ConfigMap, RBAC, and required hostPath mounts.
- File output to /var/log/kube-sentinel/falco/events.log.
- OTel receiver fragment for Falco filelog input.
- Readiness assessment and status reason handling.
- Sample SecurityAgent enabling otel_pipeline and falco.

Do not implement Tetragon, OSquery, or Trivy.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Generated Falco resources contain kube-sentinel ownership labels.
- Disabling the Falco feature removes or marks stale Falco resources for GC.
- Documentation includes the kubectl and Elasticsearch checks for shell
  execution detection.
```

## P6 - M4 Tetragon feature

```text
Implement M4 from docs/ROADMAP.md: the Tetragon feature.

Scope:

- tetragon feature config defaults and validation.
- Tetragon workload/resources needed for the PoC.
- TracingPolicy resources for process execution and container escape monitoring.
- OTel receiver fragment for Tetragon pod logs.
- Readiness assessment and status updates.

Do not implement OSquery or Trivy.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Tetragon resources contain kube-sentinel ownership labels.
- Sample SecurityAgent can enable otel_pipeline, falco, and tetragon.
- Documentation includes validation commands and expected Elasticsearch fields.
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
- OSquery inventory documents route to security-inventory.
- Sample SecurityAgent can enable otel_pipeline and osquery.
- Documentation includes query and Elasticsearch verification commands.
```

## P8 - M6 Trivy feature

```text
Implement M6 from docs/ROADMAP.md: the Trivy feature.

Scope:

- trivy feature config defaults and validation.
- Trivy Operator integration assumptions and manifests.
- VulnerabilityReport reader or CronJob design for Elasticsearch bulk upsert.
- Deterministic document ID:
  <clusterName>/<namespace>/<workloadKind>/<workloadName>/<containerName>/<vulnerabilityID>/<packageName>
- Tests for duplicate-safe upsert payload generation.
- Readiness assessment and status updates.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Duplicate Trivy fixture ingestion produces the same document ID.
- Vulnerability documents route to security-vuln.
- Documentation includes install and verification commands.
```

## P9 - M7 MITRE scenarios and Kibana dashboards

```text
Implement M7 from docs/ROADMAP.md.

Scope:

- test/pods.yaml for testbox, attacker, and target-nginx.
- test/run-ctem-scenarios.sh with MITRE scenarios from docs/PLAN.md.
- Elasticsearch query checks for each scenario.
- Kibana dashboard export or documented dashboard creation assets for events,
  inventory, and vulnerabilities.
- docs/ctem-mapping-results.md template.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Scenario script has clear pass/fail output.
- At least five scenarios are represented.
- CTEM results template maps Scope, Discovery, Priority, and Validation.
```

## P10 - M8 toggle and override validation

```text
Implement M8 from docs/ROADMAP.md.

Scope:

- End-to-end validation assets for feature enable/disable.
- Override validation for nodeAgent and feature-specific resource overrides.
- Garbage collection verification for disabled features.
- Documentation of expected kubectl diff/get output.

Acceptance criteria:

- go test ./... passes.
- go build ./... passes.
- Feature toggle tests or scripts cover at least Falco and one additional sensor.
- Override tests or scripts verify resources and tolerations are reflected in
  generated workload specs.
- Stale resource cleanup behavior is documented.
```

## Prompt quality checklist

Before running `orchestrator run`, verify the prompt has:

- A single milestone target.
- Explicit files or modules in scope.
- Explicit out-of-scope items.
- At least three acceptance criteria.
- Required verification commands.
- References to docs rather than restating the whole plan.
