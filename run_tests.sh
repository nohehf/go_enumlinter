#!/bin/bash

echo "🧪 Running Enum Linter Tests with Go Testing Framework"
echo "======================================================"

# Build the linter first
echo "📦 Building linter..."
go build -o enumlinter cmd/main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"
echo ""

# Run the Go tests
echo "🔍 Running Go tests..."
cd pkg/analyzer && go test -v

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ All tests passed!"
else
    echo ""
    echo "❌ Some tests failed"
    exit 1
fi 