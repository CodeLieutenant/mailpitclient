#!/bin/bash

# E2E Test Runner for Mailpit Go API Client
echo "=============================="
echo "Mailpit Go API Client E2E Tests"
echo "=============================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to run a test and report results
run_test() {
    local test_name="$1"
    local test_pattern="$2"

    echo -e "${YELLOW}Running $test_name...${NC}"

    if go test -v -run "$test_pattern" -timeout 60s; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        return 1
    fi
}

# Build the project first
echo -e "${YELLOW}Building project...${NC}"
if ! go build ./...; then
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"
echo

# Counter for test results
passed=0
failed=0

echo "Running core functionality tests..."
echo "===================================="

# Test core features (should work in all versions)
if run_test "Core Features" "TestE2E_CoreFeatures"; then
    ((passed++))
else
    ((failed++))
fi
echo

# Test server operations
if run_test "Server Operations" "TestE2E_ServerOperations"; then
    ((passed++))
else
    ((failed++))
fi
echo

# Test basic message operations
if run_test "Message Operations" "TestE2E_MessageOperations"; then
    ((passed++))
else
    ((failed++))
fi
echo

# Test optional features (may fail in some versions)
echo "Running optional feature tests..."
echo "==================================="
echo -e "${YELLOW}Note: These tests may show 'not available' messages - this is expected${NC}"
echo

if run_test "Optional Features" "TestE2E_OptionalFeatures"; then
    ((passed++))
else
    ((failed++))
fi
echo

if run_test "Tag Operations (Optional)" "TestE2E_TagOperations"; then
    ((passed++))
else
    ((failed++))
fi
echo

if run_test "Send Operations (Optional)" "TestE2E_SendOperations"; then
    ((passed++))
else
    ((failed++))
fi
echo

# Test deletion operations
if run_test "Message Deletion" "TestE2E_MessageDeletion"; then
    ((passed++))
else
    ((failed++))
fi
echo

# Summary
echo "=============================="
echo "E2E Test Summary"
echo "=============================="
echo -e "Tests passed: ${GREEN}$passed${NC}"
echo -e "Tests failed: ${RED}$failed${NC}"
echo -e "Total tests: $((passed + failed))"
echo

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed. Check the output above for details.${NC}"
    echo -e "${YELLOW}Note: Some failures may be expected if testing optional features.${NC}"
    exit 1
fi
