# Orchestrator 프롬프트

아래 prompt는 먼저 `orchestrator plan`으로 검토한 뒤, plan이 적절할 때만
`orchestrator run`으로 실행한다.

모든 prompt는 project root가 다음 checkout이라고 가정한다.

```text
github.com/bocopile/kube-sentinel
```

## 명령 패턴

Dry run:

```bash
orchestrator plan --project . --request "<prompt>"
```

구현 실행:

```bash
orchestrator run --project . --request "<prompt>" --auto-approve
```

## 마일스톤 매핑

| Prompt | Roadmap stage | Roadmap milestone | 목적 |
| --- | --- | --- | --- |
| P0 | Foundation | First implementation block | Go management controller skeleton과 core API contract |
| P1 | S0 | M0 | Assessment readiness check |
| P2 | S0.5 | M0.5 | Delivery artifact security assessment baseline |
| P3 | S1 | M1 | Report store, finding schema, evidence bundle, dashboard backend |
| P4 | S2 | M2 | Mgmt operator core, Feature orchestrator, security assessment scaffold |
| P5 | S2 | M3 | Security Assessment feature |
| P6 | S3 | M4 | Applied cluster configuration scan |
| P7 | S3 | M5 | Trivy delivery image scan, image integrity, optional VulnerabilityReport ingestion |
| P8 | S5 | M6 | Phase 2 optional inventory/telemetry extension |
| P9 | S4 | M7 | Final Check Dashboard |
| P10 | S4 | M8 | Final-check validation, report, exception, garbage collection |
| P11 | S4 | M9 | (선택) AI remediation advisor — Gemini advisory sidecar, redaction, provenance |

## 공통 지시 블록

복잡한 milestone prompt에는 다음 블록을 추가한다.

```text
docs/PLAN.md를 source plan으로 사용하고, docs/REQUIREMENTS.md,
docs/ARCHITECTURE.md, docs/SECURITY_ASSESSMENT.md,
docs/ASSESSMENT_SUPPORT_FEATURES.md, docs/FRONTEND_ARCHITECTURE.md,
docs/ROADMAP.md, docs/ORCHESTRATOR.md를 구현 계약으로 사용한다.
변경은 요청된 milestone 범위로 제한한다. Phase 2 inventory, telemetry,
runtime sensor, automatic remediation은 해당 milestone에 명시되지 않은 한
구현하지 않는다. 변경 후 build 가능한 상태를 유지한다. 새 logic에는
집중된 test를 추가한다. 검증에는 go test ./...와 go build ./....를 포함한다.
```

## P0 - Project skeleton 생성

```text
docs/PLAN.md, docs/REQUIREMENTS.md, docs/ARCHITECTURE.md,
docs/ROADMAP.md를 project contract로 사용한다.

docs/ROADMAP.md의 첫 kube-sentinel code block만 구현한다.

- 이 project가 github.com/bocopile/kube-sentinel module을 사용하는 Go
  Kubernetes management controller project인지 확인한다.
- api/v1alpha1 아래 ClusterTarget, SecurityAssessment, ScanRun API type을
  추가하거나 완성한다.
- 비어 있지만 build 가능한 controller reconciler를 추가한다.
- feature registry interface, Feature orchestrator skeleton, deterministic
  priority ordering을 추가한다.
- Artifact Store backend plugin interface를 추가한다.
- registry ordering과 unknown feature validation test를 추가한다.
- optional inventory, OTel manifest, LGTM integration, runtime sensor,
  security assessment job, Trivy, dashboard는 아직 구현하지 않는다.

Acceptance criteria:

- go test ./... 통과
- go build ./... 통과
- ClusterTarget이 target kubeconfigRef, targetNamespace, namespaceAllowlist,
  output, capabilities, status field를 포함
- SecurityAssessment가 selected target과 scan profile 포함
- ScanRun이 scan execution status와 target별 result 포함
- pure Go unit test로 unknown feature name을 탐지하고 보고 가능
- Registry ordering이 priority와 feature ID 기준으로 deterministic
- Artifact Store backend plugin interface가 backend 구현체와 분리됨
```

## P1 - M0 assessment readiness checks

```text
docs/ROADMAP.md S0/M0과 docs/ASSESSMENT_SUPPORT_FEATURES.md를 구현 대상으로 사용한다.

kube-sentinel assessment 준비 상태 검증 자산을 구현한다.

- kube-sentinel-system namespace manifest.
- kubeconfig 존재 여부, API 접근 가능 여부, namespace 존재 여부, read-only RBAC,
  image pull 접근, report store write 접근을 확인하는 target preflight check.
- target credential에 의도치 않은 Secret read 권한이 포함되었는지 탐지하고
  preflight risk로 보고하는 guard.
- check 실행 방법과 결과 해석 방법 문서.

runtime sensor, OTel/LGTM telemetry, privileged DaemonSet, automatic remediation은
구현하지 않는다.

수용 기준:

- Go package가 존재하면 go test ./... 통과.
- Go package가 존재하면 go build ./... 통과.
- Kubernetes YAML을 문서화된 명령으로 render 또는 apply 가능.
- Preflight가 target 환경 실패와 scanner finding을 구분.
- Secret raw value를 읽지 않음.
```

## P2 - M0.5 delivery artifact security assessment baseline

```text
docs/SECURITY_ASSESSMENT.md, docs/ASSESSMENT_SUPPORT_FEATURES.md,
docs/ROADMAP.md S0.5/M0.5를 구현 대상으로 사용한다.

1차 security assessment baseline을 구현한다.

- Semgrep/gosec, Gitleaks, Trivy/Grype, Syft, Cosign/Notation, Crane,
  kube-linter, conftest, Hadolint, ShellCheck scanner config placeholder.
- source path, image list, digest list, Helm/YAML, RBAC, Dockerfile, script를
  선언하는 artifact-input.example.yaml.
- scanner version과 vulnerability DB/rule baseline capture.
- scripts/run-security-assessment.sh orchestration skeleton.
- 승인 digest 비교를 위한 scripts/verify-image-digest.sh.
- scanner 결과 정규화용 scripts/normalize-findings.sh placeholder.
- missing artifact, unsupported target, scanner error, stale baseline,
  registry pull failure를 나타내는 scan health output.

runtime event correlation, OSQuery, OTel/LGTM, automatic remediation은 구현하지
않는다. M0.5는 scanner config, input validation, baseline capture, scan-health
skeleton만 만든다. 실제 납품 이미지 취약점 scanning은 M5에서 구현한다.

수용 기준:

- Go package가 존재하면 go test ./... 통과.
- Go package가 존재하면 go build ./... 통과.
- 필수 input 없이 assessment script를 실행하면 false pass가 아니라 scan health
  failure를 보고.
- 필수 artifact input이 문서화됨.
- Scanner baseline data가 report와 함께 기록됨.
- Secret raw value가 report에 기록되지 않음.
```

## P3 - M1 report store, schema, evidence, and dashboard backend

```text
docs/ROADMAP.md의 M1을 구현한다.

범위:

- raw scanner report, normalized finding, scan health, final decision record,
  evidence bundle을 위한 Report Store interface.
- filesystem, S3-compatible, SeaweedFS 등으로 확장 가능한 Artifact Store
  backend plugin interface.
- Security Finding Schema와 schema validator.
- stable finding ID와 deduplication helper.
- report, log, dashboard record, artifact 대상 Secret redaction guard.
- evidence bundle export 구조.
- Overview, Targets, Assessments, Findings, Reports, Governance를 위한 기본
  dashboard/read-model record.

이 milestone에서는 OTel/LGTM telemetry 또는 Grafana 전용 dashboard를 구현하지
않는다.

수용 기준:

- Go package가 존재하면 go test ./... 통과.
- Go package가 존재하면 go build ./... 통과.
- 중복 fixture finding이 같은 stable finding ID를 생성.
- 잘못된 normalized finding은 schema validation 실패.
- Evidence bundle이 raw report, normalized finding, scan health, final decision,
  exception candidate를 참조.
- Secret 형태의 fixture 값은 저장 전에 redaction 또는 reject 처리.
- Artifact Store backend 선택이 finding metadata schema를 변경하지 않음.
```

## P4 - M2 management controller core and assessment scaffold

```text
docs/ROADMAP.md의 M2를 구현한다.

범위:

- Mgmt Cluster 단일 operator 기준 ClusterTarget, SecurityAssessment, ScanRun
  reconciler core.
- finalizer handling.
- feature registry와 Feature orchestrator integration.
- desired state store.
- ClusterTarget kubeconfigRef를 사용하는 remote apply client skeleton.
- target namespace/RBAC/scanner resource에 대한 bootstrap policy handling.
- docs/ARCHITECTURE.md의 managed label과 annotation을 포함한 server-side apply
  skeleton.
- observedGeneration과 workflow condition을 포함한 status patching.
- 모든 scanner logic을 구현하지 않고 assessment Job/CronJob resource를 생성할
  수 있는 security_assessment feature scaffold.
- ScanRun 결과를 위한 report writer skeleton.

optional inventory, OTel/LGTM, runtime sensor, automatic remediation, Trivy
feature logic은 아직 구현하지 않는다.
Biz Cluster에는 kube-sentinel operator 또는 CRD를 설치하지 않는다.

수용 기준:

- go test ./... 통과.
- go build ./... 통과.
- unit test가 finalizer behavior, unknown feature status, registry ordering,
  Feature orchestrator ordering, desired state label, remote apply label
  generation, bootstrap policy guardrail, status phase calculation을 검증.
- minimal assessment deployment용 sample ClusterTarget, SecurityAssessment,
  ScanRun YAML 존재.
```

## P5 - M3 Security Assessment feature

```text
docs/ROADMAP.md의 M3, Security Assessment feature를 구현한다.

범위:

- security_assessment feature config default와 validation.
- delivery artifact scan을 위한 Assessment Job/CronJob resource.
- scanner config mount point와 report output convention.
- finding normalization invocation.
- scanner failure와 missing artifact에 대한 scan health reporting.
- artifact input manifest validation.
- scanner baseline capture.

optional inventory, Trivy delivery image scan, applied cluster configuration
scan은 아직 구현하지 않는다.

수용 기준:

- go test ./... 통과.
- go build ./... 통과.
- 생성된 assessment resource가 kube-sentinel ownership label을 포함.
- security_assessment feature 비활성화 시 stale run-scoped resource가 GC 대상
  으로 제거되거나 표시됨.
- scanner failure가 scan health finding으로 표현됨.
- Evidence bundle output이 raw report와 normalized finding reference를 포함.
```

## P6 - M4 Applied cluster configuration scan

```text
docs/ROADMAP.md의 M4, applied cluster configuration scan을 구현한다.

범위:

- 승인 namespace에 대한 read-only Kubernetes client access.
- securityContext, volume, image, ServiceAccount setting에 대한 workload spec
  inspection.
- Role, RoleBinding, ClusterRole, ClusterRoleBinding risk에 대한 RBAC inspection.
- Secret raw value를 읽지 않는 Secret reference inspection.
- 선택 warning category로 Service/Ingress exposure inspection.
- namespace allowlist validator.
- applied configuration risk에 대한 normalized finding.

optional inventory, runtime sensor, automatic remediation은 구현하지 않는다.

수용 기준:

- go test ./... 통과.
- go build ./... 통과.
- Applied cluster inspection이 read-only permission을 사용.
- Secret raw value를 읽거나 저장하지 않음.
- Sample SecurityAssessment가 applied cluster scan setting으로
  security_assessment를 활성화 가능.
- 문서가 validation command와 예상 report field를 포함.
```

## P7 - M5 Trivy delivery image scan and integrity

```text
docs/ROADMAP.md의 M5, Trivy delivery image scan과 image integrity를 구현한다.

범위:

- delivery image scanning을 위한 trivy feature config default와 validation.
- registry digest 또는 image tar scan flow.
- Syft 또는 Trivy SBOM output을 사용하는 SBOM generation.
- Crane과 승인 digest list를 사용하는 digest verification.
- Cosign 또는 Notation을 위한 선택 signature verification hook.
- CRD가 존재하고 ClusterTarget에 get/list/watch permission이 있을 때 선택적으로
  read-only Trivy Operator VulnerabilityReport ingestion.
- deterministic finding ID:
  <imageRepository>/<imageDigest>/<vulnerabilityID>/<packageName>
- direct Trivy scan과 optional VulnerabilityReport input 간 duplicate-safe
  finding generation test.

이 milestone의 일부로 Trivy Operator를 설치하거나 운영하지 않는다.
VulnerabilityReport를 사용할 수 없어도 전체 assessment를 실패 처리하지 않고,
optional input unavailable을 scan health에 기록한다.

수용 기준:

- go test ./... 통과.
- go build ./... 통과.
- 중복 Trivy fixture ingestion이 같은 finding ID를 생성.
- Optional VulnerabilityReport fixture ingestion이 동일 finding schema로 정규화.
- Vulnerability finding이 Report Store record와 evidence bundle에 기록됨.
- 문서가 설치와 독립적인 verification command를 포함.
```

## P8 - M6 Phase 2 optional inventory/telemetry extension

```text
별도 설계 검토 후 Phase 2 inventory 또는 telemetry가 승인된 경우에만
docs/ROADMAP.md의 M6를 구현한다.

범위 후보:

- OSQuery 또는 동등한 inventory sensor.
- normalized finding과 report event에서 OTel/LGTM로 export하는 path.
- runtime event 또는 drift assessment.
- long-running sensor DaemonSet model.

제품 범위가 명시적으로 요구하지 않는 한 1차 final-check PoC에서는 구현하지
않는다.

작업 시작 전에 별도 설계 문서에서 수용 기준을 정의해야 한다.
```

## P9 - M7 Final Check Dashboard

```text
docs/ROADMAP.md와 docs/FRONTEND_ARCHITECTURE.md의 M7을 구현한다.

범위:

- Overview, Targets, Assessments, Findings, Reports, Governance를 위한 Final
  Check Dashboard asset.
- Findings table 또는 문서화된 panel query convention.
- final-check report, evidence bundle, raw report, normalized finding,
  scan health summary를 위한 Reports menu.
- environment, target version/build, scan run ID, namespace, image, severity,
  category, scanner, scan status, exception status dashboard variable.

수용 기준:

- Go package가 존재하면 go test ./... 통과.
- Go package가 존재하면 go build ./... 통과.
- Dashboard asset 또는 setup instruction이 deterministic.
- screenshot 또는 문서화된 query가 각 menu를 포함.
- Reports menu가 evidence bundle과 final decision data를 노출.
```

## P10 - M8 final-check validation

```text
docs/ROADMAP.md의 M8을 구현한다.

범위:

- Code / Artifact Scan, Biz Cluster Scan, Full Final Check를 위한 end-to-end
  validation asset.
- disabled profile과 stale ScanRun에 대한 garbage collection verification.
- delivery artifact assessment validation.
- applied cluster configuration assessment validation.
- Secret redaction validation.
- evidence bundle과 exception review validation.
- no-auto-remediation guardrail validation.
- 예상 kubectl diff/get output과 final-check report output 문서화.

수용 기준:

- go test ./... 통과.
- go build ./... 통과.
- validation script가 security_assessment와 trivy를 포함.
- stale resource cleanup behavior가 문서화됨.
- final-check report output이 scan health, evidence bundle reference, exception
  status, 자동 Biz Cluster infrastructure mutation 없음 정보를 포함.
```

## P11 - M9 AI remediation advisor (선택)

```text
docs/AI_REMEDIATION.md와 docs/ROADMAP.md의 M9를 구현한다. 1차 선택 기능이며 기본 OFF다.

범위:

- SecurityAssessment.spec.aiRemediation opt-in config와 validation.
- final decision 확정 이후 동작하는 remediation_enrichment feature (priority ~250).
- egress 전 field allowlist + Secret redaction guard 재사용. secret/sast/script 제외,
  Critical/High + kubernetes/rbac/dockerfile/image_vulnerability + per-scan cap 50.
- 공개 Gemini API provider 구현과 RemediationAdvisorProvider interface.
- security.aiRemediation/v1 출력 schema 검증과 거부 시 static fallback.
- remediation-advisory.jsonl sidecar와 provenance 기록. core findings.jsonl 불변.
- API/timeout/validation 실패 시 scan non-Fail, scan_health degraded.

automatic remediation, severity/판정 변경, secret/sast/script 입력, Vertex AI,
core remediation 덮어쓰기는 구현하지 않는다.

수용 기준:

- go test ./... 통과, go build ./... 통과.
- Secret fixture 입력 시 Gemini request body에 원문 미포함.
- AI ON/OFF 동일 scan에서 finding count, severity, final decision 동일.
- Gemini 실패 fixture에서 scan Completed + scan_health degraded.
- evidence bundle에 sidecar와 provenance 포함.
```

## 프롬프트 품질 체크리스트

`orchestrator run` 실행 전에 prompt가 다음을 포함하는지 확인한다.

- 단일 milestone target.
- 범위에 포함되는 file 또는 module.
- 범위에서 제외되는 항목.
- 최소 3개 이상의 수용 기준.
- 필수 verification command.
- 전체 plan을 다시 쓰지 않고 docs를 참조하는 방식.
