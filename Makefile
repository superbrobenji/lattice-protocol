.PHONY: generate check test

generate:
	go generate ./...

# Verify generated C headers are up to date with the Go constants.
# Run this in CI after any change to opcodes/ or adapter/.
check: generate
	git diff --exit-code c/ proto/

test:
	go test ./...
