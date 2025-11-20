#!/bin/bash
# Test runner script for PAM package manager

set -e # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 Running PAM test suite..."
echo ""

# Function to print colored output
print_status() {
  if [ $1 -eq 0 ]; then
    echo -e "${GREEN}✓${NC} $2"
  else
    echo -e "${RED}✗${NC} $2"
  fi
}

# Track overall status
OVERALL_STATUS=0

# Test each package individually for better visibility
echo "Testing internal packages..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test nixconfig
echo "📦 Testing internal/nixconfig..."
if go test -v ./internal/nixconfig 2>&1 | tee /tmp/nixconfig_test.log; then
  print_status 0 "internal/nixconfig tests passed"
else
  print_status 1 "internal/nixconfig tests failed"
  OVERALL_STATUS=1
fi
echo ""

# Test search
echo "🔍 Testing internal/search..."
if go test -v ./internal/search 2>&1 | tee /tmp/search_test.log; then
  print_status 0 "internal/search tests passed"
else
  print_status 1 "internal/search tests failed"
  OVERALL_STATUS=1
fi
echo ""

# Test assets
echo "📄 Testing internal/assets..."
if go test -v ./internal/assets 2>&1 | tee /tmp/assets_test.log; then
  print_status 0 "internal/assets tests passed"
else
  print_status 1 "internal/assets tests failed"
  OVERALL_STATUS=1
fi
echo ""

# Test UI
echo "🎨 Testing internal/ui..."
if go test -v ./internal/ui 2>&1 | tee /tmp/ui_test.log; then
  print_status 0 "internal/ui tests passed"
else
  print_status 1 "internal/ui tests failed"
  OVERALL_STATUS=1
fi
echo ""

# Test setup
echo "⚙️  Testing internal/setup..."
if go test -v ./internal/setup 2>&1 | tee /tmp/setup_test.log; then
  print_status 0 "internal/setup tests passed"
else
  print_status 1 "internal/setup tests failed"
  OVERALL_STATUS=1
fi
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Run all tests with coverage
echo "📊 Generating coverage report..."
if go test -coverprofile=coverage.out ./internal/... >/dev/null 2>&1; then
  COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
  echo -e "${GREEN}Total coverage: $COVERAGE${NC}"

  # Generate HTML coverage report
  go tool cover -html=coverage.out -o coverage.html
  echo "📈 HTML coverage report generated: coverage.html"
else
  echo -e "${YELLOW}⚠ Coverage report generation failed${NC}"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $OVERALL_STATUS -eq 0 ]; then
  echo -e "${GREEN}✓ All tests passed!${NC}"
else
  echo -e "${RED}✗ Some tests failed${NC}"
  echo ""
  echo "Test logs saved to /tmp/*_test.log"
  echo "Fix the failing tests and run './test.sh' again"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

exit $OVERALL_STATUS
