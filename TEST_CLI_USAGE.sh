#!/bin/bash

# Quick commands for testing distributed state store

echo "=== AxiomNizam Test CLI Commands ==="
echo ""

# Test 1: Run with in-memory store only
echo "1. Test with in-memory store (no etcd needed):"
echo "   docker-compose --profile test run --rm test-cli -cmd all -store memory"
echo ""

# Test 2: Run with etcd
echo "2. Test with etcd integration:"
echo "   docker-compose --profile test up etcd"
echo "   docker-compose --profile test run --rm test-cli -cmd all -store etcd -etcd etcd:2379"
echo ""

# Test 3: Specific tests
echo "3. Run specific tests:"
echo "   docker-compose --profile test run --rm test-cli -cmd store-test -store memory"
echo "   docker-compose --profile test run --rm test-cli -cmd lock-test -store memory"
echo "   docker-compose --profile test run --rm test-cli -cmd counter-test -store memory"
echo "   docker-compose --profile test run --rm test-cli -cmd election-test -store etcd -etcd etcd:2379"
echo ""

# Test 4: Help
echo "4. Show help:"
echo "   docker-compose --profile test run --rm test-cli -cmd help"
echo ""

# Test 5: Run all services including test
echo "5. Run full stack with tests:"
echo "   docker-compose --profile test up -d"
echo "   docker-compose --profile test run test-cli -cmd all -store etcd -etcd etcd:2379"
echo ""

# Test 6: View logs
echo "6. View test logs:"
echo "   docker-compose --profile test logs test-cli"
echo ""

# Test 7: Cleanup
echo "7. Stop and remove containers:"
echo "   docker-compose --profile test down"
echo ""

echo "=== Available Commands ==="
echo "  store-test     - Basic store operations"
echo "  cas-test       - Compare-and-swap"
echo "  counter-test   - Distributed counter"
echo "  set-test       - Distributed set"
echo "  queue-test     - Distributed queue"
echo "  lock-test      - Distributed locks"
echo "  election-test  - Leader election"
echo "  cache-test     - Cached store"
echo "  all            - All tests"
echo "  help           - Show help"
echo ""
