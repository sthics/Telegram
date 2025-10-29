#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)    OS="macos";;
        Linux*)     OS="linux";;
        CYGWIN*|MINGW*|MSYS*) OS="windows";;
        *)          OS="unknown";;
    esac
    log_info "Detected OS: $OS"
}

# Install Homebrew (macOS)
install_homebrew() {
    if ! command -v brew &> /dev/null; then
        log_info "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

        # Add Homebrew to PATH for Apple Silicon
        if [[ $(uname -m) == "arm64" ]]; then
            echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
            eval "$(/opt/homebrew/bin/brew shellenv)"
        fi

        log_success "Homebrew installed"
    else
        log_success "Homebrew already installed"
    fi
}

# Install Go
install_go() {
    if command -v go &> /dev/null; then
        CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        REQUIRED_VERSION="1.22"

        if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$CURRENT_VERSION" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
            log_success "Go $CURRENT_VERSION already installed (>= $REQUIRED_VERSION)"
            return
        else
            log_warning "Go $CURRENT_VERSION found, but version >= $REQUIRED_VERSION required"
        fi
    fi

    log_info "Installing Go 1.22..."

    if [ "$OS" = "macos" ]; then
        brew install go@1.22 || brew install go
        log_success "Go installed via Homebrew"
    elif [ "$OS" = "linux" ]; then
        # Download and install Go for Linux
        GO_VERSION="1.22.0"
        wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
        rm /tmp/go.tar.gz

        # Add to PATH
        if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
            echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
        fi
        export PATH=$PATH:/usr/local/go/bin

        log_success "Go $GO_VERSION installed"
    else
        log_error "Please install Go manually from https://go.dev/dl/"
        exit 1
    fi
}

# Install Docker
install_docker() {
    if command -v docker &> /dev/null; then
        log_success "Docker already installed ($(docker --version))"
        return
    fi

    log_info "Installing Docker..."

    if [ "$OS" = "macos" ]; then
        log_info "Installing Docker Desktop for Mac..."

        # Check if running on Apple Silicon or Intel
        if [[ $(uname -m) == "arm64" ]]; then
            log_info "Detected Apple Silicon Mac"
            curl -L "https://desktop.docker.com/mac/main/arm64/Docker.dmg" -o /tmp/Docker.dmg
        else
            log_info "Detected Intel Mac"
            curl -L "https://desktop.docker.com/mac/main/amd64/Docker.dmg" -o /tmp/Docker.dmg
        fi

        # Mount and install
        sudo hdiutil attach /tmp/Docker.dmg
        sudo /Volumes/Docker/Docker.app/Contents/MacOS/install
        sudo hdiutil detach /Volumes/Docker
        rm /tmp/Docker.dmg

        log_success "Docker Desktop installed. Please start Docker Desktop manually."
        log_warning "After starting Docker Desktop, run this script again to continue."

    elif [ "$OS" = "linux" ]; then
        # Install Docker on Linux
        log_info "Installing Docker on Linux..."

        # Update package index
        sudo apt-get update -qq

        # Install prerequisites
        sudo apt-get install -y -qq \
            apt-transport-https \
            ca-certificates \
            curl \
            gnupg \
            lsb-release

        # Add Docker's official GPG key
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

        # Set up stable repository
        echo \
          "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
          $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

        # Install Docker Engine
        sudo apt-get update -qq
        sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io

        # Add user to docker group
        sudo usermod -aG docker $USER

        log_success "Docker installed. You may need to log out and back in for group changes to take effect."

    else
        log_error "Please install Docker manually from https://www.docker.com/get-started"
        exit 1
    fi
}

# Install Docker Compose
install_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        log_success "Docker Compose already installed ($(docker-compose --version))"
        return
    fi

    # Check if docker compose (v2) is available
    if docker compose version &> /dev/null; then
        log_success "Docker Compose v2 already available via Docker CLI"
        # Create alias for docker-compose
        if [ "$OS" = "macos" ]; then
            echo 'alias docker-compose="docker compose"' >> ~/.zshrc
        else
            echo 'alias docker-compose="docker compose"' >> ~/.bashrc
        fi
        return
    fi

    log_info "Installing Docker Compose..."

    if [ "$OS" = "macos" ]; then
        brew install docker-compose
        log_success "Docker Compose installed via Homebrew"
    elif [ "$OS" = "linux" ]; then
        # Install Docker Compose v2
        COMPOSE_VERSION="2.24.0"
        sudo curl -L "https://github.com/docker/compose/releases/download/v${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
        log_success "Docker Compose v${COMPOSE_VERSION} installed"
    fi
}

# Install OpenSSL (usually pre-installed)
check_openssl() {
    if command -v openssl &> /dev/null; then
        log_success "OpenSSL already installed ($(openssl version))"
    else
        log_info "Installing OpenSSL..."
        if [ "$OS" = "macos" ]; then
            brew install openssl
        elif [ "$OS" = "linux" ]; then
            sudo apt-get install -y openssl
        fi
        log_success "OpenSSL installed"
    fi
}

# Install Go tools
install_go_tools() {
    log_info "Installing Go development tools..."

    # goose (database migrations)
    if ! command -v goose &> /dev/null; then
        log_info "Installing goose..."
        go install github.com/pressly/goose/v3/cmd/goose@latest
        log_success "goose installed"
    else
        log_success "goose already installed"
    fi

    # buf (protobuf)
    if ! command -v buf &> /dev/null; then
        log_info "Installing buf..."
        go install github.com/bufbuild/buf/cmd/buf@latest
        log_success "buf installed"
    else
        log_success "buf already installed"
    fi

    # golangci-lint (optional, for linting)
    if ! command -v golangci-lint &> /dev/null; then
        log_info "Installing golangci-lint..."
        if [ "$OS" = "macos" ]; then
            brew install golangci-lint
        elif [ "$OS" = "linux" ]; then
            curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
        fi
        log_success "golangci-lint installed"
    else
        log_success "golangci-lint already installed"
    fi

    # Add GOPATH/bin to PATH if not already there
    GOPATH=$(go env GOPATH)
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
        if [ "$OS" = "macos" ]; then
            echo "export PATH=\$PATH:$GOPATH/bin" >> ~/.zshrc
        else
            echo "export PATH=\$PATH:$GOPATH/bin" >> ~/.bashrc
        fi
        export PATH=$PATH:$GOPATH/bin
        log_info "Added $GOPATH/bin to PATH"
    fi
}

# Install k6 (load testing)
install_k6() {
    if command -v k6 &> /dev/null; then
        log_success "k6 already installed"
        return
    fi

    log_info "Installing k6 for load testing..."

    if [ "$OS" = "macos" ]; then
        brew install k6
        log_success "k6 installed via Homebrew"
    elif [ "$OS" = "linux" ]; then
        sudo gpg -k
        sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
        echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
        sudo apt-get update
        sudo apt-get install k6
        log_success "k6 installed"
    fi
}

# Verify Docker is running
verify_docker_running() {
    log_info "Verifying Docker is running..."

    if ! docker ps &> /dev/null; then
        log_warning "Docker daemon is not running."
        if [ "$OS" = "macos" ]; then
            log_info "Starting Docker Desktop..."
            open -a Docker
            log_info "Waiting for Docker to start (this may take a minute)..."

            # Wait for Docker to be ready
            for i in {1..30}; do
                if docker ps &> /dev/null; then
                    log_success "Docker is now running"
                    return
                fi
                sleep 2
                echo -n "."
            done

            log_error "Docker failed to start. Please start Docker Desktop manually."
            exit 1
        else
            log_error "Please start Docker daemon: sudo systemctl start docker"
            exit 1
        fi
    else
        log_success "Docker daemon is running"
    fi
}

# Create directory structure
create_directories() {
    log_info "Creating necessary directories..."
    mkdir -p secrets
    mkdir -p bin
    log_success "Directories created"
}

# Summary
print_summary() {
    echo ""
    echo "=============================================="
    log_success "All dependencies installed successfully!"
    echo "=============================================="
    echo ""
    echo "Installed tools:"
    echo "  - Go:             $(go version | awk '{print $3}')"
    echo "  - Docker:         $(docker --version | awk '{print $3}' | sed 's/,//')"
    echo "  - Docker Compose: $(docker-compose --version 2>/dev/null | awk '{print $4}' || docker compose version --short)"
    echo "  - OpenSSL:        $(openssl version | awk '{print $2}')"
    echo "  - goose:          $(goose --version 2>/dev/null || echo 'installed')"
    echo "  - buf:            $(buf --version 2>/dev/null || echo 'installed')"
    echo "  - k6:             $(k6 version 2>/dev/null | head -n1 || echo 'installed')"
    echo ""
    echo "Next steps:"
    echo "  1. Source your shell config to update PATH:"
    if [ "$OS" = "macos" ]; then
        echo "     source ~/.zshrc"
    else
        echo "     source ~/.bashrc"
    fi
    echo ""
    echo "  2. Run the setup script:"
    echo "     cd $(pwd)"
    echo "     ./scripts/setup.sh"
    echo ""
    echo "  3. Or start development immediately:"
    echo "     make deps          # Download Go dependencies"
    echo "     make docker-up     # Start infrastructure services"
    echo "     make migrate-up    # Run database migrations"
    echo "     make run-gateway   # Start the gateway service"
    echo ""
    echo "=============================================="
}

# Main installation flow
main() {
    echo ""
    echo "=============================================="
    echo "  Mini-Telegram Dependency Installer"
    echo "=============================================="
    echo ""

    detect_os

    if [ "$OS" = "macos" ]; then
        install_homebrew
    fi

    install_go
    check_openssl
    install_docker
    install_docker_compose
    verify_docker_running
    install_go_tools
    install_k6
    create_directories

    print_summary
}

# Run main function
main
