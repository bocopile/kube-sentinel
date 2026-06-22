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
| PATCH | `/api/v1/scan-runs/{id}/retry` | ScanRun workflow 부분 재실행 (trigger) |
| GET | `/api/v1/scan-runs/{id}` | ScanRun 단건 |
| GET | `/api/v1/scan-runs/{id}/status` | phase 폴링 (5초 주기) |
| GET | `/api/v1/scan-runs/{id}/findings` | finding 목록 (필터/페이지) |
| GET | `/api/v1/scan-runs/{id}/findings/{findingId}` | finding 단건 |
| GET | `/api/v1/scan-runs/{id}/findings/{findingId}/raw-report` | raw scanner 출력 |
| GET | `/api/v1/scan-runs/{id}/health` | scan health 기록 |
| GET | `/api/v1/scan-runs/{id}/artifacts` | artifact 목록 |
| GET | `/api/v1/scan-runs/{id}/artifacts/{artifactId}/download` | artifact 다운로드 URL |
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
      "kubernetes_version": "1.35.0",
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

ScanRun CR을 생성해 scan을 트리거한다. backend는 Mgmt k8s API에 ScanRun CR만 apply하고(PostgreSQL `scan_runs` 초기 row는 operator reconciler가 upsert)
결과 id를 반환한다.

**요청 body:**

```json
{
  "assessment_name": "final-check-20260618",
  "targets": ["biz-a", "biz-b"],
  "profiles": ["SourceSecurity", "ImageSupplyChain", "KubernetesConfig"]
}
```

| 필드 | 형식 | 필수 | 설명 |
|------|------|------|------|
| `assessment_name` | string | ✓ | 연결할 SecurityAssessment 이름 |
| `targets` | string[] | | override 검사 대상 ClusterTarget 이름 목록. `ScanRun.spec.targets`에 매핑되며 생략 시 SecurityAssessment.spec.targets 사용 |
| `profiles` | string[] | | override scan profile list. `ScanRun.spec.profiles`에 매핑되며 생략 시 SecurityAssessment 기본값 사용 |

**응답 `201`:**

```json
{
  "id": "scanrun-abc123",
  "assessment_name": "final-check-20260618",
  "phase": "Pending",
  "created_at": "2026-06-18T11:00:00Z"
}
```

**응답 `400`:** assessment_name 없음, targets 참조 오류 또는 profiles enum 형식 오류. (reconcile 단계에서만 발견되는 unknown profile은 HTTP 400으로 중복 거부하지 않고 ScanRun `status.features[]`에 `ConfigError`로 기록한 뒤 해당 profile만 무시한다.)

---

### PATCH /api/v1/scan-runs/{id}/retry

기존 ScanRun 안에서 선택 workflow만 재실행한다. 새 ScanRun을 만들지 않고 동일 id의 phase/finalDecision을 갱신한다. backend는 `ScanRun.spec`을 변경하지 않고 ScanRun `metadata.annotations`에 `security.kube-sentinel.io/retry-scope`(+ `retry-request-id`/`retry-requested-at`)를 patch하며, operator reconciler가 선택 phase만 재실행한 뒤 annotation을 observed 처리한다.

**요청 body:**

```json
{
  "scope": "ArtifactOnly",
  "reason": "registry pull failure resolved"
}
```

| 필드 | 형식 | 필수 | 설명 |
|------|------|------|------|
| `scope` | string (enum) | ✓ | `Full`, `ArtifactOnly`, `ClusterOnly`, `FinalDecisionOnly` |
| `reason` | string | | audit용 사유. retry-request annotation/status condition에 기록 |

scope 동작:

| `scope` | 재실행 대상 | 보존 |
|---|---|---|
| `Full` | Code / Artifact Scan + Biz Cluster Scan 전체 + 재판정 | — |
| `ArtifactOnly` | Code / Artifact Scan(Mgmt-local Job)만 + 재판정 | 기존 `cluster_scan_phase`·cluster finding |
| `ClusterOnly` | Biz Cluster Scan(read-only + optional Biz-remote Job)만 + 재판정 | 기존 `artifact_scan_phase`·artifact finding |
| `FinalDecisionOnly` | scan 재실행 없이 `finalDecision`만 재계산 | 두 scan phase finding 그대로 사용 |

**응답 `202`:**

```json
{
  "id": "scanrun-abc123",
  "scope": "ArtifactOnly",
  "phase": "Running",
  "artifact_scan_phase": "Pending",
  "cluster_scan_phase": "Completed"
}
```

**응답 `400`:** `scope` 누락 또는 enum 형식 오류.
**응답 `404`:** ScanRun 없음.
**응답 `409`:** 해당 ScanRun이 `Canceled`이거나 선택 workflow가 이미 `Running`(이전 retry-scope annotation이 아직 observed 처리되지 않음).

재실행으로 생성/갱신되는 finding은 PostgreSQL `findings`에 finding_id 기준 멱등 upsert되어 중복 집계되지 않는다(raw 정본=PostgreSQL, dedup 규칙은 DATABASE.md §findings).

---

### GET /api/v1/scan-runs/{id}

단건. list items와 동일 스키마.

**응답 `404`:** ScanRun 없음.

---

### GET /api/v1/scan-runs/{id}/status

phase 폴링용 경량 엔드포인트. frontend는 5초마다 호출한다. `scan_runs` row의 정본 write 주체는 operator reconciler이므로, `POST /api/v1/scan-runs`로 CR을 apply한 직후 reconcile이 초기 row를 만들기 전까지는 `404 NOT_FOUND`를 반환할 수 있고 frontend는 이 초기 404를 일시적 상태로 처리하고 폴링을 계속한다.

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

`final_decision`은 `phase = Completed`일 때만 non-null이며, `ScanRun.status.finalDecision.status`(`Pass`/`Fail`/`Warning`)를 평면화한 문자열(DATABASE `scan_runs.final_decision`과 동일)이다. 실패 근거 목록(`finalDecision.reasons[]`)은 이 경량 polling 응답에 포함하지 않고 `GET /api/v1/scan-runs/{id}` 및 Overview drill-down에서 제공한다.

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
| `target_cluster` | string | 정확 일치. Biz applied finding의 ClusterTarget 이름 |
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
      "target_cluster": null,
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
      "path": "reports/final-check-20260618/scanrun-abc123/evidence/evidence-bundle.tar.gz",
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
      "path": "reports/final-check-20260618/scanrun-abc123/sbom/sha256-abc123.cyclonedx.json",
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

### GET /api/v1/scan-runs/{id}/artifacts/{artifactId}/download

`artifact_index.id`로 식별되는 단일 Artifact Store 파일의 presigned download URL을 반환한다.
`artifactId`는 `GET /api/v1/scan-runs/{id}/artifacts` 응답의 `items[].id`를 사용한다. 같은
`artifact_type`(digest별 SBOM, verificationType별 integrity report 등)에 파일이 여러 개여도
이 값으로 단일 파일을 선택한다. `artifact_type`은 list/filter 용도이며 download 식별자로
쓰지 않는다. path convention은 [ARCHITECTURE.md](./ARCHITECTURE.md) §Artifact path convention을 따른다.
Filesystem store의 경우 backend가 stream proxy로 동작한다.

**경로 파라미터:**

| 파라미터 | 설명 |
|----------|------|
| `artifactId` | `GET /api/v1/scan-runs/{id}/artifacts` 응답의 `id` (동일 `artifact_type` 다중 파일 구분용) |

**응답 `200`:**

```json
{
  "url": "https://storage.example.com/bucket/reports/final-check-20260618/scanrun-abc123/evidence/evidence-bundle.tar.gz?sig=...",
  "expires_at": "2026-06-18T13:00:00Z"
}
```

Filesystem store: `url`이 backend proxy 경로 (`/api/v1/artifacts/proxy/...`).

**응답 `404`:** 해당 `artifactId` artifact 없음.

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
| `Rejected` | `Required` | 재스캔에서 동일 finding 재보고 시 재평가(재신청 가능) |
| `Expired` | `Required` | 재스캔/만료 후 동일 finding 재보고 시 재평가(재신청 가능) |

`Rejected`/`Expired` → `Required` 전환은 재스캔 carry-over 재평가([DATABASE.md](./DATABASE.md) §exception_reviews 재스캔 carry-over 규칙)에 따라 operator가 새 ScanRun row를 만들 때 수행하며 사용자 PATCH로도 허용한다. 그 외 전환은 `409 CONFLICT`.

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

재스캔으로 동일 `finding_id`가 새 ScanRun에 등장하면, 유효한 `Approved`(미만료)는 새 row로 carry-over하고 `Expired`/`Rejected`는 `Required`로 재평가한다(상세는 DATABASE.md §exception_reviews 재스캔 carry-over 규칙).

### scan profiles enum

| 값 | 설명 | 스캔 방식 |
|----|------|---------|
| `SourceSecurity` | 소스 보안 스캔 | Code / Artifact Scan |
| `ImageSupplyChain` | 이미지 공급망 스캔 | Code / Artifact Scan |
| `KubernetesConfig` | K8s 매니페스트 & RBAC 스캔 | Code / Artifact Scan |
| `RBACAndSecretReference` | 적용된 RBAC & Secret 참조 스캔 | Biz Cluster Scan |
| `BuildAndDeploy` | 빌드 & 배포 스캔 | Code / Artifact Scan |

`profiles[]`는 base feature set을 결정하고 `features[]`가 enable/disable·config override를 적용한다. profile→registry feature ID 정본은 [ARCHITECTURE.md](./ARCHITECTURE.md) §Profile / features → registry feature ID 매핑이다. backend는 enum/참조 값 검증만 하고, 실제 병합·resolve와 unknown profile의 `ConfigError` 처리는 operator가 deterministic하게 수행한다.

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
