# 로드맵

PoC는 수직 slice 단계로 구현한다. 각 단계가 끝난 뒤 리포지터리는 build 및
test 가능한 상태여야 한다.

## Stage gate

| Stage | 범위 | Exit criteria |
| --- | --- | --- |
| S0 | Assessment prerequisite | Mgmt namespace, Biz Cluster kubeconfig Secret, target preflight, bootstrap policy, read-only RBAC, image access, report store write test 통과 |
| S0.5 | Delivery artifact security assessment runner baseline | artifact input manifest를 기준으로 실제 scanner를 실행하고 raw report, minimal normalized finding, scan health file 생성 |
| S1 | Report store와 finding schema spike | S0.5 raw report fixture가 stable finding ID, metadata record, report artifact, evidence bundle로 적재 |
| S2 | Mgmt 단일 operator와 assessment vertical slice | `ClusterTarget`, `SecurityAssessment`, `ScanRun`이 Feature orchestrator/registry, remote apply/read-only inspection으로 assessment Job과 report record 생성 |
| S3 | Controller-integrated assessment capability | S0.5 scanner report convention, Trivy image report integration, applied cluster config scan을 각각 enable/disable, assess, verify 가능 |
| S4 | Final-check validation과 dashboard | Delivery artifact scan, applied cluster configuration scan, dashboard, report, exception review 통과 |
| S5 | Phase 2 telemetry/inventory | OTel/LGTM/OSQuery/runtime telemetry는 제품 요구사항이 된 경우에만 설계 후 추가 |

## Milestone

| Milestone | 설명 | 예상 기간 | Exit criteria |
| --- | --- | ---: | --- |
| M0 | Assessment readiness check | 1일 | Mgmt namespace, Biz Cluster kubeconfig Secret, target preflight, bootstrap policy, read-only RBAC, image access, report store write test 통과 |
| M0.5 | Delivery artifact security assessment runner baseline | 2-3일 | required artifact, artifact input manifest, scanner version, vulnerability DB baseline, image access, digest list, SAST/Secret/Image/SBOM/Integrity/Manifest/RBAC/Dockerfile/Script raw report, minimal normalized finding, scan health file 검증 |
| M1 | Report store와 dashboard backend | 2일 | M0.5 report fixture를 report artifact, stable ID가 있는 normalized finding, metadata index, Artifact Store backend plugin, scan health, final decision record, evidence bundle, 기본 dashboard retrieval view로 적재 |
| M2 | Mgmt operator core + Feature orchestrator scaffold | 3-4일 | ClusterTarget/SecurityAssessment/ScanRun CRD, feature registry, feature orchestrator, desired state store, remote apply, bootstrap policy, SSA, finalizer, report writer, assessment scaffold 동작 |
| M3 | Security Assessment feature integration | 1-2일 | controller-managed assessment Job/CronJob이 M0.5 scanner runner/report convention을 사용하고 scan health를 포함한 normalized finding을 Report Store로 연결 |
| M4 | Applied cluster configuration scan | 2일 | read-only cluster inspection이 Workload, RBAC, ServiceAccount, Secret reference 위험을 보고 |
| M5 | Trivy feature + image integrity integration | 2일 | M0.5 delivery image CVE/SBOM/digest report 경로를 controller feature로 통합하고 optional Trivy Operator `VulnerabilityReport` 입력이 중복 finding ID 없이 정규화 |
| M6 | Optional telemetry/inventory extension | 선택 | Phase 2 전용. OTel/LGTM/OSQuery/runtime telemetry는 별도 설계 검토 후 도입 |
| M7 | Final-check dashboard | 2-3일 | Overview, Targets, Assessments, Findings, Reports, Governance 메뉴 캡처 |
| M8 | Final-check validation | 1일 | delivery artifact scan, applied cluster configuration scan, report generation, Secret redaction, exception status, evidence bundle, no-auto-remediation guardrail end-to-end 검증 |

## 첫 구현 블록

첫 코드 블록은 모든 sensor를 한 번에 구현하지 않는다. 다음 항목을 생성한다.

- Go module과 controller-runtime project skeleton
- `ClusterTarget`, `SecurityAssessment`, `ScanRun` API type
- status patching이 포함된 빈 reconciler
- feature registry interface와 Feature orchestrator skeleton
- Artifact Store backend plugin interface
- registry ordering과 unknown feature validation unit test
- `orchestrator init`으로 생성한 `.orchestrator/config.yaml`

이후 S0, S0.5, S1 순서로 진행한다. S0.5에서 scanner 실행과 raw report 산출물
규칙을 먼저 고정하고, S1에서 이를 Report Store와 dashboard read model에 적재한다.
