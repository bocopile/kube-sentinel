# 요구사항

## 목표

kube-sentinel은 Mgmt Cluster 기반 Kubernetes 최종점검 보안 평가 PoC입니다.
Mgmt Cluster에는 단일 `kube-sentinel-operator`를 설치하고, Biz Cluster에는 kube-sentinel operator/CRD를 설치하지
않습니다.
Mgmt Cluster CRD는 Biz Cluster를 등록하고, 납품 산출물 보안 평가를 실행하며, remote apply/read-only 접근으로 Biz Cluster 적용
설정을 점검하고, report/evidence 결과를 dashboard에 게시해야 합니다.

## 성공 기준

| ID | 요구사항 | 검증 방법 |
| --- | --- | --- |
| G1 | `ClusterTarget`/`SecurityAssessment`가 존재하고 backend `POST /api/v1/scan-runs` 또는 수동 `ScanRun` CR apply로 실행을 트리거하면 management controller가 Code / Artifact Scan은 Mgmt-local Job으로 실행하고, Biz Cluster Scan은 read-only inspection(옵션으로 허용된 remote scanner Job)으로 실행한다. | ScanRun trigger 후 Mgmt Cluster에서 `kubectl get clustertarget,securityassessment,scanrun`과 Mgmt-local Code / Artifact Scan Job 확인. Biz Cluster Scan을 remote Job으로 실행하도록 구성한 경우에만 Biz Cluster target namespace에서 `kubectl get job,cronjob,cm,sa,role,rolebinding` 확인 |
| G2 | Feature toggle은 feature별 managed resource를 생성하거나 제거한다. | `spec.features[].enabled` patch 후 resource 생성/삭제 확인 |
| G3 | allowlist 기반 scan resource config로 선택된 scan Job의 resource와 scheduling field를 변경할 수 있다. | `spec.scanResources` patch 후 생성된 workload spec과 거부된 금지 field 확인 |
| G4 | Trivy와 security assessment data는 PostgreSQL query record와 evidence/export artifact로 정규화된다. | `raw_reports`, `findings`, scan health, evidence/export artifact 검토 |
| G5 | Dashboard view는 finding, vulnerability, scan health, final-check security assessment 결과를 노출한다. | dashboard screenshot 캡처 |
| G6 | Final Check Dashboard는 의사결정 중심 메뉴로 assessment result를 노출한다. | dashboard screenshot 캡처 |
| G7 | scope, discovery, priority, validation, exception review에 대한 evidence/decision mapping이 통과한다. | evidence bundle, final decision summary, exception review artifact 검토 |
| G8 | 소스코드 정적분석으로 risky code와 security anti-pattern을 식별한다. | Semgrep/gosec report 검토 |
| G9 | hardcoded secret, token, credential, account information을 탐지한다. | Gitleaks report 검토 |
| G10 | 컨테이너 이미지 Critical vulnerability와 위험한 base image를 탐지한다. | Trivy/Grype image scan report 검토 |
| G11 | image digest, SBOM, signature/integrity 검증 결과를 생성한다. | Syft/Cosign/Crane output 검토 |
| G12 | Helm/YAML, RBAC, Dockerfile, deployment script의 고위험 설정을 탐지한다. | kube-linter/conftest/hadolint/shellcheck report 검토 |
| G13 | Kubernetes YAML, RBAC, Secret reference에 대한 적용된 개발 cluster 설정 위험을 탐지한다. | rendered/applied workload spec, RBAC, ServiceAccount, Service/Ingress, Secret reference path 확인 |
| G14 | scanner failure, unsupported scan, missing required artifact는 failed scan health finding으로 노출된다. | security assessment summary 검토 |
| G15 | 1차 assessment support feature는 재현 가능한 입력, scanner baseline, stable finding ID, Secret redaction, evidence bundle, exception review, scan health summary를 제공한다. | `artifact-input.yaml`, scanner baseline report, normalized finding, evidence bundle, exception review artifact, scan health summary 검토 |
| G16 | Trivy Operator `VulnerabilityReport`는 존재할 때 선택적 read-only 입력으로 지원되며 필수 의존성이 아니다. | optional `VulnerabilityReport` ingestion 또는 `optional input unavailable` scan health status 확인 |
| G17 | Mgmt Cluster 단일 operator가 Feature-as-Plugin registry를 통해 검사 기능을 오케스트레이션한다. | Reconciler 변경 없이 feature enable/disable, priority ordering, status reporting 확인 |
| G18 | Biz Cluster Scan 전 preflight가 누락된 bootstrap 항목을 식별하고, 정책상 허용된 항목만 설치한다. | namespace/RBAC/image pull/report upload/optional CRD check 결과와 bootstrap audit 확인 |
| G19 | Artifact Store는 backend plugin으로 교체 가능하며 S3/MinIO에 고정되지 않는다. | Filesystem 또는 SeaweedFS/S3-compatible backend 설정 전환 후 report artifact 조회 확인 |
| G20 | AI remediation advisor는 기본 OFF opt-in이며, ON 시 advisory sidecar, provenance, redaction, `scan_health=Warning` (reason=`ai_advisor_unavailable`) 기록을 제공한다. AI 실패는 scan Fail이 아니다. | AI ON/OFF scan에서 sidecar/provenance 생성, redaction fixture, Gemini 실패 시 scan Completed 확인 ([AI_REMEDIATION.md](./AI_REMEDIATION.md)) |
| G21 | AI ON/OFF 동일 scan에서 finding count, severity, final decision이 동일하다(판정 비개입). | AI ON/OFF A/B 결과 비교 |

G1~G19는 1차 필수 성공 기준이다.
G20/G21은 AI remediation advisor opt-in 시에만 적용하는 1차 선택 기준이며, 상세는
[AI_REMEDIATION.md](./AI_REMEDIATION.md)를 따른다.

## 비목표

- Biz Cluster별 operator 설치
- 자동 인프라 remediation 또는 고객 application workload mutation
- inline blocking 또는 policy enforcement
- Kafka 또는 streaming middleware
- 완전한 OCSF normalization
- production-grade high availability

## 필수 프로젝트 capability

- controller-runtime 기반 Go Kubernetes operator
- `security.kube-sentinel.io/v1alpha1` 아래 `ClusterTarget`, `SecurityAssessment`, `ScanRun` CRD
- Mgmt Cluster에서 실행되는 단일 `kube-sentinel-operator`
- Code / Artifact Scan, Biz Cluster Scan, Final Decision을 위한 assessment workflow architecture
- Reconciler가 Feature를 오케스트레이션하고 scanner별 세부 구현은 Feature plugin이 담당하는 Feature-as-Plugin architecture
- priority 기반 Feature plugin registry와 deterministic ordering
- Mgmt-local resource와 Biz-remote scan resource에 한정된 server-side apply
- Biz Cluster Scan 전 preflight와 허용된 bootstrap resource 설치
- feature readiness, config error, apply error, degraded runtime state에 대한 status reporting
- raw scanner output, normalized finding, scan health, final decision, exception candidate를 위한
  Report Store와 Evidence Bundle 생성
- PostgreSQL `raw_reports`/`findings` 정본, evidence bundle용 normalized JSONL/JSON export, metadata
  index, stable artifact reference, dashboard/API read model을 이용한 결과 저장 및 조회
- Filesystem, S3-compatible, MinIO, SeaweedFS, NFS/PVC 등으로 교체 가능한 Artifact Store backend plugin
  interface
- source, secret, image, SBOM, integrity, Kubernetes manifest, RBAC, Dockerfile, script risk에 대한
  delivery artifact security assessment
- Pod security setting, RBAC, Secret reference, ServiceAccount token behavior, Service/Ingress
  exposure에 대한 applied cluster configuration assessment
- artifact failure, cluster connectivity failure, RBAC denied, skipped cluster scan이 섞이지 않도록 Code /
  Artifact Scan과 Biz Cluster Scan phase 분리
- scanner error, unsupported target, missing required artifact, stale vulnerability database 또는
  policy rule에 대한 scan health reporting
- 납품 전 승인이 필요한 finding에 대한 exception 및 remediation tracking
- finding, evidence, remediation recommendation, scan health, exception candidate를 포함한 report
  generation.
  PoC는 Biz Cluster infrastructure를 자동 수정하지 않는다.
- [ASSESSMENT_SUPPORT_FEATURES.md](./ASSESSMENT_SUPPORT_FEATURES.md)에 정의된 assessment support
  feature.
  1차 필수 기능은 optional telemetry 또는 inventory extension보다 먼저 구현한다.
- Final Check Dashboard asset

## 환경 가정

- Mgmt Cluster는 Biz Cluster kubeconfig Secret을 저장하고 kube-sentinel management controller를 실행한다.
- Code / Artifact Scan scanner Job은 Mgmt Cluster에만 생성한다.
  Biz Cluster remote scanner Job은 Biz Cluster Scan profile + bootstrap/capability 허용 시에만 생성한다.
- Biz Cluster 누락 항목은 preflight에서 먼저 식별한다.
  Mgmt operator는 `ClusterTarget.spec.bootstrapPolicy`가 허용한 kube-sentinel 전용 리소스만 생성한다.
- Report Store와 Dashboard storage는 Mgmt Cluster 안에 있거나 Mgmt Cluster에서 접근 가능하다.
- dashboard/API filtering은 PostgreSQL `raw_reports`/`findings`/summary records를 사용한다.
  Artifact Store의 `manifest.json`은 `artifact_index` 재생성에만 사용하며 raw/finding 정본을 대체하지 않는다.
- image scanner와 필요한 scanner image는 Mgmt Cluster scanner Job image 또는 해당 Job이 접근 가능한
  registry/artifact path에서 실행 가능하다.
- Biz Cluster는 승인된 namespace와 applied configuration assessment에 필요한 cluster-level RBAC resource 범위의
  read-only credential로 조회할 수 있다.
- target kubeconfig Secret data는 status, dashboard, log, report에 노출하지 않는다.
- private registry access, approved image digest list, optional offline image tar artifact는 image
  vulnerability와 integrity check에 사용할 수 있다.
- vulnerability database와 scanner rule set은 final-check 실행 전에 승인된 baseline date로 update 또는 pin한다.
- Secret value는 수집하거나 report에 기록하지 않는다.
  Secret reference, mount, environment reference, ServiceAccount token setting만 평가한다.
- Trivy Operator `VulnerabilityReport`는 CRD와 read-only permission이 이미 존재할 때만 선택적 Biz Cluster 입력으로 읽을
  수 있다.
  Trivy Operator 설치 또는 운영은 1차 범위 요구사항이 아니다.
