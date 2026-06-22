# Security Assessment

이 문서는 고객사 인프라 적용 전 내부 최종점검 환경에서 실제 납품 대상 산출물과 Biz Cluster에 적용된 설정을 기준으로 보안 위험을 점검하기 위한 실행 환경, 입력 산출물, 대시보드 메뉴, 판정 정책을 정의한다.

용어:

- Mgmt Cluster: kube-sentinel 솔루션이 설치되는 관리 클러스터.
- Biz Cluster: 보안 점검 대상이 되는 업무/개발/검수 클러스터.

## Scope

현재 버전의 최종점검은 납품 산출물과 Biz Cluster에 실제 적용된 Kubernetes 설정을 기준으로 수행한다. 실시간 런타임 이벤트 탐지와 산출물-런타임 drift 분석은 다음 버전 범위로 둔다.

| 영역 | 현재 버전 포함 여부 | 주요 대상 |
|------|------------------|----------|
| Delivery Artifact Assessment | 포함 | 소스코드, 컨테이너 이미지, Helm/YAML, Dockerfile, RBAC, Secret 참조 방식, 배포 스크립트, SBOM, digest |
| Applied Cluster Configuration Assessment | 포함 | Biz Cluster에 적용된 Workload YAML, RBAC, ServiceAccount, Secret 참조, securityContext, volume 설정 |
| Runtime Event/Drift Assessment | Next Version | runtime sensor event, 산출물-클러스터 drift, 행위 기반 탐지 |

## Required Environment

| 구분 | 필요 항목 | 비고 |
|------|----------|------|
| Execution host | Mgmt Cluster 내 Mgmt-local scanner Job (Code / Artifact Scan); Biz Cluster는 read-only inspection + 옵션 remote scanner Job만 | Code / Artifact Scan은 operator가 Mgmt Cluster에 Job 생성; Biz Cluster Scan은 Mgmt operator read-only 조회 + bootstrap 허용 시 Biz remote Job. CI runner/점검 VM은 `artifactInput`을 Artifact Store/PVC에 준비하는 외부 사전 단계로만 사용 |
| Common tools | `kubectl`, `helm`, `jq`, `yq`, Docker 또는 `nerdctl` | Helm render, applied YAML 조회, image pull/digest 조회 |
| Source scanners | Semgrep, gosec | SAST 및 위험 코드 패턴 탐지 |
| Secret scanner | Gitleaks | hardcoded secret/token/account 탐지 |
| Image scanners | Trivy, Grype, Syft, Cosign 또는 Notation, Crane | CVE, SBOM, digest, 서명/무결성 확인 |
| Manifest scanners | kube-linter, conftest | Kubernetes 고위험 설정 및 RBAC policy 확인 |
| Build/script scanners | Hadolint, ShellCheck | Dockerfile 및 배포 스크립트 위험 확인 |
| Cluster access | `ClusterTarget` + Mgmt Cluster kubeconfig Secret | Biz Cluster에 적용된 YAML, RBAC, Secret 참조 방식 검수 |
| Report backend | PostgreSQL Report Store + Artifact Store evidence backend | finding 저장, 집계, evidence bundle, dashboard 조회 |

## Platform Components

최종점검은 단일 scanner가 아니라 여러 보안 도구와 결과 정규화, 대시보드, 예외 검토를 묶은 assessment platform으로 본다.

| 구성요소 | 역할 | 현재 버전 사용 방식 |
|----------|------|------------------|
| SonarQube | 소스코드 정적분석, code smell, security hotspot | Findings의 Source & Secrets 탭에 연계 |
| Semgrep/gosec | 보안 취약 패턴과 위험 코드 탐지 | SonarQube 보완 scanner로 사용 |
| Secret Scanner | 하드코딩 Secret, Token, 계정 정보 탐지 | Gitleaks 기준으로 source, values, YAML, script 검사 |
| Trivy/Grype | 컨테이너 이미지 CVE 및 Critical 취약점 탐지 | Findings의 Images & Integrity 탭에 집계 |
| Trivy Operator | Biz Cluster에 적용된 workload image 취약점 확인 | 현재 선택 입력. CRD와 read-only 권한이 있을 때 `VulnerabilityReport`를 보조 증적으로 정규화 |
| Kubernetes Policy Engine | Helm/YAML, RBAC, Pod security policy 검증 | conftest/OPA, kube-linter, Kyverno/Gatekeeper 정책으로 검사 |
| Image Signing/Verification | 이미지 digest, 서명, 무결성 검증 | Cosign/Notation/Crane으로 승인 digest와 서명 확인 |
| SBOM Store | 이미지 digest별 SBOM 저장, 추적, 재분석 기준 제공 | Syft/Trivy SBOM 결과를 digest 기준으로 보관 |
| Finding Normalizer | scanner별 결과를 공통 finding schema로 변환 | dashboard와 최종 판정이 동일 schema를 사용 |
| Result Dashboard | 점검 결과 조회, drill-down, 예외/조치 추적 | `Final Check Dashboard` 단일 화면으로 구성 |

## Assessment Support Feature Scope

Scanner만 실행하면 최종점검 결과의 신뢰성을 보장할 수 없다. 현재 버전은
보조 기능을 다음 세 범위로 나누고, 1차 필수 기능을 먼저 구현한다. 상세
기준은 [ASSESSMENT_SUPPORT_FEATURES.md](./ASSESSMENT_SUPPORT_FEATURES.md)를
따른다.

| 범위 | 포함 기능 | 현재 버전 판단 |
|------|-----------|----------------|
| 1차 필수 | Target preflight, Artifact input manifest, scanner baseline capture, stable finding ID/deduplication, Secret redaction guard, Evidence bundle export, Exception review artifact, Scan health summary | 최종점검 결과의 재현성과 증적성을 위해 필수 |
| 1차 선택 | Policy severity mapping, Finding schema validator, Namespace allowlist validator, Image digest resolver, read-only RBAC manifest generator, Trivy Operator `VulnerabilityReport`, Markdown/JSON report export | MVP에 포함 가능하되 필수 기능 뒤에 구현 |
| 후순위/2차 | Applied vs delivery manifest comparison, dashboard deep link metadata, audit log, OSQuery, OTel/LGTM telemetry, runtime event/drift, long-running DaemonSet sensor | 1차 안정화 뒤 별도 설계 |

Trivy Operator `VulnerabilityReport`는 유지한다. 단, Biz Cluster에 Trivy
Operator가 이미 설치되어 있고 해당 CRD에 대한 `get/list/watch` 권한이
있는 경우에만 보조 입력으로 사용한다. `VulnerabilityReport`가 없거나
권한이 없으면 전체 scan을 실패로 보지 않고 `scan_health`에 optional input
unavailable로 기록한다.

## Scan Profiles

사용자는 카테고리별로 관련 검사를 한 번에 실행할 수 있어야 한다.

`SecurityAssessment.spec.profiles[]` CRD 값과 프로파일 명칭 매핑:

| `spec.profiles[]` 값 | 프로파일 명칭 | 검사 그룹 |
|----------------------|--------------|----------|
| `SourceSecurity` | Source Security Scan | Code / Artifact Scan |
| `ImageSupplyChain` | Image Supply Chain Scan | Code / Artifact Scan |
| `KubernetesConfig` | Manifest & RBAC Manifest Scan | Code / Artifact Scan |
| `RBACAndSecretReference` | Applied RBAC & Secret Reference Scan | Biz Cluster Scan |
| `BuildAndDeploy` | Build & Deploy Scan | Code / Artifact Scan |

각 profile이 enable하는 내부 registry feature ID와 생성 `findings.category` 요약(정본은
[ARCHITECTURE.md](./ARCHITECTURE.md) §Profile / features → registry feature ID 매핑):

| `spec.profiles[]` | registry feature ID | `findings.category` |
|---|---|---|
| `SourceSecurity` | `source_security`, `secret_scan` | `sast`, `secret` |
| `ImageSupplyChain` | `image_vulnerability`, `image_integrity`, `sbom` | `image_vulnerability`, `integrity`, `sbom` |
| `KubernetesConfig` | `kubernetes_manifest`, `rbac_review` | `kubernetes`, `rbac` |
| `BuildAndDeploy` | `dockerfile_scan`, `script_scan` | `dockerfile`, `script` |
| `RBACAndSecretReference` | `applied_cluster_config`, `rbac_review`, `secret_reference` | `kubernetes`, `rbac`, `secret_ref`, `network` |

`spec.features[]`는 위 base set에 enable/disable·config override를 적용하고, unknown profile은
`ConfigError`로 기록하고 무시한다. 상세 병합 규칙은 ARCHITECTURE.md를 따른다.

검사 절차는 권한 모델과 실패 원인이 다르므로 `Code / Artifact Scan`과
`Biz Cluster Scan`으로 분리한다.

| 검사 그룹 | 목적 | 실행 위치 (정본) | artifact 전달 | 대표 산출물 |
|----------|------|------------------|---------------|------------|
| Code / Artifact Scan | 납품 산출물 자체의 보안 위험 확인 | Mgmt-local Job (`kube-sentinel-system`). Biz Cluster 접근 불필요 | init container가 `artifactInput`을 `emptyDir`로 clone/fetch | SAST report, secret report, image CVE/SBOM/digest report, manifest/RBAC/Dockerfile/script report → PostgreSQL `raw_reports` |
| Biz Cluster Scan | Biz Cluster에 실제 적용된 설정과 권한 상태 확인 | Mgmt controller read-only inspection(기본) + 옵션 remote scanner Job(Biz `kube-sentinel-system`) | kubeconfig 기반 API 조회; remote Job은 label 추적 | applied config report, RBAC risk report, Secret reference report, cluster scan health → PostgreSQL `raw_reports` |
| Full Final Check | 두 검사 그룹을 순차 실행하고 최종 판정 생성 | Mgmt Cluster `ScanRun` reconcile | — | final-check summary, normalized findings, exception candidates |

실행 토폴로지는 위 표가 단일 정본이다. 검사 그룹이 실행 위치를 deterministic하게 결정하므로
`SecurityAssessment`/`ScanRun` spec에 runner placement(`mgmt-local`|`biz-remote`) 선택 필드를 추가하지 않는다.

Code / Artifact Scan 프로파일:

| 프로파일 | 포함 검사 | 결과 메뉴 |
|----------|----------|----------|
| Source Security Scan | SonarQube, Semgrep, gosec, Gitleaks | Source & Secrets |
| Image Supply Chain Scan | Trivy/Grype, Syft, Cosign/Notation, Crane | Images & Integrity |
| Manifest & RBAC Manifest Scan | Helm render, kube-linter, conftest, RBAC manifest policy | Kubernetes Config & RBAC |
| Build & Deploy Scan | Hadolint, ShellCheck, deploy script inspection | Dockerfile & Scripts |

Biz Cluster Scan 프로파일:

| 프로파일 | 포함 검사 | 결과 메뉴 |
|----------|----------|----------|
| Applied Workload Config Scan | Pod/Deployment/DaemonSet/StatefulSet spec inspection | Kubernetes Config & RBAC |
| Applied RBAC Scan | Role, RoleBinding, ClusterRole, ClusterRoleBinding inspection | Kubernetes Config & RBAC |
| Secret Reference Scan | env/envFrom/volume Secret reference, ServiceAccount token automount inspection | Source & Secrets, Kubernetes Config & RBAC |
| Exposure Scan | Service/Ingress external exposure inspection | Kubernetes Config & RBAC |

`Full Final Check`는 Code / Artifact Scan을 먼저 실행하고, 산출물 누락이나
scanner 실행 실패를 `scan_health=Fail`로 기록한 뒤 Biz Cluster Scan으로
진행한다. Biz Cluster 접속 실패, RBAC denied, namespace allowlist 위반은
Code / Artifact Scan 결과와 분리된 cluster scan health로 기록한다.

## Assessment Workflows

워크플로우 문서는 검사 그룹별로 분리한다. 하나의 `Full Final Check`는 두
워크플로우를 순차 실행하지만, 각 워크플로우는 독립 실행, 재실행, 실패
분석이 가능해야 한다.

### Code / Artifact Workflow

목적은 Biz Cluster 접근 없이 납품 산출물 자체의 위험을 판단하는 것이다.

Code / Artifact Scan은 Mgmt-local Job으로 실행한다. `SecurityAssessment.spec.artifactInput`을
init container가 Job의 `emptyDir` 공유 volume으로 전달한 뒤 scanner container가 read-only로 mount한다.
전달 규약:

| `artifactInput` 필드 | 전달 방식 |
|----------------------|-----------|
| `sourceRef.path` / `manifestRef.path` | operator가 준비한 입력 mount에서 init container가 `emptyDir`로 복사 |
| `sourceRef.artifactStorePath` / `manifestRef.artifactStorePath` | init container가 Artifact Store에서 fetch 후 checksum 검증 |
| `imageList[].tarPath` | offline image tar를 `emptyDir`로 fetch |
| `imageList[].image` / `digestList[]` | registry digest 조회·pull(registry credential) |

```text
1. artifactInput preflight (Mgmt Cluster, Biz Cluster 미접속)
   - manifestRef/sourceRef/imageList/digestList 존재 확인
   - source, Dockerfile, Helm/YAML, RBAC manifest, script 존재 확인
   - image list, approved digest list, registry credential 확인
   - checksum이 있으면 fetch 전후 digest 일치 검증 (누락·불일치 → scan_health=Fail)
2. Mgmt-local Job init container가 artifactInput을 emptyDir 공유 volume으로 clone/fetch
3. scanner 기준선 확인
   - scanner version
   - vulnerability DB/rule 기준일
   - policy bundle version
4. source/secret scan 실행 (emptyDir mount)
5. image/SBOM/digest/signature scan 실행
6. manifest/RBAC/Dockerfile/script scan 실행
7. raw report 저장(PostgreSQL raw_reports)
8. finding 정규화
9. `ScanRun.status.artifactScan` 갱신
```

실패 기준:

- 필수 산출물 누락
- scanner 실행 실패
- 취약점 DB 또는 rule 기준일 미확인
- registry pull/digest 조회 실패
- Critical finding, Secret 노출, digest mismatch, 서명 검증 실패

### Biz Cluster Workflow

목적은 Biz Cluster에 실제 적용된 설정과 권한 상태를 확인하는 것이다.

```text
1. ClusterTarget 조회
2. kubeconfig Secret 참조 확인
3. Biz Cluster preflight
   - API server 연결 확인
   - namespace allowlist 확인
   - read-only RBAC 확인
   - optional CRD/bootstrap capability 확인
4. applied workload spec 조회
5. applied RBAC/ServiceAccount 조회
6. Secret reference, ServiceAccount token automount, Service/Ingress 조회
7. 필요 시 허용된 scanner Job remote apply
8. raw inspection report 저장(PostgreSQL raw_reports)
9. finding 정규화
10. `ClusterTarget.status`와 `ScanRun.status.clusterScan` 갱신
```

실패 기준:

- kubeconfig Secret 누락 또는 인증 실패
- API server unreachable
- namespace allowlist 위반
- read-only RBAC denied
- Secret raw data 조회 시도
- 승인되지 않은 privileged/hostPath/hostNetwork/RBAC wildcard/cluster-admin

### Full Final Check Workflow

```text
1. ScanRun 생성
2. Code / Artifact Scan 실행
3. Biz Cluster preflight
   - kubeconfig Secret 참조 확인
   - API server 연결 확인
   - namespace allowlist 확인
   - read-only RBAC 확인
   - optional CRD/bootstrap capability 확인
4. Biz Cluster Scan 실행
   - applied workload/RBAC/ServiceAccount/Secret reference 조회
   - 필요 시 허용된 scanner Job remote apply
5. finding 정규화
6. Code 결과와 Biz Cluster 결과 상관 분석
7. 최종 판정과 예외 검토 항목 생성
```

`Full Final Check`는 Code / Artifact Workflow 실패를 Biz Cluster Workflow
결과로 덮어쓰지 않는다. 두 워크플로우의 실패 원인은 `artifactScan`,
`clusterScan`, `scan_health`에서 각각 확인되어야 한다.

현재 버전의 최종 산출물은 보고서와 대시보드 판정이다. 시스템은 finding,
증적, 개선 권고, 예외 검토 후보를 생성하지만 Biz Cluster의 애플리케이션
workload, RBAC, Service, Ingress, Secret을 자동으로 수정하지 않는다.

## Required Inputs

| 입력 | 필수 여부 | 설명 |
|------|----------|------|
| Source code | 필수 | 실제 납품 기준 전체 소스 |
| Dockerfile | 필수 | 이미지 빌드 기준 파일 |
| Helm chart and values | 필수 | 고객사 적용 전 렌더링 기준 |
| Kubernetes YAML | 필수 | Helm 외 직접 적용 산출물 |
| RBAC manifests | 필수 | Role, ClusterRole, RoleBinding, ClusterRoleBinding |
| Deployment scripts | 필수 | install, upgrade, rollback, bootstrap script |
| Image list | 필수 | 납품 대상 image repository/tag 목록 |
| Approved image digest list | 필수 | 납품 승인 digest 기준 |
| Registry credentials | 필요 시 | private registry pull 및 digest 조회용 |
| Offline image tar | 필요 시 | 폐쇄망 또는 registry 접근 불가 시 분석 대상 |
| SBOM | 선택 | 외부 생성 SBOM이 있으면 scanner 결과와 비교 |
| Existing exception review file | 선택 | 이전에 승인된 예외, 만료일, 승인자, 사유를 입력으로 재사용 |
| Generated exception review artifact | 필수 산출물 | 이번 ScanRun에서 생성된 예외 후보, owner, 사유, 만료일, 승인 상태 추적 |
| ClusterTarget | 필요 시 | read-only kubeconfig Secret을 참조하는 Biz Cluster 등록 정보 |

위 입력은 `SecurityAssessment.spec.artifactInput`(`sourceRef`/`imageList`/`digestList`/`manifestRef`)으로
선언하고, Code / Artifact Scan Mgmt-local Job의 init container가 `emptyDir` 공유 volume으로 전달한다.
preflight는 입력 manifest의 존재와 checksum을 검증하며, 누락·checksum 불일치는 `scan_health=Fail`(필수
산출물 누락)로 기록한다. 대용량 원문은 CRD에 인라인하지 않고 Artifact Store `artifact-input.yaml` 및 입력
경로 참조만 둔다.

## Cluster Access Policy

Biz Cluster 접근은 현재 버전에 포함하되, 실시간 런타임 탐지 목적이 아니라 적용된 설정 검수 목적으로 제한한다. Dashboard와 API는 Mgmt Cluster의 `ClusterTarget` 목록과 status만 조회하고 kubeconfig Secret data를 노출하지 않는다.

| 권한 범위 | 필요 권한 | 제한 |
|----------|----------|------|
| Workload spec | Pod, Deployment, DaemonSet, StatefulSet, ReplicaSet 조회 | spec, securityContext, volume, image, ServiceAccount 확인 |
| RBAC | Role, RoleBinding, ClusterRole, ClusterRoleBinding 조회 | wildcard, cluster-admin, 민감 리소스 권한 확인 |
| ServiceAccount | ServiceAccount 조회 | token automount, binding 관계 확인 |
| ConfigMap/Secret reference | Workload의 env/envFrom/volume 참조 확인 | Secret raw data 조회 금지 |
| Service/Ingress | Service, Ingress 조회 | 외부 노출 설정 확인 |

Secret 값은 어떤 경우에도 report, log, dashboard에 저장하지 않는다. 점검 대상은 하드코딩된 값과 Kubernetes 리소스의 Secret 참조 방식이다.

## Current Check Items

| # | 점검 항목 | 기본 도구 | 실패 기준 |
|---|----------|----------|----------|
| 1 | 소스코드 정적분석 기반 보안 취약 패턴 및 위험 코드 존재 여부 | Semgrep, gosec | Critical/High rule 존재 |
| 2 | 하드코딩된 Secret, Token, 계정 정보 등 민감정보 노출 여부 | Gitleaks, applied YAML inspection | verified/high confidence secret 존재 또는 Secret 값 직접 포함 |
| 3 | 컨테이너 이미지 취약점 및 Critical 취약점 존재 여부 | Trivy, Grype | Critical CVE 존재 또는 fixable High 과다 |
| 4 | 이미지 digest 및 무결성 불일치 여부 | Syft, Cosign/Notation, Crane | 승인 digest 불일치, 서명 검증 실패, SBOM 누락 |
| 5 | `privileged`, `hostPath` 등 고위험 Kubernetes 설정 여부 | kube-linter, conftest, applied YAML inspection | 고위험 policy 위반 |
| 6 | RBAC 과권한 및 불필요한 권한 부여 여부 | conftest, rbac-police, applied RBAC inspection | wildcard, cluster-admin, 민감 리소스 과권한 |
| 7 | Dockerfile 및 배포 스크립트 내 보안 위험 요소 | Hadolint, ShellCheck | High 이상 rule 또는 shellcheck error 존재 |
| 8 | 스캔 실패, 분석 불가, 필수 산출물 누락 여부 | security-assessment orchestrator | 필수 report 누락 또는 scanner error |
| 9 | 개선 권고 및 예외 검토 필요 항목 | exception review | 미승인/만료 예외, 개선 권고 누락 |

## Finding Categories

| Category | 대상 | 대표 실패 조건 |
|----------|------|---------------|
| `sast` | Source code | Critical/High rule 탐지 |
| `secret` | Source, values, scripts, applied YAML | verified/high confidence secret 탐지 또는 Secret 값 직접 포함 |
| `image_vulnerability` | Container image | Critical CVE 또는 fixable High 과다 |
| `sbom` | Image SBOM | 납품 이미지별 SBOM 누락 |
| `integrity` | Image digest/signature | 승인 digest 불일치, 서명 검증 실패 |
| `kubernetes` | Helm/YAML, applied workload spec | privileged, hostPath, hostNetwork, hostPID, hostIPC, root 실행 |
| `rbac` | RBAC manifests, applied RBAC | wildcard, cluster-admin, secrets get/list/watch, pods/exec 과권한 |
| `secret_ref` | Workload env/envFrom/volume Secret 참조, ServiceAccount token automount | Secret raw value 조회 시도, automount 미비활성화, 과도한 token projection |
| `network` | Service, Ingress 외부 노출 설정 | 승인되지 않은 NodePort/LoadBalancer, TLS 미적용, 과도한 Ingress rule |
| `dockerfile` | Dockerfile | root user, floating tag, unsafe package install, secret copy |
| `script` | Deployment scripts | unsafe shell, unchecked command, secret echo |
| `scan_health` | Scanner pipeline | scanner error, unsupported target, missing artifact, stale DB/rule |

## Dashboard Menus

대시보드는 여러 개로 분리하지 않고 `Final Check Dashboard` 하나로 구성한다. 메뉴는 도구별이 아니라 최종점검 판단 흐름 기준으로 나눈다.

상단 공통 필터:

- Environment: `dev`, `final-check`
- Target version/build
- Scan run ID
- Namespace
- Image
- Severity
- Category
- Exception status

| 메뉴 | 주요 질문 | 주요 위젯 | 기본 액션 |
|------|----------|----------|----------|
| Overview | 지금 납품 가능한 상태인가? | Overall status, Critical/High count, failed scans, missing artifacts, exception-required count, last scan time | 실패 원인 Top 5로 drill-down |
| Targets | 어떤 Biz Cluster를 점검할 수 있는가? | ClusterTarget, connection phase, namespace allowlist, capability, last validation time | cluster add/import, preflight 실패 원인 확인 |
| Assessments | 어떤 검사가 실행됐고 어디서 실패했는가? | Code / Artifact Scan, Biz Cluster Scan, Full Final Check, retry/resume state | 실패 workflow 재실행 |
| Findings | 어떤 보안 위험이 발견됐는가? | Source & Secrets, Images & Integrity, Kubernetes Config & RBAC, Dockerfile & Scripts 탭 | finding 상세와 개선 가이드 확인 |
| Reports | 어떤 보고서와 증적을 남길 수 있는가? | final-check report, evidence bundle, raw reports, normalized findings, scan health summary | report/evidence export |
| Governance | 무엇을 고치거나 승인해야 하는가? | remediation list, exception candidates, approved exceptions, expired exceptions, rescan status | 예외 승인/만료/재점검 상태 확인 |

Source & Secrets, Images & Integrity, Kubernetes Config & RBAC, Dockerfile &
Scripts는 top-level 메뉴가 아니라 Findings 또는 Assessments 내부 탭으로
제공한다.

## Decision Policy

| 조건 | 기본 판정 |
|------|----------|
| Critical finding 존재 | Fail |
| Secret 노출 | Fail |
| 이미지 digest 불일치 | Fail |
| 서명 필수 이미지의 서명 검증 실패 | Fail |
| scanner 실패 또는 분석 불가 | Fail |
| 필수 산출물 누락 | Fail |
| 승인되지 않은 privileged/hostPath/hostNetwork/hostPID/hostIPC | Fail |
| 승인되지 않은 RBAC wildcard 또는 cluster-admin | Fail |
| High finding 존재 | 개선 또는 예외 승인 필요 |
| 예외 만료 또는 미승인 예외 | Fail |

`final-check` 환경에서는 scanner exit code 하나에만 의존하지 않는다. PostgreSQL `findings`, scan health, 필수 산출물 존재 여부, DB/rule 기준일, 예외 승인 상태를 함께 평가한다.

AI remediation advisor(선택 기능)가 활성화되어도 최종 판정은 deterministic 입력만
사용한다. AI가 생성한 조치 가이드는 advisory이며 severity, Pass/Fail,
`exception_required`에 영향을 주지 않는다. 외부 Gemini 전송 시 마스킹과 provenance가
필수이며 상세는 [AI_REMEDIATION.md](./AI_REMEDIATION.md)를 따른다.

## Next Version: Runtime Assessment

다음 버전에서는 실제 최종점검 클러스터의 런타임 상태를 포함한다.

| 후보 항목 | 설명 |
|----------|------|
| Runtime image drift | 실제 실행 image digest와 승인 digest 비교 |
| Runtime event correlation | runtime sensor 이벤트와 산출물 finding 연결 |
| Runtime behavior validation | exec, privilege escalation, suspicious API call 등 행위 기반 검증 |
| Runtime event | runtime sensor 이벤트 기반 고위험 행위 확인 |
