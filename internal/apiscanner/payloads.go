package apiscanner

var defaultSQLPayloads = []string{
	"' OR '1'='1",
	"'; DROP TABLE users;--",
	"1' OR '1'='1",
	"admin'--",
	"' UNION SELECT NULL--",
}

var defaultNoSQLPayloads = []string{
	"{$ne:null}",
	"{$gt:''}",
	"{$or:[{},{}]}",
	"{$where:'sleep(1000)'}",
	"{$regex:'.*'}",
}

var defaultXSSPayloads = []string{
	"<script>alert('XSS')</script>",
	"\"><script>alert('XSS')</script>",
	"<img src=x onerror=alert('XSS')>",
	"<svg onload=alert('XSS')>",
	"javascript:alert('XSS')",
}

var sqlErrorSignatures = []string{
	"sql syntax",
	"syntax error at or near",
	"unclosed quotation mark",
	"mysql",
	"postgresql",
	"sqlite",
	"ora-",
	"odbc",
	"sqlstate",
}

var noSQLErrorSignatures = []string{
	"mongoerror",
	"mongodb",
	"bson",
	"invalid operator",
	"cannot deserialize",
	"e11000",
	"$where",
	"nosql",
}
