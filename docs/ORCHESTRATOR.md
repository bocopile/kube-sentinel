# orchestrator 사용 가이드

`~/IdeaProjects/orchestrator`는 kube-sentinel의 source template가 아니라
workflow runner로 사용한다. 내장 scaffold command는 현재 Node/TypeScript
project를 대상으로 하며, kube-sentinel은 Go Kubernetes operator다.

## 현재 bootstrap 상태

이 리포지터리는 pre-skeleton Go 상태다. 임시 root `go.mod`(첫 PR에서 `operator/go.mod`로
대체·제거 예정)와 계획 문서는 존재하지만 3-모듈 정본의 `operator/go.mod`와 Kubebuilder가
생성한 `operator/cmd/`, `operator/api/`, `operator/internal/`, `operator/config/`,
`operator/PROJECT`, `.orchestrator/config.yaml` 파일은 아직 없다.

따라서 유효한 실행 방식은 두 가지다.

- local orchestrator `plan`과 `run`을 사용할 수 있으면
  [PROMPTS.md](./PROMPTS.md)를 orchestrator로 실행한다.
- orchestrator가 아직 foundation/skeleton 단계이거나 이 리포지터리에서
  실행에 실패하면, 먼저 P0-P3 prompt를 Claude Code에서 직접 사용한다.
  Go project를 안정적으로 초기화하고 실행할 수 있게 된 뒤 orchestrator로
  돌아온다.

## 로컬 선행 조건

현재 머신에 이미 있는 항목:

- Node 24
- `orchestrator`
- Go
- `claude`
- `codex`
- `cursor-agent`
- `kubectl`
- `helm`

operator workflow에 추가로 필요한 항목:

- `kubebuilder`
- `controller-gen`
- `kustomize`
- `kind` or `minikube`

## 권장 순서

프로젝트는 모노레포 3-모듈 구조다. Kubebuilder는 `operator/` 안에서 초기화한다.

```bash
# 1. operator 모듈 초기화 (최초 1회)
mkdir -p operator
cd operator/
kubebuilder init --domain kube-sentinel.io --repo github.com/bocopile/kube-sentinel/operator
kubebuilder create api --group security --version v1alpha1 --kind ClusterTarget --namespaced=false
kubebuilder create api --group security --version v1alpha1 --kind SecurityAssessment --namespaced=false
kubebuilder create api --group security --version v1alpha1 --kind ScanRun --namespaced=false
go test ./...
go build ./...
cd ..

# 2. backend 모듈 초기화 (최초 1회)
cd backend/
go mod init github.com/bocopile/kube-sentinel/backend
cd ..

# 3. frontend 모듈 초기화 (최초 1회)
cd frontend/
npm create next-app@latest . --typescript --tailwind --app
cd ..

# 4. orchestrator 초기화 (root에서)
orchestrator init --project . --yes
```

이미 초기화된 repo에서 생성 파일을 확인하지 않고 Kubebuilder command를 두 번
실행하지 않는다. `operator/go.mod`, `operator/PROJECT`, `operator/api/`, `operator/config/`가 이미 존재하면
재초기화하지 말고 해당 상태에 맞는 prompt부터 이어서 진행한다.

그 다음 milestone별로 orchestrator를 실행한다.

```bash
orchestrator plan --project . --request "Implement the first kube-sentinel code block from docs/ROADMAP.md"
orchestrator run --project . --request "Implement the first kube-sentinel code block from docs/ROADMAP.md" --auto-approve
```

모든 stage에서 `run` 전에 `plan`을 사용한다. `plan` command는 코드 변경 전에
멈추고, 요청에 충분한 acceptance criteria와 project context가 있는지
확인할 수 있어 유용하다.

## 권장 요청 방식

넓은 요청 대신 좁은 milestone 요청을 사용한다.

좋은 예:

```text
Implement M2 management controller core from docs/ROADMAP.md: CRD type,
assessment registry, desired state store, target kubeconfig loader, remote
apply skeleton, SSA apply skeleton, report writer skeleton, and tests. Do not
implement runtime sensors yet.
```

피해야 할 예:

```text
Build kube-sentinel.
```

## 검증 명령

이 리포지터리의 orchestrator config는 Go command를 사용해야 한다.

```bash
(cd operator && go test ./... && go build ./...)
(cd backend  && go test ./... && go build ./...)
```

Cluster stage는 milestone request에 명시적인 manual check를 추가해야 한다.
예:

```bash
kubectl --context mgmt get clustertarget,securityassessment,scanrun -A
kubectl --context mgmt logs -n kube-sentinel-system deploy/kube-sentinel-controller-manager

kubectl --context biz-a get namespace kube-sentinel-system
kubectl --context biz-a get job,cronjob,cm,sa,role,rolebinding -n kube-sentinel-system
```

모든 milestone request에서 kubeconfig context를 명시한다. Mgmt Cluster
command는 kube-sentinel CRD, controller log, status를 확인한다. Biz Cluster
command는 remote apply된 resource와 read-only scan target만 확인한다.

Report stage는 PostgreSQL `raw_reports` record, normalized finding record, scan health
summary, final decision record, evidence bundle, dashboard screenshot에 대한
구체적인 검증을 포함해야 한다.
