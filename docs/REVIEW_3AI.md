# kube-sentinel 문서 3-AI 교차검증 · 협의 결과

> 작성일: 2026-06-22 · 참여 AI: Claude(C\*) · Codex/OpenAI(X\*) · Cursor(U\*)
> 절차: 개별 분석 → 교차검증(각 발견사항을 다른 2개 AI가 원문 대조) → 협의(병합·조정·우선순위화)

## 0. 검증 방법 및 집계

- **3개 AI**: Claude(C*), Codex/OpenAI(X*), Cursor(U*) — 동일한 14개 문서 전수 검토
- **개별 분석**: 42개 발견사항 (Claude 14 / Codex 13 / Cursor 15)
- **교차검증**: 각 발견사항을 자신을 제외한 다른 2개 AI가 원문 대조로 적대적 검증 (총 84 verdict)
- **결과 분포**:
  - 만장일치 confirmed: 다수
  - refined(핵심 인정, 범위/표현 수정): 13건
  - **기각(REJECTED)**: 1건 (C7 — 두 검증자 모두 refuted)
  - **강등(DOWNGRADED)**: 1건 (X2 — 실제 결함 아님, 용어 명확화로 축소)
- **중복 병합** 후 고유 이슈 약 20개(+ 후속 사용자 검토로 추가된 I-23 1건). 핵심은 "구현 착수를 막는 **저장 정본 충돌**과 **모듈 구조 충돌**, 그리고 **CRD 계약 누락 다발**".

---

## 1. CRITICAL — 구현 착수 차단 (최우선)

### [I-1] raw report·normalized findings 저장 정본 충돌 (PostgreSQL ↔ Artifact Store)
- **출처/검증**: X3 + U1(critical) + U12 — 만장일치 confirmed (ARCHITECTURE 내부 자기모순)
- **근거**: ARCHITECTURE가 같은 문서 안에서 raw/normalized를 Artifact Store(mermaid line 87, line 410/447/651, "Normalized finding JSONL is the canonical" line 757)와 PostgreSQL(`raw_reports`/`findings`, line 783~785 "raw/ 경로는 Artifact Store에 없다")에 **동시 단언**. DATABASE는 PostgreSQL 정본. MODULES 통신표·PROMPTS P3는 "ArtifactStore write: raw scanner output". PROMPTS P3는 "ArtifactStore write … JSONB 또는 TEXT"로 한 문장 안에서도 모순.
- **영향**: operator report writer, backend raw-report API, evidence bundle 구성, AI sidecar 입력이 모두 이 결정에 의존. P3(report store) 구현 자체가 막힘.
- **합의 권고**: PostgreSQL `raw_reports`/`findings` = **query 정본**, Artifact Store = SBOM/evidence bundle/baseline/artifact-input 전용. `normalized/findings.jsonl`은 evidence bundle용 **export(파생)**으로 명문화. ARCHITECTURE mermaid·MODULES 통신표·PROMPTS P3·AI_REMEDIATION의 "canonical" 표현을 일괄 수정.

### [I-2] 3-모듈 구조 vs 루트 단일 go.mod 충돌
- **출처/검증**: X1 + U4 + C3 — 만장일치 confirmed
- **근거**: README "다음 구현 단계"(line 82 "module을 `github.com/bocopile/kube-sentinel`로 초기화") + 실제 `go.mod` = 루트 단일 module. 그러나 MODULES/ORCHESTRATOR/PROMPTS = `operator/`,`backend/`,`frontend/` 3개 독립 모듈 + `github.com/bocopile/kube-sentinel/operator` + `cd operator && go test ./...`.
- **영향**: 첫 구현 PR(P0)의 module path, Kubebuilder 실행 위치, 검증 명령이 정면 모순. README "리포지터리 상태"의 `cmd/`,`api/`,`internal/`,`config/` 목록도 모노레포에선 `operator/cmd` 등이어야 함.
- **합의 권고**: 3-모듈을 정본으로 채택 → README 단락을 operator/ 기준으로 수정하고 루트 `go.mod`를 `operator/`로 이동(또는 `go.work` 도입). 충돌 진원지는 README "다음 구현 단계" 단락 하나.

---

## 2. HIGH — CRD/API 구현 계약 누락·미정의

### [I-3] CRD spec 필드 누락 다발 (Go 타입 ↔ YAML/API/prompt 불일치)
- **병합**: C2 + X5 + C8 + U7 + U5 — 전부 만장일치 confirmed
- **누락 목록**:
  - `SecurityAssessmentSpec.aiRemediation` — AI_REMEDIATION/PROMPTS P11이 전제하나 Go 타입에 없음 (C2)
  - `ScanRunSpec.Profiles` — API `POST /scan-runs`의 `profiles` override를 담을 필드 없음 (X5)
  - `POST /api/v1/scan-runs`의 `targets` — multi-target 부분 실행 불가 (C8)
  - `ClusterTargetSpec.BootstrapPolicy` — YAML 샘플·prose는 쓰는데 Go 타입에 없음 (U7)
  - `SecurityAssessmentSpec.artifactInput` — 납품 산출물(소스/이미지/digest/manifest) 입력 경로를 CRD/API에 연결할 필드 없음 (U5)
- **합의 권고**: PLAN(및 ARCHITECTURE) 핵심 Go 타입에 위 필드를 일괄 추가하고 YAML 샘플·API·prompt와 동기화. P0 skeleton 생성 전 필수.

### [I-4] ScanRun/SecurityAssessment status 모델 불일치
- **병합**: C6 + U9 — 만장일치 confirmed
- **근거**: PLAN `ScanRunStatus`에 `Canceled` phase·`remoteResources[]` 없음, `SecurityAssessmentStatus` 타입 자체 부재. ARCHITECTURE(status.remoteResources[], lastRunRef)·DATABASE·API는 포함. `finalDecision`이 string인지 object인지도 문서마다 다름(API string ↔ FRONTEND `finalDecision.status`).
- **합의 권고**: `FinalDecision` struct(`status`,`reasons[]`,`decidedAt`)와 `ScanRunStatus`/`SecurityAssessmentStatus`를 ARCHITECTURE 기준으로 PLAN에 확정, API 응답·FRONTEND state model 동일 스키마로 정렬.

### [I-5] profiles[] → feature 매핑 및 features[]/profiles[] 병합 규칙 미정의
- **출처/검증**: U6 — 만장일치 confirmed(medium 신뢰)
- **근거**: ARCHITECTURE는 `features[].name` umbrella → registry feature ID 확장 예시(trivy→image_vulnerability…)만 제시. `SourceSecurity`/`KubernetesConfig` 등 profile이 어떤 feature를 enable하는지 표가 없고, profiles[]와 features[]가 둘 다 spec에 존재할 때 병합/우선순위 규칙이 없음.
- **합의 권고**: profile→feature ID 매핑 표 + 병합 규칙 + unknown profile 처리(`ConfigError`)를 ARCHITECTURE/PLAN에 단일 정본으로 추가.

### [I-6] Code/Artifact Scan 실행 토폴로지 및 artifact input 전달 경로 미정의
- **출처/검증**: U2(critical, refined) + U5 — confirmed(범위 조정)
- **근거**: G1은 "controller가 Biz Cluster에 assessment workload 생성"이라 cluster-side 실행을 시사하나, SECURITY_ASSESSMENT는 Code/Artifact Scan을 "Mgmt Cluster Job/별도 runner/CI runner/검수 VM"으로 둠. 소스/이미지 등 입력을 어느 runner로 어떻게 전달하는지(PVC/Git ref/init container) 미정의. (협의: "Biz remote Job과 정면 충돌"은 과장 → "runner placement + artifact input 전달 계약 미정의"로 축소.)
- **합의 권고**: `SecurityAssessment`/`ScanRun` spec에 runner placement(`mgmt-local`|`biz-remote`) 필드 추가, artifact input 마운트/동기화 규약을 reconcile flow에 명시.

### [I-7] PLAN Security Finding Schema category enum이 secret_ref·network 누락
- **병합**: C4 + X7 — 만장일치 confirmed
- **근거**: PLAN category 목록(line 479)에 `secret_ref`,`network` 없음. DATABASE/SECURITY_ASSESSMENT/FRONTEND/ARCHITECTURE는 포함. PLAN을 정본 삼은 schema validator가 secret_reference·exposure finding을 조용히 reject.
- **합의 권고**: PLAN enum에 `secret_ref`,`network` 추가하여 DATABASE와 일치.

### [I-8] finding_id 생성 규칙 불완전·충돌 (dedup 신뢰성)
- **병합**: C1 + X8 + U8(refined)
  - `image_vulnerability` 규칙에 `<scanner>` prefix 없음 → Trivy/Grype 동일 이미지·CVE·패키지가 `UNIQUE(finding_id, scan_run_id)` 충돌. 게다가 API_DESIGN 예시는 `trivy/…`로 **규칙-예시 자기모순** (C1)
  - `kubernetes`/`rbac` 규칙에 cluster/target 식별자 없음 → 다중 Biz Cluster 동일 리소스 충돌 (X8)
  - `sbom`/`integrity`/`secret_ref`/`network` finding_id 규칙 부재 (U8, 범위: 최소 secret_ref/network/integrity)
- **합의 권고**: DATABASE finding_id 표를 `<scanner>/…`로 통일 + 다중 target 시 cluster 식별자 포함 + 누락 category 규칙 보강(또는 기존 규칙 재사용 명시). API 예시·PLAN 예시를 규칙과 일치.

### [I-9] workflow 독립 재실행(retry/resume) 트리거 미정의
- **출처/검증**: U3 — confirmed(refined)
- **근거**: SECURITY_ASSESSMENT "각 workflow는 독립 재실행 가능", FRONTEND "Code/Artifact Scan만 재실행" 요구. 그러나 `POST /scan-runs`는 신규 생성만, ScanRun spec에 rerun scope/resume 필드·reconciler 진입점 없음. (ScanRun.status가 artifactScan/clusterScan phase는 분리 관리 → 재실행 단위 모델 일부 존재.)
- **합의 권고**: `ScanRun.spec.rerunScope`(Full|ArtifactOnly|ClusterOnly|FinalDecisionOnly) 또는 `PATCH /scan-runs/{id}/retry` 정의 + reconciler 분기 추가.

### [I-10] M0.5 exit criteria가 M5 기능(Trivy/SBOM/Integrity)과 겹쳐 보임
- **출처/검증**: X6 — confirmed(refined)
- **근거**: ROADMAP S0.5 "Image, SBOM, Integrity … 생성 및 정규화" vs PROMPTS P5/M3 "Trivy delivery image scan 아직 구현 안 함", M5에서 구현. (협의: P2/M0.5 실제 범위는 scanner config placeholder + baseline capture → "M0.5가 M5 기능을 요구"는 과장. 실제 결함은 ROADMAP S0.5 문구가 placeholder보다 넓게 읽히는 점.)
- **합의 권고**: ROADMAP S0.5 exit 문구를 placeholder/baseline 수준으로 한정. G10/G11/G16이 실제 충족되는 milestone 명시.

---

## 3. MEDIUM — 정합성·완결성

- **[I-11]** README 문서 목록·PLAN docs 트리에 DATABASE/API_DESIGN(및 MODULES) 누락 — C5+X13+U11 만장일치. (협의: PLAN 트리엔 MODULES는 이미 있음 → DATABASE/API_DESIGN만 추가.)
- **[I-12]** reconcile 순서 역전: evidence bundle/final-decision을 **생성(step 18~19) 전에 저장(step 15)** — X4 만장일치. ARCHITECTURE+PLAN+AI_REMEDIATION 흐름을 `Normalize → Final Decision → (remediation enrich) → report export/evidence 생성 → metadata 확정` 순으로 재정렬.
- **[I-13]** `scan_runs` 초기 row·`cluster_targets` 미러 write 주체 모호 — C10+U15 만장일치. operator가 reconcile 시 upsert(정본), backend POST는 CR apply만, status polling의 row-부재 404 처리 명시.
- **[I-14]** artifact download API가 `{artifactType}`만 받아 다중 SBOM/integrity 구분 불가(X10) + Artifact Store path 규칙 불일치(U14: `scanrun-abc123/reports/…` vs `reports/<assessment>/<run>/evidence/…`, `.spdx` vs `.cyclonedx`) — 만장일치. download를 artifact id/path 기반으로, path convention 단일화.
- **[I-15]** AI advisor 확장 모델 정책 불명확 — C9(typed `aiRemediation` vs `FeatureSpec.Config` RawExtension "CRD 변경 없음" 원칙 충돌, refined) + X9("scan_health degraded"는 enum 값이 아니라 `Warning`+`reason=ai_advisor_unavailable` 서술 → 용어 통일, refined). 확장 모델 정책과 용어를 문서로 명문화.
- **[I-16]** Environment enum 불일치 — C12 만장일치. SECURITY_ASSESSMENT/FRONTEND는 `dev`,`final-check`만, PLAN/DB/API는 `prod` 포함. DATABASE 기준 `dev|final-check|prod`로 고정(필요 시 "PoC 미사용" 명시).
- **[I-17]** Kyverno/Gatekeeper/rbac-police가 baseline/feature/milestone에 없음 — U10(refined; SonarQube는 이미 optional 계약 있어 제외). 3개 도구를 필수/optional/비목표로 분류.
- **[I-18]** remote scanner Job → Report Store upload 경로 미정의 — X12 만장일치. (I-1 저장 정본 해소와 함께 다룰 것: upload token/Secret/egress vs controller pull.)
- **[I-19]** G1 검증 기준이 ScanRun 생성 주체/순서를 모호하게 둠 — U13(refined). G1 절차에 ScanRun 생성 주체(사용자 vs controller 자동) 명시.
- **[I-23]** MODULES.md "모듈 간 경계" 토폴로지 다이어그램이 `frontend (React SPA)`를 **Mgmt Cluster 박스 밖**에 배치(Biz Cluster scanner Jobs와 같은 높이) — ARCHITECTURE.md 정본(line 14·25 "Dashboard/API … in Mgmt", mermaid line 53~65에서 `dashboard`/`assessment_api`/`metadata_store`/`artifact_store` 모두 `subgraph mgmt` 내부)과 모순. *(3-AI 자동 검토가 놓치고 사용자 후속 검토에서 발견 → 이 브랜치에서 다이어그램 수정 완료: operator·PostgreSQL·Artifact Store·backend·frontend 전부 Mgmt Cluster 박스 안에 배치, Biz Cluster는 scanner Job/RBAC/namespace만 별도 박스로 분리.)*

---

## 4. LOW / NIT

- **[I-20]** API cluster-targets 예시 `kubernetes_version: "1.31.0"`이 baseline `v1.34~v1.36` 밖 — C11 만장일치. 예시를 `1.35.0`으로.
- **[I-21]** exception 상태머신 재스캔 carry-over 규칙 부재 — C13(refined). "유효 Approved 유지, Expired/Rejected는 Required로 재평가" 규칙 추가.
- **[I-22]** 첫 구현 PR 범위(`.orchestrator/config.yaml` 포함 여부, 3모듈 동시 초기화 여부)가 README/ROADMAP/ORCHESTRATOR마다 상이 — C14 만장일치. ROADMAP "첫 구현 블록"을 단일 정본으로.

---

## 5. 협의에서 기각·강등된 지적

- **C7 (기각)**: "P8의 S5/M6 매핑이 stage gate와 불일치". → Codex·Cursor 둘 다 refuted. ROADMAP에서 S5=Phase2 telemetry/inventory, M6=Optional telemetry/inventory로 **의미상 일치**하며, "다른 행은 1:1 매핑"이라는 전제 자체가 틀림(PROMPTS는 S2→M2/M3, S4→M7/M8/M9도 매핑). 결함 아님.
- **X2 (강등→용어 명확화)**: "read-only credential과 remote-apply write 권한이 같은 모델에 혼재". → Cursor refuted, Claude refined. ARCHITECTURE guardrail(line 880~893)이 write를 `managed-by=kube-sentinel` 라벨 리소스로 한정하고 customer config는 read-only, Secret read 미부여로 **이미 분리**. 실제 결함이 아니라 capability 표의 "read-only" 단독 표현이 오해를 부르는 **용어 혼선** → "read-only inspection of customer config + write of kube-sentinel-owned scan resources"로 표현만 명확화(minor).

---

## 6. 문서별 수정 요약

| 문서 | 주요 수정 |
|---|---|
| **README.md** | 모듈 구조(3-모듈) · 첫 PR 범위 · 문서 목록(DATABASE/API_DESIGN 추가)(I-2,I-11,I-22) |
| **PLAN.md** | Spec 필드(aiRemediation/Profiles/BootstrapPolicy/artifactInput) · status 타입 · category enum(secret_ref/network) · reconcile 순서 · docs 트리(I-3,4,7,12,11) |
| **ARCHITECTURE.md** | 저장 정본 단일화 · reconcile 순서 · status model · profile→feature 매핑표 · path convention · 도구 분류(I-1,4,5,12,14,17) |
| **DATABASE.md** | finding_id 규칙(scanner prefix/cluster id/누락 category) · scan_health 용어 · exception carry-over(I-8,15,21) |
| **API_DESIGN.md** | POST body(targets/profiles/artifactInput) · retry endpoint · artifact download(id/path) · k8s_version 예시 · auth 최소 모델 · path convention(I-3,9,14,20) |
| **MODULES.md** | cluster_targets sync owner · raw report 저장 정본 정렬(I-1,13) |
| **SECURITY_ASSESSMENT.md** | 실행 토폴로지 · scanner 분류 · profile 매핑(I-6,17) |
| **PROMPTS.md** | P3 storage · module path 검증 명령 · M0.5/M5 정렬(I-1,2,10) |
| **AI_REMEDIATION.md** | "degraded" 용어 · "canonical" 표현 · pipeline 순서(I-1,12,15) |
| **ROADMAP.md / ORCHESTRATOR.md** | 첫 블록 범위 · S↔M 매핑 표 · S0.5 exit 문구(I-10,22) |

## 7. 우선 조치 순서 (권장)

1. **[I-1] 저장 정본 결정** — 가장 많은 문서가 의존하며 I-12/13/14/18의 선행조건. 여기부터.
2. **[I-2] 모듈 구조 확정** — P0 skeleton 착수 차단 해제.
3. **[I-3,4,5] CRD spec/status 정본 확정** — P0/P11 구현 계약.
4. **[I-7,8] category enum + finding_id 규칙 통일** — normalization/dedup 신뢰성.
5. **[I-6,9] 실행 토폴로지·artifact input·retry 트리거** — workflow 구현.
6. **나머지 MEDIUM/LOW 정합성 정리** (I-11~22).
