# Database Schema

kube-sentinel은 PostgreSQL 18.x를 메타데이터 저장소로 사용한다.
모든 dashboard/API 쿼리는 이 DB에서 수행한다.
raw scanner output도 `raw_reports` 테이블의 JSONB 컬럼에 저장한다.
SBOM, evidence bundle, human report, scanner baseline은 Artifact Store(파일) 에 저장하고 `artifact_index`
테이블로 참조한다.

PostgreSQL을 선택한 이유:

- `JSONB + GIN 인덱스`: raw scanner output 내부 필드를 인덱스로 조회 가능
- `TOAST 자동 오프로드`: 대형 JSONB가 row scan 속도에 영향을 주지 않음
- `TEXT[] 네이티브 배열`: `target_names`, `namespace_allowlist` 별도 테이블 불필요
- `LISTEN/NOTIFY`: 추후 scan 상태 push 확장 경로
- MariaDB는 GIN 인덱스와 ARRAY 타입이 없어 이 설계에 부적합

---

## 테이블 목록

| 테이블 | 역할 |
|--------|------|
| `scan_runs` | ScanRun 실행 단위, phase, summary 집계 |
| `raw_reports` | raw scanner 출력 (JSONB / TEXT) |
| `findings` | normalized finding (stable ID, 필터 인덱스) |
| `scan_health` | scanner 실패, unsupported target, stale baseline 기록 |
| `exception_reviews` | finding 예외 승인 이력과 status machine |
| `artifact_index` | Artifact Store 파일 참조 (SBOM, evidence bundle 등) |
| `cluster_targets` | ClusterTarget CR k8s 미러 (status 캐시) |

---

## 테이블 상세 명세

### scan_runs

ScanRun 하나의 실행 단위.
`summary` JSONB에 집계 카운터를 선계산해 Overview API 응답을 빠르게 한다.
`scan_runs` row 생성/갱신의 정본 write 주체는 operator `ScanRun` reconciler다.
backend `POST /api/v1/scan-runs`는 Mgmt k8s API에 ScanRun CR만 apply하고 PostgreSQL `scan_runs`를 insert하지
않는다.
operator는 reconcile 시작 시 `Pending` 초기 row를 upsert하고 이후 phase/final_decision/summary를 갱신한다.

```sql
CREATE TABLE scan_runs (
    id                   VARCHAR(255) PRIMARY KEY,
    assessment_name      VARCHAR(255) NOT NULL,
    target_names         TEXT[]       NOT NULL,         -- 검사 대상 ClusterTarget 이름 목록
    phase                VARCHAR(50)  NOT NULL,          -- Pending | Running | Completed | Failed | Canceled
    artifact_scan_phase  VARCHAR(50),                   -- Pending | Running | Completed | Failed | Skipped
    cluster_scan_phase   VARCHAR(50),                   -- Pending | Running | Completed | Failed | Skipped
    final_decision       VARCHAR(50),                   -- Pass | Fail | Warning (= summary.final_decision.status projection, 필터/인덱스용)
    summary              JSONB        NOT NULL DEFAULT '{}',
    -- {
    --   "critical_count": 0,
    --   "high_count": 0,
    --   "exception_required_count": 0,
    --   "scan_health_fail_count": 0,
    --   "scanner_baseline_date": "2026-06-18",
    --   "final_decision": {        -- security.finalDecision/v1 object snapshot
    --     "status": "Fail",
    --     "reasons": [ { "code": "critical_finding", "severity": "Critical", "count": 3 } ],
    --     "decided_at": "2026-06-18T12:00:00Z"
    --   }
    -- }
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    started_at           TIMESTAMPTZ,
    finished_at          TIMESTAMPTZ
);

CREATE INDEX idx_scan_runs_phase       ON scan_runs(phase);
CREATE INDEX idx_scan_runs_decision    ON scan_runs(final_decision);
CREATE INDEX idx_scan_runs_created_at  ON scan_runs(created_at DESC);
CREATE INDEX idx_scan_runs_assessment  ON scan_runs(assessment_name);
```

---

### raw_reports

scanner가 생성한 원본 출력.
JSON/SARIF는 `data JSONB`에, 비구조화 텍스트는 `data_text TEXT`에 저장한다.
GIN 인덱스로 JSONB 내부 필드를 직접 조회할 수 있다.

```sql
CREATE TABLE raw_reports (
    id           BIGSERIAL    PRIMARY KEY,
    scan_run_id  VARCHAR(255) NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
    scanner      VARCHAR(100) NOT NULL,
    -- trivy | grype | semgrep | gosec | gitleaks | kube-linter | conftest
    -- hadolint | shellcheck | cosign | notation | crane | syft
    target_name  TEXT,
    -- image ref (registry.example.com/app:sha256:...), file path, k8s resource name
    format       VARCHAR(20)  NOT NULL CHECK (format IN ('json', 'sarif', 'text')),
    data         JSONB,                     -- format = json | sarif
    data_text    TEXT,                      -- format = text (비구조화 fallback)
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_raw_reports_data CHECK (
        (format = 'text' AND data IS NULL AND data_text IS NOT NULL) OR
        (format IN ('json', 'sarif') AND data IS NOT NULL AND data_text IS NULL)
    )
);

CREATE INDEX idx_raw_reports_scanrun   ON raw_reports(scan_run_id, scanner);
CREATE INDEX idx_raw_reports_target    ON raw_reports(target_name);
CREATE INDEX idx_raw_reports_data_gin  ON raw_reports USING GIN (data)
    WHERE data IS NOT NULL;
```

scanner별 `format` 값:

| Scanner | 권장 format | 비고 |
|---------|------------|------|
| Trivy | `json` | `trivy image --format json` |
| Grype | `json` | `grype --output json` |
| Semgrep | `sarif` | `semgrep --sarif` |
| gosec | `sarif` | `gosec -fmt sarif` |
| Gitleaks | `json` | `gitleaks detect --report-format json` |
| kube-linter | `json` | `kube-linter lint --format json` |
| conftest | `json` | `conftest test --output json` |
| Hadolint | `json` | `hadolint --format json` |
| ShellCheck | `json` | `shellcheck --format json` |
| Cosign/Notation | `json` | verification result JSON |
| Crane | `json` | digest metadata JSON |

---

### findings

normalized finding.
`finding_id`는 scanner + target + rule 조합의 deterministic stable ID이며 생성 규칙은 아래 표가 단일 정본이다(category별
구성 요소가 다르다).
같은 report를 2회 처리하거나 같은 workflow를 재실행해도 동일 finding_id로 멱등 upsert되어 `UNIQUE (finding_id, scan_run_id)`
기준 중복 집계가 발생하지 않는다(M5 dedup).
`exception_status`는 `exception_reviews`와 sync해 join 없이 필터 가능하다.

```sql
CREATE TABLE findings (
    id                 BIGSERIAL    PRIMARY KEY,
    finding_id         VARCHAR(512) NOT NULL,          -- stable deterministic ID
    scan_run_id        VARCHAR(255) NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
    raw_report_id      BIGINT       REFERENCES raw_reports(id),
    -- raw scanner 출력과의 연결. 재정규화 시 동일 raw_report_id 재사용
    scanner            VARCHAR(100) NOT NULL,
    category           VARCHAR(100) NOT NULL,
    -- sast | secret | image_vulnerability | sbom | integrity | kubernetes
    -- rbac | secret_ref | network | dockerfile | script | scan_health
    severity           VARCHAR(50)  NOT NULL CHECK (severity IN ('Critical','High','Medium','Low','Info')),
    target_type        VARCHAR(100),                   -- source | image | helm | yaml | dockerfile | script | kubernetes | rbac | secret_ref | network
    target_name        TEXT,                           -- 파일 경로, 이미지 ref, k8s resource name
    target_cluster     VARCHAR(255),                   -- Biz applied finding의 ClusterTarget 이름. Code / Artifact finding은 NULL
    namespace          VARCHAR(255),
    image_digest       VARCHAR(255),                   -- sha256:...
    rule_id            VARCHAR(255),                   -- CVE ID, semgrep rule ID, policy ID
    message            TEXT         NOT NULL,
    remediation        TEXT,                           -- static catalog (AI sidecar는 별도)
    exception_required BOOLEAN      NOT NULL DEFAULT FALSE,
    exception_status   VARCHAR(50)  NOT NULL DEFAULT 'None'
        CHECK (exception_status IN ('None','Required','Requested','Approved','Expired','Rejected')),
    scan_status        VARCHAR(50)  NOT NULL
        CHECK (scan_status IN ('Pass','Fail','Error','Skipped','Unsupported')),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    details            JSONB,                          -- scanner 특화 추가 필드

    UNIQUE (finding_id, scan_run_id)
);

CREATE INDEX idx_findings_scan_run     ON findings(scan_run_id);
CREATE INDEX idx_findings_filter       ON findings(scan_run_id, category, severity, exception_status, scan_status);
CREATE INDEX idx_findings_digest       ON findings(image_digest) WHERE image_digest IS NOT NULL;
CREATE INDEX idx_findings_namespace    ON findings(namespace)    WHERE namespace IS NOT NULL;
CREATE INDEX idx_findings_target_cluster ON findings(target_cluster) WHERE target_cluster IS NOT NULL;
CREATE INDEX idx_findings_raw_report   ON findings(raw_report_id);
CREATE INDEX idx_findings_details_gin  ON findings USING GIN (details) WHERE details IS NOT NULL;
```

finding_id 생성 규칙 (DATABASE가 단일 정본.
PLAN/API_DESIGN 예시는 이 표를 따른다):

| category | 구성 요소 |
|----------|----------|
| `image_vulnerability` | `<scanner>/<imageRepository>/<imageDigest>/<vulnerabilityID>/<packageName>` |
| `sbom` | `<scanner>/<imageRepository>/<imageDigest>/<purl>` |
| `integrity` | `<scanner>/<imageDigest>/<verificationType>` |
| `sast`, `secret` | `<scanner>/<filePath>/<ruleID>/<lineHash>` |
| `dockerfile`, `script` | `<scanner>/<filePath>/<ruleID>/<lineHash>` (sast/secret 규칙 재사용) |
| `kubernetes`, `rbac` (Code / Artifact manifest) | `<scanner>/_artifact/<filePathHash>/<namespaceOr_cluster>/<resourceKind>/<resourceName>/<ruleID>` |
| `kubernetes`, `rbac` (Biz applied) | `<scanner>/<clusterTarget>/<namespaceOr_cluster>/<resourceKind>/<resourceName>/<ruleID>` |
| `secret_ref`, `network` | `<scanner>/<clusterTarget>/<namespace>/<resourceKind>/<resourceName>/<ruleID>` |
| `scan_health` | `scan_health/<scanRunID>/<scanner>/<errorCode>` |

규칙 메모:
- `<scanner>`는 `findings.scanner` 정규화 값이다.
  Trivy CLI 직접 image scan과 Trivy Operator `VulnerabilityReport` ingestion은 둘 다 `scanner=trivy`로
  정규화해, 입력 경로가 달라도 동일 이미지·CVE·패키지가 같은 finding_id가 되어 dedup된다.
  Grype는 `grype`.
  따라서 같은 이미지·CVE를 Trivy와 Grype가 모두 보고하면 finding_id가 달라 `UNIQUE (finding_id, scan_run_id)` 충돌 없이 둘 다
  기록된다.
- `<clusterTarget>`는 `ScanRun.spec.targets[]`의 ClusterTarget `metadata.name`(display name 아님)이며
  `findings.target_cluster` 컬럼에도 저장한다.
  한 ScanRun이 여러 Biz Cluster target을 검사할 때 동일 namespace/resource/rule이 target 간 충돌하지 않도록 한다.
  Code / Artifact Scan(Mgmt-local, target 비종속)의 manifest finding은 `_artifact` + `<filePathHash>`를
  사용하고 `target_cluster`는 NULL이다.
- cluster-scoped 리소스(ClusterRole, ClusterRoleBinding 등 namespace 없음)는 `<namespaceOr_cluster>` 자리에
  리터럴 `_cluster`를 쓴다.
- `<verificationType>`은 `digest_mismatch`, `signature_invalid`, `sbom_missing` 등 무결성 점검 코드다.
- finding_id는 `VARCHAR(512)` 한계 안에 인코딩한다.
  target+namespace+kind+name 조합이 길면 normalizer가 각 segment 길이를 제한하거나 초과분을 hash로 축약한다.
- 동일 `(finding_id, scan_run_id)`는 멱등 upsert되어 중복 집계되지 않는다(M5 dedup).
  원본 scanner 출력은 `raw_reports`에 보존한다.

---

### scan_health

scanner 실행 실패, unsupported target, stale DB/rule, missing artifact를 별도 카테고리로 기록한다.
취약점 없음으로 오판하지 않도록 강제한다.

```sql
CREATE TABLE scan_health (
    id           BIGSERIAL    PRIMARY KEY,
    scan_run_id  VARCHAR(255) NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
    scanner      VARCHAR(100),                         -- NULL이면 전체 pipeline 레벨 오류
    target_name  TEXT,
    status       VARCHAR(50)  NOT NULL CHECK (status IN ('OK','Warning','Fail','Skipped')),
    reason       VARCHAR(100),
    -- scanner_error | unsupported_target | missing_artifact | stale_db
    -- stale_rules | registry_pull_failure | rbac_denied | optional_input_unavailable
    -- ai_advisor_unavailable | ai_output_rejected
    message      TEXT,
    details      JSONB,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_scan_health_scanrun ON scan_health(scan_run_id);
CREATE INDEX idx_scan_health_status  ON scan_health(scan_run_id, status);
```

---

### exception_reviews

finding별 예외 승인 이력.
status machine을 강제한다.

```
Required → Requested → Approved
                     → Rejected
Approved → Expired  (expires_at 기준 자동)
```

```sql
CREATE TABLE exception_reviews (
    id           BIGSERIAL    PRIMARY KEY,
    finding_id   VARCHAR(512) NOT NULL,
    scan_run_id  VARCHAR(255) NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
    status       VARCHAR(50)  NOT NULL
        CHECK (status IN ('Required','Requested','Approved','Expired','Rejected')),
    owner        VARCHAR(255),                         -- 예외 신청자 또는 담당자
    reason       TEXT,
    expires_at   TIMESTAMPTZ,                          -- NULL이면 만료 없음
    approved_by  VARCHAR(255),
    approved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_exceptions_finding    ON exception_reviews(finding_id);
CREATE INDEX idx_exceptions_status     ON exception_reviews(status);
CREATE INDEX idx_exceptions_expiry     ON exception_reviews(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_exceptions_scanrun    ON exception_reviews(scan_run_id);
```

`findings.exception_status`와 동기화 규칙:

- `exception_reviews.status` 변경 시 대응하는 `findings.exception_status`를 같은 트랜잭션에서 업데이트한다.
- `expires_at < now()` 이면 background job이 `status = 'Expired'`로 전환하고 `findings.exception_status`도
  갱신한다.

재스캔 carry-over 규칙(같은 `finding_id`가 새 ScanRun에서 다시 보고될 때, operator가 적용):

- 직전 ScanRun에서 `Approved`이고 `expires_at`이 아직 유효하면 새 finding의 `exception_status`를 `Approved`로
  carry-over하고 `exception_reviews` row(owner/reason/expires_at/approved_by/approved_at)를 같은
  `finding_id`/새 `scan_run_id`로 복제한다.
- 직전 상태가 `Expired` 또는 `Rejected`이거나 `Approved`라도 `expires_at < now()`이면 carry-over하지 않고 새 finding을
  `Required`로 재평가한다(다시 승인 절차를 거친다).
- 직전 `Requested`(미결)는 carry-over하지 않고 새 ScanRun에서 `Required`로 시작한다.
- finding이 재스캔에서 사라지면(해결됨) 새 row를 만들지 않는다.

---

### artifact_index

Artifact Store 파일 참조.
SBOM, evidence bundle, human report, scanner baseline, artifact-input.yaml의 경로와 메타데이터를 기록한다.

```sql
CREATE TABLE artifact_index (
    id               BIGSERIAL    PRIMARY KEY,
    scan_run_id      VARCHAR(255) NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
    artifact_type    VARCHAR(100) NOT NULL,
    -- sbom | integrity_report | evidence_bundle | human_report
    -- exception_review_yaml | scanner_baseline | artifact_input
    -- remediation_advisory | remediation_provenance
    path             TEXT         NOT NULL,            -- Artifact Store 내 경로
    checksum         VARCHAR(255),                     -- sha256:<hex>
    schema_version   VARCHAR(50),                      -- security.finding/v1 등
    scanner          VARCHAR(100),
    scanner_version  VARCHAR(100),
    db_baseline_date DATE,                             -- 취약점 DB 기준일
    size_bytes       BIGINT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_artifact_index_scanrun ON artifact_index(scan_run_id, artifact_type);
CREATE INDEX idx_artifact_index_path    ON artifact_index(path);
```

---

### cluster_targets

ClusterTarget CR의 k8s 미러.
operator `ClusterTarget` reconciler/watch가 PostgreSQL `cluster_targets`를 upsert(정본 write 주체)하며,
backend는 read-only query로만 사용한다.
dashboard Targets 메뉴의 list/get 응답 속도를 보장한다.
미러 row가 아직 없는 ClusterTarget은 backend가 k8s API 직접 조회로 fallback하거나 `404 NOT_FOUND`를 반환한다.

```sql
CREATE TABLE cluster_targets (
    name                VARCHAR(255) PRIMARY KEY,       -- ClusterTarget.metadata.name
    display_name        VARCHAR(255),
    environment         VARCHAR(100),                   -- dev | final-check | prod
    phase               VARCHAR(50),
    -- Pending | Ready | Degraded | AuthFailed | Unreachable | PermissionDenied
    kubernetes_version  VARCHAR(50),
    capabilities        JSONB,
    -- {
    --   "scannerJobs": true,
    --   "readOnlyInspection": true,
    --   "trivyOperatorReports": false,
    --   "hostPath": false,
    --   "imageAccess": true,
    --   "reportUpload": true
    -- }
    namespace_allowlist TEXT[],
    conditions          JSONB,                          -- []metav1.Condition
    last_validated_at   TIMESTAMPTZ,
    last_credential_rotation_at TIMESTAMPTZ,
    synced_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_cluster_targets_phase ON cluster_targets(phase);
CREATE INDEX idx_cluster_targets_env   ON cluster_targets(environment);
```

---

## 마이그레이션 정책

- 마이그레이션 도구: `golang-migrate/migrate` 또는 `goose`
- 파일 위치: `backend/internal/db/migrations/`
- 명명 규칙: `YYYYMMDDHHMMSS_<description>.up.sql` / `.down.sql`
- 운영 환경에서 `down` 마이그레이션은 수동 실행만 허용한다.
- PostgreSQL 버전: `18.x` 권장.
  `17.x` 이상에서 동작 확인 필요.

---

## 데이터 보안 정책

- Secret raw value는 어떤 컬럼에도 저장하지 않는다.
- `raw_reports.data JSONB`에 Secret redaction guard를 통과한 출력만 삽입한다.
- `findings.message`, `findings.details`에도 redaction 적용 후 저장한다.
- `cluster_targets` 테이블에 kubeconfig Secret 값을 저장하지 않는다.
  `phase`, `capabilities`, `conditions` status 필드만 저장한다.
- PostgreSQL encryption at rest는 Mgmt Cluster 인프라 레벨에서 적용한다.
- `raw_reports` 접근 권한은 `backend` 서비스 계정에만 부여한다.
  dashboard UI는 backend API를 통해서만 raw report를 조회한다.
