package sqlfilter

import "strings"

// Classify returns the QueryClass for a SQL keyword.
func Classify(keyword string) QueryClass {
	kw := strings.ToUpper(strings.TrimSpace(keyword))
	if kw == "" {
		return ClassUnknown
	}

	if _, ok := readKeywords[kw]; ok {
		return ClassRead
	}
	if _, ok := writeKeywords[kw]; ok {
		return ClassWrite
	}
	if _, ok := ddlKeywords[kw]; ok {
		return ClassDDL
	}
	if _, ok := controlKeywords[kw]; ok {
		return ClassControl
	}
	if _, ok := execKeywords[kw]; ok {
		return ClassExec
	}

	return ClassUnknown
}

var readKeywords = map[string]struct{}{
	"SELECT":   {},
	"WITH":     {},
	"SHOW":     {},
	"DESCRIBE": {},
	"DESC":     {},
	"EXPLAIN":  {},
	"TABLE":    {}, // MySQL 8 TABLE statement (read-only)
	"VALUES":   {}, // MySQL 8 VALUES statement
}

var writeKeywords = map[string]struct{}{
	"INSERT":  {},
	"UPDATE":  {},
	"DELETE":  {},
	"REPLACE": {},
	"UPSERT":  {},
	"MERGE":   {},
}

var ddlKeywords = map[string]struct{}{
	"CREATE":   {},
	"ALTER":    {},
	"DROP":     {},
	"TRUNCATE": {},
	"RENAME":   {},
	"COMMENT":  {},
}

var controlKeywords = map[string]struct{}{
	"BEGIN":     {},
	"START":     {},
	"COMMIT":    {},
	"ROLLBACK":  {},
	"SAVEPOINT": {},
	"RELEASE":   {},
	"GRANT":     {},
	"REVOKE":    {},
	"SET":       {},
	"USE":       {},
	"LOCK":      {},
	"UNLOCK":    {},
}

var execKeywords = map[string]struct{}{
	"CALL":    {},
	"EXEC":    {},
	"EXECUTE": {},
	"DO":      {},
}
