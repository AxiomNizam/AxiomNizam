@echo off
REM AxiomNizam CLI Helper for Windows

echo === AxiomNizam Distributed State Store CLI ===
echo.

if "%1"=="" (
    echo Usage: cli.bat [command] [store-type]
    echo.
    echo Commands:
    echo   all              - Run all tests
    echo   store-test       - Basic store operations
    echo   cas-test         - Compare-and-swap
    echo   counter-test     - Distributed counter
    echo   set-test         - Distributed set
    echo   queue-test       - Distributed queue
    echo   lock-test        - Distributed locks
    echo   election-test    - Leader election
    echo   cache-test       - Cached store
    echo   help             - Show help
    echo.
    echo Store types: memory (default), etcd
    echo.
    echo Examples:
    echo   cli.bat all
    echo   cli.bat lock-test memory
    echo   cli.bat counter-test etcd
    echo.
    exit /b 1
)

set CMD=%1
set STORE=%2
if "%STORE%"=="" set STORE=memory

echo Running: %CMD% with store: %STORE%
echo.

docker exec axiomnizamctl ./cli -cmd %CMD% -store %STORE%

pause
