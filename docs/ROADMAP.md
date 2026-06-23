# 로드맵

PoC는 수직 slice 단계로 구현한다.
각 단계가 끝난 뒤 리포지터리는 build 및 test 가능한 상태여야 한다.

## Stage gate

| Stage | 범위 | Exit criteria |
| --- | --- | --- |
| S0 | Assessment prerequisite | Mgmt namespace, Biz Cluster kubeconfig Secret, target preflight, bootstrap policy, read-only RBAC, image access, report store write test 통과 |
| S0.5 | Delivery artifact security assessment baseline | SAST/Secret/Manifest/RBAC/Dockerfile/Script finding 생성·정규화와 artifact input manifest·scanner baseline·scan health report 검증. Image/SBOM/Integrity는 scanner config·fixture·baseline placeholder만 두고, 실제 Trivy delivery image scan·SBOM·integrity 생성은 S3/M5(P7)에서 구현한다 |
| S1 | Report store와 finding schema spike | Security Assessment/Trivy fixture가 PostgreSQL `raw_reports`/`findings`, stable finding ID, metadata record, evidence export artifact, evidence bundle로 변환 |
| S2 | Mgmt 단일 operator와 assessment vertical slice | `ClusterTarget`, `SecurityAssessment`, `ScanRun`이 Feature orchestrator/registry, remote apply/read-only inspection으로 assessment Job과 report record 생성 |
| S3 | 나머지 assessment capability | Trivy delivery image scan과 applied cluster config scan을 각각 enable/disable, assess, verify 가능 |
| S4 | Final-check validation과 dashboard | Delivery artifact scan, applied cluster configuration scan, dashboard, report, exception review 통과 |
| S5 | Phase 2 telemetry/inventory | OTel/LGTM/OSQuery/runtime telemetry는 제품 요구사항이 된 경우에만 설계 후 추가 |

## Milestone

| Milestone | 설명 | 예상 기간 | Exit criteria |
| --- | --- | ---: | --- |
| M0 | Assessment readiness check | 1일 | Mgmt namespace, Biz Cluster kubeconfig Secret, target preflight, bootstrap policy, read-only RBAC, image access, report store write test 통과 |
| M0.5 | Delivery artifact security assessment baseline | 1일 | required artifact, artifact input manifest, scanner version, vulnerability DB baseline, image access, digest list, scan health report 검증 |
| M1 | Report store와 dashboard backend | 1-2일 | PostgreSQL `raw_reports`/`findings`, stable ID가 있는 normalized finding, metadata index, Artifact Store backend plugin, scan health, final decision record, evidence bundle, 기본 dashboard retrieval view 동작 |
| M2 | Mgmt operator core + Feature orchestrator scaffold | 3-4일 | ClusterTarget/SecurityAssessment/ScanRun CRD, feature registry, feature orchestrator, desired state store, remote apply, bootstrap policy, SSA, finalizer, report writer, assessment scaffold 동작 |
| M3 | Security Assessment feature | 2-3일 | delivery artifact scanner report가 scan health를 포함한 normalized finding으로 변환 |
| M4 | Applied cluster configuration scan | 2일 | read-only cluster inspection이 Workload, RBAC, ServiceAccount, Secret reference 위험을 보고 |
| M5 | Trivy feature + image integrity | 2일 | delivery image CVE/SBOM/digest finding과 optional Trivy Operator `VulnerabilityReport` 입력이 중복 finding ID 없이 정규화 |
| M6 | Optional telemetry/inventory extension | 선택 | Phase 2 전용. OTel/LGTM/OSQuery/runtime telemetry는 별도 설계 검토 후 도입 |
| M7 | Final-check dashboard | 2-3일 | Overview, Targets, Assessments, Findings, Reports, Governance 메뉴 캡처 |
| M8 | Final-check validation | 1일 | delivery artifact scan, applied cluster configuration scan, report generation, Secret redaction, exception status, evidence bundle, no-auto-remediation guardrail end-to-end 검증 |
| M9 | AI remediation advisor (선택) | 선택 | 기본 OFF opt-in. ON 시 redaction·advisory sidecar·provenance·`scan_health=Warning` (reason=`ai_advisor_unavailable`) 생성, AI ON/OFF 판정 동일. 상세는 [AI_REMEDIATION.md](./AI_REMEDIATION.md) |

## 첫 구현 블록

이 절이 첫 구현 PR/블록(P0) 범위의 단일 정본이다(README "다음 구현 단계", ORCHESTRATOR/PROMPTS P0는 이 절을 참조한다).
첫 코드 블록은 모든 sensor를 한 번에 구현하지 않으며, 모노레포 3-모듈 중 `operator/` skeleton만 초기화한다.

P0에서 생성 완료된 항목(`go build ./... && go test ./...` 통과):

- ✅ `operator/` Go module(`github.com/bocopile/kube-sentinel/operator`)과 controller-runtime skeleton.
  임시 root `go.mod` placeholder는 `operator/go.mod`로 대체하고 root `go.mod`는 제거함.
- ✅ `ClusterTarget`, `SecurityAssessment`, `ScanRun` API type (cluster-scoped) + deepcopy + 생성된 CRD/RBAC/sample.
- ✅ 빈 reconciler 3종(status subresource 활성화; 실제 status patch와 discovery는 M0/M2).
- ✅ feature registry interface와 Feature orchestrator skeleton(`internal/feature`).
- ✅ Artifact Store backend plugin interface(`internal/artifactstore`, write+read; 실제 backend 구현은 M1).
- ✅ registry ordering + `profiles[]`/`features[]` merge resolver(`MergeFeatures`, 정본 profile→feature/umbrella 표 코드화) +
  unknown profile/unknown feature 분리 unit test.

P0에서 stub/placeholder로 두고 후속 milestone으로 미룬 항목:

- scanResources allowlist 확장 필드(`resources`/`nodeSelector`/`tolerations`/`scanResources.trivy.*`) — CRD에는 현재
  `securityAssessment.ttlSecondsAfterFinished`만 존재. 나머지는 M2/M5에서 타입·CRD에 추가한다.
- `ClusterTarget.status.capabilities` 관측 필드(`imageAccess`/`reportUpload`/`hostPath`) — discovery 구현(M0/M2) 시 추가.
- normalized Finding의 canonical `security.finding/v1` 전체 스키마(`target_type`/`rule_id`/`namespace`/`details`/`raw_report_id` 등) —
  operator `feature.Finding`은 placeholder이며 정본은 DATABASE `findings`; `internal/normalizer` DTO는 M3에서 구현한다.
- `TargetRunStatus` 실패 taxonomy(`AuthFailed`/`Unreachable`/`PermissionDenied`의 phase/reason 매핑) — M2에서 정의·구현.
- operator PostgreSQL 연결(`pgx` pool)과 마이그레이션 조율(backend 소유 migration 선행 적용) — M1에서 배선한다.

`.orchestrator/config.yaml`은 생성하지 않는다(orchestrator는 선택적 runner이며 정본 진행은 Claude Code 직접 구현이다.
[ORCHESTRATOR.md](./ORCHESTRATOR.md) 참조).

`backend/`/`frontend/` 모듈 디렉터리는 첫 블록 범위가 아니며 각각 후속 milestone에서 초기화한다.
이후 S0과 S1을 진행한다.
