#!/bin/sh

# Run initial test
echo "=== Running Initial Tests ==="
axiomnizamctl -cmd all -store memory
echo ""
echo "=== Tests Complete - Container Running ==="
echo "Use: axiomnizamctl [options]"
echo "Examples:"
echo "  axiomnizamctl -cmd lock-test -store memory"
echo "  axiomnizamctl -cmd counter-test -store memory"
echo "  axiomnizamctl -cmd help"
echo ""

# Keep container alive
tail -f /dev/null
