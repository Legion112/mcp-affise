GO ?= $(shell if [ -x "$(HOME)/sdk/go1.27.0/bin/go" ]; then echo "$(HOME)/sdk/go1.27.0/bin/go"; else echo go; fi)
export GOTOOLCHAIN ?= go1.27.0

BINARY := bin/mcp-affise
GEN_BINARY := bin/affise-gen
OAPI_CODEGEN ?= $(shell $(GO) env GOPATH)/bin/oapi-codegen
DOCS_HTML ?= internal/docsgen/testdata/docs3.1.html
DOCS_URL ?= https://api.affise.com/docs3.1/
PATHS ?= /3.0/offers,/3.0/offer/{ID}
COVERPROFILE := coverage.out

.PHONY: build build-gen test vet fmt lint clean \
	fetch-docs openapi generate generate-check capture-sample

build: generate
	mkdir -p bin
	$(GO) build -o $(BINARY) ./cmd/mcp-affise

build-gen:
	mkdir -p bin
	$(GO) build -o $(GEN_BINARY) ./cmd/affise-gen

fetch-docs:
	mkdir -p internal/docsgen/testdata
	curl -fsSL "$(DOCS_URL)" -o "$(DOCS_HTML)"

openapi: build-gen
	$(GEN_BINARY) openapi \
		-i "$(DOCS_HTML)" \
		-o openapi/affise.gen.yaml \
		--overlay openapi/overlays.yaml \
		--final openapi/affise.yaml \
		--paths "$(PATHS)"

$(OAPI_CODEGEN):
	$(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0

generate: openapi $(OAPI_CODEGEN)
	mkdir -p internal/affiseapi
	$(OAPI_CODEGEN) -generate types,client -package affiseapi -o internal/affiseapi/client.gen.go openapi/affise.yaml

generate-check: generate
	git diff --exit-code -- openapi/affise.gen.yaml openapi/affise.yaml internal/affiseapi/client.gen.go

capture-sample: build-gen
	@test -n "$(TENANT)" || (echo "TENANT= required" >&2; exit 1)
	@test -n "$(PATH)" || (echo "PATH= required e.g. PATH=/3.0/offers" >&2; exit 1)
	$(GEN_BINARY) capture-sample --tenant "$(TENANT)" --path "$(PATH)" --out openapi/captures

test:
	$(GO) test -race -count=1 ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

clean:
	rm -rf bin $(COVERPROFILE) coverage.html
