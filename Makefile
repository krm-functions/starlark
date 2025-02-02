KO_DOCKER_REPO ?= ko.local
STARLARK ?= ./starlark
STARLARK_IMAGE ?= $(KO_DOCKER_REPO)/starlark:latest

.EXPORT_ALL_VARIABLES:

.PHONY: build
build:
	go build -o starlark starlark.go

.PHONY: lint
lint:
	golangci-lint run -v  --timeout 10m

.PHONY: container
container:
	ko build --base-import-paths	

.PHONY: test-bin
test-bin:
	rm -rf _tmp _results
	kpt fn source examples | kpt fn eval - --truncate-output=false --exec $(STARLARK) --fn-config example-function-config/set-annotation.yaml | kpt fn sink _tmp
	make do-tests

.PHONY: test-container
test-container:
	rm -rf _tmp _results
	kpt fn source examples | kpt fn eval --results-dir _results - --image $(STARLARK_IMAGE) --fn-config example-function-config/set-annotation.yaml | kpt fn sink _tmp
	make do-tests

.PHONY: do-tests
do-tests:
	cat _tmp/deployment.yaml | yq '.metadata.annotations.foo' | grep bar
