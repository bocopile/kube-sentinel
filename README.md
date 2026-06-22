# kube-sentinel

kube-sentinel은 Mgmt Cluster 기반 Kubernetes 최종점검 보안 평가 PoC입니다.
Mgmt Cluster는 이 솔루션이 설치되는 관리 클러스터이고, Biz Cluster는
솔루션이 점검하는 업무/애플리케이션 클러스터입니다. 현재 리포지터리는
계획 및 pre-skeleton 단계입니다. Go module과 구현 계약 문서는 존재하지만
Kubebuilder controller skeleton은 아직 생성되지 않았습니다.

## 현재 범위

PoC는 Biz Cluster를 등록하고 remote apply 방식으로 최종점검 scan을 실행하는
Mgmt Cluster CRD를 중심으로 설계합니다.

- `ClusterTarget`: Biz Cluster kubeconfig 참조, capability, status 관리
- `SecurityAssessment`: 평가 template와 선택 target 정의
- `ScanRun`: 한 번의 실행과 target별 결과 기록
- Trivy 기반 납품 이미지 취약점 평가. OSquery inventory는 필수 최종점검
  항목이 아니라 이후 선택 확장입니다.
- 소스, Secret, 컨테이너 이미지, SBOM/무결성, Kubernetes YAML, RBAC,
  Dockerfile, 배포 스크립트 대상 납품 산출물 보안 평가
- read-only 접근 기반 Biz Cluster 적용 설정 평가
- scanner report, normalized finding, scan health, final decision,
  exception review candidate를 위한 Report Store와 Evidence Bundle 생성
- 검토, scan health, finding, 예외 추적을 위한 Final Check Dashboard
- target preflight, artifact input manifest, scanner baseline capture,
  stable finding ID, Secret redaction, evidence bundle export,
  exception review artifact, scan health summary 같은 평가 신뢰성 보조 기능
- Kubernetes, Kubebuilder, controller-runtime, Trivy, Semgrep, Gitleaks,
  SBOM/signing 도구, Kubernetes policy scanner에 대한 미들웨어 및 scanner
  버전 기준선

Biz Cluster에는 kube-sentinel operator를 실행하지 않으며 kube-sentinel CRD도
설치하지 않습니다. Runtime event correlation과 runtime drift validation은
이후 확장으로 둡니다.

## 리포지터리 상태

현재 포함된 항목:

- `go.mod`
- `docs/` 아래 계획 및 아키텍처 문서

아직 포함되지 않은 항목:

- Kubebuilder `PROJECT`
- `cmd/`
- `api/`
- `internal/`
- `config/`

## 문서

다음 순서로 문서를 읽습니다.

1. [docs/PLAN.md](docs/PLAN.md)
2. [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md)
3. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
4. [docs/SECURITY_ASSESSMENT.md](docs/SECURITY_ASSESSMENT.md)
5. [docs/ASSESSMENT_SUPPORT_FEATURES.md](docs/ASSESSMENT_SUPPORT_FEATURES.md)
6. [docs/FRONTEND_ARCHITECTURE.md](docs/FRONTEND_ARCHITECTURE.md)
7. [docs/ROADMAP.md](docs/ROADMAP.md)
8. [docs/ORCHESTRATOR.md](docs/ORCHESTRATOR.md)
9. [docs/PROMPTS.md](docs/PROMPTS.md)
10. [docs/AI_REMEDIATION.md](docs/AI_REMEDIATION.md)

문서 정책:

- [docs/PLAN.md](docs/PLAN.md)는 상위 계획 문서로 유지합니다.
- 구현 계약은 검토와 orchestrator prompt 재사용이 쉽도록 목적별 문서에
  분리합니다.
- 사용자 검토와 범위 결정을 위해 문서는 한국어를 기본으로 합니다. 코드,
  명령어, API 필드명, CRD kind 같은 기술 식별자는 원문을 유지합니다.
- 모든 구현 milestone은 명령, Kubernetes object inspection, report artifact,
  evidence bundle, dashboard screenshot, status field로 검증 가능한 exit
  criterion을 가져야 합니다.
- 아키텍처에서 벗어나는 변경은 코드 변경 전에 관련 문서에 먼저 기록합니다.

## 다음 구현 단계

첫 구현 PR은 Kubebuilder skeleton과 핵심 API 계약을 생성해야 합니다.

- module을 `github.com/bocopile/kube-sentinel`로 초기화
- `ClusterTarget`, `SecurityAssessment`, `ScanRun` API 생성
- build 가능한 reconciler skeleton 추가
- feature registry interface 추가
- registry ordering과 unknown feature validation test 추가

코드가 생성된 이후 예상 검증:

```bash
go test ./...
go build ./...
```
