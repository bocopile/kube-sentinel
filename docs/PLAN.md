# kube-sentinel — PoC 구현 계획서 v1.1

> **도구**: Falco · Tetragon · OSquery · Trivy  
> **파이프라인**: OTel Node Collector → OTel Gateway → Elasticsearch  
> **아키텍처**: Feature-as-Plugin (배열 기반 CRD + 자기등록 Registry)  
> **프레임워크**: CTEM Scope / Discovery / Priority / Validation 매핑  
> **비목표**: 멀티클러스터 · 인라인 차단 · Kafka · 완벽한 OCSF 정규화

---

## 1. 목표 & 성공 기준

| # | 성공 기준 | 검증 방법 |
|---|----------|----------|
| G1 | `SecurityAgent` CRD 1개 적용으로 4개 센서 + OTel 파이프라인 자동 배치 | `kubectl apply` 후 DS/Deploy 자동 생성 확인 |
| G2 | Feature 토글(`enabled: false`)로 개별 DaemonSet 생성/삭제 | CRD 수정 → DS 사라짐/생성 확인 |
| G3 | Override 적용으로 특정 도구의 리소스/tolerations 변경 | CRD override 수정 → DS spec 변경 확인 |
| G4 | 4개 도구 이벤트가 ES에 CTEM 용도별 인덱스로 적재 | Kibana에서 3개 인덱스 조회 |
| G5 | Kibana 대시보드 3개 동작 | 스크린샷 |
| G6 | MITRE ATT&CK 테스트 시나리오 5개 실행 시 탐지 | ES 쿼리로 이벤트 존재 확인 |
| G7 | CTEM 매핑 검증 체크리스트 통과 | 검증 항목 Pass/Fail |

### 구현 Stage Gate

PoC는 한 번에 4개 센서를 모두 붙이지 않고, 다음 세로 경로를 먼저 고정한다.

| Stage | 범위 | 통과 기준 |
|-------|------|----------|
| S0 | 클러스터 권한/BTF/hostPath/ES auth 사전 검증 | privileged DS 실행, `/sys/kernel/btf/vmlinux` 확인, ES 테스트 문서 적재 |
| S1 | OTel Node/Gateway + 수동 샘플 로그 3종 parser + Trivy fixture 검증 | Falco/Tetragon/OSquery 샘플이 `security-events`/`security-inventory`로 라우팅, Trivy fixture가 `security-vuln`에 upsert |
| S2 | `SecurityAgent` -> OTel Pipeline -> Falco -> `security-events` | CRD 1개로 Falco 이벤트가 ES에 유입 |
| S3 | Tetragon, OSquery, Trivy를 순차 추가 | 각 Feature 별 enable/disable, status, ES 문서 검증 |
| S4 | MITRE 시나리오 + Kibana + Override/GC 검증 | G2~G7 통과 |

---

## 2. 전체 아키텍처

```
SecurityAgent CRD (단일 진입점)
        │
        ▼
  kube-sentinel Operator (controller-runtime)
  ┌─────────────────────────────────────────┐
  │  Feature Registry (우선순위 기반)        │
  │  ├── otel_pipeline  (Priority 10)       │
  │  ├── falco          (Priority 100)      │
  │  ├── tetragon       (Priority 100)      │
  │  ├── osquery        (Priority 100)      │
  │  └── trivy          (Priority 200)      │
  └─────────────────────────────────────────┘
        │  Contribute() → DesiredStateStore
        │  SSA Apply (Server-Side Apply)
        ▼
  ┌──────────────────────────────────────────────────────┐
  │ Worker Node                                          │
  │  ┌─────────┐  ┌──────────┐  ┌─────────┐            │
  │  │  Falco  │  │ Tetragon │  │ OSquery │            │
  │  │ (eBPF)  │  │  (eBPF)  │  │  (SQL)  │            │
  │  └────┬────┘  └────┬─────┘  └────┬────┘            │
  │       │file        │stdout        │file             │
  │       ▼            ▼             ▼                  │
  │  ┌──────────────────────────────────────────────┐   │
  │  │ OTel Node Collector (DaemonSet)              │   │
  │  │ filelog/falco · filelog/tetragon · filelog/osquery│
  │  │ k8sattributes → transform → batch → otlp    │   │
  │  └───────────────────┬──────────────────────────┘   │
  └──────────────────────┼──────────────────────────────┘
                         │ OTLP/gRPC
                         ▼
               ┌─────────────────────┐
               │  OTel Gateway       │
               │  transform/severity │
               │  routing processor  │
               └──────┬──────────────┘
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
   security-events  security-   security-
   (Falco/Tetragon) inventory   vuln
   CTEM Validation  (OSquery)   (Trivy)
                   CTEM Scope  CTEM Discovery

  ※ Trivy: VulnerabilityReport CRD → CronJob → security-vuln (직접 적재)
```

---

## 3. CRD 설계

### SecurityAgent 샘플

```yaml
apiVersion: security.kube-sentinel.io/v1alpha1
kind: SecurityAgent
metadata:
  name: kube-sentinel
spec:
  global:
    clusterName: my-cluster
    targetNamespace: kube-sentinel-system

  features:
    - name: otel_pipeline
      enabled: true
      config:
        mode: "node-to-gateway-to-es"
        nodeLogBasePath: "/var/log/kube-sentinel"
        gatewayReplicas: 1

    - name: falco
      enabled: true
      config:
        driver: "modern_ebpf"
        jsonOutput: true

    - name: tetragon
      enabled: true
      config:
        exportMode: "stdout"
        policies: ["process-exec", "container-escape-monitor"]

    - name: osquery
      enabled: true
      config:
        intervalSeconds: 60
        packs: ["scope-minimal"]

    - name: trivy
      enabled: true
      config:
        scanSchedule: "@every 6h"
        severityThreshold: "HIGH"

  output:
    elasticsearch:
      endpoints: ["https://kube-sentinel-es-http.kube-sentinel-system:9200"]
      authSecretRef: { name: "es-credentials" }
      caSecretRef:   { name: "es-ca" }
      indices:
        inventory:       "security-inventory"   # CTEM Scope
        events:          "security-events"       # CTEM Validation
        vulnerabilities: "security-vuln"         # CTEM Discovery/Priority

  override:
    nodeAgent:
      tolerations:
        - key: "security.kube-sentinel.io/agent"
          operator: "Equal"
          value: "enabled"
          effect: "NoSchedule"
    falco:
      resources:
        limits:
          memory: "2Gi"
```

### Go 타입 정의 (핵심)

```go
type SecurityAgentSpec struct {
    Global   GlobalSpec    `json:"global,omitempty"`
    Features []FeatureSpec `json:"features"`
    Output   OutputSpec    `json:"output"`
    Override *OverrideSpec `json:"override,omitempty"`
    Tests    *TestSpec     `json:"tests,omitempty"`
}

// RawExtension → 스키마 변경 없이 새 도구 config 추가 가능
type FeatureSpec struct {
    Name    string               `json:"name"`
    Enabled bool                 `json:"enabled"`
    Config  runtime.RawExtension `json:"config,omitempty"`
}

type IndexSpec struct {
    Inventory       string `json:"inventory"`
    Events          string `json:"events"`
    Vulnerabilities string `json:"vulnerabilities"`
}
```

### Status 및 검증 정책

`RawExtension` 기반 config는 확장성은 높지만 CRD schema 검증이 약하므로, 런타임 검증 실패를 status에 명확히 노출한다.

```go
type SecurityAgentStatus struct {
    ObservedGeneration int64              `json:"observedGeneration,omitempty"`
    Phase              string             `json:"phase,omitempty"` // Ready, Degraded, Progressing
    Features           []FeatureCondition `json:"features,omitempty"`
    ManagedResources   []ManagedResource  `json:"managedResources,omitempty"`
}

type FeatureCondition struct {
    Name               string      `json:"name"`
    Enabled            bool        `json:"enabled"`
    Ready              bool        `json:"ready"`
    Reason             string      `json:"reason,omitempty"`  // Disabled, Ready, ConfigError, ApplyError, NotReady
    Message            string      `json:"message,omitempty"`
    ObservedGeneration int64       `json:"observedGeneration,omitempty"`
    LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}
```

검증 정책:

- 알 수 없는 `features[].name`은 `ConfigError`로 status에 기록하고 해당 feature는 적용하지 않는다.
- `Configure()` 실패는 전체 reconcile 실패로 처리하되, 이미 정상 적용된 리소스를 무리하게 삭제하지 않는다.
- feature별 기본값은 각 feature package에 두고, sample YAML은 기본값을 설명하는 용도로만 사용한다.

---

## 4. Feature-as-Plugin 아키텍처

### Feature 인터페이스

```go
type Feature interface {
    ID()        FeatureID
    Configure(raw []byte) error
    Contribute(ctx context.Context, store *DesiredStateStore) error
    OTelConfig() *OTelReceiverConfig  // nil이면 OTel 수집 불필요 (Trivy)
    Assess(ctx context.Context, c client.Client, ns string) FeatureCondition
}
```

### 우선순위 Registry

```
Priority 10   otel_pipeline  ← 수집 인프라가 센서보다 먼저 Ready
Priority 100  falco, tetragon, osquery
Priority 200  trivy
```

### 새 도구 추가 = 3단계

```
1. internal/feature/<newtool>/feature.go 생성
   → Feature 인터페이스 구현
   → init()에서 feature.Register() 호출

2. cmd/main.go에 import 1줄 추가

3. 끝. Reconciler 코드 변경 없음. CRD 스키마 변경 없음.
```

---

## 5. OTel 파이프라인 — 데이터 흐름

### Node Collector 설정 (자동 합성)

| 도구 | 출력 방식 | OTel 수집 경로 | k8s 메타 소스 |
|------|-----------|---------------|--------------|
| Falco | 파일 출력 (`/var/log/kube-sentinel/falco/events.log`) | `filelog/falco` | Falco JSON 내장 필드 → `transform/falco_meta` 리매핑 |
| Tetragon | stdout | `filelog/tetragon` (`/var/log/pods/...`) | `k8sattributes` 자동 부착 |
| OSquery | 파일 출력 (`/var/log/kube-sentinel/osquery/results.log`) | `filelog/osquery` | `hostIdentifier` → node 메타 |

### Gateway — Severity 통합 매핑

```
통합 severity    Falco              Tetragon         Trivy
5 (Critical)     Emergency/Alert    -                CRITICAL
4 (High)         Critical           kprobe match     HIGH
3 (Medium)       Error/Warning      -                MEDIUM
2 (Low)          Notice             process_exec     LOW
1 (Info)         Info/Debug         -                UNKNOWN
```

### 인덱스 라우팅

```
security_tool: osquery   → security-inventory  (CTEM Scope)
security_tool: falco     → security-events     (CTEM Validation)
security_tool: tetragon  → security-events     (CTEM Validation)
security_tool: trivy     → security-vuln       (CTEM Discovery/Priority, 직접 적재)
```

### Trivy 적재 방식

Trivy는 OTel filelog 경로에 억지로 태우지 않고, Trivy Operator가 생성한 `VulnerabilityReport`를 별도 Job/Controller가 읽어 Elasticsearch에 upsert한다.

```
Trivy Operator
  → VulnerabilityReport CRD
  → kube-sentinel trivy-ingestor CronJob
  → bulk upsert
  → security-vuln
```

upsert document id:

```
<clusterName>/<namespace>/<workloadKind>/<workloadName>/<containerName>/<vulnerabilityID>/<packageName>
```

M6 통과 조건은 같은 report를 2회 적재해도 document count가 증가하지 않는 것이다.

---

## 6. Reconciler 흐름

```
Reconcile()
  │
  ├── 1. Finalizer 등록
  ├── 2. Spec 변경, owned resource 변경, 주기적 health check 이벤트 수신
  ├── 3. BuildActiveFeatures() → 우선순위 순서로 Feature 목록 구성
  ├── 4. 각 Feature.Contribute() → DesiredStateStore에 리소스 기여
  ├── 5. OTelConfig() 수집 → OTel Node ConfigMap 자동 합성
  ├── 6. Override 적용 (공통 nodeAgent → 도구별 순서)
  ├── 7. SSA Apply (OwnerReference 자동 설정)
  ├── 8. 비활성 Feature GC (stale 리소스 정리)
  └── 9. Status 갱신 (ObservedGeneration + FeatureCondition)
```

### Watch / Drift / Status 원칙

- `GenerationChangedPredicate`만 사용하면 Pod 상태 변화나 generated resource drift를 놓칠 수 있으므로, `For(SecurityAgent)`와 함께 주요 owned resource를 `Owns()`로 watch한다.
- spec 변경 reconcile과 status 점검 reconcile을 분리한다. spec 미변경 이벤트에서는 desired state 재합성은 허용하되, 불필요한 apply는 object hash 비교로 줄인다.
- 외부 상태 점검은 `RequeueAfter`로 주기 실행한다. PoC 기본값은 60초이며, ES 연결 실패나 Pod NotReady는 `Degraded`로 표시한다.
- Status update는 별도 patch로 수행하고, `observedGeneration`은 spec 기반 apply가 성공한 뒤에만 갱신한다.

### SSA / GC 리소스 소유 전략

모든 생성 리소스에는 다음 label/annotation을 붙인다.

```yaml
metadata:
  labels:
    app.kubernetes.io/managed-by: kube-sentinel
    security.kube-sentinel.io/instance: kube-sentinel
    security.kube-sentinel.io/feature: falco
  annotations:
    security.kube-sentinel.io/spec-hash: "<sha256>"
```

- SSA field manager는 `kube-sentinel/<feature>` 형식으로 분리한다.
- apply conflict는 기본적으로 status `ApplyError`로 보고하고 강제 적용하지 않는다.
- GC는 ownerReference와 label selector를 함께 사용한다. CRD처럼 ownerReference가 부적절하거나 cluster-scoped인 리소스는 label 기반으로만 정리한다.
- 비활성 feature GC는 `security.kube-sentinel.io/feature=<id>` 리소스 중 desired set에 없는 항목만 삭제한다.

---

## 7. 프로젝트 디렉터리 구조

```
kube-sentinel/
├── cmd/
│   └── main.go                          # Feature import + healthz + metrics
│
├── api/v1alpha1/
│   ├── securityagent_types.go           # CRD 타입 정의
│   └── zz_generated.deepcopy.go
│
├── internal/
│   ├── controller/
│   │   └── reconciler.go               # Feature-agnostic Reconciler
│   └── feature/
│       ├── feature.go                  # Feature 인터페이스
│       ├── types.go                    # OTelReceiverConfig, FeatureCondition
│       ├── registry.go                 # 우선순위 기반 Registry
│       ├── store.go                    # DesiredStateStore
│       ├── override/override.go        # 2단계 Override
│       ├── otel/config_builder.go      # OTel ConfigMap 자동 합성
│       ├── otel_pipeline/feature.go    # Priority 10
│       ├── falco/feature.go            # Priority 100
│       ├── tetragon/feature.go         # Priority 100
│       ├── osquery/feature.go          # Priority 100
│       └── trivy/feature.go            # Priority 200
│
├── config/
│   ├── crd/bases/
│   ├── elasticsearch/
│   │   ├── elasticsearch.yaml
│   │   ├── kibana.yaml
│   │   └── index-templates.sh
│   └── samples/
│       ├── securityagent_full.yaml
│       └── securityagent_minimal.yaml
│
├── test/
│   ├── pods.yaml                       # testbox + attacker + target-nginx
│   └── run-ctem-scenarios.sh           # MITRE 시나리오 자동화
│
└── docs/
    ├── PLAN.md                         # 이 문서
    └── ctem-mapping-results.md         # 검증 결과 (M7 이후 작성)
```

---

## 8. 마일스톤

| 마일스톤 | 내용 | 기간 | Exit Criteria |
|---------|------|:---:|--------------|
| **M0** | 인프라 준비 (네임스페이스, PSA, BTF 확인, 로그 디렉터리) | 1일 | privileged Pod 배포 가능, `/sys/kernel/btf/vmlinux` 존재 |
| **M0.5** | 로그/OTel/ES 스파이크 (샘플 로그 parser + Trivy fixture upsert) | 1일 | OTel 샘플 로그가 events/inventory에 적재되고 Trivy fixture가 vuln에 upsert |
| **M1** | Elasticsearch + Kibana (ECK) + 인덱스 템플릿 3개 | 1~2일 | ES green, 3개 인덱스 템플릿 생성, 테스트 문서 삽입 성공 |
| **M2** | Operator Core (CRD + Registry + Store + Override + SSA + Finalizer) + OTel Pipeline Feature | 3~4일 | CRD 적용 시 OTel Gateway/Node DS 자동 생성, OTLP → ES 적재 확인 |
| **M3** | Falco Feature (파일 출력 + k8s meta + 로테이션 + 노이즈 제어) | 2~3일 | `kubectl exec -- sh` → security-events에 Falco 이벤트 적재 |
| **M4** | Tetragon Feature (stdout + TracingPolicy CRD) | 2일 | 시나리오 B 교차검증 (Falco + Tetragon 양쪽) |
| **M5** | OSquery Feature (CTEM Scope 쿼리 팩) | 2일 | security-inventory에 system_info 문서 유입 |
| **M6** | Trivy Feature (Operator + CronJob upsert) | 2일 | security-vuln에 CVE 적재, 중복 없음 |
| **M7** | MITRE 시나리오 5개 + Kibana 대시보드 3개 + CTEM 검증 | 2~3일 | 4개+ 탐지, 교차검증 2개+, 대시보드 스크린샷 |
| **M8** | Feature 토글 + Override 최종 검증 | 1일 | 토글 시 DS 생성/삭제, Override 반영 확인 |

**총 예상 기간: 3주+ (16 working days, parser/권한 이슈 발생 시 4주 버퍼)**

```
Week 1                Week 2                Week 3
──────────────────────────────────────────────────
M0 ■                  M3 ■■■                M6 ■■
M0.5 ■                M4 ■■                 M7 ■■■
M1 ■■                 M5 ■■                 M8 ■
M2 ■■■■
```

---

## 9. CTEM 프레임워크 매핑

| CTEM 단계 | 담당 도구 | ES 인덱스 | 비고 |
|----------|---------|----------|------|
| **Scope** | OSquery | `security-inventory` | 노드 OS/커널/포트/컨테이너 인벤토리 |
| **Discovery** | Trivy | `security-vuln` | CVE ID/CVSS/fixedVersion |
| **Priority** | Trivy + Falco | `security-vuln`, `security-events` | CVSS + Falco priority 통합 severity |
| **Validation** | Falco + Tetragon | `security-events` | MITRE 시나리오 탐지 + 교차검증 |
| **Mobilization** | Kibana Alert | — | PoC 선택사항 |

---

## 10. MITRE ATT&CK 테스트 시나리오

| # | 기법 | 실행 Pod | 탐지 도구 | Pass 기준 |
|---|------|---------|---------|---------|
| A | CTEM Scope 수집 | (자동) | OSquery | 15분 내 inventory 유입 |
| B | T1059.004 Unix Shell | testbox (alpine) | Falco + Tetragon | 양쪽 이벤트 존재 |
| C | T1552.001 + K8s API | testbox + attacker | Falco + Tetragon | 양쪽 이벤트 존재 |
| D | T1611 Container Escape | testbox + attacker | Tetragon (kprobe) | TracingPolicy 동작 확인 |
| E | T1053.003 Cron Persistence | testbox | Falco + Tetragon | 즉시 탐지 |

테스트 전제:

- `testbox`와 `attacker` Pod의 RBAC, PSA, privileged/capability 설정을 `test/pods.yaml`에 명시한다.
- 클러스터 정책 때문에 D/E 시나리오 실행이 차단되면, 차단 이벤트 자체를 기록하고 대체 시나리오를 실행한다.
- Pass 기준은 "탐지 이벤트 존재"와 함께 `security_tool`, `rule_name` 또는 `policy_name`, `k8s.namespace.name`, `k8s.pod.name`, `severity_number` 필드 존재를 포함한다.

---

## 11. 리스크

| # | 리스크 | 확률 | 회피 전략 |
|---|--------|:---:|----------|
| R1 | OTel filelog JSON 파싱 실패 | 높음 | 사전 로그 샘플 수집 후 파서 테스트 |
| R2 | k8sattributes가 Falco 파일 로그에 Pod 메타 못 붙임 | 높음 | 하이브리드: Falco=transform 리매핑, Tetragon=k8sattributes |
| R3 | Falco + Tetragon eBPF 공존 시 커널 충돌 | 낮음 | Ubuntu 6.1+ 사전 검증, 문제 시 Tetragon 단독 테스트 |
| R4 | Falco 노이즈로 ES 폭발 | 높음 | PoC 전용 falco_rules.local.yaml (노이즈 Top5 비활성화) |
| R5 | Reconcile 무한 루프 | 중간 | spec hash 비교 + status patch 분리 + ObservedGeneration 패턴 |
| R6 | ES 매핑 충돌 (dot notation vs 객체) | 중간 | 객체 구조 컴포넌트 템플릿 + flattened 타입 |
| R7 | 로그 로테이션 미설정으로 디스크 풀 | 중간 | Falco rotate max_size:100MB×3 |
| R8 | Predicate 과사용으로 drift/status 변경 미감지 | 중간 | owned resource watch + RequeueAfter health check |
| R9 | Trivy report 중복 적재 | 중간 | deterministic document id + bulk upsert |
| R10 | 테스트 Pod 권한 부족으로 MITRE 시나리오 실패 | 중간 | RBAC/PSA 전제 명시 + 대체 시나리오 |

---

## 12. 리소스 사이징 (노드당)

```
컴포넌트           CPU req/lim     Memory req/lim
─────────────────────────────────────────────────
Falco DS           100m / 500m     512Mi / 1Gi
Tetragon DS        100m / 500m     256Mi / 512Mi
OSquery DS          50m / 200m     128Mi / 256Mi
OTel Node DS       100m / 300m     128Mi / 256Mi
─────────────────────────────────────────────────
노드당 합계        350m / 1500m    1Gi / 2Gi
```

**최소 클러스터**: Control Plane 2vCPU/4GiB + Worker×2 (4vCPU/8GiB), 커널 5.15+ (BTF)
