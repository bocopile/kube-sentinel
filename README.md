# kube-sentinel

`kube-sentinel`은 Kubernetes 환경의 납품 전 보안 최종점검을 자동화하기 위한 Mgmt Cluster 기반 PoC 프로젝트입니다.

관리 클러스터(Mgmt Cluster)에 단일 operator를 설치하고, 점검 대상 업무 클러스터(Biz Cluster)는 별도 agent나 kube-sentinel CRD 없이
remote apply/read-only 방식으로 검사합니다.
목표는 소스, Secret, 컨테이너 이미지, SBOM/무결성, Kubernetes manifest, RBAC, Dockerfile, 배포 스크립트, 실제 적용된 클러스터 설정을
한 번의 `ScanRun`으로 평가하고, 결과를 report/evidence/dashboard로 남기는 것입니다.

## 핵심 아이디어

- Mgmt Cluster 중심의 단일 operator 구조
- `ClusterTarget`, `SecurityAssessment`, `ScanRun` CRD 기반 실행 모델
- 검사 기능을 Reconciler에 하드코딩하지 않는 Feature-as-Plugin 아키텍처
- feature priority 기반 deterministic orchestration
- scanner output, normalized finding, scan health, final decision, evidence bundle 저장
- Artifact Store backend를 filesystem, S3-compatible, MinIO, SeaweedFS 등으로 교체 가능한 구조
- Biz Cluster에는 kube-sentinel operator/CRD를 설치하지 않는 경량 점검 방식

## 현재 상태

현재 리포지터리는 계획 및 pre-skeleton 단계입니다.

포함된 항목:

- `go.mod`: 임시 root placeholder
- `docs/`: 요구사항, 아키텍처, 모듈 구조, API, roadmap 문서

아직 생성되지 않은 항목:

- `operator/`
- `backend/`
- `frontend/`
- Kubebuilder controller skeleton
- 실제 Feature registry 및 plugin 구현체

첫 구현 블록에서는 `operator/` Go module과 Kubebuilder skeleton을 만들고, `ClusterTarget`, `SecurityAssessment`,
`ScanRun` API 및 Feature registry interface를 추가할 예정입니다.

## 문서

프로젝트를 이해하려면 아래 문서를 순서대로 보면 됩니다.

1. [docs/PLAN.md](docs/PLAN.md)
2. [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md)
3. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
4. [docs/MODULES.md](docs/MODULES.md)
5. [docs/DATABASE.md](docs/DATABASE.md)
6. [docs/API_DESIGN.md](docs/API_DESIGN.md)
7. [docs/ROADMAP.md](docs/ROADMAP.md)

## 예상 모듈 구조

```text
kube-sentinel/
├── operator/    # Mgmt Cluster operator, CRD, Feature plugin orchestration
├── backend/     # Report/API server, metadata query, artifact read
├── frontend/    # Final Check Dashboard
└── docs/        # 설계 및 구현 계약 문서
```

## 검증 목표

구현 이후 기본 검증은 다음 흐름을 기준으로 합니다.

```bash
cd operator
go test ./...
go build ./...
```
