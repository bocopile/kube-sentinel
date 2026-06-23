# kube-sentinel operator — build/test/codegen targets.
# Requires Go and controller-gen (installed via `make controller-gen`).

IMG ?= kube-sentinel-operator:latest
CONTROLLER_GEN_VERSION ?= v0.16.5
LOCALBIN ?= $(shell pwd)/bin
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

CRD_OPTIONS ?= crd

.PHONY: all
all: build

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(LOCALBIN) ## Install controller-gen into ./bin.
	@test -s $(CONTROLLER_GEN) || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

.PHONY: generate
generate: controller-gen ## Generate DeepCopy methods.
	$(CONTROLLER_GEN) object:headerFile=/dev/null paths="./api/..."

.PHONY: manifests
manifests: controller-gen ## Generate CRD and RBAC manifests.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./api/..." output:crd:dir=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=kube-sentinel-operator-role paths="./internal/..." output:rbac:dir=config/rbac

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: test
test: generate manifests fmt vet ## Run unit tests.
	go test ./... -coverprofile cover.out

.PHONY: build
build: generate fmt vet ## Build the operator binary.
	go build -o bin/manager ./cmd

.PHONY: run
run: generate fmt vet ## Run the operator against the configured cluster.
	go run ./cmd

.PHONY: docker-build
docker-build: ## Build the operator container image.
	docker build -t $(IMG) .

.PHONY: install
install: manifests ## Install CRDs into the cluster.
	kubectl apply -f config/crd/bases

.PHONY: uninstall
uninstall: ## Remove CRDs from the cluster.
	kubectl delete -f config/crd/bases
