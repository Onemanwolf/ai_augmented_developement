#!/bin/bash
# check-prerequisites.sh - Verify all required tools are installed with correct versions

set -e

echo "Checking prerequisites for E-Commerce Microservices Platform..."
echo "=============================================================="
echo ""

ERRORS=0

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_command() {
    local cmd=$1
    local min_version=$2
    local install_hint=$3

    if ! command -v "$cmd" &> /dev/null; then
        echo -e "${RED}[MISSING]${NC} $cmd not installed"
        echo "         Install: $install_hint"
        ERRORS=$((ERRORS + 1))
        return 1
    fi
    return 0
}

check_version() {
    local cmd=$1
    local version_cmd=$2
    local min_version=$3
    local install_hint=$4

    if ! check_command "$cmd" "$min_version" "$install_hint"; then
        return 1
    fi

    local actual_version
    actual_version=$(eval "$version_cmd" 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)

    if [ -z "$actual_version" ]; then
        echo -e "${YELLOW}[WARNING]${NC} $cmd installed but version could not be determined"
        return 0
    fi

    echo -e "${GREEN}[OK]${NC} $cmd: $actual_version (min: $min_version)"
    return 0
}

echo "Required Tools:"
echo "---------------"

# Go
check_version "go" "go version" "1.21" "https://go.dev/dl/"

# Docker
check_version "docker" "docker --version" "24.0" "https://docker.com"

# Docker Compose
if docker compose version &> /dev/null; then
    version=$(docker compose version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo -e "${GREEN}[OK]${NC} docker compose: $version (min: 2.20)"
elif command -v docker-compose &> /dev/null; then
    check_version "docker-compose" "docker-compose --version" "2.20" "Included with Docker Desktop"
else
    echo -e "${RED}[MISSING]${NC} docker compose not installed"
    echo "         Install: Included with Docker Desktop"
    ERRORS=$((ERRORS + 1))
fi

# kubectl
check_version "kubectl" "kubectl version --client -o yaml 2>/dev/null | grep gitVersion | head -1" "1.28" "brew install kubectl"

# Helm
check_version "helm" "helm version --short" "3.12" "brew install helm"

# Terraform
check_version "terraform" "terraform --version" "1.5" "brew install terraform"

# Azure CLI
check_version "az" "az --version | head -1" "2.50" "brew install azure-cli"

# golangci-lint
check_version "golangci-lint" "golangci-lint --version" "1.54" "brew install golangci-lint"

# pre-commit
check_version "pre-commit" "pre-commit --version" "3.3" "pip install pre-commit"

echo ""
echo "Optional Tools:"
echo "---------------"

# jq (for parsing tasks.json)
if check_command "jq" "1.6" "brew install jq"; then
    version=$(jq --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo -e "${GREEN}[OK]${NC} jq: $version"
fi

# mongosh (for MongoDB debugging)
if command -v mongosh &> /dev/null; then
    echo -e "${GREEN}[OK]${NC} mongosh installed"
else
    echo -e "${YELLOW}[OPTIONAL]${NC} mongosh not installed (useful for MongoDB debugging)"
fi

# istioctl (for Istio management)
if command -v istioctl &> /dev/null; then
    version=$(istioctl version --remote=false 2>/dev/null | head -1)
    echo -e "${GREEN}[OK]${NC} istioctl: $version"
else
    echo -e "${YELLOW}[OPTIONAL]${NC} istioctl not installed (needed for Phase 5.7)"
fi

# argocd CLI
if command -v argocd &> /dev/null; then
    echo -e "${GREEN}[OK]${NC} argocd CLI installed"
else
    echo -e "${YELLOW}[OPTIONAL]${NC} argocd CLI not installed (useful for GitOps management)"
fi

echo ""
echo "=============================================================="

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Prerequisites check FAILED: $ERRORS required tool(s) missing${NC}"
    echo "Please install the missing tools before proceeding."
    exit 1
else
    echo -e "${GREEN}All required prerequisites are installed!${NC}"
    exit 0
fi
