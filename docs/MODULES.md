# 모듈 구조

kube-sentinel은 모노레포 안에 3개의 독립 모듈로 구성된다. 각 모듈은 자체 빌드
단위와 배포 단위를 가진다.

## 모듈 개요

| 모듈 | 경로 | 언어 | Go module path | 역할 |
|------|------|------|----------------|------|
| operator | `operator/` | Go | `github.com/bocopile/kube-sentinel/operator` | Mgmt Cluster operator, CRD, Feature plugin, remote apply, ArtifactStore write |
| backend | `backend/` | Go | `github.com/bocopile/kube-sentinel/backend` | REST API 서버, PostgreSQL query, k8s CR 조회, ArtifactStore read |
| frontend | `frontend/` | TypeScript | `kube-sentinel-frontend` (npm) | Final Check Dashboard, React SPA |

```
kube-sentinel/
├── operator/
├── backend/
├── frontend/
└── docs/
```

---

## operator 모듈

**역할**: Mgmt Cluster에 설치되는 단일 operator. CRD 정의, Reconciler, Feature
orchestrator, remote apply, Finding normalization, Report Artifact Store write.

**Go module**: `github.com/bocopile/kube-sentinel/operator`

```
operator/
├── go.mod
├── go.sum
├── cmd/
│   └── main.go                          # operator 진입점, Feature import
│
├── api/
│   └── v1alpha1/
│       ├── clustertarget_types.go
│       ├── securityassessment_types.go
│       ├── scanrun_types.go
│       └── zz_generated.deepcopy.go
│
├── internal/
│   ├── controller/
│   │   ├── clustertarget_controller.go
│   │   ├── securityassessment_controller.go
│   │   └── scanrun_controller.go
│   │
│   ├── feature/
│   │   ├── feature.go                   # Feature interface
│   │   ├── registry.go                  # priority registry
│   │   ├── store.go                     # DesiredStateStore
│   │   ├── types.go                     # FeatureCondition, scan resource config
│   │   ├── target_preflight/feature.go  # Priority 10
│   │   ├── bootstrap/feature.go         # Priority 20
│   │   ├── source_security/feature.go   # Priority 50
│   │   ├── secret_scan/feature.go       # Priority 50
│   │   ├── image_vulnerability/feature.go # Priority 100
│   │   ├── image_integrity/feature.go   # Priority 100
│   │   ├── sbom/feature.go              # Priority 100
│   │   ├── kubernetes_manifest/feature.go # Priority 150
│   │   ├── rbac_review/feature.go       # Priority 150
│   │   ├── applied_cluster_config/feature.go # Priority 200
│   │   ├── secret_reference/feature.go  # Priority 200
│   │   ├── trivy_operator_reports/feature.go # Priority 200
│   │   ├── remediation_enrichment/feature.go # Priority 250 (선택, AI)
│   │   └── report_export/feature.go     # Priority 300
│   │
│   ├── target/
│   │   ├── kubeconfig.go                # Mgmt Secret → k8s client
│   │   ├── remote_apply.go              # Biz Cluster SSA
│   │   ├── discovery.go                 # RBAC/capability 검사
│   │   └── bootstrap.go                 # 허용된 namespace/RBAC/scanner resource 생성
│   │
│   ├── normalizer/
│   │   ├── finding_id.go                # stable finding ID / dedup
│   │   ├── schema_validator.go          # security.finding/v1 검증
│   │   └── secret_redaction.go          # Secret raw value 차단
│   │
│   ├── report/
│   │   ├── store.go                     # Metadata DB + Artifact Store 연동
│   │   ├── evidence_bundle.go           # Evidence Bundle export
│   │   └── exception_review.go          # Exception review artifact writer
│   │
│   └── artifactstore/
│       ├── store.go                     # ArtifactStore interface (write + read)
│       ├── filesystem/
│       ├── s3compatible/
│       ├── seaweedfs/
│       └── pvc/
│
├── config/
│   ├── crd/bases/
│   └── samples/
│       ├── clustertarget_dev.yaml
│       ├── securityassessment_final_check.yaml
│       └── scanrun_sample.yaml
│
└── security/
    ├── scanners/
    ├── scripts/
    └── inputs/
```

**빌드 및 테스트:**

```bash
cd operator/
go test ./...
go build ./...
```

**Kubebuilder 초기화 (최초 1회):**

```bash
cd operator/
kubebuilder init --domain kube-sentinel.io --repo github.com/bocopile/kube-sentinel/operator
kubebuilder create api --group security --version v1alpha1 --kind ClusterTarget --namespaced=false
kubebuilder create api --group security --version v1alpha1 --kind SecurityAssessment --namespaced=false
kubebuilder create api --group security --version v1alpha1 --kind ScanRun --namespaced=false
```

---

## backend 모듈

**역할**: REST API 서버. PostgreSQL 메타데이터 조회, k8s API 직접 조회(ClusterTarget ·
ScanRun CR), Report Artifact Store read-only 조회, dashboard 응답 생성.

**Go module**: `github.com/bocopile/kube-sentinel/backend`

**k8s 접근 방식**: `k8s.io/client-go` dynamic client로 Mgmt Cluster API를 직접
조회한다. CRD Go 타입은 operator 모듈과 공유하지 않고, backend가 필요한 필드만
자체 경량 struct로 정의한다. Secret raw value는 조회하지 않는다.

**ArtifactStore**: operator의 write 인터페이스 전체가 아닌 read-only 서브셋만
자체 정의한다.

```go
// backend/internal/artifactstore/store.go
type ArtifactReader interface {
    GetArtifact(ctx context.Context, ref ArtifactRef) (io.ReadCloser, error)
    ListArtifacts(ctx context.Context, prefix string) ([]ArtifactRef, error)
    GenerateDownloadURL(ctx context.Context, ref ArtifactRef) (string, error)
}
```

```
backend/
├── go.mod
├── go.sum
├── cmd/
│   └── main.go                          # API 서버 진입점
│
└── internal/
    ├── handler/
    │   ├── overview.go                  # GET /api/v1/overview
    │   ├── cluster_targets.go           # GET /api/v1/cluster-targets
    │   ├── scan_runs.go                 # GET/POST /api/v1/scan-runs
    │   ├── findings.go                  # GET /api/v1/scan-runs/{id}/findings
    │   ├── scan_health.go               # GET /api/v1/scan-runs/{id}/health
    │   ├── exceptions.go                # GET/PATCH /api/v1/exceptions
    │   ├── artifacts.go                 # GET /api/v1/scan-runs/{id}/artifacts
    │   └── governance.go                # GET /api/v1/governance/summary
    │
    ├── db/
    │   ├── postgres.go                  # connection pool, migration
    │   ├── scan_runs.go
    │   ├── findings.go
    │   ├── exceptions.go
    │   ├── scan_health.go
    │   └── artifact_index.go
    │
    ├── k8s/
    │   ├── client.go                    # dynamic client 초기화
    │   ├── types.go                     # 경량 CR struct (필드 subset)
    │   ├── cluster_targets.go           # ClusterTarget list/get
    │   └── scan_runs.go                 # ScanRun create/get
    │
    ├── artifactstore/
    │   ├── store.go                     # ArtifactReader interface
    │   ├── filesystem/
    │   └── s3compatible/
    │
    └── middleware/
        ├── cors.go
        └── logging.go
```

**빌드 및 테스트:**

```bash
cd backend/
go test ./...
go build ./...
```

---

## frontend 모듈

**역할**: React SPA. Final Check Dashboard (Overview, Targets, Assessments, Findings,
Reports, Governance). backend REST API만 호출한다. k8s API와 직접 통신하지 않는다.

**npm package**: `kube-sentinel-frontend`

```
frontend/
├── package.json
├── tsconfig.json
├── next.config.ts                       # 또는 vite.config.ts
│
└── src/
    ├── app/                             # Next.js App Router 또는 React Router
    │   ├── layout.tsx
    │   ├── overview/page.tsx
    │   ├── targets/page.tsx
    │   ├── assessments/page.tsx
    │   ├── findings/page.tsx
    │   ├── reports/page.tsx
    │   └── governance/page.tsx
    │
    ├── components/
    │   ├── filter-bar/                  # 공통 필터 (severity, category, ...)
    │   ├── finding-table/               # finding 목록 + drill-down drawer
    │   ├── scan-status/                 # workflow phase 표시
    │   ├── exception-drawer/            # exception review CRUD
    │   └── evidence-download/           # artifact download 버튼
    │
    ├── api/
    │   ├── client.ts                    # fetch wrapper, base URL, error handling
    │   ├── overview.ts
    │   ├── cluster-targets.ts
    │   ├── scan-runs.ts
    │   ├── findings.ts
    │   ├── exceptions.ts
    │   └── artifacts.ts
    │
    └── types/
        ├── finding.ts
        ├── scan-run.ts
        ├── cluster-target.ts
        └── exception.ts
```

**빌드 및 테스트:**

```bash
cd frontend/
npm install
npm run dev       # 개발 서버
npm run build     # 프로덕션 빌드
npm run test
```

---

## 모듈 간 경계

```
┌─────────────────────────────────────────────────────────────┐
│                      Mgmt Cluster                           │
│                                                             │
│  ┌─────────────┐   writes   ┌──────────────────────────┐   │
│  │  operator   │──────────▶│  PostgreSQL (metadata)   │   │
│  │             │            │  Artifact Store (files)  │   │
│  │  CRD types  │            └──────────┬───────────────┘   │
│  │  reconciler │                       │ reads             │
│  │  feature    │            ┌──────────▼───────────────┐   │
│  │  normalizer │            │  backend (API server)    │   │
│  └──────┬──────┘            │                          │   │
│         │ remote apply      │  k8s dynamic client      │   │
│         │                   │  REST /api/v1/           │   │
└─────────┼───────────────────┴──────────┬───────────────┘   │
          │                              │ HTTP               │
          ▼                              ▼                    │
     Biz Cluster                   ┌──────────┐              │
     scanner Jobs                  │ frontend │              │
                                   │ React SPA│              │
                                   └──────────┘              │
```

| 통신 방향 | 방식 | 비고 |
|----------|------|------|
| operator → PostgreSQL | `lib/pq` 또는 `pgx` | finding, scan_run, artifact_index write |
| operator → Artifact Store | ArtifactStore interface | raw report, JSONL, evidence bundle write |
| backend → PostgreSQL | `pgx` | read-only query |
| backend → Artifact Store | ArtifactReader interface | GetArtifact, GenerateDownloadURL |
| backend → Mgmt k8s API | dynamic client | ClusterTarget, ScanRun CR get/list/create |
| frontend → backend | HTTP REST `/api/v1/` | polling 5s for scan status |
| operator → Biz k8s API | target kubeconfig | remote apply (scanner Jobs, RBAC) |

---

## 관련 문서

- [DATABASE.md](./DATABASE.md) — 전체 PostgreSQL 테이블 DDL 및 인덱스
- [API_DESIGN.md](./API_DESIGN.md) — REST API 엔드포인트 명세

---

## 설계 결정

| 결정 | 선택 | 이유 |
|------|------|------|
| CRD 타입 공유 | 공유 안 함. backend는 dynamic client + 경량 struct | 모듈 간 Go 의존성 없앰. backend가 operator를 import하면 controller-runtime 등 불필요한 의존성 포함 |
| ArtifactStore | operator full interface / backend read-only 서브셋 각자 정의 | PoC 범위에서 인터페이스가 작아 중복 허용. 추후 shared module로 추출 가능 |
| frontend ↔ k8s | 직접 통신 없음. backend 경유 | RBAC 단순화, kubeconfig frontend 노출 금지 |
| scan 진행 상황 | 5초 polling (`GET /api/v1/scan-runs/{id}/status`) | SSE는 Phase 2 확장 경로로 예약 |
| ClusterTarget 미러 | PostgreSQL `cluster_targets` 테이블에 캐시, k8s watch sync | dashboard list 응답 속도 보장 |

---

## 빌드 전체 실행

```bash
# root에서 전체 모듈 검증
(cd operator && go build ./... && go test ./...)
(cd backend  && go build ./... && go test ./...)
(cd frontend && npm run build)
```
