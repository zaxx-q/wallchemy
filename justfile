# Justfile for wallchemy

default:
    @just --list

version := `git describe --tags --always 2>/dev/null || echo dev`
flags := "-s -w -X main.version=" + version
prefix := env_var_or_default("PREFIX", env_var("HOME") / ".local/bin")

# Build local binary
build:
    go build -ldflags '{{flags}}' -o wallchemy .

# Run test suite
test:
    go test ./...

# Build and install to ~/.local/bin (or custom PREFIX)
install: test
    @mkdir -p "{{prefix}}"
    go build -ldflags '{{flags}}' -o "{{prefix}}/wallchemy" .
    @echo "Installed wallchemy ({{version}}) to {{prefix}}/wallchemy"

# Install system-wide to /usr/local/bin (requires sudo)
install-system: test
    sudo go build -ldflags '{{flags}}' -o /usr/local/bin/wallchemy .
    @echo "Installed wallchemy ({{version}}) to /usr/local/bin/wallchemy"

# Clean local build artifacts
clean:
    rm -f wallchemy
