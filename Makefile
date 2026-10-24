.PHONY: all check templ-fmt templ-check lint test generate new-module remove-module init-project

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
	go test ./...
	go test -C cmd/jrmp ./...

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
