package handlers

import "testing"

func TestNormalizeSQLPolicyMode(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{in: "", want: "", wantOK: true},
		{in: "compat", want: "compat", wantOK: true},
		{in: "STRICT", want: "strict", wantOK: true},
		{in: "  compat  ", want: "compat", wantOK: true},
		{in: "invalid", want: "", wantOK: false},
	}

	for _, tc := range cases {
		got, ok := normalizeSQLPolicyMode(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("normalizeSQLPolicyMode(%q) = (%q,%v), want (%q,%v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestStrictReadOnlyQueryPolicyModes(t *testing.T) {
	if !isStrictReadOnlyQuery("SELECT id FROM users WHERE id = ?", "compat") {
		t.Fatalf("expected compat to allow basic SELECT")
	}

	if isStrictReadOnlyQuery("SELECT * FROM users; DELETE FROM users", "compat") {
		t.Fatalf("expected stacked statements to be rejected")
	}

	if isStrictReadOnlyQuery("INSERT INTO users(id) VALUES (?)", "compat") {
		t.Fatalf("expected write statement to be rejected")
	}

	query := "SELECT * FROM users INTO OUTFILE '/tmp/users.csv'"
	if !isStrictReadOnlyQuery(query, "compat") {
		t.Fatalf("expected compat mode to allow legacy read-like template")
	}
	if isStrictReadOnlyQuery(query, "strict") {
		t.Fatalf("expected strict mode to reject INTO OUTFILE")
	}
}
