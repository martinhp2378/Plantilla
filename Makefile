.PHONY: all check templ-fmt templ-check lint test generate new-module remove-module init-project vulncheck trivy security

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
	cd cmd/jrmp && golangci-lint run -c ../../.golangci.yml ./...

# Run automated tests
test:
	go test -mod=vendor ./...
	go test -C cmd/jrmp ./...

# Run govulncheck across project and cmd/jrmp
vulncheck:
	@echo "==> Scanning root project with govulncheck..."
	@(which govulncheck >/dev/null 2>&1 && govulncheck ./... || go run golang.org/x/vuln/cmd/govulncheck@latest ./...) || true
	@echo "==> Scanning cmd/jrmp with govulncheck..."
	@(cd cmd/jrmp && (which govulncheck >/dev/null 2>&1 && govulncheck ./... || go run golang.org/x/vuln/cmd/govulncheck@latest ./...)) || true

# Run Trivy security scanner (uses local trivy if available, otherwise docker)
trivy:
	@if which trivy >/dev/null 2>&1; then \
		trivy fs --skip-dirs vendor --severity HIGH,CRITICAL .; \
	elif docker info >/dev/null 2>&1; then \
		docker run --rm -v $$(pwd):/app -w /app aquasec/trivy:latest fs --skip-dirs vendor --severity HIGH,CRITICAL .; \
	else \
		echo "Warning: neither trivy nor docker is available to run trivy scan"; \
	fi

# Run all security scans (govulncheck + trivy)
security: vulncheck trivy

# Initialize project module name: make init-project MODULE=MiNuevoProyecto
init-project:
	@go run -C cmd/jrmp . init -module="$(MODULE)"

# Generate a new CRUD module: make new-module NAME=proveedor [SCHEMA=sistema] [TITLE="Proveedores"]
new-module:
	@go run -C cmd/jrmp . new-module -name="$(NAME)" $(if $(SCHEMA),-schema="$(SCHEMA)",) $(if $(TITLE),-title="$(TITLE)",)

# Remove a scaffolded CRUD module: make remove-module NAME=proveedor
remove-module:
	@go run -C cmd/jrmp . remove-module -name="$(NAME)"

# Full verification: Templ check, Go linter, and unit tests
check: templ-check lint test
