.PHONY: build test test-race vet fmt check tools download-specs generate regenerate clean help

# Build all packages (confirms everything compiles)
build:
	go build ./...

# Run all unit tests
test:
	go test ./...

# Run tests with race detector
test-race:
	go test ./... -race

# go vet
vet:
	go vet ./...

# Format
fmt:
	go fmt ./...

# fmt + vet + build + test-race
check: fmt vet build test-race

# Install tool dependencies
tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0

# Download OpenAPI specs from dev.proof.com and apply fixup scripts
download-specs:
	@mkdir -p openapi
	@echo "Downloading Business API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-business-api-specification.json" -o openapi/business.json
	@echo "Downloading Real Estate API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-real-estate-api-specification.json" -o openapi/realestate.json
	@echo "Downloading SCIM API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-scim-api-specification.json" -o openapi/scim.json
	@echo "Downloading Logs API spec..."
	@curl -sL "https://dev.proof.com/openapi/proof-logs-api-specification.json" -o openapi/logs.json
	@echo "Downloading Certificates API spec..."
	@curl -sL "https://dev.proof.com/openapi/organization-certificates-openapi-specification.json" -o openapi/certificates.json
	@echo "Fixing deep \$$ref references in specs..."
	@python3 scripts/fix-openapi-refs.py openapi/business.json
	@python3 scripts/fix-openapi-refs.py openapi/realestate.json
	@python3 scripts/fix-openapi-refs.py openapi/scim.json
	@python3 scripts/fix-openapi-refs.py openapi/logs.json
	@python3 scripts/fix-openapi-refs.py openapi/certificates.json
	@echo "Fixing SCIM operationIds..."
	@python3 scripts/fix-scim-operation-ids.py openapi/scim.json
	@echo "All specs downloaded and fixed!"

# oapi-codegen command - use go run to avoid PATH issues
OAPI_CODEGEN = go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0

# Generate SDK clients from OpenAPI specs
generate:
	@echo "Generating Business SDK..."
	@$(OAPI_CODEGEN) --config business/oapi-codegen.yaml openapi/business.json
	@echo "Generating Real Estate SDK..."
	@$(OAPI_CODEGEN) --config realestate/oapi-codegen.yaml openapi/realestate.json
	@echo "Generating SCIM SDK..."
	@$(OAPI_CODEGEN) --config scim/oapi-codegen.yaml openapi/scim.json
	@echo "Generating Logs SDK..."
	@$(OAPI_CODEGEN) --config logs/oapi-codegen.yaml openapi/logs.json
	@echo "Generating Certificates SDK..."
	@$(OAPI_CODEGEN) --config certificates/oapi-codegen.yaml openapi/certificates.json
	@echo "SDK generation complete!"

# Download specs, regenerate all SDKs, then build + test
regenerate: download-specs generate build test

# Clean openapi specs and generated clients (everything regenerate would recreate)
clean:
	rm -rf openapi/
	rm -f */client.gen.go

help:
	@echo "Available targets:"
	@echo "  build          Build all packages"
	@echo "  test           Run all tests"
	@echo "  test-race      Run tests with -race"
	@echo "  vet            Run go vet"
	@echo "  fmt            Run go fmt"
	@echo "  check          fmt + vet + build + test-race"
	@echo "  tools          Install oapi-codegen tool dependency"
	@echo "  download-specs Download OpenAPI specs and apply fixups"
	@echo "  generate       Regenerate SDK clients from OpenAPI specs"
	@echo "  regenerate     download-specs + generate + build + test"
	@echo "  clean          Remove openapi/ and generated client files"
