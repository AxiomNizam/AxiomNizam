package sqlfilter

import "strings"

// containsBlockedKeywords checks if the query contains any blocked SQL keywords.
func containsBlockedKeywords(query string, mode PolicyMode) bool {
	normalized := strings.ToUpper(strings.Join(strings.Fields(query), " "))
	padded := " " + normalized + " "

	for _, blocked := range commonBlocklist {
		if strings.Contains(padded, blocked) {
			return true
		}
	}

	if mode == PolicyStrict {
		for _, blocked := range strictBlocklist {
			if strings.Contains(padded, blocked) {
				return true
			}
		}
	}

	return false
}

// commonBlocklist — always blocked regardless of mode.
var commonBlocklist = []string{
	" UPDATE ",
	" DELETE ",
	" DROP ",
	" ALTER ",
	" TRUNCATE ",
	" CREATE ",
	" REPLACE ",
	" MERGE ",
	" GRANT ",
	" REVOKE ",
	" CALL ",
	" EXEC ",
	" EXECUTE ",
	" UPSERT ",
	" DO ",
	" SET ",
	" USE ",
}

// strictBlocklist — blocked only in strict mode.
var strictBlocklist = []string{
	" FOR UPDATE ",
	" LOCK IN SHARE MODE ",
	" INTO OUTFILE ",
	" INTO DUMPFILE ",
	" LOAD DATA ",
	" COPY ",
}

// legacyReadOnlyHeuristic is the fallback for unclassified queries
// in compat mode. Returns true if the query looks safe.
func legacyReadOnlyHeuristic(query string) bool {
	upper := strings.ToUpper(strings.TrimSpace(query))
	if upper == "" {
		return false
	}

	// If it starts with something that looks like a read, allow it.
	readPrefixes := []string{"SELECT", "WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN"}
	for _, prefix := range readPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}

	return false
}
