#!/bin/bash

echo "Building test CLI image..."
docker build -f Dockerfile.test-cli -t axiom-test-cli .

echo "Running distributed store tests..."

echo ""
echo "=== Test 1: In-Memory Store ==="
docker run --rm axiom-test-cli -cmd all -store memory

echo ""
echo "=== Test 2: etcd with Docker Network ==="
docker run --rm --network axiom-nizam_app-network axiom-test-cli -cmd all -store etcd -etcd etcd:2379

echo ""
echo "✓ All tests completed!"
