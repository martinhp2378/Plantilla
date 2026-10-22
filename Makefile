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

# Full verification: Templ check, Go linter, and unit tests
check: templ-check lint test
