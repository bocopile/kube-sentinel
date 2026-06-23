# Frontend Architecture

이 문서는 최종점검 결과를 조회하고 카테고리별 검사를 실행하기 위한 `Final Check Dashboard` 프론트 화면 구조를 정의한다.

## Goal

프론트 화면의 목적은 scanner 결과를 단순 나열하는 것이 아니라 고객사 적용 전 납품 가능 여부를 판단하도록 돕는 것이다.

사용자는 다음 작업을 수행할 수 있어야 한다.

- 검사 프로파일을 선택해 관련 검사를 한 번에 실행한다.
- Code / Artifact Scan과 Biz Cluster Scan의 실행 상태와 실패 원인을 분리해서 확인한다.
- 전체 finding을 severity, category, target, scanner, exception 상태로 필터링한다.
- 각 카테고리별 상세 결과와 증적을 확인한다.
- 최종 판정 실패 원인을 확인한다.
- 검사 결과 보고서와 evidence bundle을 조회/내보낸다.
- 개선 권고, 예외 승인 후보, 만료 예외, 재점검 상태를 추적한다.

## Product Shape

현재 버전은 Report Store 기반 `Final Check Dashboard`를 기본으로 한다.
Grafana/LGTM은 2차 telemetry export가 필요할 때 선택적으로 연동한다.

| 단계 | UI 형태 | 목적 |
|------|--------|------|
| MVP | Report Store backed dashboard | 빠른 PoC 검증, finding 집계, 증적 조회, report export |
| Product UI | React/Next.js 또는 동등한 SPA | 검사 실행, finding 상세, 예외 workflow, 최종 판정 관리 |

## Information Architecture

대시보드는 하나의 제품 화면으로 두고, 내부 메뉴를 검사 영역 기준으로 나눈다.

```text
Final Check Dashboard
├── Overview
├── Targets
├── Assessments
├── Findings
├── Reports
└── Governance
```

## Menus

| 메뉴 | 목적 | 주요 기능 |
|------|------|----------|
| Overview | 납품 가능 여부 요약 | Pass/Fail, Critical/High, failed scans, missing artifacts, exception-required count |
| Targets | Biz Cluster 등록/상태 확인 | ClusterTarget list, cluster add/import, connection phase, namespace allowlist, last validation time |
| Assessments | 검사 실행과 workflow 상태 확인 | Code / Artifact Scan, Biz Cluster Scan, Full Final Check, preflight status, retry/resume |
| Findings | 보안 도메인별 finding 분석 (기본 OPEN 프리셋) | 5개 보안 도메인 탭 — 1) 소스 저장소 `sast,secret,dockerfile,script,kubernetes/rbac(target_cluster IS NULL)`, 2) 컨테이너 이미지 `image_vulnerability,sbom`, 3) 무결성·공급망 `integrity`, 4) K8s 실행 환경 `kubernetes/rbac(target_cluster IS NOT NULL),secret_ref,network`, 5) 스캔 상태·산출물 `scan_health`. 공통 severity/scanner/status 필터 공유 |
| Reports | 보고서 export와 증적 | final-check report(Markdown source, optional PDF), evidence bundle, raw report/normalized finding download(SARIF/JSON), scan health summary |
| 예외 관리 (Governance) | 예외 워크플로와 개선 권고 추적 | finding별 [예외 요청]/[오탐]/[조치 완료], 상태머신(Required→Requested→Approved/Rejected→Expired), owner/reason/expiry/approver, expired 예외 재평가, remediation 추적. 개선 권고·예외 검토 PDF export |

## Scan Profiles

`Run Scan` 메뉴는 도구명이 아니라 검사 그룹과 검사 영역 기준으로 실행 버튼을 제공한다.

| 검사 그룹 | 프로파일 | 실행 도구/방식 | 결과 메뉴 |
|----------|----------|---------------|----------|
| Code / Artifact Scan | Source Security Scan | SonarQube, Semgrep, gosec, Gitleaks | Assessments, Findings > 소스 저장소 |
| Code / Artifact Scan | Image Supply Chain Scan | Trivy/Grype, Syft, Cosign/Notation, Crane | Assessments, Findings > 컨테이너 이미지 / 무결성·공급망 |
| Code / Artifact Scan | Manifest & RBAC Manifest Scan | Helm render, kube-linter, conftest, RBAC manifest policy | Assessments, Findings > 소스 저장소 (`kubernetes,rbac` / `target_cluster IS NULL`) |
| Code / Artifact Scan | Build & Deploy Scan | Hadolint, ShellCheck | Assessments, Findings > 소스 저장소 (`dockerfile,script`) |
| Biz Cluster Scan | Applied Workload Config Scan | read-only applied workload inspection | Assessments, Findings > K8s 실행 환경 |
| Biz Cluster Scan | Applied RBAC Scan | read-only applied RBAC inspection | Assessments, Findings > K8s 실행 환경 |
| Biz Cluster Scan | Secret Reference Scan | env/envFrom/volume/ServiceAccount token reference inspection | Assessments, Findings > K8s 실행 환경 (`secret_ref`) |
| Biz Cluster Scan | Exposure Scan | Service/Ingress exposure inspection | Assessments, Findings > K8s 실행 환경 (`network`) |
| Full Final Check | Full Final Check | Code / Artifact Scan 이후 Biz Cluster preflight와 Biz Cluster Scan 실행 | Overview, Findings, Reports, Governance |

Biz Cluster Scan 실행 버튼은 선택한 `ClusterTarget`이 `Ready`가 아니면 비활성화한다.
비활성 사유는 kubeconfig Secret 누락, API 연결 실패, RBAC denied, namespace allowlist 미설정, optional
CRD/bootstrap capability 부족 중 하나로 표시한다.

## Common Filters

모든 메뉴는 같은 필터 모델을 공유한다.

| 필터 | 설명 |
|------|------|
| Environment | `dev`, `final-check`, `prod` (PoC에서 `prod` 미사용) |
| Target version/build | 납품 대상 버전 또는 build ID |
| Scan run ID | 검사 실행 단위 |
| Scan group | Code / Artifact, Biz Cluster, Full Final Check |
| Namespace | Biz Cluster 적용 설정 검수 범위 |
| Image | 이미지 repository, tag, digest |
| Severity | Critical, High, Medium, Low, Info |
| Category | `sast`, `secret`, `image_vulnerability`, `sbom`, `integrity`, `kubernetes`, `rbac`, `secret_ref`, `network`, `dockerfile`, `script`, `scan_health` |
| Scanner | SonarQube, Semgrep, gosec, Gitleaks, Trivy, Grype, Syft, Cosign, kube-linter, conftest, Hadolint, ShellCheck |
| Scan status | Pass, Fail, Error, Skipped, Unsupported |
| Exception status | None, Required, Requested, Approved, Expired, Rejected |
| View preset | **OPEN(기본)** = 조치 필요(`scan_status IN (Fail,Error)` AND `exception_status IN (None,Required,Requested,Rejected,Expired)`). Approved 포함·전체 토글 가능. OPEN은 새 컬럼이 아니라 쿼리 프리셋이며 finding을 숨기지 않는다 |
| Report domain / Target source | 5개 보안 도메인(category 프리셋) · Code/Artifact(`target_cluster IS NULL`) / Biz applied(`target_cluster IS NOT NULL`) / All |

## Cluster List

`Targets` 메뉴는 Mgmt Cluster의 `ClusterTarget` CR과 status를 조회해 Biz Cluster 목록을 표시한다.
kubeconfig Secret data는 어떤 UI/API에서도 노출하지 않는다.

| 컬럼 | 출처 | 설명 |
|------|------|------|
| Name | `metadata.name` | 내부 Biz Cluster ID |
| Display name | `spec.displayName` | 사용자 표시 이름 |
| Environment | `spec.environment` | dev, final-check, prod (PoC에서 prod target 미사용일 수 있음) |
| Phase | `status.phase` | Ready, Degraded, AuthFailed, Unreachable, PermissionDenied |
| Kubernetes version | `status.kubernetesVersion` | Biz Cluster discovery 결과 |
| Last validated | `status.lastValidatedAt` | 마지막 연결/RBAC 검증 시각 |
| Capabilities | `status.capabilities` | read-only inspection, scanner Job, report upload, image pull, optional Trivy Operator report 가능 여부 |
| Namespace allowlist | `spec.namespaceAllowlist` | 검사 허용 namespace |

Cluster detail 화면은 connectivity error, RBAC denied, egress blocked, image pull failure 같은 상태를
remediation과 함께 보여준다.
Credential rotation은 상태와 마지막 회전 시각만 표시하고 Secret 값은 표시하지 않는다.

## Finding Detail

모든 finding은 상세 drawer 또는 상세 페이지에서 같은 구조로 보여준다.

| 영역 | 필드 |
|------|------|
| Identity | finding ID, scan run ID, category, scanner, rule ID/CVE ID |
| Severity | severity, CVSS, exploitability, fixable 여부 |
| Target | file path, image digest, Kubernetes resource, namespace, ServiceAccount |
| Evidence | scanner message, matched location, applied YAML snippet metadata, command output reference |
| Remediation | 개선 권고, fixed version, policy recommendation |
| Exception | exception required, approval status, owner, expiry date, approval reason |
| Traceability | original report path, normalized finding path, dashboard link, rescan history |

Secret 값은 상세 화면에 표시하지 않는다.
Secret finding은 파일 위치, 키 이름, 탐지 rule, confidence, remediation만 표시한다.

## Data Flow

```text
Scanner / Cluster Inspector
  -> Raw reports
  -> Finding Normalizer
  -> Normalized findings
  -> Final decision summary
  -> PostgreSQL raw_reports / findings (query 정본) + Artifact Store evidence exports
  -> Evidence Bundle
  -> Final Check Dashboard
```

조회 흐름:

```text
Dashboard
  -> assessment-api
  -> PostgreSQL metadata query
  -> artifact reference lookup
  -> SBOM / evidence bundle / exported report download
```

Dashboard 목록, 필터, 집계와 raw scanner report·normalized finding 조회는 PostgreSQL
`raw_reports`/`findings`에서 수행한다.
SBOM, exported report, evidence bundle은 artifact store의 stable path를 참조해 다운로드한다.
PostgreSQL이 query 정본이며, artifact store의 `manifest.json`은 `artifact_index` 재생성에만 사용하고 raw/finding 정본을
대체하지 않는다.

## Workflow View

`Run Scan`과 `Overview`는 검사 결과를 도구별이 아니라 workflow별로 먼저 보여준다.

| Workflow | UI 상태 | 대표 실패 원인 | 재실행 단위 |
|----------|---------|----------------|-------------|
| Code / Artifact Workflow | `artifactScan.phase` | missing artifact, artifactInput checksum mismatch, scanner error, stale DB/rule, registry pull failure, digest mismatch | Code / Artifact Scan만 재실행 (Mgmt-local Job) → `PATCH /scan-runs/{id}/retry` scope `ArtifactOnly` |
| Biz Cluster Workflow | `clusterScan.phase` | kubeconfig missing, API unreachable, RBAC denied, namespace allowlist violation, optional CRD/bootstrap unavailable | Biz Cluster Scan만 재실행 (read-only + optional Biz-remote Job) → scope `ClusterOnly` |
| Final Decision Workflow | `finalDecision.status` | Critical finding, Secret exposure, unapproved exception, expired exception | 전체 또는 실패 workflow 재실행 후 재판정 → scope `Full` / `FinalDecisionOnly` |

`Assessments` 메뉴의 "실패 workflow 재실행" 버튼은 해당 workflow scope로 `PATCH /api/v1/scan-runs/{id}/retry`를
호출하고, 응답 `202` 후 `GET /api/v1/scan-runs/{id}/status` 폴링으로 갱신된 phase를 표시한다.
새 ScanRun id를 만들지 않고 동일 id의 phase가 갱신된다(retry/resume state).

워크플로우 상세 화면은 같은 `ScanRun` 안에서 단계별 timestamps, conditions, raw report 링크, normalized finding 링크를
보여준다.
사용자는 Code / Artifact 실패와 Biz Cluster 실패를 같은 실패로 보지 않고, 어느 절차를 다시 실행해야 하는지 즉시 판단할 수 있어야 한다.

제품형 UI로 확장할 때는 다음 API 계층을 둔다.

| API | 역할 |
|-----|------|
| `assessment-api` | scan run, finding, exception, 최종 판정 결과 조회 |
| `scanner-runner` | 검사 profile 실행 요청과 상태 추적 |
| `artifact-store` | SBOM, scanner baseline, evidence bundle, exported report 저장 (raw report·normalized finding은 PostgreSQL) |

## State Model

| 객체 | 설명 |
|------|------|
| ScanRun | 한 번의 검사 실행 단위. profile, target version, status, timestamps를 가진다. |
| ScanPhase | `artifactScan`, `clusterScan` 등 검사 절차별 phase, timestamps, conditions를 가진다. |
| Finding | normalized finding. category, scanner, target, severity, status를 가진다. |
| FinalDecision | scan run별 최종 판정 객체. `status`(Pass/Fail/Warning), `reasons[]`(code, message, severity, category, count, findingID), `decidedAt`로 구성하며 PLAN.md `FinalDecision` struct·`ScanRun.status.finalDecision`과 동일 스키마다. REST 목록/polling 응답의 `final_decision` 문자열은 이 객체의 status를 평면화한 값이다. |
| ExceptionReview | finding별 예외 승인 상태, 승인자, 사유, 만료일. MVP 정본은 per-finding `exception_reviews`(PostgreSQL); 예외 정책(패턴/scope 매칭)·`fingerprint`·`REVOKED`/`FALSE_POSITIVE`는 Phase2 plugin. |
| Artifact | SBOM, digest verification report, scanner baseline, evidence bundle, exported report. raw report와 normalized finding은 PostgreSQL에서 조회한다. |
| EvidenceBundle | raw report, normalized findings, scan health, final decision, exception candidates를 묶은 검수 증적. |

## UI Guardrails

- 최종 판정 실패 원인을 Overview에서 바로 보여준다.
- 스캔 실패와 필수 산출물 누락은 취약점 없음으로 처리하지 않고 Scan Health에서 Fail로 표시한다.
- Secret 원문 값은 UI, log, artifact 어디에도 표시하지 않는다.
- 5개 보안 도메인(소스 저장소/컨테이너 이미지/무결성·공급망/K8s 실행환경/스캔 상태·산출물)은 top-level
  메뉴가 아니라 Findings 내부 탭으로 제공하고, `findings.category`+`target_cluster`(NULL=매니페스트, 값=applied) 프리셋 필터로 구분한다.
- 예외 승인은 finding을 숨기지 않는다. 기본 Findings 뷰는 OPEN 프리셋(조치 필요)만 보이지만 이는 쿼리 필터일 뿐이며,
  `Approved`/`Expired` finding도 DB·예외 관리 메뉴·상세·PDF·evidence에 그대로 유지된다(스캔 단계 제외·DB 삭제 없음). 상태만 `Approved`로 바꾸고 만료일을 표시한다.
- 예외 관리는 finding별 [예외 요청]→`Requested` PATCH, [오탐]은 별도 enum이 아니라 reason 분류, [조치 완료]는 finding 삭제가 아니라 재스캔(`PATCH /scan-runs/{id}/retry`)으로 처리한다.
  예외 정책 패턴 매칭·`fingerprint` 컬럼·`REVOKED`/`FALSE_POSITIVE` enum은 Phase2 plugin이다.
- 개선 권고와 remediation은 보고서/추적 정보로만 제공한다.
  현재 버전 UI는 Biz Cluster 인프라 자동 수정 액션을 제공하지 않는다.
- AI remediation advisor(선택)가 생성한 조치 가이드는 "AI generated / advisory / non-binding" 라벨과
  provenance(model, 생성 시각)와 함께 표시하고, 정적 remediation과 구분한다.
  AI 가이드에도 자동 수정 액션은 제공하지 않는다.
  상세는 [AI_REMEDIATION.md](./AI_REMEDIATION.md).
