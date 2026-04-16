package s3client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	signV4Algorithm = "AWS4-HMAC-SHA256"
	iso8601Format   = "20060102T150405Z"
	yyyymmdd        = "20060102"
	unsignedPayload = "UNSIGNED-PAYLOAD"
)

// Credentials holds S3 access and secret keys.
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

// sumHMAC returns the HMAC-SHA256 of data using the given key.
func sumHMAC(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// sha256Hex returns the hex-encoded SHA-256 hash of the given data.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// getSigningKey derives the AWS Signature V4 signing key.
// SigningKey = HMAC(HMAC(HMAC(HMAC("AWS4"+secret, date), region), "s3"), "aws4_request")
func getSigningKey(secretKey string, t time.Time, region string) []byte {
	dateKey := sumHMAC([]byte("AWS4"+secretKey), []byte(t.Format(yyyymmdd)))
	regionKey := sumHMAC(dateKey, []byte(region))
	serviceKey := sumHMAC(regionKey, []byte("s3"))
	signingKey := sumHMAC(serviceKey, []byte("aws4_request"))
	return signingKey
}

// getScope returns the credential scope string: date/region/s3/aws4_request
func getScope(t time.Time, region string) string {
	return strings.Join([]string{
		t.Format(yyyymmdd),
		region,
		"s3",
		"aws4_request",
	}, "/")
}

// getCanonicalHeaders builds the canonical headers string from signed headers.
func getCanonicalHeaders(headers http.Header, signedHeaders []string) string {
	var buf strings.Builder
	for _, h := range signedHeaders {
		vals := headers.Values(h)
		buf.WriteString(strings.ToLower(h))
		buf.WriteByte(':')
		buf.WriteString(strings.TrimSpace(strings.Join(vals, ",")))
		buf.WriteByte('\n')
	}
	return buf.String()
}

// getSignedHeadersList returns a sorted, semicolon-separated list of header names to sign.
func getSignedHeadersList(headers http.Header) []string {
	var keys []string
	for k := range headers {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)
	return keys
}

// getCanonicalQueryString URL-encodes query parameters in sorted order.
func getCanonicalQueryString(query url.Values) string {
	if len(query) == 0 {
		return ""
	}
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		vals := query[k]
		sort.Strings(vals)
		for _, v := range vals {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(parts, "&")
}

// getCanonicalRequest builds the canonical request per AWS Sig V4 spec.
func getCanonicalRequest(method, path string, query url.Values, headers http.Header, signedHeaders []string, payloadHash string) string {
	if path == "" {
		path = "/"
	}
	return strings.Join([]string{
		method,
		uriEncode(path, false),
		getCanonicalQueryString(query),
		getCanonicalHeaders(headers, signedHeaders),
		strings.Join(signedHeaders, ";"),
		payloadHash,
	}, "\n")
}

// getStringToSign builds the string to sign per AWS Sig V4 spec.
func getStringToSign(canonicalRequest string, t time.Time, scope string) string {
	return strings.Join([]string{
		signV4Algorithm,
		t.Format(iso8601Format),
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
}

// getSignature returns the hex-encoded HMAC-SHA256 signature.
func getSignature(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))
}

// SignRequest signs an HTTP request using AWS Signature Version 4.
// The request must already have the Host header set.
// payloadHash should be the SHA-256 hex hash of the body, or "UNSIGNED-PAYLOAD".
func SignRequest(req *http.Request, cred Credentials, payloadHash string) {
	now := time.Now().UTC()

	// Set required headers
	req.Header.Set("X-Amz-Date", now.Format(iso8601Format))
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.Host)
	}

	// Build signed headers list
	signedHeaders := getSignedHeadersList(req.Header)

	scope := getScope(now, cred.Region)

	// Build canonical request
	canonicalReq := getCanonicalRequest(
		req.Method,
		req.URL.Path,
		req.URL.Query(),
		req.Header,
		signedHeaders,
		payloadHash,
	)

	// Build string to sign
	stringToSign := getStringToSign(canonicalReq, now, scope)

	// Derive signing key and compute signature
	signingKey := getSigningKey(cred.SecretAccessKey, now, cred.Region)
	signature := getSignature(signingKey, stringToSign)

	// Build Authorization header
	auth := signV4Algorithm + " " +
		"Credential=" + cred.AccessKeyID + "/" + scope + ", " +
		"SignedHeaders=" + strings.Join(signedHeaders, ";") + ", " +
		"Signature=" + signature

	req.Header.Set("Authorization", auth)
}

// PresignURL generates a pre-signed URL for the given method, path, and expiry.
func PresignURL(method string, endpoint *url.URL, bucket, key string, cred Credentials, expires time.Duration) string {
	now := time.Now().UTC()
	scope := getScope(now, cred.Region)

	path := "/" + bucket
	if key != "" {
		path += "/" + key
	}

	query := url.Values{}
	query.Set("X-Amz-Algorithm", signV4Algorithm)
	query.Set("X-Amz-Credential", cred.AccessKeyID+"/"+scope)
	query.Set("X-Amz-Date", now.Format(iso8601Format))
	expSec := int(expires.Seconds())
	if expSec < 1 {
		expSec = 900
	}
	query.Set("X-Amz-Expires", formatInt(expSec))
	query.Set("X-Amz-SignedHeaders", "host")

	// Build canonical request for presigning (no Authorization header, no payload)
	headers := http.Header{}
	headers.Set("Host", endpoint.Host)

	signedHeaders := []string{"host"}

	canonicalReq := getCanonicalRequest(
		method,
		path,
		query,
		headers,
		signedHeaders,
		unsignedPayload,
	)

	stringToSign := getStringToSign(canonicalReq, now, scope)
	signingKey := getSigningKey(cred.SecretAccessKey, now, cred.Region)
	signature := getSignature(signingKey, stringToSign)

	query.Set("X-Amz-Signature", signature)

	u := *endpoint
	u.Path = path
	u.RawQuery = query.Encode()
	return u.String()
}

// uriEncode encodes a URI path component per AWS S3 spec.
func uriEncode(path string, encodeSep bool) string {
	var buf strings.Builder
	for _, ch := range []byte(path) {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || ch == '~' || ch == '.' {
			buf.WriteByte(ch)
		} else if ch == '/' && !encodeSep {
			buf.WriteByte('/')
		} else {
			buf.WriteString("%" + strings.ToUpper(hex.EncodeToString([]byte{ch})))
		}
	}
	return buf.String()
}

// formatInt converts an int to its string representation.
func formatInt(n int) string {
	return strconv.Itoa(n)
}
