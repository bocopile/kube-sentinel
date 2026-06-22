# API Design

kube-sentinel backend REST API 명세. base path는 `/api/v1/`. 모든 응답은
`Content-Type: application/json`이다.

---

## 공통 규칙

### Pagination

list 엔드포인트는 offset/limit 기반 페이지네이션을 사용한다.

| 파라미터 | 기본값 | 최대값 | 설명 |
|----------|--------|--------|------|
| `offset` | `0` | — | 건너뛸 레코드 수 |
| `limit` | `20` | `500` | 반환할 최대 레코드 수 |

응답 공통 wrapper:

```json
{
  "items": [...],
  "total": 1234,
  "offset": 0,
  "limit": 20
}
```

### 에러 응답

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "scan run not found: abc-123"
  }
}
```

| HTTP Status | 코드 | 조건 |
|-------------|------|------|
| `400` | `INVALID_PARAM` | 필수 파라미터 누락, 형식 오류 |
| `404` | `NOT_FOUND` | 리소스 없음 |
| `409` | `CONFLICT` | 상태 전환 불가 (exception status machine 위반) |
| `500` | `INTERNAL_ERROR` | 서버 내부 오류 |
| `503` | `UNAVAILABLE` | k8s API, PostgreSQL 연결 불가 |

### 날짜 형식

모든 timestamp는 RFC 3339 UTC (`2026-06-18T12:34:56Z`).

---

## 엔드포인트 목록

| 메서드 | 경로 | 요약 |
|--------|------|------|
| GET | `/api/v1/overview` | 전체 요약 (카운터, 최근 scan) |
| GET | `/api/v1/cluster-targets` | ClusterTarget 목록 |
| GET | `/api/v1/cluster-targets/{name}` | ClusterTarget 단건 |
| GET | `/api/v1/scan-runs` | ScanRun 목록 |
| POST | `/api/v1/scan-runs` | ScanRun 생성 (trigger) |
| GET | `/api/v1/scan-runs/{id}` | ScanRun 단건 |
| GET | `/api/v1/scan-runs/{id}/status` | phase 폴링 (5초 주기) |
| GET | `/api/v1/scan-runs/{id}/findings` | finding 목록 (필터/페이지) |
| GET | `/api/v1/scan-runs/{id}/findings/{findingId}` | finding 단건 |
| GET | `/api/v1/scan-runs/{id}/findings/{findingId}/raw-report` | raw scanner 출력 |
| GET | `/api/v1/scan-runs/{id}/health` | scan health 기록 |
| GET | `/api/v1/scan-runs/{id}/artifacts` | artifact 목록 |
| GET | `/api/v1/scan-runs/{id}/artifacts/{artifactType}/download` | artifact 다운로드 URL |
| GET | `/api/v1/exceptions` | 예외 검토 목록 |
| PATCH | `/api/v1/exceptions/{id}` | 예외 상태 전환 |
| GET | `/api/v1/governance/summary` | 거버넌스 요약 |

---

## 엔드포인트 상세

---

### GET /api/v1/overview

dashboard Overview 화면용 집계 데이터.

**응답 `200`:**

```json
{
  "latest_scan_run": {
    "id": "scanrun-abc123",
    "assessment_name": "final-check-20260618",
    "phase": "Completed",
    "final_decision": "Fail",
    "finished_at": "2026-06-18T12:00:00Z"
  },
  "summary": {
    "critical_count": 3,
    "high_count": 12,
    "exception_required_count": 5,
    "scan_health_fail_count": 1,
    "scanner_baseline_date": "2026-06-18"
  },
  "trend": [
    {
      "scan_run_id": "scanrun-abc123",
      "final_decision": "Fail",
      "critical_count": 3,
      "high_count": 12,
      "finished_at": "2026-06-18T12:00:00Z"
    }
  ]
}
```

`trend`는 최근 10회 ScanRun의 decision + severity 카운터.

---

### GET /api/v1/cluster-targets

**쿼리 파라미터:**

| 파라미터 | 형식 | 설명 |
|----------|------|------|
| `phase` | string | `Ready`, `Degraded`, `AuthFailed` 등으로 필터 |
| `environment` | string | `dev`, `final-check`, `prod` |
| `offset` | int | |
| `limit` | int | |

**응답 `200`:**

```json
{
  "items": [
    {
      "name": "biz-a",
      "display_name": "Biz Cluster A",
      "environment": "final-check",
      "phase": "Ready",
      "kubernetes_version": "1.31.0",
      "capabilities": {
        "scannerJobs": true,
        "readOnlyInspection": true,
        "trivyOperatorReports": false,
        "hostPath": false,
        "imageAccess": true,
        "reportUpload": true
      },
      "namespace_allowlist": ["default", "kube-system"],
      "last_validated_at": "2026-06-18T11:00:00Z",
      "conditions": []
    }
  ],
  "total": 3,
  "offset": 0,
  "limit": 20
}
```

---

### GET /api/v1/cluster-targets/{name}

단건 조회. `name`은 ClusterTarget.metadata.name.

**응답 `200`:** 위 list items 단건과 동일 스키마.

**응답 `404`:** ClusterTarget 없음.

---

### GET /api/v1/scan-runs

**쿼리 파라미터:**

| 파라미터 | 형식 | 설명 |
|----------|------|------|
| `assessment_name` | string | |
| `phase` | string | `Pending`, `Running`, `Completed`, `Failed`, `Canceled` |
| `final_decision` | string | `Pass`, `Fail`, `Warning` |
| `offset` | int | |
| `limit` | int | |

**응답 `200`:**

```json
{
  "items": [
    {
      "id": "scanrun-abc123",
      "assessment_name": "final-check-20260618",
      "target_names": ["biz-a", "biz-b"],
      "phase": "Completed",
      "artifact_scan_phase": "Completed",
      "cluster_scan_phase": "Completed",
      "final_decision": "Fail",
      "summary": {
        "critical_count": 3,
        "high_count": 12,
        "exception_required_count": 5,
        "scan_health_fail_count": 1,
        "scanner_baseline_date": "2026-06-18"
      },
      "created_at": "2026-06-18T11:00:00Z",
      "started_at": "2026-06-18T11:00:05Z",
      "finished_at": "2026-06-18T12:00:00Z"
    }
  ],
  "total": 5,
  "offset": 0,
  "limit": 20
}
```

---

### POST /api/v1/scan-runs

ScanRun CR을 생성해 scan을 트리거한다. 내부적으로 k8s API에 ScanRun CR를 apply하고
결과 id를 반환한다.

**요청 body:**

```json
{
  "assessment_name": "final-check-20260618",
  "profiles": ["SourceSecurity", "ImageSupplyChain", "KubernetesConfig"]
}
```

| 필드 | 형식 | 필수 | 설명 |
|------|------|------|------|
| `assessment_name` | string | ✓ | 연결할 SecurityAssessment 이름 |
| `profiles` | string[] | | override scan profile list. 생략 시 SecurityAssessment 기본값 사용 |

**응답 `201`:**

```json
{
  "id": "scanrun-abc123",
  "assessment_name": "final-check-20260618",
  "phase": "Pending",
  "created_at": "2026-06-18T11:00:00Z"
}
```

**응답 `400`:** assessment_name 없음 또는 profiles enum 오류.

---

### GET /api/v1/scan-runs/{id}

단건. list items와 동일 스키마.

**응답 `404`:** ScanRun 없음.

---

### GET /api/v1/scan-runs/{id}/status

phase 폴링용 경량 엔드포인트. frontend는 5초마다 호출한다.

**응답 `200`:**

```json
{
  "id": "scanrun-abc123",
  "phase": "Running",
  "artifact_scan_phase": "Completed",
  "cluster_scan_phase": "Running",
  "final_decision": null
}
```

`final_decision`은 `phase = Completed`일 때만 non-null.

---

### GET /api/v1/scan-runs/{id}/findings

**쿼리 파라미터:**

| 파라미터 | 형식 | 설명 |
|----------|------|------|
| `category` | string[] (쉼표 구분) | `sast,secret,image_vulnerability,...` |
| `severity` | string[] (쉼표 구분) | `Critical,High,Medium,Low,Info` |
| `exception_status` | string[] (쉼표 구분) | `None,Required,Requested,Approved,Expired,Rejected` |
| `scan_status` | string[] (쉼표 구분) | `Pass,Fail,Error,Skipped,Unsupported` |
| `target_name` | string | 부분 일치 (`LIKE %value%`) |
| `namespace` | string | 정확 일치 |
| `scanner` | string | 정확 일치 |
| `offset` | int | |
| `limit` | int | |
| `sort` | string | `severity_desc` (기본값), `created_at_desc` |

**응답 `200`:**

```json
{
  "items": [
    {
      "id": 1,
      "finding_id": "trivy/registry.example.com/app/sha256:abc.../CVE-2024-1234/openssl",
      "scan_run_id": "scanrun-abc123",
      "raw_report_id": 42,
      "scanner": "trivy",
      "category": "image_vulnerability",
      "severity": "Critical",
      "target_type": "image",
      "target_name": "registry.example.com/app:latest",
      "namespace": "default",
      "image_digest": "sha256:abc123...",
      "rule_id": "CVE-2024-1234",
      "message": "openssl 3.0.2 has a critical vulnerability",
      "remediation": "Update openssl to >= 3.0.3",
      "exception_required": true,
      "exception_status": "Required",
      "scan_status": "Fail",
      "created_at": "2026-06-18T12:00:00Z",
      "details": {
        "package_version": "3.0.2",
        "fixed_version": "3.0.3",
        "cvss_score": 9.8,
        "references": ["https://nvd.nist.gov/vuln/detail/CVE-2024-1234"]
      }
    }
  ],
  "total": 156,
  "offset": 0,
  "limit": 20
}
```

---

### GET /api/v1/scan-runs/{id}/findings/{findingId}

단건. 위 items 스키마와 동일.

---

### GET /api/v1/scan-runs/{id}/findings/{findingId}/raw-report

해당 finding을 생성한 raw scanner 출력을 반환한다.
`format = text`이면 `data_text`를, `json/sarif`이면 `data`(JSONB)를 반환한다.

**응답 `200`:**

```json
{
  "id": 42,
  "scanner": "trivy",
  "target_name": "registry.example.com/app:latest",
  "format": "json",
  "data": { ... },
  "created_at": "2026-06-18T12:00:00Z"
}
```

`format = text`일 때:

```json
{
  "id": 43,
  "scanner": "shellcheck",
  "target_name": "scripts/deploy.sh",
  "format": "text",
  "data": null,
  "data_text": "In scripts/deploy.sh line 12:\n  rm -rf $DIR\n  ...",
  "created_at": "2026-06-18T12:00:00Z"
}
```

**응답 `404`:** finding 또는 raw_report 없음.

---

### GET /api/v1/scan-runs/{id}/health

scan_health 기록. scanner 실패, unsupported target, stale baseline 등.

**쿼리 파라미터:**

| 파라미터 | 형식 | 설명 |
|----------|------|------|
| `status` | string[] (쉼표 구분) | `OK,Warning,Fail,Skipped` |
| `scanner` | string | 정확 일치 |
| `offset` | int | |
| `limit` | int | |

**응답 `200`:**

```json
{
  "items": [
    {
      "id": 1,
      "scan_run_id": "scanrun-abc123",
      "scanner": "cosign",
      "target_name": "registry.example.com/app:latest",
      "status": "Fail",
      "reason": "registry_pull_failure",
      "message": "cosign: MANIFEST_UNKNOWN: manifest unknown",
      "details": { "exit_code": 1 },
      "created_at": "2026-06-18T12:00:00Z"
    }
  ],
  "total": 3,
  "offset": 0,
  "limit": 20
}
```

---

### GET /api/v1/scan-runs/{id}/artifacts

Artifact Store에 저장된 파일 목록.

**쿼리 파라미터:**

| 파라미터 | 형식 | 설명 |
|----------|------|------|
| `artifact_type` | string | `sbom`, `evidence_bundle`, `human_report` 등 |

**응답 `200`:**

```json
{
  "items": [
    {
      "id": 1,
      "artifact_type": "evidence_bundle",
      "path": "scanrun-abc123/reports/evidence-bundle.tar.gz",
      "checksum": "sha256:deadbeef...",
      "schema_version": "security.finding/v1",
      "scanner": null,
      "scanner_version": null,
      "db_baseline_date": "2026-06-18",
      "size_bytes": 4096000,
      "created_at": "2026-06-18T12:00:00Z"
    },
    {
      "id": 2,
      "artifact_type": "sbom",
      "path": "scanrun-abc123/sbom/registry.example.com_app_sha256_abc.spdx.json",
      "checksum": "sha256:cafebabe...",
      "schema_version": "SPDX-2.3",
      "scanner": "syft",
      "scanner_version": "1.4.0",
      "db_baseline_date": null,
      "size_bytes": 102400,
      "created_at": "2026-06-18T12:00:00Z"
    }
  ],
  "total": 6
}
```

---

### GET /api/v1/scan-runs/{id}/artifacts/{artifactType}/download

Artifact Store 파일의 presigned download URL을 반환한다.
Filesystem store의 경우 backend가 stream proxy로 동작한다.

**경로 파라미터:**

| 파라미터 | 설명 |
|----------|------|
| `artifactType` | `evidence_bundle`, `sbom`, `human_report`, `exception_review_yaml`, `remediation_advisory` |

**응답 `200`:**

```json
{
  "url": "https://storage.example.com/bucket/scanrun-abc123/reports/evidence-bundle.tar.gz?sig=...",
  "expires_at": "2026-06-18T13:00:00Z"
}
```

Filesystem store: `url`이 backend proxy 경로 (`/api/v1/artifacts/proxy/...`).

**응답 `404`:** 해당 artifact_type 없음.

---

### GET /api/v1/exceptions

예외 검토 목록. finding과 join해 finding 정보 포함.

**쿼리 파라미터:**

| 파라미터 | 형식 | 설명 |
|----------|------|------|
| `status` | string[] (쉼표 구분) | `Required,Requested,Approved,Expired,Rejected` |
| `scan_run_id` | string | |
| `offset` | int | |
| `limit` | int | |

**응답 `200`:**

```json
{
  "items": [
    {
      "id": 1,
      "finding_id": "trivy/registry.example.com/.../CVE-2024-1234/openssl",
      "scan_run_id": "scanrun-abc123",
      "status": "Requested",
      "owner": "bob@example.com",
      "reason": "패치가 3분기 이후 예정. 네트워크 격리로 위험 완화됨.",
      "expires_at": "2026-09-30T00:00:00Z",
      "approved_by": null,
      "approved_at": null,
      "created_at": "2026-06-18T13:00:00Z",
      "updated_at": "2026-06-18T13:10:00Z",
      "finding": {
        "scanner": "trivy",
        "category": "image_vulnerability",
        "severity": "Critical",
        "rule_id": "CVE-2024-1234",
        "message": "openssl 3.0.2 has a critical vulnerability",
        "target_name": "registry.example.com/app:latest"
      }
    }
  ],
  "total": 5,
  "offset": 0,
  "limit": 20
}
```

---

### PATCH /api/v1/exceptions/{id}

예외 상태를 전환한다. status machine 위반 시 `409`를 반환한다.

**허용 전환:**

| 현재 status | 요청 status | 설명 |
|-------------|-------------|------|
| `Required` | `Requested` | 예외 신청 |
| `Requested` | `Approved` | 예외 승인 |
| `Requested` | `Rejected` | 예외 거부 |
| `Approved` | `Expired` | 만료 (자동 또는 수동) |

**요청 body:**

```json
{
  "status": "Approved",
  "owner": "alice@example.com",
  "reason": "패치 일정 확인, 네트워크 격리 승인",
  "expires_at": "2026-12-31T00:00:00Z",
  "approved_by": "alice@example.com"
}
```

| 필드 | 필수 | 설명 |
|------|------|------|
| `status` | ✓ | 목표 status |
| `owner` | | 담당자 |
| `reason` | | 예외 사유 |
| `expires_at` | | 만료일. `Approved` 전환 시 권장 |
| `approved_by` | | `Approved` 전환 시 필수 |

**응답 `200`:** 변경된 exception 단건.

**응답 `409`:** 허용되지 않은 상태 전환.

---

### GET /api/v1/governance/summary

Governance 메뉴용 집계. 최근 ScanRun의 final decision 추이와 카테고리별 현황.

**응답 `200`:**

```json
{
  "latest_decision": "Fail",
  "latest_scan_run_id": "scanrun-abc123",
  "decision_trend": [
    { "scan_run_id": "scanrun-abc123", "decision": "Fail", "finished_at": "2026-06-18T12:00:00Z" },
    { "scan_run_id": "scanrun-abc001", "decision": "Fail", "finished_at": "2026-05-18T12:00:00Z" },
    { "scan_run_id": "scanrun-abc000", "decision": "Pass", "finished_at": "2026-04-18T12:00:00Z" }
  ],
  "category_summary": [
    { "category": "image_vulnerability", "critical": 3, "high": 8, "exception_approved": 1 },
    { "category": "sast",               "critical": 0, "high": 4, "exception_approved": 0 },
    { "category": "kubernetes",          "critical": 0, "high": 0, "exception_approved": 0 }
  ],
  "exception_summary": {
    "required": 5,
    "requested": 3,
    "approved": 2,
    "expired": 1
  }
}
```

---

## 타입 참조

### ScanRun.phase 상태 흐름

```
Pending → Running → Completed
                  → Failed
                  → Canceled
```

### Finding.exception_status 상태 흐름

```
None
Required → Requested → Approved → Expired
                     → Rejected
```

### scan profiles enum

| 값 | 설명 | 스캔 방식 |
|----|------|---------|
| `SourceSecurity` | 소스 보안 스캔 | Code / Artifact Scan |
| `ImageSupplyChain` | 이미지 공급망 스캔 | Code / Artifact Scan |
| `KubernetesConfig` | K8s 매니페스트 & RBAC 스캔 | Code / Artifact Scan |
| `RBACAndSecretReference` | 적용된 RBAC & Secret 참조 스캔 | Biz Cluster Scan |
| `BuildAndDeploy` | 빌드 & 배포 스캔 | Code / Artifact Scan |

---

## 구현 노트

- **라우터**: `net/http` + `chi` 또는 `gorilla/mux`
- **DB 쿼리**: `pgx/v5` 직접 또는 `sqlc` 코드 생성
- **k8s 조회**: `k8s.io/client-go` dynamic client. `POST /api/v1/scan-runs`에서 ScanRun CR apply 시 사용
- **CORS**: `frontend` origin 허용. backend middleware로 처리
- **인증**: PoC 단계에서는 bearer token 또는 IP allowlist. 문서에 추후 정책 명시
- **SSE**: `GET /api/v1/scan-runs/{id}/status` 는 현재 polling. Phase 2에서 SSE로 교체 예약
- **raw-report 접근 제한**: dashboard에서 raw scanner 출력을 직접 렌더링할 때
  Secret redaction guard를 통과한 데이터만 응답한다. backend handler에서
  `findings/{findingId}/raw-report` 응답 전 재검증한다.
