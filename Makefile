.PHONY: all check templ-fmt templ-check lint test generate

all: check

# Format all .templ files
templ-fmt:
	templ fmt .

# Verify .templ formatting and compilation integrity
templ-check:
	templ fmt -fail .
	templ generate

# Run golangci-lint
lint:
	golangci-lint run ./...

# Run automated tests
test:
	go test ./...

# Generate a new CRUD module: make new-module NAME=proveedor [SCHEMA=sistema] [TITLE="Proveedores"]
new-module:
	@go run cmd/tools/scaffold/main.go -name="$(NAME)" $(if $(SCHEMA),-schema="$(SCHEMA)",) $(if $(TITLE),-title="$(TITLE)",)

# Full verification: Templ check, Go linter, and unit tests
check: templ-check lint test
