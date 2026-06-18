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
- 개선 권고, 예외 승인 후보, 만료 예외, 재점검 상태를 추적한다.

## Product Shape

현재 버전은 Grafana 기반 `Final Check Dashboard`를 기본으로 한다. 별도 Web UI는 다음 단계에서 `assessment-api`를 붙여 제품형 화면으로 확장한다.

| 단계 | UI 형태 | 목적 |
|------|--------|------|
| MVP | Grafana dashboard | 빠른 PoC 검증, finding 집계, 증적 조회 |
| Product UI | React/Next.js 또는 동등한 SPA | 검사 실행, finding 상세, 예외 workflow, 최종 판정 관리 |

## Information Architecture

대시보드는 하나의 제품 화면으로 두고, 내부 메뉴를 검사 영역 기준으로 나눈다.

```text
Final Check Dashboard
├── Overview
├── Clusters
├── Run Scan
├── Findings
├── Code / Artifacts
├── Biz Cluster
├── Source & Secrets
├── Images & Integrity
├── Kubernetes Config & RBAC
├── Dockerfile & Scripts
├── Scan Health
└── Exceptions & Remediation
```

## Menus

| 메뉴 | 목적 | 주요 기능 |
|------|------|----------|
| Overview | 납품 가능 여부 요약 | Pass/Fail, Critical/High, failed scans, missing artifacts, exception-required count |
| Clusters | Biz Cluster 등록 상태 확인 | ClusterTarget list, connection phase, capability status, namespace allowlist, last validation time |
| Run Scan | 검사 그룹별 실행 | Code / Artifact Scan, Biz Cluster Scan, Full Final Check, preflight status |
| Findings | 전체 finding 통합 목록 | severity/category/scanner/target/status 필터, finding 상세 drawer |
| Code / Artifacts | 납품 산출물 검사 결과 | source, image, SBOM, digest, manifest, Dockerfile, script scan status |
| Biz Cluster | Biz Cluster 적용 상태 검사 결과 | applied workload, RBAC, ServiceAccount, Secret reference, Service/Ingress scan status |
| Source & Secrets | 코드와 민감정보 위험 분석 | SonarQube/Semgrep/gosec/Gitleaks 결과, 파일 위치, rule, remediation |
| Images & Integrity | 이미지와 공급망 위험 분석 | CVE, SBOM, digest mismatch, signature verification, base image risk |
| Kubernetes Config & RBAC | YAML, applied config, 권한 분석 | privileged, hostPath, hostNetwork, Secret references, RBAC wildcard, cluster-admin |
| Dockerfile & Scripts | 빌드/배포 위험 분석 | root user, floating tag, unsafe package install, unchecked shell, secret echo |
| Scan Health | 분석 신뢰도 확인 | scanner failure, unsupported target, missing artifact, stale DB/rule, registry pull failure |
| Exceptions & Remediation | 조치와 예외 추적 | remediation owner, due date, exception approval, expired exception, rescan result |

## Scan Profiles

`Run Scan` 메뉴는 도구명이 아니라 검사 그룹과 검사 영역 기준으로 실행 버튼을
제공한다.

| 검사 그룹 | 프로파일 | 실행 도구/방식 | 결과 메뉴 |
|----------|----------|---------------|----------|
| Code / Artifact Scan | Source Security Scan | SonarQube, Semgrep, gosec, Gitleaks | Code / Artifacts, Source & Secrets |
| Code / Artifact Scan | Image Supply Chain Scan | Trivy/Grype, Syft, Cosign/Notation, Crane | Code / Artifacts, Images & Integrity |
| Code / Artifact Scan | Manifest & RBAC Manifest Scan | Helm render, kube-linter, conftest, RBAC manifest policy | Code / Artifacts, Kubernetes Config & RBAC |
| Code / Artifact Scan | Build & Deploy Scan | Hadolint, ShellCheck | Code / Artifacts, Dockerfile & Scripts |
| Biz Cluster Scan | Applied Workload Config Scan | read-only applied workload inspection | Biz Cluster, Kubernetes Config & RBAC |
| Biz Cluster Scan | Applied RBAC Scan | read-only applied RBAC inspection | Biz Cluster, Kubernetes Config & RBAC |
| Biz Cluster Scan | Secret Reference Scan | env/envFrom/volume/ServiceAccount token reference inspection | Biz Cluster, Source & Secrets |
| Biz Cluster Scan | Exposure Scan | Service/Ingress exposure inspection | Biz Cluster, Kubernetes Config & RBAC |
| Full Final Check | Full Final Check | Code / Artifact Scan 이후 Biz Cluster preflight와 Biz Cluster Scan 실행 | Overview, Findings, Scan Health, Exceptions & Remediation |

Biz Cluster Scan 실행 버튼은 선택한 `ClusterTarget`이 `Ready`가 아니면
비활성화한다. 비활성 사유는 kubeconfig Secret 누락, API 연결 실패, RBAC
denied, namespace allowlist 미설정, optional CRD/bootstrap capability 부족 중
하나로 표시한다.

## Common Filters

모든 메뉴는 같은 필터 모델을 공유한다.

| 필터 | 설명 |
|------|------|
| Environment | `dev`, `final-check` |
| Target version/build | 납품 대상 버전 또는 build ID |
| Scan run ID | 검사 실행 단위 |
| Scan group | Code / Artifact, Biz Cluster, Full Final Check |
| Namespace | Biz Cluster 적용 설정 검수 범위 |
| Image | 이미지 repository, tag, digest |
| Severity | Critical, High, Medium, Low, Info |
| Category | `sast`, `secret`, `image_vulnerability`, `sbom`, `integrity`, `kubernetes`, `rbac`, `dockerfile`, `script`, `scan_health` |
| Scanner | SonarQube, Semgrep, gosec, Gitleaks, Trivy, Grype, Syft, Cosign, kube-linter, conftest, Hadolint, ShellCheck |
| Scan status | Pass, Fail, Error, Skipped, Unsupported |
| Exception status | None, Required, Requested, Approved, Expired, Rejected |

## Cluster List

`Clusters` 메뉴는 Mgmt Cluster의 `ClusterTarget` CR과 status를 조회해
Biz Cluster 목록을 표시한다. kubeconfig Secret data는 어떤 UI/API에서도
노출하지 않는다.

| 컬럼 | 출처 | 설명 |
|------|------|------|
| Name | `metadata.name` | 내부 Biz Cluster ID |
| Display name | `spec.displayName` | 사용자 표시 이름 |
| Environment | `spec.environment` | dev, final-check 등 |
| Phase | `status.phase` | Ready, Degraded, AuthFailed, Unreachable, PermissionDenied |
| Kubernetes version | `status.kubernetesVersion` | Biz Cluster discovery 결과 |
| Last validated | `status.lastValidatedAt` | 마지막 연결/RBAC 검증 시각 |
| Capabilities | `status.capabilities` | privileged, hostPath, BTF, egress, image pull 가능 여부 |
| Namespace allowlist | `spec.namespaceAllowlist` | 검사 허용 namespace |

Cluster detail 화면은 connectivity error, RBAC denied, egress blocked, image
pull failure 같은 상태를 remediation과 함께 보여준다. Credential rotation은
상태와 마지막 회전 시각만 표시하고 Secret 값은 표시하지 않는다.

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

Secret 값은 상세 화면에 표시하지 않는다. Secret finding은 파일 위치, 키 이름, 탐지 rule, confidence, remediation만 표시한다.

## Data Flow

```text
Scanner / Cluster Inspector
  -> Raw reports
  -> Finding Normalizer
  -> Normalized findings
  -> Final decision summary
  -> Report artifact + LGTM metrics/logs
  -> Final Check Dashboard
```

## Workflow View

`Run Scan`과 `Overview`는 검사 결과를 도구별이 아니라 workflow별로 먼저
보여준다.

| Workflow | UI 상태 | 대표 실패 원인 | 재실행 단위 |
|----------|---------|----------------|-------------|
| Code / Artifact Workflow | `artifactScan.phase` | missing artifact, scanner error, stale DB/rule, registry pull failure, digest mismatch | Code / Artifact Scan만 재실행 |
| Biz Cluster Workflow | `clusterScan.phase` | kubeconfig missing, API unreachable, RBAC denied, namespace allowlist violation, optional CRD/bootstrap unavailable | Biz Cluster Scan만 재실행 |
| Final Decision Workflow | `finalDecision.status` | Critical finding, Secret exposure, unapproved exception, expired exception | 전체 또는 실패 workflow 재실행 후 재판정 |

워크플로우 상세 화면은 같은 `ScanRun` 안에서 단계별 timestamps,
conditions, raw report 링크, normalized finding 링크를 보여준다. 사용자는
Code / Artifact 실패와 Biz Cluster 실패를 같은 실패로 보지 않고, 어느
절차를 다시 실행해야 하는지 즉시 판단할 수 있어야 한다.

제품형 UI로 확장할 때는 다음 API 계층을 둔다.

| API | 역할 |
|-----|------|
| `assessment-api` | scan run, finding, exception, 최종 판정 결과 조회 |
| `scanner-runner` | 검사 profile 실행 요청과 상태 추적 |
| `artifact-store` | raw report, SBOM, normalized finding, dashboard snapshot 저장 |
| `exception-store` | 예외 승인 이력, 만료일, 승인자, 사유 저장 |

## State Model

| 객체 | 설명 |
|------|------|
| ScanRun | 한 번의 검사 실행 단위. profile, target version, status, timestamps를 가진다. |
| ScanPhase | `artifactScan`, `clusterScan` 등 검사 절차별 phase, timestamps, conditions를 가진다. |
| Finding | normalized finding. category, scanner, target, severity, status를 가진다. |
| FinalDecision | scan run별 최종 Pass/Fail/Warning 판정과 주요 실패 원인 목록. |
| ExceptionReview | finding별 예외 승인 상태, 승인자, 사유, 만료일. |
| Artifact | raw report, SBOM, digest verification report, normalized finding report. |

## UI Guardrails

- 최종 판정 실패 원인을 Overview에서 바로 보여준다.
- 스캔 실패와 필수 산출물 누락은 취약점 없음으로 처리하지 않고 Scan Health에서 Fail로 표시한다.
- Secret 원문 값은 UI, log, artifact 어디에도 표시하지 않는다.
- 카테고리 메뉴는 상세 분석용이고, 모든 finding은 Findings 메뉴에서 통합 조회할 수 있어야 한다.
- 예외 승인은 finding을 숨기지 않는다. 상태만 `Approved`로 바꾸고 만료일을 표시한다.
