@echo off
REM AxiomNizam Distributed State Store Testing Guide for Windows

echo === AxiomNizam Distributed State Store Testing Guide ===
echo.

echo SETUP:
echo 1. You already have etcd in docker-compose.yml
echo 2. etcd client is already in go.mod (go.etcd.io/etcd/client/v3 v3.5.9)
echo.

echo QUICK START - Option 1: Unit Tests (Local)
echo ==========================================
echo Run all unit tests:
echo   cd c:\Users\office\Documents\AxiomNizam\AxiomNizam
echo   go test -v ./internal/distributedstate/...
echo.
echo Run specific test:
echo   go test -v -run TestDistributedCounter ./internal/distributedstate/...
echo.
echo Run with coverage:
echo   go test -cover ./internal/distributedstate/...
echo.
echo Run benchmarks:
echo   go test -bench=. -benchmem ./internal/distributedstate/...
echo.

echo QUICK START - Option 2: CLI Testing (Local)
echo =========================================
echo Test with in-memory store:
echo   cd c:\Users\office\Documents\AxiomNizam\AxiomNizam
echo   go run cmd/axiomnizamctl/test_cli.go -cmd all -store memory
echo.
echo Test specific commands:
echo   go run cmd/axiomnizamctl/test_cli.go -cmd lock-test -store memory
echo   go run cmd/axiomnizamctl/test_cli.go -cmd counter-test -store memory
echo   go run cmd/axiomnizamctl/test_cli.go -cmd cas-test -store memory
echo.
echo Verbose output:
echo   go run cmd/axiomnizamctl/test_cli.go -cmd all -v
echo.

echo QUICK START - Option 3: Docker Testing
echo ====================================
echo First, ensure docker-compose services are running:
echo   docker-compose up -d etcd mysql postgres
echo.
echo Build test CLI image:
echo   docker build -f Dockerfile.test-cli -t axiom-test-cli .
echo.
echo Run tests with in-memory store:
echo   docker run --rm axiom-test-cli -cmd all -store memory
echo.
echo Run all tests (in-memory and etcd):
echo   run-tests.bat
echo.
echo Run specific test with etcd:
echo   docker run --rm --network axiom-nizam_app-network axiom-test-cli -cmd lock-test -store etcd -etcd etcd:2379
echo.

echo AVAILABLE TEST COMMANDS:
echo =======================
echo   store-test     - Basic Put/Get/Delete/List operations
echo   cas-test       - Compare-and-Swap atomic operations
echo   counter-test   - Distributed counter with increment/decrement
echo   set-test       - Distributed set with add/remove/contains
echo   queue-test     - Distributed queue with enqueue/dequeue
echo   lock-test      - Distributed locks
echo   election-test  - Leader election
echo   cache-test     - Cached store layer
echo   all            - Run all tests
echo   help           - Show help
echo.

echo ETCD INFORMATION:
echo =================
echo Port: 2379 (client)
echo Port: 2380 (peer)
echo Container name: etcd
echo Network: app-network
echo.
echo Check etcd status:
echo   docker exec etcd etcdctl --endpoints=localhost:2379 member list
echo   docker exec etcd etcdctl --endpoints=localhost:2379 get "" --prefix
echo.

echo TROUBLESHOOTING:
echo ===============
echo.
echo If etcd connection fails:
echo   1. Check docker-compose is running: docker ps ^| findstr etcd
echo   2. Check network: docker network ls ^| findstr app-network
echo   3. Check etcd logs: docker logs etcd
echo   4. Try connecting directly: etcdctl --endpoints=localhost:2379 member list
echo.
echo If go build fails:
echo   1. Run: go mod tidy
echo   2. Run: go mod download
echo   3. Clean: go clean -modcache
echo.

echo INTEGRATION EXAMPLE:
echo ===================
echo.
echo Use in your code:
echo.
echo   package main
echo   import (
echo     "context"
echo     "example.com/axiomnizam/internal/distributedstate"
echo   )
echo.
echo   store, _ := distributedstate.NewEtcdStateStore([]string{"localhost:2379"})
echo   manager := distributedstate.NewDistributedManager(store, "myapp")
echo.
echo   ctx := context.Background()
echo   manager.PutState(ctx, "config/db", "postgres")
echo   value, _ := manager.GetState(ctx, "config/db")
echo.

echo.
echo ✓ Setup complete!
