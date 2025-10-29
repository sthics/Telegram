#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "=============================================="
echo "  Mini-Telegram Dependency Checker"
echo "=============================================="
echo ""

MISSING_DEPS=0

# Check Go
echo -n "Checking Go... "
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.22"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
        echo -e "${GREEN}✓${NC} Go $GO_VERSION (>= 1.22)"
    else
        echo -e "${YELLOW}⚠${NC} Go $GO_VERSION found, but >= 1.22 required"
        MISSING_DEPS=1
    fi
else
    echo -e "${RED}✗${NC} Not installed"
    MISSING_DEPS=1
fi

# Check Docker
echo -n "Checking Docker... "
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    echo -e "${GREEN}✓${NC} Docker $DOCKER_VERSION"

    # Check if Docker daemon is running
    if docker ps &> /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} Docker daemon is running"
    else
        echo -e "  ${YELLOW}⚠${NC} Docker daemon is not running"
        MISSING_DEPS=1
    fi
else
    echo -e "${RED}✗${NC} Not installed"
    MISSING_DEPS=1
fi

# Check Docker Compose
echo -n "Checking Docker Compose... "
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version | awk '{print $4}' | sed 's/,//')
    echo -e "${GREEN}✓${NC} Docker Compose $COMPOSE_VERSION"
elif docker compose version &> /dev/null; then
    COMPOSE_VERSION=$(docker compose version --short)
    echo -e "${GREEN}✓${NC} Docker Compose v2 $COMPOSE_VERSION"
else
    echo -e "${RED}✗${NC} Not installed"
    MISSING_DEPS=1
fi

# Check OpenSSL
echo -n "Checking OpenSSL... "
if command -v openssl &> /dev/null; then
    OPENSSL_VERSION=$(openssl version | awk '{print $2}')
    echo -e "${GREEN}✓${NC} OpenSSL $OPENSSL_VERSION"
else
    echo -e "${RED}✗${NC} Not installed"
    MISSING_DEPS=1
fi

# Check goose
echo -n "Checking goose (migrations)... "
if command -v goose &> /dev/null; then
    echo -e "${GREEN}✓${NC} Installed"
else
    echo -e "${YELLOW}⚠${NC} Not installed (optional, needed for migrations)"
fi

# Check buf
echo -n "Checking buf (protobuf)... "
if command -v buf &> /dev/null; then
    echo -e "${GREEN}✓${NC} Installed"
else
    echo -e "${YELLOW}⚠${NC} Not installed (optional, needed for protobuf generation)"
fi

# Check k6
echo -n "Checking k6 (load testing)... "
if command -v k6 &> /dev/null; then
    echo -e "${GREEN}✓${NC} Installed"
else
    echo -e "${YELLOW}⚠${NC} Not installed (optional, needed for load testing)"
fi

# Check golangci-lint
echo -n "Checking golangci-lint... "
if command -v golangci-lint &> /dev/null; then
    echo -e "${GREEN}✓${NC} Installed"
else
    echo -e "${YELLOW}⚠${NC} Not installed (optional, for code linting)"
fi

# Check Make
echo -n "Checking Make... "
if command -v make &> /dev/null; then
    echo -e "${GREEN}✓${NC} Installed"
else
    echo -e "${YELLOW}⚠${NC} Not installed (optional, but recommended)"
fi

# Check git
echo -n "Checking Git... "
if command -v git &> /dev/null; then
    GIT_VERSION=$(git --version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Git $GIT_VERSION"
else
    echo -e "${YELLOW}⚠${NC} Not installed"
fi

# Check for JWT keys
echo -n "Checking JWT keys... "
if [ -f "secrets/es256.key" ]; then
    echo -e "${GREEN}✓${NC} Found at secrets/es256.key"
else
    echo -e "${YELLOW}⚠${NC} Not found (run 'make generate-keys' or setup.sh)"
fi

# Check for .env file
echo -n "Checking .env file... "
if [ -f ".env" ]; then
    echo -e "${GREEN}✓${NC} Found"
else
    echo -e "${YELLOW}⚠${NC} Not found (copy from .env.example)"
fi

echo ""
echo "=============================================="

if [ $MISSING_DEPS -eq 0 ]; then
    echo -e "${GREEN}All critical dependencies are installed!${NC}"
    echo ""
    echo "You're ready to start development. Run:"
    echo "  ./scripts/setup.sh       # Complete setup"
    echo "  make docker-up           # Start services"
    echo "  make run-gateway         # Run gateway"
else
    echo -e "${RED}Some dependencies are missing!${NC}"
    echo ""
    echo "To install all dependencies automatically, run:"
    echo "  ./scripts/install-dependencies.sh"
    echo ""
    echo "Or install manually:"
    echo "  Go 1.22+:        https://go.dev/dl/"
    echo "  Docker:          https://www.docker.com/get-started"
    echo "  Docker Compose:  https://docs.docker.com/compose/install/"
fi

echo "=============================================="
echo ""

exit $MISSING_DEPS
