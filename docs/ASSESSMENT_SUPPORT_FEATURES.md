# Assessment Support Features

이 문서는 최종점검 MVP에서 보안 취약점 자체를 탐지하는 scanner 외에, 점검 결과의 신뢰도, 재현성, 증적성을 보장하기 위해 필요한 보조 기능을 범위별로 분리한다.

1차 범위의 기준은 다음과 같다.

- 고객사 인프라 적용 전 납품 산출물과 Biz Cluster 적용 설정을 점검한다.
- 결과는 보고서, 증적, 예외 검토 항목으로 남긴다.
- Biz Cluster의 애플리케이션 workload, RBAC, Service, Ingress, Secret을 자동 수정하지 않는다.
- OSQuery, OTel, LGTM full stack, runtime event, long-running DaemonSet은 2차 확장으로 둔다.
- Trivy Operator `VulnerabilityReport`는 2차 기능이 아니라 1차 선택 입력으로 유지한다.

## 1차 필수 기능

| 기능 | 목적 | 산출물/상태 |
|------|------|-------------|
| Target preflight check | kubeconfig 누락, API unreachable, RBAC denied, namespace 누락, 금지된 Secret read 권한을 실제 scan 전에 분리한다. | `ClusterTarget.status`, `ScanRun.status.clusterScan`, preflight report |
| Artifact input manifest | `SecurityAssessment.spec.artifactInput` 및 `artifact-input.yaml`로 source path, image list, digest list, Helm/YAML, RBAC, Dockerfile, script 위치와 checksum을 선언한다. preflight에서 존재·checksum 검증 후 Code / Artifact Scan Mgmt-local Job의 init container가 Artifact Store/PVC에서 `emptyDir`로 staging한다. | `artifact-input.yaml`, input validation report, Mgmt-local Job staging mount |
| Scanner version / DB baseline capture | Trivy DB, Grype DB, Semgrep rule, Gitleaks rule, policy bundle 기준일을 기록한다. | scanner baseline report |
| Finding stable ID / deduplication | 같은 CVE/rule finding이 재스캔 때 중복 집계되지 않도록 안정 ID를 만든다. | deterministic `finding_id`, dedup summary |
| Secret redaction guard | report, log, dashboard, artifact에 Secret 원문이 섞이지 않도록 마지막 방어선을 둔다. | redaction check result, blocked output log |
| Evidence bundle export | raw report, normalized findings, summary, scan health, exception 후보를 묶어 납품/검수 증적으로 남긴다. | evidence bundle archive or directory |
| Exception review artifact | 자동 remediation 대신 이번 ScanRun의 예외 후보, 승인 상태, 만료일, 사유, owner를 추적한다. 기존 승인 예외 파일은 선택 입력으로 병합할 수 있다. | `exception-review.yaml` 또는 `exception-review.md` |
| Scan health summary | scanner 실패, unsupported target, missing artifact, stale DB를 취약점 없음으로 오판하지 않게 한다. | `scan_health` finding, summary report |

## 1차 선택 기능

1차 선택 기능은 MVP에 포함해도 범위가 과도하게 커지지 않지만, 구현 순서는 필수 기능 뒤로 둔다.

| 기능 | 목적 | 적용 기준 |
|------|------|-----------|
| Policy severity mapping | scanner별 severity를 내부 Critical/High/Fail 기준으로 통일한다. | scanner 2개 이상 결과를 같은 dashboard/report에서 비교할 때 필요 |
| Finding schema validator | 정규화 결과가 schema를 벗어나 dashboard/report가 조용히 틀어지는 것을 방지한다. | Finding Normalizer 구현 시 포함 권장 |
| Namespace allowlist validator | Biz Cluster scan이 의도하지 않은 namespace를 조회하지 않도록 제한한다. | Biz Cluster Scan 활성화 시 포함 권장 |
| Image digest resolver | tag 입력을 실제 digest로 고정해 취약점/무결성 판단 기준을 명확히 한다. | image list가 tag를 포함할 때 필요 |
| Read-only RBAC manifest generator | Biz Cluster별 최소 권한 ServiceAccount/Role/ClusterRole 예시를 제공한다. | 자동 적용이 아니라 bootstrap 템플릿 생성으로 제한 |
| Trivy Operator VulnerabilityReport ingestion | Biz Cluster에 Trivy Operator가 이미 설치되어 있으면 `VulnerabilityReport`를 read-only로 보조 입력으로 사용한다. | CRD 존재와 get/list/watch 권한이 확인된 경우만 사용 |
| Report format export | 사람이 보는 Markdown/HTML과 시스템 연동용 JSON을 제공한다. | SARIF는 외부 연동 요구가 있을 때 추가 |
| AI remediation advisor | finding 조치 가이드를 AI(공개 Gemini API)로 보강해 보고서에 advisory로 추가한다. 상세는 [AI_REMEDIATION.md](./AI_REMEDIATION.md). | 기본 OFF opt-in, egress 허용 환경에서만. 판정 불개입, 마스킹·provenance·실패격리 필수 |

## 후순위 또는 2차 기능

아래 기능은 유용하지만 1차 최종점검 MVP의 필수 기능으로 두지 않는다.

| 기능 | 후순위 사유 |
|------|------------|
| Applied vs delivery manifest comparison | 가치가 높지만 Helm render, Kustomize, applied object defaulting 차이를 정규화해야 하므로 기본 scan 안정화 뒤 진행한다. |
| Dashboard deep link metadata | finding 상세 검토 시간을 줄이지만 report schema와 artifact path 규칙이 안정된 뒤 붙이는 것이 안전하다. |
| Audit log for scan execution | 제품화 단계에서 중요하지만 PoC 최소 범위는 ScanRun metadata와 report timestamp로 시작할 수 있다. |
| OSQuery inventory | host/node inventory 성격이 강하므로 2차 optional inventory 확장으로 둔다. |
| OTel/LGTM full telemetry | 최종점검 보고서의 source of truth가 아니므로 2차 export/observability 확장으로 둔다. |
| Runtime event/drift assessment | 실시간 행위 탐지와 runtime drift는 현재 산출물 최종점검과 별도 버전으로 분리한다. |
| Long-running DaemonSet sensor model | 1차는 read-only API inspection과 scan Job 중심이며 장수 센서는 2차 설계 후 도입한다. |

## Trivy Operator VulnerabilityReport 정책

Trivy Operator `VulnerabilityReport`는 유지한다.
단, 1차 범위에서의 역할은 다음으로 제한한다.

- 기본 경로는 납품 이미지 registry digest 또는 image tar를 직접 scan하는 Code / Artifact Scan이다.
- `VulnerabilityReport`는 Biz Cluster에 Trivy Operator가 이미 설치되어 있고, read-only 권한이 확인된 경우에만 보조 입력으로
  사용한다.
- `VulnerabilityReport`가 없거나 권한이 없다는 이유만으로 전체 scan을 실패 처리하지 않는다.
  대신 `scan_health`에 optional input unavailable로 기록한다.
- `VulnerabilityReport`의 결과는 delivery image scan 결과와 같은 `finding_id` 규칙으로 정규화하고 중복 집계하지 않는다.
- Trivy Operator 설치, 운영, 자동 remediation은 1차 범위에 포함하지 않는다.
