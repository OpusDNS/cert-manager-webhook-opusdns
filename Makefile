GO ?= $(shell which go)
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

IMAGE_NAME := ghcr.io/opusdns/cert-manager-webhook-opusdns
IMAGE_TAG := latest

OUT := $(shell pwd)/_out

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

## Tool Binaries
ENVTEST ?= $(LOCALBIN)/setup-envtest

ENVTEST_VERSION ?= $(shell v='$(call gomodver,sigs.k8s.io/controller-runtime)'; \
  printf '%s\n' "$$v" | sed -E 's/^v?([0-9]+)\.([0-9]+).*/release-\1.\2/')

ENVTEST_K8S_VERSION ?= $(shell v='$(call gomodver,k8s.io/api)'; \
  printf '%s\n' "$$v" | sed -E 's/^v?[0-9]+\.([0-9]+).*/1.\1/')

.PHONY: all
all: build

.PHONY: build
build:
	CGO_ENABLED=0 $(GO) build -o webhook -ldflags '-w -extldflags "-static"' .

.PHONY: test
test: envtest
	@echo "Setting up envtest binaries for Kubernetes $(ENVTEST_K8S_VERSION)..."
	@KUBEBUILDER_ASSETS="$$("$(ENVTEST)" use $(ENVTEST_K8S_VERSION) --bin-dir "$(LOCALBIN)" -p path)"; \
	TEST_ASSET_ETCD="$$KUBEBUILDER_ASSETS/etcd" \
	TEST_ASSET_KUBE_APISERVER="$$KUBEBUILDER_ASSETS/kube-apiserver" \
	TEST_ASSET_KUBECTL="$$KUBEBUILDER_ASSETS/kubectl" \
	$(GO) test -v .

.PHONY: verify
verify: lint test

.PHONY: envtest
envtest: $(ENVTEST)
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: clean
clean:
	chmod -R u+w $(LOCALBIN) $(OUT) 2>/dev/null || true
	rm -rf $(LOCALBIN) $(OUT) webhook

.PHONY: lint
lint:
	golangci-lint run

.PHONY: docker-build
docker-build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

.PHONY: docker-push
docker-push: docker-build
	docker push "$(IMAGE_NAME):$(IMAGE_TAG)"

.PHONY: docker-buildx
docker-buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

.PHONY: helm-template
helm-template: $(OUT)/rendered-manifest.yaml

$(OUT)/rendered-manifest.yaml: | $(OUT)
	helm template cert-manager-webhook-opusdns \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) \
		deploy/cert-manager-webhook-opusdns > $@

$(OUT):
	mkdir -p $@

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$$(realpath "$(1)-$(3)")" "$(1)"
endef

define gomodver
$(shell go list -m -f '{{if .Replace}}{{.Replace.Version}}{{else}}{{.Version}}{{end}}' $(1) 2>/dev/null)
endef
