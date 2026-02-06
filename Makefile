# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

ENSURE_GARDENER_MOD := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
NAME := diki-operator
IMAGE := $(NAME)
REPO_ROOT := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR := $(REPO_ROOT)/hack
GARDENER_HACK_DIR := $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack
LD_FLAGS := "-w $(shell bash $(GARDENER_HACK_DIR)/get-build-ld-flags.sh k8s.io/component-base $(REPO_ROOT)/VERSION $(NAME))"

GOCMD?= go
GOOS := $(shell $(GOCMD) env GOOS)
GOARCH := $(shell $(GOCMD) env GOARCH)
GO_TOOL := $(GOCMD) tool

VERSION := $(shell cat VERSION)
REVISION := $(shell git rev-parse --short HEAD)
EFFECTIVE_VERSION := $(VERSION)-$(REVISION)
ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

# Kubernetes code-generator tools
#
# https://github.com/kubernetes/code-generator
K8S_GEN_TOOLS := deepcopy-gen defaulter-gen register-gen conversion-gen
K8S_GEN_TOOLS_LOG_LEVEL ?= 0

# Common options for the `addlicense' tool
ADDLICENSE_OPTS ?= -f $(HACK_DIR)/LICENSE_BOILERPLATE.txt \
			-ignore "dev/**" \
			-ignore "**/*.md" \
			-ignore "**/*.html" \
			-ignore "**/*.yaml" \
			-ignore "**/*.yml" \
			-ignore "**/Dockerfile"

# Path in which to generate the API reference docs
API_REF_DOCS ?= $(REPO_ROOT)/docs/api-reference

# Run a command.
#
# When used with `foreach' the result is concatenated, so make sure to preserve
# the empty whitespace at the end of this function.
#
# https://www.gnu.org/software/make/manual/html_node/Foreach-Function.html
define run-command
$(1)

endef

TOOLS_DIR := $(REPO_ROOT)/hack/tools
include $(GARDENER_HACK_DIR)/tools.mk

# TODO: Uncomment these with CMD
# .PHONY: start
# start:
# 	go run ./cmd/diki-operator/main.go \
# 	    --config=$(REPO_ROOT)/examples/config.yaml \
# 		--kubeconfig $(KUBECONFIG)

# .PHONY: install
# install:
# 	@LD_FLAGS=$(LD_FLAGS) EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
# 		bash $(GARDENER_HACK_DIR)/install.sh ./...

# .PHONY: docker-images
# docker-images:
# 	@docker build --build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) --build-arg TARGETARCH=$(GOARCH) -t $(IMAGE):$(EFFECTIVE_VERSION) -t $(IMAGE):latest -f Dockerfile --target $(NAME) . --memory 6g


.PHONY: api-ref-docs
api-ref-docs:
	@mkdir -p $(API_REF_DOCS)
	@$(GO_TOOL) crd-ref-docs \
		--config $(REPO_ROOT)/api-ref-docs.yaml \
		--output-mode group \
		--output-path $(API_REF_DOCS) \
		--renderer markdown \
		--source-path $(REPO_ROOT)/pkg/apis

.PHONY: clean
clean:
	@bash $(GARDENER_HACK_DIR)/clean.sh ./cmd/... ./pkg/... ./internal/...

.PHONY: generate
generate:
	@echo "Running code-generator tools ..."
	$(foreach gen_tool,$(K8S_GEN_TOOLS),$(call run-command,$(GO_TOOL) $(gen_tool) -v=$(K8S_GEN_TOOLS_LOG_LEVEL) ./pkg/apis/diki/v1alpha1 ./pkg/apis/diki ))
	@echo "Generating CRDs ..."
	@$(GO_TOOL) controller-gen crd:crdVersions=v1 paths=./pkg/apis/diki/v1alpha1 output:crd:dir=./pkg/apis/diki/crds
#	@cp ./pkg/apis/diki/crds/*.yaml ./charts/diki/crds/
	@$(MAKE) api-ref-docs

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(HELM) $(YQ) $(TYPOS) 
	go vet ./...
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/... ./internal/...
	@bash $(GARDENER_HACK_DIR)/check-typos.sh
	@bash $(GARDENER_HACK_DIR)/check-file-names.sh
	@bash $(GARDENER_HACK_DIR)/check-charts.sh ./charts

.PHONY: get
get:
	@$(GOCMD) mod download
	@$(GOCMD) mod tidy

.PHONY: tidy
tidy:
	@$(GOCMD) mod tidy

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg ./internal/...

.PHONY: sast
sast: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh --exclude-dirs dev

.PHONY: sast-report
sast-report: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh --gosec-report true --exclude-dirs dev

.PHONY: test
test: $(REPORT_COLLECTOR)
	@bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/... ./internal/...

.PHONY: test-cov
test-cov:
	@bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/... ./internal/...

.PHONY: test-clean
test-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: addlicense
addlicense:
	@$(GO_TOOL) addlicense $(ADDLICENSE_OPTS) .

.PHONY: checklicense
checklicense:
	@files=$$( $(GO_TOOL) addlicense -check $(ADDLICENSE_OPTS) .) || { \
		echo "Missing license headers in the following files:"; \
		echo "$${files}"; \
		echo "Run 'make addlicense' in order to fix them."; \
		exit 1; \
	}

.PHONY: verify
verify: check format test sast

.PHONY: verify-extended
verify-extended: check-generate check format test test-cov test-clean sast-report checklicense
