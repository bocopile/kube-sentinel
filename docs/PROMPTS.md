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

## 모노레포 3-모듈 구조

이 프로젝트는 모노레포 안에 3개의 독립 빌드 단위로 구성된다. 상세는
`docs/MODULES.md` 참조.

| 모듈 | 경로 | 언어 | 역할 |
|------|------|------|------|
| operator | `operator/` | Go | Mgmt Cluster operator, CRD, Feature plugin, ArtifactStore write |
| backend | `backend/` | Go | REST API 서버, PostgreSQL query, k8s CR 조회 |
| frontend | `frontend/` | TypeScript | Final Check Dashboard (React SPA) |

모듈별 검증 명령:

```bash
# operator
(cd operator && go build ./... && go test ./...)

# backend
(cd backend && go build ./... && go test ./...)

# frontend
(cd frontend && npm run build && npm test)

# 전체
(cd operator && go build ./... && go test ./...) && \
(cd backend  && go build ./... && go test ./...) && \
(cd frontend && npm run build)
```

## 마일스톤 매핑

| Prompt | Roadmap stage | Roadmap milestone | 목적 | 주 모듈 |
|--------|--------------|-------------------|------|---------|
| P0 | Foundation | — | 모노레포 + operator skeleton | operator |
| P1 | S0 | M0 | Assessment readiness check | operator |
| P2 | S0.5 | M0.5 | Delivery artifact security assessment baseline | operator |
| P3 | S1 | M1 | Report store, PostgreSQL schema, evidence bundle, backend skeleton | operator, backend |
| P4 | S2 | M2 | Mgmt operator core, Feature orchestrator, assessment scaffold | operator |
| P5 | S2 | M3 | Security Assessment feature | operator |
| P6 | S3 | M4 | Applied cluster configuration scan | operator |
| P7 | S3 | M5 | Trivy delivery image scan, image integrity | operator |
| P8 | S5 | M6 | Phase 2 optional inventory/telemetry extension | operator |
| P9 | S4 | M7 | Final Check Dashboard frontend + backend REST API | backend, frontend |
| P10 | S4 | M8 | Final-check validation, report, exception, GC | operator, backend |
| P11 | S4 | M9 | (선택) AI remediation advisor | operator |

## 공통 지시 블록

복잡한 milestone prompt에는 다음 블록을 추가한다.

```text
docs/PLAN.md를 source plan으로 사용하고, docs/REQUIREMENTS.md,
docs/ARCHITECTURE.md, docs/SECURITY_ASSESSMENT.md,
docs/ASSESSMENT_SUPPORT_FEATURES.md, docs/FRONTEND_ARCHITECTURE.md,
docs/ROADMAP.md, docs/ORCHESTRATOR.md, docs/MODULES.md,
docs/DATABASE.md, docs/API_DESIGN.md를 구현 계약으로 사용한다.
변경은 요청된 milestone 범위로 제한한다. Phase 2 inventory, telemetry,
runtime sensor, automatic remediation은 해당 milestone에 명시되지 않은 한
구현하지 않는다. 변경 후 build 가능한 상태를 유지한다. 새 logic에는
집중된 test를 추가한다.

이 프로젝트는 모노레포 3-모듈 구조다 (operator/, backend/, frontend/).
각 모듈은 독립 Go module 또는 npm package이며 서로 Go import로 의존하지 않는다.
작업 모듈의 검증 명령:
  operator: cd operator && go test ./... && go build ./...
  backend:  cd backend  && go test ./... && go build ./...
  frontend: cd frontend && npm run build && npm test
```

---

## P0 - 모노레포 초기화 및 operator skeleton

```text
docs/PLAN.md, docs/REQUIREMENTS.md, docs/ARCHITECTURE.md,
docs/ROADMAP.md, docs/MODULES.md를 project contract로 사용한다.

이 프로젝트는 operator/, backend/, frontend/ 3개 모듈로 구성된 모노레포다.
P0에서는 operator/ 모듈 skeleton만 생성한다.

operator/ 초기화 (operator/ 디렉터리 안에서):
- Go module: github.com/bocopile/kube-sentinel/operator
- Kubebuilder 초기화 도메인: kube-sentinel.io
- CRD: ClusterTarget, SecurityAssessment, ScanRun (모두 cluster-scoped)
- api/v1alpha1 아래 ClusterTarget, SecurityAssessment, ScanRun API type 추가.
- ClusterTargetStatus에 ObservedGeneration, Capabilities, Namespaces 포함
  (docs/PLAN.md ClusterTargetStatus 정의 참조).
- 빈 controller reconciler skeleton.
- Feature interface (ID, Priority, Validate, Preflight, Build, Collect, Normalize),
  feature registry, priority-ordered orchestrator skeleton.
- ArtifactStore interface (filesystem / S3-compatible plugin).
- registry ordering과 unknown feature validation unit test.

optional inventory, OTel, LGTM, runtime sensor, backend 모듈, frontend 모듈은
아직 구현하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- ClusterTarget이 kubeconfigRef, bootstrapPolicy, namespaceAllowlist,
  capabilities, status field를 포함.
- SecurityAssessment가 target 선택과 scan profile (SourceSecurity,
  ImageSupplyChain, KubernetesConfig, RBACAndSecretReference, BuildAndDeploy) 포함.
- ScanRun이 scan execution status와 target별 result 포함.
- Feature registry ordering이 priority → feature ID 기준으로 deterministic.
- ArtifactStore interface가 filesystem, S3-compatible 구현체와 분리됨.
```

---

## P1 - M0 Assessment readiness checks

```text
docs/ROADMAP.md S0/M0과 docs/ASSESSMENT_SUPPORT_FEATURES.md를 구현 대상으로 사용한다.
작업 디렉터리: operator/

kube-sentinel assessment 준비 상태 검증 자산을 구현한다.

- kube-sentinel-system namespace manifest.
- kubeconfig 존재 여부, API 접근 가능 여부, namespace 존재 여부, read-only RBAC,
  image pull 접근, report store write 접근을 확인하는 target preflight check.
- target credential에 의도치 않은 Secret read 권한이 포함되었는지 탐지하고
  preflight risk로 보고하는 guard.
- check 실행 방법과 결과 해석 방법 문서.

runtime sensor, OTel/LGTM, privileged DaemonSet, automatic remediation은
구현하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- Kubernetes YAML을 문서화된 명령으로 render 또는 apply 가능.
- Preflight가 target 환경 실패와 scanner finding을 구분.
- Secret raw value를 읽지 않음.
```

---

## P2 - M0.5 Delivery artifact security assessment baseline

```text
docs/SECURITY_ASSESSMENT.md, docs/ASSESSMENT_SUPPORT_FEATURES.md,
docs/ROADMAP.md S0.5/M0.5를 구현 대상으로 사용한다.
작업 디렉터리: operator/

1차 security assessment baseline을 구현한다.

- Semgrep/gosec, Gitleaks, Trivy/Grype, Syft, Cosign/Notation, Crane,
  kube-linter, conftest, Hadolint, ShellCheck scanner config placeholder.
  format 규칙: docs/DATABASE.md raw_reports 테이블 scanner format 컬럼 참조
  (json/sarif/text 구분).
- source path, image list, digest list, Helm/YAML, RBAC, Dockerfile, script를
  선언하는 artifact-input.example.yaml.
- scanner version과 vulnerability DB/rule baseline capture.
- scripts/run-security-assessment.sh orchestration skeleton.
- 승인 digest 비교를 위한 scripts/verify-image-digest.sh.
- scanner 결과 정규화용 scripts/normalize-findings.sh placeholder.
- missing artifact, unsupported target, scanner error, stale baseline,
  registry pull failure를 나타내는 scan health output.
  scan_health reason 열거: scanner_error | unsupported_target |
  missing_artifact | stale_db | stale_rules | registry_pull_failure |
  rbac_denied | optional_input_unavailable (docs/DATABASE.md 참조).

runtime event correlation, OSQuery, OTel/LGTM, automatic remediation은 구현하지
않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- 필수 input 없이 assessment script를 실행하면 false pass가 아니라 scan health
  failure를 보고.
- 필수 artifact input이 문서화됨.
- Scanner baseline data가 report와 함께 기록됨.
- Secret raw value가 report에 기록되지 않음.
```

---

## P3 - M1 Report store, PostgreSQL schema, evidence bundle, backend skeleton

```text
docs/ROADMAP.md의 M1을 구현한다.
docs/DATABASE.md와 docs/MODULES.md를 구현 계약으로 사용한다.

작업 디렉터리: operator/ (report/normalizer), backend/ (DB init, migration)

operator/ 범위:

- Security Finding Schema (security.finding/v1)와 schema validator.
- stable finding ID와 deduplication helper.
  finding_id 생성 규칙: docs/DATABASE.md findings 테이블 참조.
- report, log, dashboard record, artifact 대상 Secret redaction guard.
- evidence bundle export 구조.
- ArtifactStore write: raw scanner output을 JSONB 또는 TEXT로 저장하는
  raw_report writer (docs/DATABASE.md raw_reports 테이블 스키마 기준).

backend/ 범위:

- Go module 초기화: github.com/bocopile/kube-sentinel/backend
- PostgreSQL 연결 풀 (pgx/v5).
- 마이그레이션 파일 생성 (backend/internal/db/migrations/):
  docs/DATABASE.md 전체 테이블 DDL 기준으로 migration 파일 작성.
  테이블: scan_runs, raw_reports, findings, scan_health,
          exception_reviews, artifact_index, cluster_targets.
  GIN 인덱스: raw_reports.data, findings.details 포함.
- DB 레코드 write helper (scan_run, finding, raw_report, scan_health insert).
- backend REST API skeleton: net/http + chi 라우터.
  엔드포인트 목록: docs/API_DESIGN.md 참조.
  이 milestone에서는 라우터 등록과 핸들러 stub만 구현.

OTel/LGTM telemetry, Grafana dashboard, frontend는 이 milestone에서 구현하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- cd backend  && go test ./... 통과.
- cd backend  && go build ./... 통과.
- 중복 fixture finding이 같은 stable finding ID를 생성.
- 잘못된 normalized finding은 schema validation 실패.
- Evidence bundle이 raw report, normalized finding, scan health, final decision,
  exception candidate를 참조.
- Secret 형태의 fixture 값은 저장 전에 redaction 또는 reject 처리.
- backend DB migration이 docs/DATABASE.md 전체 테이블 DDL을 충족.
```

---

## P4 - M2 Management controller core and assessment scaffold

```text
docs/ROADMAP.md의 M2를 구현한다.
작업 디렉터리: operator/

- Mgmt Cluster 단일 operator 기준 ClusterTarget, SecurityAssessment, ScanRun
  reconciler core.
- finalizer handling.
- feature registry와 Feature orchestrator integration.
- desired state store.
- ClusterTarget kubeconfigRef를 사용하는 remote apply client skeleton.
- bootstrapPolicy handling: 허용 namespace/RBAC/scanner resource 생성.
- docs/ARCHITECTURE.md의 managed label, annotation 포함 server-side apply skeleton.
- observedGeneration과 workflow condition 포함 status patching.
- security_assessment feature scaffold (scanner logic 없이 Job/CronJob 생성).
- ScanRun 결과 report writer skeleton:
  scan_run record와 findings를 PostgreSQL에 insert (backend DB 연동).

optional inventory, OTel/LGTM, runtime sensor, automatic remediation, Trivy feature
logic은 아직 구현하지 않는다.
Biz Cluster에는 kube-sentinel operator 또는 CRD를 설치하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- unit test가 finalizer behavior, unknown feature status, registry ordering,
  Feature orchestrator ordering, desired state label, remote apply label
  generation, bootstrap policy guardrail, status phase calculation을 검증.
- minimal assessment deployment용 sample ClusterTarget, SecurityAssessment,
  ScanRun YAML 존재 (operator/config/samples/).
```

---

## P5 - M3 Security Assessment feature

```text
docs/ROADMAP.md의 M3, Security Assessment feature를 구현한다.
작업 디렉터리: operator/

- security_assessment feature config default와 validation.
- delivery artifact scan을 위한 Assessment Job/CronJob resource.
- scanner config mount point와 report output convention.
  raw scanner output을 PostgreSQL raw_reports 테이블에 저장
  (docs/DATABASE.md format 컬럼 규칙 준수: json/sarif/text).
- finding normalization invocation.
- scanner failure와 missing artifact에 대한 scan health reporting.
  reason enum: docs/DATABASE.md scan_health 테이블 참조.
- artifact input manifest validation.
- scanner baseline capture → artifact_index 테이블 기록.

optional inventory, Trivy delivery image scan, applied cluster configuration
scan은 아직 구현하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- 생성된 assessment resource가 kube-sentinel ownership label을 포함.
- security_assessment feature 비활성화 시 stale run-scoped resource가 GC 대상.
- scanner failure가 scan health finding으로 표현됨.
- Evidence bundle output이 raw report와 normalized finding reference를 포함.
```

---

## P6 - M4 Applied cluster configuration scan

```text
docs/ROADMAP.md의 M4, applied cluster configuration scan을 구현한다.
작업 디렉터리: operator/

- 승인 namespace에 대한 read-only Kubernetes client access.
- securityContext, volume, image, ServiceAccount setting에 대한 workload spec
  inspection.
- Role, RoleBinding, ClusterRole, ClusterRoleBinding risk에 대한 RBAC inspection.
  finding category: rbac, secret_ref (docs/DATABASE.md findings.category 참조).
- Secret raw value를 읽지 않는 Secret reference inspection.
- 선택 warning category로 Service/Ingress exposure inspection.
  finding category: network.
- namespace allowlist validator.
- applied configuration risk에 대한 normalized finding → findings 테이블 insert.

optional inventory, runtime sensor, automatic remediation은 구현하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- Applied cluster inspection이 read-only permission을 사용.
- Secret raw value를 읽거나 저장하지 않음.
- Sample SecurityAssessment가 applied cluster scan setting으로 활성화 가능.
- 문서가 validation command와 예상 report field를 포함.
```

---

## P7 - M5 Trivy delivery image scan and integrity

```text
docs/ROADMAP.md의 M5, Trivy delivery image scan과 image integrity를 구현한다.
작업 디렉터리: operator/

- delivery image scanning을 위한 trivy feature config default와 validation.
- registry digest 또는 image tar scan flow.
  raw output → raw_reports 테이블 (scanner='trivy', format='json').
- Syft 또는 Trivy SBOM output을 사용하는 SBOM generation.
  SBOM 파일 → Artifact Store (artifact_type='sbom').
  artifact_index 테이블에 path, checksum, scanner_version, db_baseline_date 기록.
- Crane과 승인 digest list를 사용하는 digest verification.
- Cosign 또는 Notation을 위한 선택 signature verification hook.
  raw output → raw_reports (scanner='cosign' or 'notation', format='json').
- CRD 존재 시 선택 Trivy Operator VulnerabilityReport read-only ingestion.
- deterministic finding ID:
  imageRepository/imageDigest/vulnerabilityID/packageName
  (docs/DATABASE.md findings finding_id 생성 규칙 참조).
- direct Trivy scan과 optional VulnerabilityReport 간 duplicate-safe test.

이 milestone에서 Trivy Operator를 설치하거나 운영하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- 중복 Trivy fixture ingestion이 같은 finding ID를 생성.
- Optional VulnerabilityReport fixture ingestion이 동일 finding schema로 정규화.
- Vulnerability finding이 findings 테이블과 evidence bundle에 기록됨.
- 문서가 설치와 독립적인 verification command를 포함.
```

---

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
않는다. 작업 시작 전에 별도 설계 문서에서 수용 기준을 정의해야 한다.
```

---

## P9 - M7 Final Check Dashboard (frontend + backend REST API)

```text
docs/ROADMAP.md M7, docs/FRONTEND_ARCHITECTURE.md, docs/API_DESIGN.md,
docs/MODULES.md를 구현 대상으로 사용한다.
작업 디렉터리: backend/ (REST API 완성), frontend/ (React SPA)

backend/ 범위:

- docs/API_DESIGN.md 전체 엔드포인트 구현:
  GET  /api/v1/overview
  GET  /api/v1/cluster-targets, /cluster-targets/{name}
  GET  /api/v1/scan-runs, POST /api/v1/scan-runs
  GET  /api/v1/scan-runs/{id}, /scan-runs/{id}/status
  GET  /api/v1/scan-runs/{id}/findings
  GET  /api/v1/scan-runs/{id}/findings/{findingId}
  GET  /api/v1/scan-runs/{id}/findings/{findingId}/raw-report
  GET  /api/v1/scan-runs/{id}/health
  GET  /api/v1/scan-runs/{id}/artifacts
  GET  /api/v1/scan-runs/{id}/artifacts/{artifactType}/download
  GET  /api/v1/exceptions, PATCH /api/v1/exceptions/{id}
  GET  /api/v1/governance/summary
- PostgreSQL query 구현 (pgx/v5): docs/DATABASE.md 인덱스 활용.
- k8s dynamic client (k8s.io/client-go): ClusterTarget list/get, ScanRun create.
  backend는 operator Go module을 import하지 않음; 경량 struct 자체 정의.
- ArtifactReader interface 구현 (filesystem / S3-compatible):
  GetArtifact, ListArtifacts, GenerateDownloadURL.
- raw-report 응답 전 Secret redaction guard 재실행.
- CORS middleware (frontend origin 허용).
- exception status machine 강제 (PATCH /api/v1/exceptions/{id} 전환 규칙).

frontend/ 범위 (docs/FRONTEND_ARCHITECTURE.md 기준):

- Next.js App Router (TypeScript + Tailwind).
- 메뉴: Overview, Targets, Assessments, Findings, Reports, Governance.
- src/api/: docs/API_DESIGN.md 엔드포인트 호출 fetch wrapper.
- src/types/: docs/API_DESIGN.md 응답 스키마 기반 TypeScript 타입 정의.
- Finding 필터: severity, category (sast, secret, image_vulnerability, sbom,
  integrity, kubernetes, rbac, secret_ref, network, dockerfile, script, scan_health),
  exception_status, scan_status, scanner, namespace.
- 5초 polling: GET /api/v1/scan-runs/{id}/status.
- Exception review drawer: status 전환 UI (Required → Requested → Approved/Rejected).
- Evidence bundle 다운로드: /scan-runs/{id}/artifacts/evidence_bundle/download.

frontend는 k8s API와 직접 통신하지 않는다. backend API 경유만 허용.

수용 기준:

- cd backend  && go test ./... 통과.
- cd backend  && go build ./... 통과.
- cd frontend && npm run build 통과.
- GET /api/v1/scan-runs/{id}/status 가 phase polling에서 정확한 phase를 반환.
- PATCH /api/v1/exceptions/{id} 허용되지 않은 상태 전환 시 409 반환.
- GET /api/v1/scan-runs/{id}/findings/x/raw-report 응답에 Secret 원문 미포함.
- frontend Overview 화면이 최신 scan summary를 표시.
- frontend Findings 화면이 severity·category·exception_status 필터 동작.
```

---

## P10 - M8 Final-check validation

```text
docs/ROADMAP.md의 M8을 구현한다.
작업 디렉터리: operator/, backend/

- Code / Artifact Scan, Biz Cluster Scan, Full Final Check를 위한 end-to-end
  validation asset.
- disabled profile과 stale ScanRun에 대한 garbage collection verification.
- delivery artifact assessment validation.
- applied cluster configuration assessment validation.
- Secret redaction validation.
- evidence bundle과 exception review validation.
  exception status machine: Required → Requested → Approved/Rejected → Expired.
  docs/DATABASE.md exception_reviews 동기화 규칙 준수.
- no-auto-remediation guardrail validation.
- 예상 kubectl diff/get output과 final-check report output 문서화.
- scan_runs.summary JSONB 집계 카운터 업데이트 검증.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- cd backend  && go test ./... 통과.
- cd backend  && go build ./... 통과.
- validation script가 security_assessment와 trivy를 포함.
- stale resource cleanup behavior가 문서화됨.
- final-check report output이 scan health, evidence bundle reference, exception
  status, 자동 Biz Cluster infrastructure mutation 없음 정보를 포함.
- GET /api/v1/governance/summary 응답이 decision_trend와 exception_summary를 포함.
```

---

## P11 - M9 AI remediation advisor (선택)

```text
docs/AI_REMEDIATION.md와 docs/ROADMAP.md의 M9를 구현한다. 1차 선택 기능이며 기본 OFF다.
작업 디렉터리: operator/

- SecurityAssessment.spec.aiRemediation opt-in config와 validation.
- final decision 확정 이후 동작하는 remediation_enrichment feature (priority ~250).
- egress 전 field allowlist + Secret redaction guard 재사용.
  secret/sast/script 제외, Critical/High + kubernetes/rbac/dockerfile/
  image_vulnerability + per-scan cap 50.
- 공개 Gemini API provider 구현과 RemediationAdvisorProvider interface.
- security.aiRemediation/v1 출력 schema 검증과 거부 시 static fallback.
- artifact_index에 artifact_type='remediation_advisory' 기록.
  remediation-advisory sidecar와 provenance. core findings 테이블 불변.
- scan_health reason: ai_advisor_unavailable | ai_output_rejected.
- API/timeout/validation 실패 시 scan non-Fail, scan_health degraded.

automatic remediation, severity/판정 변경, secret/sast/script 입력, Vertex AI,
core remediation 덮어쓰기는 구현하지 않는다.

수용 기준:

- cd operator && go test ./... 통과.
- cd operator && go build ./... 통과.
- Secret fixture 입력 시 Gemini request body에 원문 미포함.
- AI ON/OFF 동일 scan에서 finding count, severity, final decision 동일.
- Gemini 실패 fixture에서 scan Completed + scan_health degraded.
- evidence bundle에 sidecar와 provenance 포함.
```

---

## 프롬프트 품질 체크리스트

`orchestrator run` 실행 전에 prompt가 다음을 포함하는지 확인한다.

- 단일 milestone target.
- 작업 디렉터리 명시 (operator/ / backend/ / frontend/).
- 범위에 포함되는 file 또는 module.
- 범위에서 제외되는 항목.
- 최소 3개 이상의 수용 기준.
- 필수 verification command (모듈별 `cd <module> && go test/build` 또는 `npm run build`).
- 전체 plan을 다시 쓰지 않고 docs를 참조하는 방식.
- docs/DATABASE.md 테이블/컬럼 기준 참조 (DB 관련 milestone).
- docs/API_DESIGN.md 엔드포인트 기준 참조 (API/frontend milestone).
