package storage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestGenerateAndValidatePresignedSignature(t *testing.T) {
	secret := "test-secret-key"
	urlPath, err := GeneratePresignedURLWithHost(http.MethodGet, "tenant-bucket", "folder/object.txt", 5*time.Minute, "AXAK123", secret, "example.com")
	if err != nil {
		t.Fatalf("GeneratePresignedURLWithHost() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, urlPath, nil)
	req.Host = "example.com"

	if !isPresignedRequest(req) {
		t.Fatalf("expected request to be detected as presigned")
	}
	if isExpired(req) {
		t.Fatalf("expected presigned request to be unexpired")
	}
	if err := ValidatePresignedSignature(req, secret); err != nil {
		t.Fatalf("ValidatePresignedSignature() error = %v", err)
	}
}

func TestValidatePresignedSignatureRejectsMethodTampering(t *testing.T) {
	secret := "test-secret-key"
	urlPath, err := GeneratePresignedURLWithHost(http.MethodGet, "tenant-bucket", "folder/object.txt", 5*time.Minute, "AXAK123", secret, "example.com")
	if err != nil {
		t.Fatalf("GeneratePresignedURLWithHost() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, urlPath, nil)
	req.Host = "example.com"

	if err := ValidatePresignedSignature(req, secret); err == nil {
		t.Fatalf("expected signature validation failure for method tampering")
	}
}

func TestValidatePresignedSignatureRejectsPathTampering(t *testing.T) {
	secret := "test-secret-key"
	urlPath, err := GeneratePresignedURLWithHost(http.MethodGet, "tenant-bucket", "folder/object.txt", 5*time.Minute, "AXAK123", secret, "example.com")
	if err != nil {
		t.Fatalf("GeneratePresignedURLWithHost() error = %v", err)
	}

	u, err := url.Parse(urlPath)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	u.Path = "/api/v1/storage/buckets/tenant-bucket/objects/folder/other.txt"

	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	req.Host = "example.com"

	if err := ValidatePresignedSignature(req, secret); err == nil {
		t.Fatalf("expected signature validation failure for path tampering")
	}
}

func TestIsExpired(t *testing.T) {
	u := &url.URL{Path: "/api/v1/storage/buckets/b/objects/k"}
	q := u.Query()
	q.Set("X-Amz-Algorithm", sigV4Algorithm)
	q.Set("X-Amz-Credential", "AXAK123/20260410/us-east-1/s3/aws4_request")
	q.Set("X-Amz-Signature", "deadbeef")
	q.Set("X-Amz-Date", "20260410T000000Z")
	q.Set("X-Amz-Expires", "60")
	u.RawQuery = q.Encode()

	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	if !isExpired(req) {
		t.Fatalf("expected request to be expired")
	}
}
