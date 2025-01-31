
# needed if we use do not use --logstderr=false
SHELL := bash

REV=$(shell git describe --long --tags --match='v*' --dirty 2>/dev/null || git rev-list -n1 HEAD)

URL ?= http://localhost:8100
DOWNLOAD_DIR ?= ./testdata-download
UPLOAD_DIR ?= ./testdata
ASSET_DIR ?= ./internal/webapp/assets

# This is the default. It can be overridden in the main Makefile after
# including build.make.
REGISTRY_NAME=zot.lion.act3-ace.ai
IMAGE_REPO ?= $(REGISTRY_NAME)/ace/data/telemetry

# WebApp Dependencies
BOOTSTRAP_VERSION  ?= 5.2.2
BOOTSTRAP_ICONS_VERSION ?= 1.11.0
LEADER_LINE_VERSION ?= 1.0.7
MATHJAX_VERSION ?= 3.2.2
REQUIREJS_VERSION ?= 2.3.6

# Verbosity while running locally
V ?= 8

# Tool verisons
CONTROLLER_GEN_VERSION?=v0.14.0
CRD_REF_DOCS_VERSION?=v0.0.11
KO_VERSION?=v0.15.2
GOLANGCILINT_VERSION?=v1.58.2

all: build

.PHONY: generate apidoc
generate: tool/controller-gen
	go generate ./...

.PHONY: build
build: deps
	@mkdir -p bin
	go build -tags "sqlite_fts5" -o bin/telemetry ./cmd/telemetry 

.PHONY: build-debug
build-debug: deps
	@mkdir -p bin
	go build -gcflags="all=-N -l" -tags "sqlite_fts5" -o bin/telemetry ./cmd/telemetry 

.PHONY: build-linux
build-linux: deps
	@mkdir -p ci-dist
	GOOS=linux GOARCH=amd64 go build -o ci-dist/telemetry/linux/amd64/bin/telemetry ./cmd/telemetry

.PHONY: test
test:

.PHONY: test-go
test: test-go
test-go: template
	go test ./... -tags "sqlite_fts5"

.PHONY: lint
test: lint
lint: tool/golangci-lint
	tool/golangci-lint run

.PHONY: cover
cover: template
	go clean -testcache
	- rm coverage.txt
	go test ./... -coverprofile coverage.txt -coverpkg=$(shell go list)/...
	./filter-coverage.sh < coverage.txt > coverage.txt.filtered
	go tool cover -func coverage.txt.filtered

.PHONY: clean
clean: clean-deps
	- rm -rf bin/*
	- rm test.db
	- rm tool/*
	- rm -f testdata/event/event*.json
	- rm -f testdata/manifest/manifest*.json
	- rm -f testdata/bottle/bottle*.json
	# go clean -cache

.PHONY: run
# ./bin/telemetry -v=$(V) serve $(RUN_ARGS) 2> >(jq -j -f log.jq)
# ./bin/telemetry -v=$(V) serve --logtostderr=false $(RUN_ARGS) | jq -j -f log.jq
run: build
	./bin/telemetry -v=$(V) serve $(RUN_ARGS) 2> >(jq -j -f log.jq)

.PHONY: template
template: build
	./bin/telemetry -v=$(V) template ./testdata

.PHONY: upload
upload: template build
	./bin/telemetry -v=$(V) client upload $(UPLOAD_DIR) $(URL) --all --continue 2> >(jq -j -f log.jq)

# .PHONY: reload
# reload: template build
# 	- rm test.db
# 	./bin/telemetry -v=$(V) client upload $(UPLOAD_DIR) --all 2> >(jq -j -f log.jq)
	
.PHONY: download
download: build
	@mkdir -p testdata-download
	./bin/telemetry -v=$(V) client download $(DOWNLOAD_DIR) $(URL) --all --from-latest 2> >(jq -j -f log.jq)

.PHONY: test-webapp
test-webapp: template
	hack/test-webapp.sh $(URL)

.PHONY: hub
hub:
	@echo "*** Ensure you have an .act3_token file with an ACT3 Gitlab token ***"
	podman build --secret=id=act3_token,src=.act3_token -t $(IMAGE_REPO)/hub:$(REV) -t $(IMAGE_REPO)/hub:latest .acehub
	podman push $(IMAGE_REPO)/hub:$(REV)
	podman push $(IMAGE_REPO)/hub:latest

.PHONY: image
image: build-linux
	podman build . -t $(IMAGE_REPO):$(REV) --label version=$(REV)
	podman push $(IMAGE_REPO):$(REV)

.PHONY: install
install:
	go install ./cmd/telemetry

# NOTE This does not support jupyter (we shell out to jupter)
.PHONY: ko
ko: tool/ko
	VERSION=$(REV) KO_DOCKER_REPO=$(IMAGE_REPO) tool/ko build -B --platform=all --image-label version=$(REV) ./cmd/telemetry

tool/controller-gen: tool/.controller-gen.$(CONTROLLER_GEN_VERSION)
	GOBIN=$(PWD)/tool go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

tool/.controller-gen.$(CONTROLLER_GEN_VERSION):
	@rm -f tool/.controller-gen.*
	@mkdir -p tool
	touch $@


tool/crd-ref-docs: tool/.crd-ref-docs.$(CRD_REF_DOCS_VERSION)
	GOBIN=$(PWD)/tool go install github.com/elastic/crd-ref-docs@$(CRD_REF_DOCS_VERSION)

tool/.crd-ref-docs.$(CRD_REF_DOCS_VERSION):
	@rm -f tool/.crd-ref-docs.*
	@mkdir -p tool
	touch $@


tool/ko: tool/.ko.$(KO_VERSION)
	GOBIN=$(PWD)/tool go install github.com/google/ko@$(KO_VERSION)

tool/.ko.$(KO_VERSION):
	@rm -f tool/.ko.*
	@mkdir -p tool
	touch $@


tool/golangci-lint: tool/.golangci-lint.$(GOLANGCILINT_VERSION)
	GOBIN=$(PWD)/tool go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCILINT_VERSION)

tool/.golangci-lint.$(GOLANGCILINT_VERSION):
	@rm -f tool/.golangci-lint.*
	@mkdir -p tool
	touch $@


.PHONY: tool
tool: tool/controller-gen tool/crd-ref-docs tool/ko tool/golangci-lint

.PHONY: gendoc
gendoc: build-linux
	- rm -r docs/cli/*
	HOME=HOMEDIR ci-dist/telemetry/linux/amd64/bin/telemetry gendocs md --only-commands docs/cli/

.PHONY: apidoc
apidoc: $(addsuffix .md, $(addprefix docs/apis/config.telemetry.act3-ace.io/, v1alpha2))
docs/apis/%.md: tool/crd-ref-docs $(wildcard pkg/apis/$*/*_types.go) 
	@mkdir -p $(@D)
	tool/crd-ref-docs --config=apidocs.yaml --renderer=markdown --source-path=pkg/apis/$* --output-path=$@

.PHONY: clean-deps
clean-deps:
	- rm -rf $(ASSET_DIR)/.downloads
	- rm -rf $(ASSET_DIR)/static/libs

.PHONY: deps
deps: bootstrap bootstrap-icons leader-line mathjax requirejs

$(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist.zip:
	@mkdir -p $(ASSET_DIR)/.downloads
	curl -sSLo $(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist.zip \
		https://github.com/twbs/bootstrap/releases/download/v$(BOOTSTRAP_VERSION)/bootstrap-$(BOOTSTRAP_VERSION)-dist.zip

$(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/js/bootstrap.bundle.min.js \
$(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/js/bootstrap.bundle.min.js.map \
$(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/css/bootstrap.min.css \
$(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/css/bootstrap.min.css.map &: $(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist.zip
	- rm -rf $(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist
	unzip -q $(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist.zip -d $(ASSET_DIR)/.downloads

	- rm -rf $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)
	@mkdir -p $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/css $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/js
	cp \
		$(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist/css/bootstrap.min.css \
		$(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist/css/bootstrap.min.css.map \
		$(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/css
	cp $(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist/js/bootstrap.bundle.min.js $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/js
	cp $(ASSET_DIR)/.downloads/bootstrap-$(BOOTSTRAP_VERSION)-dist/js/bootstrap.bundle.min.js.map $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/js

.PHONY: bootstrap
bootstrap: $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/js/bootstrap.bundle.min.js
bootstrap: $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/js/bootstrap.bundle.min.js.map
bootstrap: $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/css/bootstrap.min.css
bootstrap: $(ASSET_DIR)/static/libs/bootstrap/$(BOOTSTRAP_VERSION)/css/bootstrap.min.css.map

$(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION).zip:
	@mkdir -p $(ASSET_DIR)/.downloads
	curl -sSLo $(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION).zip \
		https://github.com/twbs/icons/releases/download/v$(BOOTSTRAP_ICONS_VERSION)/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION).zip

$(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/fonts/bootstrap-icons.woff \
$(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/fonts/bootstrap-icons.woff2 \
$(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/bootstrap-icons.css &: $(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION).zip
	- rm -rf $(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION)
	unzip -q $(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION).zip -d $(ASSET_DIR)/.downloads

	- rm -rf $(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)
	@mkdir -p $(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/fonts
	cp \
		$(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION)/fonts/bootstrap-icons.woff \
		$(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION)/fonts/bootstrap-icons.woff2 \
		$(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/fonts
	cp $(ASSET_DIR)/.downloads/bootstrap-icons-$(BOOTSTRAP_ICONS_VERSION)/bootstrap-icons.css $(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/

.PHONY: boostrap-icons
bootstrap-icons: $(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/fonts/bootstrap-icons.woff
bootstrap-icons: $(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/fonts/bootstrap-icons.woff2
bootstrap-icons: $(ASSET_DIR)/static/libs/bootstrap-icons/$(BOOTSTRAP_ICONS_VERSION)/bootstrap-icons.css

$(ASSET_DIR)/.downloads/leader-line-$(LEADER_LINE_VERSION).tar.gz:
	@mkdir -p $(ASSET_DIR)/.downloads
	curl -sSLo $(ASSET_DIR)/.downloads/leader-line-$(LEADER_LINE_VERSION).tar.gz https://github.com/anseki/leader-line/archive/refs/tags/$(LEADER_LINE_VERSION).tar.gz

$(ASSET_DIR)/static/libs/leader-line/$(LEADER_LINE_VERSION)/leader-line.min.js: $(ASSET_DIR)/.downloads/leader-line-$(LEADER_LINE_VERSION).tar.gz
	- rm -rf $(ASSET_DIR)/.downloads/leader-line-$(LEADER_LINE_VERSION)
	tar -xf $(ASSET_DIR)/.downloads/leader-line-$(LEADER_LINE_VERSION).tar.gz -C $(ASSET_DIR)/.downloads

	- rm -rf $(ASSET_DIR)/static/libs/leader-line/$(LEADER_LINE_VERSION)
	@mkdir -p $(ASSET_DIR)/static/libs/leader-line/$(LEADER_LINE_VERSION)
	cp $(ASSET_DIR)/.downloads/leader-line-$(LEADER_LINE_VERSION)/leader-line.min.js $(ASSET_DIR)/static/libs/leader-line/$(LEADER_LINE_VERSION)/leader-line.min.js

.PHONY: leader-line
leader-line: $(ASSET_DIR)/static/libs/leader-line/$(LEADER_LINE_VERSION)/leader-line.min.js

$(ASSET_DIR)/.downloads/mathjax-$(MATHJAX_VERSION).tar.gz:
	@mkdir -p $(ASSET_DIR)/.downloads
	curl -sSLo $(ASSET_DIR)/.downloads/mathjax-$(MATHJAX_VERSION).tar.gz https://github.com/mathjax/MathJax/archive/$(MATHJAX_VERSION).tar.gz


$(ASSET_DIR)/static/libs/mathjax/$(MATHJAX_VERSION): $(ASSET_DIR)/.downloads/mathjax-$(MATHJAX_VERSION).tar.gz
	- rm -rf $(ASSET_DIR)/.downloads/MathJax-$(MATHJAX_VERSION)
	tar -xf $(ASSET_DIR)/.downloads/mathjax-$(MATHJAX_VERSION).tar.gz -C $(ASSET_DIR)/.downloads

	- rm -rf $(ASSET_DIR)/static/libs/mathjax/$(MATHJAX_VERSION)
	@mkdir -p $(ASSET_DIR)/static/libs/mathjax/$(MATHJAX_VERSION)
	cp -r $(ASSET_DIR)/.downloads/MathJax-$(MATHJAX_VERSION)/* $(ASSET_DIR)/static/libs/mathjax/$(MATHJAX_VERSION)/

.PHONY: mathjax
mathjax: $(ASSET_DIR)/static/libs/mathjax/$(MATHJAX_VERSION)

$(ASSET_DIR)/static/libs/requirejs/$(REQUIREJS_VERSION)/require.min.js:
	@mkdir -p $(ASSET_DIR)/static/libs/requirejs/$(REQUIREJS_VERSION)
	curl -sSLo $(ASSET_DIR)/static/libs/requirejs/$(REQUIREJS_VERSION)/require.min.js https://requirejs.org/docs/release/$(REQUIREJS_VERSION)/minified/require.js

.PHONY: requirejs
requirejs: $(ASSET_DIR)/static/libs/requirejs/$(REQUIREJS_VERSION)/require.min.js

.PHONY: swagger 
swagger: #template
	yq eval --inplace '.paths."/api/bottle".put.requestBody.content."application/json".examples.bottleJSON.value=load_str("testdata/bottle/bottle1.json")' swagger.yml
	yq eval --inplace '.paths."/api/manifest".put.requestBody.content."application/json".examples.manifestJSON.value=load_str("testdata/manifest/manifest1.json")' swagger.yml
	yq eval --inplace '.paths."/api/event".put.requestBody.content."application/json".examples.eventJSON.value=load_str("testdata/event/push1.json")' swagger.yml

.PHONY: siggen
siggen:
	openssl ecparam -name secp521r1 -genkey -noout -out testdata/signature/priv.pem
	openssl ec -in testdata/signature/priv.pem -pubout > testdata/signature/pub.pem
	openssl dgst -sha256 -sign testdata/signature/priv.pem -out testdata/signature/signature.raw testdata/signature/data-to-sign.txt
	openssl dgst -sha256 -verify testdata/signature/pub.pem -signature testdata/signature/signature.raw testdata/signature/data-to-sign.txt
	openssl base64 -in testdata/signature/signature.raw -out testdata/signature/signature.base64 -A
