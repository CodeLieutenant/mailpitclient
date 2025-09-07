#!/bin/bash

# API Coverage Maintenance Script
# This script helps maintain the API coverage test by providing utilities
# to check coverage and generate stubs for missing methods.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to run the API coverage test
run_coverage_test() {
    log_info "Running API coverage test..."
    cd "$PROJECT_DIR"

    if go test -v -run TestAPIRouteCoverage -timeout=3m; then
        log_success "API coverage test passed!"
        return 0
    else
        log_error "API coverage test failed!"
        return 1
    fi
}

# Function to run only the offline coverage test (faster)
run_offline_coverage_test() {
    log_info "Running offline API coverage test..."
    cd "$PROJECT_DIR"

    if go test -v -run TestAPIRouteCoverageOffline -timeout=30s; then
        log_success "Offline API coverage test passed!"
        return 0
    else
        log_error "Offline API coverage test failed!"
        return 1
    fi
}

# Function to update the Mailpit OpenAPI spec URL in the test
update_spec_url() {
    local new_url="$1"
    if [ -z "$new_url" ]; then
        log_error "Please provide a new URL for the OpenAPI specification"
        echo "Usage: $0 update-spec-url <new-url>"
        exit 1
    fi

    log_info "Updating OpenAPI specification URL to: $new_url"

    sed -i.bak "s|mailpitSwaggerURL = \".*\"|mailpitSwaggerURL = \"$new_url\"|" \
        "$PROJECT_DIR/e2e_api_coverage_test.go"

    log_success "Updated OpenAPI specification URL"
    log_info "Backup saved as e2e_api_coverage_test.go.bak"
}

# Function to generate method stubs for missing routes
generate_method_stubs() {
    log_info "This feature will be implemented to generate method stubs for missing routes"
    log_warning "For now, check the test output to see which routes are missing"
    log_info "Example stub generation would create methods like:"
    echo ""
    echo "// RenameTag renames a tag"
    echo "func (c *client) RenameTag(ctx context.Context, oldTag, newTag string) error {"
    echo "    // Implementation needed"
    echo "    return errors.New(\"not implemented\")"
    echo "}"
    echo ""
}

# Function to check if all imports are available
check_dependencies() {
    log_info "Checking Go dependencies..."
    cd "$PROJECT_DIR"

    if go mod tidy; then
        log_success "Go dependencies are up to date"
    else
        log_error "Failed to update Go dependencies"
        return 1
    fi

    if go mod verify; then
        log_success "Go module verification passed"
    else
        log_error "Go module verification failed"
        return 1
    fi
}

# Function to show usage information
show_usage() {
    echo "Mailpit API Coverage Maintenance Script"
    echo ""
    echo "Usage: $0 <command> [arguments]"
    echo ""
    echo "Commands:"
    echo "  test              Run the full API coverage test"
    echo "  test-offline      Run the offline API coverage test (faster)"
    echo "  update-spec-url   Update the OpenAPI specification URL"
    echo "  generate-stubs    Generate method stubs for missing routes"
    echo "  check-deps        Check and update Go dependencies"
    echo "  help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 test"
    echo "  $0 test-offline"
    echo "  $0 update-spec-url https://example.com/new-swagger.json"
    echo "  $0 generate-stubs"
    echo "  $0 check-deps"
}

# Main script logic
case "${1:-help}" in
    "test")
        run_coverage_test
        ;;
    "test-offline")
        run_offline_coverage_test
        ;;
    "update-spec-url")
        update_spec_url "$2"
        ;;
    "generate-stubs")
        generate_method_stubs
        ;;
    "check-deps")
        check_dependencies
        ;;
    "help"|"--help"|"-h")
        show_usage
        ;;
    *)
        log_error "Unknown command: $1"
        echo ""
        show_usage
        exit 1
        ;;
esac
