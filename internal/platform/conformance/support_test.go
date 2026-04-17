package conformance

import (
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the current working directory to find the
// module root (the directory containing go.mod).  This lets the test
// run both via `go test ./...` from the repo root and via `go test`
// from the package directory itself.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found starting from %s", dir)
		}
		dir = parent
	}
}

// skipTests is a go/parser filter that excludes *_test.go files so the
// conformance checks inspect production code only.
func skipTests(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

// exprText renders an ast.Expr back to source text for cheap textual
// matching.  Good enough for "does this field mention clientv3.Client?".
func exprText(e ast.Expr) string {
	var sb strings.Builder
	_ = printer.Fprint(&sb, token.NewFileSet(), e)
	return sb.String()
}
