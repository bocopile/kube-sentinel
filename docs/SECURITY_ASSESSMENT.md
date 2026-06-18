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
| Execution host | Linux 점검 VM 또는 CI runner | scanner 실행, report 생성, artifact mount 가능 |
| Common tools | `kubectl`, `helm`, `jq`, `yq`, Docker 또는 `nerdctl` | Helm render, applied YAML 조회, image pull/digest 조회 |
| Source scanners | Semgrep, gosec | SAST 및 위험 코드 패턴 탐지 |
| Secret scanner | Gitleaks | hardcoded secret/token/account 탐지 |
| Image scanners | Trivy, Grype, Syft, Cosign 또는 Notation, Crane | CVE, SBOM, digest, 서명/무결성 확인 |
| Manifest scanners | kube-linter, conftest | Kubernetes 고위험 설정 및 RBAC policy 확인 |
| Build/script scanners | Hadolint, ShellCheck | Dockerfile 및 배포 스크립트 위험 확인 |
| Cluster access | `ClusterTarget` + Mgmt Cluster kubeconfig Secret | Biz Cluster에 적용된 YAML, RBAC, Secret 참조 방식 검수 |
| Report backend | 파일 report, Loki/Mimir/Grafana 또는 CI artifact 저장소 | finding 저장, 집계, dashboard 조회 |

## Platform Components

최종점검은 단일 scanner가 아니라 여러 보안 도구와 결과 정규화, 대시보드, 예외 검토를 묶은 assessment platform으로 본다.

| 구성요소 | 역할 | 현재 버전 사용 방식 |
|----------|------|------------------|
| SonarQube | 소스코드 정적분석, code smell, security hotspot | Source & Secrets 메뉴에 연계 |
| Semgrep/gosec | 보안 취약 패턴과 위험 코드 탐지 | SonarQube 보완 scanner로 사용 |
| Secret Scanner | 하드코딩 Secret, Token, 계정 정보 탐지 | Gitleaks 기준으로 source, values, YAML, script 검사 |
| Trivy/Grype | 컨테이너 이미지 CVE 및 Critical 취약점 탐지 | Images & Integrity 메뉴에 집계 |
| Trivy Operator | Biz Cluster에 적용된 workload image 취약점 확인 | 선택 기능. 현재 필수 M6 범위는 납품 이미지 registry digest/image tar 스캔 |
| Kubernetes Policy Engine | Helm/YAML, RBAC, Pod security policy 검증 | conftest/OPA, kube-linter, Kyverno/Gatekeeper 정책으로 검사 |
| Image Signing/Verification | 이미지 digest, 서명, 무결성 검증 | Cosign/Notation/Crane으로 승인 digest와 서명 확인 |
| SBOM Store | 이미지 digest별 SBOM 저장, 추적, 재분석 기준 제공 | Syft/Trivy SBOM 결과를 digest 기준으로 보관 |
| Finding Normalizer | scanner별 결과를 공통 finding schema로 변환 | dashboard와 최종 판정이 동일 schema를 사용 |
| Result Dashboard | 점검 결과 조회, drill-down, 예외/조치 추적 | `Final Check Dashboard` 단일 화면으로 구성 |

## Scan Profiles

사용자는 카테고리별로 관련 검사를 한 번에 실행할 수 있어야 한다.

| 프로파일 | 포함 검사 | 대표 산출물 |
|----------|----------|------------|
| Source Security Scan | SonarQube, Semgrep, gosec, Gitleaks | SAST report, secret report |
| Image Supply Chain Scan | Trivy/Grype, Syft, Cosign/Notation, Crane | CVE report, SBOM, digest/signature verification |
| Kubernetes Config Scan | Helm render, kube-linter, conftest, applied YAML inspection | Kubernetes policy report |
| RBAC & Secret Reference Scan | RBAC manifest scan, applied RBAC inspection, ServiceAccount/Secret reference inspection | RBAC risk report, secret reference report |
| Build & Deploy Scan | Hadolint, ShellCheck, deploy script inspection | Dockerfile/script report |
| Full Final Check | 위 모든 profile 실행 후 최종 판정 요약 생성 | final-check summary, normalized findings, exception candidates |

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
| Exception review file | 선택 | 승인된 예외 및 만료일 관리 |
| ClusterTarget | 필요 시 | read-only kubeconfig Secret을 참조하는 Biz Cluster 등록 정보 |

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
| Source & Secrets | 코드와 설정에 직접 노출된 위험이 있는가? | SAST finding, hardcoded secret, risky code pattern, applied YAML secret exposure | 파일/리소스 위치와 개선 가이드 확인 |
| Images & Integrity | 이미지가 안전하고 승인된 것인가? | Critical CVE, fixable High, SBOM status, digest mismatch, signature verification, base image risk | 이미지별 CVE/SBOM/digest 상세 확인 |
| Kubernetes Config & RBAC | 배포 산출물과 Biz Cluster 적용 설정의 권한이 과도한가? | privileged, hostPath, hostNetwork, capabilities, Secret references, RBAC wildcard, cluster-admin | manifest와 applied resource 비교 확인 |
| Dockerfile & Scripts | 빌드/배포 과정에 위험 요소가 있는가? | root user, floating tag, unsafe package install, unchecked shell command, secret echo | 빌드/배포 파일별 조치 항목 확인 |
| Scan Health | 분석 결과를 신뢰할 수 있는가? | scanner status, missing artifacts, unsupported target, stale DB/rule, registry pull failure | 실패 scanner 재실행 또는 누락 산출물 확인 |
| Exceptions & Remediation | 무엇을 고치거나 승인해야 하는가? | remediation list, exception candidates, approved exceptions, expired exceptions, rescan status | 예외 승인/만료/재점검 상태 확인 |

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

`final-check` 환경에서는 scanner exit code 하나에만 의존하지 않는다. normalized finding, scan health, 필수 산출물 존재 여부, DB/rule 기준일, 예외 승인 상태를 함께 평가한다.

## Next Version: Runtime Assessment

다음 버전에서는 실제 최종점검 클러스터의 런타임 상태를 포함한다.

| 후보 항목 | 설명 |
|----------|------|
| Runtime image drift | 실제 실행 image digest와 승인 digest 비교 |
| Runtime event correlation | runtime sensor 이벤트와 산출물 finding 연결 |
| Runtime behavior validation | exec, privilege escalation, suspicious API call 등 행위 기반 검증 |
| Runtime event | runtime sensor 이벤트 기반 고위험 행위 확인 |
