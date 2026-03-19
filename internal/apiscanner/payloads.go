package apiscanner

var defaultSQLPayloads = []string{
	"' OR '1'='1",
	"\" OR \"1\"=\"1",
	"1 OR 1=1",
	"'; DROP TABLE users;--",
	"'; WAITFOR DELAY '0:0:5'--",
	"1' OR '1'='1",
	"admin'--",
	"' UNION SELECT NULL--",
	"' UNION SELECT NULL,NULL--",
}

var defaultNoSQLPayloads = []string{
	"{$ne:null}",
	"{$gt:''}",
	"{$or:[{},{}]}",
	"{$where:'sleep(1000)'}",
	"{$regex:'.*'}",
	"{\"$ne\": null}",
	"{\"$gt\": \"\"}",
	"{\"$where\": \"return true\"}",
}

var defaultXSSPayloads = []string{
	"<script>alert('XSS')</script>",
	"\"><script>alert('XSS')</script>",
	"<img src=x onerror=alert('XSS')>",
	"<svg onload=alert('XSS')>",
	"<body onload=alert('XSS')>",
	"<iframe srcdoc=\"<script>alert('XSS')</script>\"></iframe>",
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
	"unterminated quoted string",
	"syntax error",
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
	"cast to objectid failed",
	"unknown top level operator",
}
