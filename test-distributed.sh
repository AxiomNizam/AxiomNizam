#!/bin/bash

echo "Running Go tests for distributed state store..."
echo ""

cd "$(dirname "$0")"

echo "=== Unit Tests ==="
go test -v ./internal/distributedstate/... -timeout 30s

echo ""
echo "=== Benchmarks ==="
go test -bench=. -benchmem ./internal/distributedstate/... -timeout 60s

echo ""
echo "=== Coverage ==="
go test -cover ./internal/distributedstate/...

echo ""
echo "✓ Testing complete"
