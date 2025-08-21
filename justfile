# Default recipe - show available commands
default:
    @just --list

# Variables for platform detection
goos := env_var_or_default('GOOS', `go env GOOS`)
goarch := env_var_or_default('GOARCH', `go env GOARCH`)

# Build the application with platform auto-detection
build:
    @echo "Building for {{goos}}/{{goarch}}..."
    GOOS={{goos}} GOARCH={{goarch}} go build -o builds/gim .
    @echo "Built gim for {{goos}}/{{goarch}}"

# Install the application
install:
    @echo "Installing for {{goos}}/{{goarch}}..."
    @if [ "{{goos}}" = "linux" ]; then \
        echo "Linux detected. Building and copying to ~/.local/bin..."; \
        just build; \
        mkdir -p ~/.local/bin; \
        cp builds/gim ~/.local/bin/; \
        echo "Copied gim to ~/.local/bin/"; \
    else \
        echo "Non-Linux detected. Installing via go install..."; \
        GOOS={{goos}} GOARCH={{goarch}} go install github.com/nicholasflintwillow/github-issue-manager@latest; \
        echo "Installed gim for {{goos}}/{{goarch}}"; \
    fi
# Build for specific platform (optional convenience recipes)
build-linux:
    GOOS=linux GOARCH=amd64 go build -o builds/gim-linux-amd64 .

build-windows:
    GOOS=windows GOARCH=amd64 go build -o builds/gim-windows-amd64.exe .

build-darwin:
    GOOS=darwin GOARCH=amd64 go build -o builds/gim-darwin-amd64 .

# Build for all major platforms
build-all: build-linux build-windows build-darwin
    @echo "Built for all major platforms"

# Clean build artifacts
clean:
    rm -f builds/gim builds/gim-* builds/*.exe