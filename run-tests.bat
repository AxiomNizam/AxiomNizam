@echo off
REM Build test CLI image
echo Building test CLI image...
docker build -f Dockerfile.test-cli -t axiom-test-cli .

if errorlevel 1 (
    echo Build failed
    exit /b 1
)

echo.
echo === Test 1: In-Memory Store ===
docker run --rm axiom-test-cli -cmd all -store memory

echo.
echo === Test 2: Basic Store Operations ===
docker run --rm axiom-test-cli -cmd store-test -store memory

echo.
echo === Test 3: Compare-and-Swap ===
docker run --rm axiom-test-cli -cmd cas-test -store memory

echo.
echo === Test 4: Distributed Counter ===
docker run --rm axiom-test-cli -cmd counter-test -store memory

echo.
echo === Test 5: Distributed Lock ===
docker run --rm axiom-test-cli -cmd lock-test -store memory

echo.
echo === Test 6: Leader Election ===
docker run --rm axiom-test-cli -cmd election-test -store memory

echo.
echo === Test 7: With Docker Network (etcd) ===
docker run --rm --network axiom-nizam_app-network axiom-test-cli -cmd all -store etcd -etcd etcd:2379

echo.
echo ✓ All tests completed!
