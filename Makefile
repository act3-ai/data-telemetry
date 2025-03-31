
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

.PHONY: cover
cover:
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

.PHONY: install
install:
	go install ./cmd/telemetry

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

.PHONY: siggen
siggen:
	openssl ecparam -name secp521r1 -genkey -noout -out testdata/signature/priv.pem
	openssl ec -in testdata/signature/priv.pem -pubout > testdata/signature/pub.pem
	openssl dgst -sha256 -sign testdata/signature/priv.pem -out testdata/signature/signature.raw testdata/signature/data-to-sign.txt
	openssl dgst -sha256 -verify testdata/signature/pub.pem -signature testdata/signature/signature.raw testdata/signature/data-to-sign.txt
	openssl base64 -in testdata/signature/signature.raw -out testdata/signature/signature.base64 -A
