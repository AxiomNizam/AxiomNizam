# AxiomNizam — Code-Level Security Audit

**Date:** 2026-04-29  
**Scope:** All Go backend code (`internal/`, `main.go`) + frontend (`frontend/templates/*.js`)  
**Auditor:** Platform Architecture Team  
**Codebase:** ~200K+ lines Go, ~30 JS files

---

## Executive Summary

A code-level security audit of the AxiomNizam platform identified **38 findings** across 9 categories. The most critical issues are hardcoded credentials in the `.env` file, XSS vectors in the frontend via `innerHTML`, and command injection risk in the certificate handler. The platform has solid foundations (rate limiting, JWT auth, RBAC, encryption at rest) but needs hardening in credential management, frontend output encoding, and input validation.

| Severity | Count | Category |
|----------|-------|----------|
| CRITICAL | 3 | Hardcoded credentials, RSA key in .env, demo accounts |
| HIGH | 7 | XSS via innerHTML, MD5/SHA1 usage, weak token fallback, command injection risk |
| MEDIUM | 12 | localStorage tokens, missing CSRF, SSRF risk, unbounded reads, error info leaks |
| LOW | 8 | math/rand usage, missing security headers, path traversal risk |
| INFO | 8 | Positive findings (rate limiting, RBAC, encryption, sanitization functions) |

---

## Positive Security Controls (Already Present)

Before listing vulnerabilities, these controls are already in place:

| Control | Location | Status |
|---------|----------|--------|
| JWT authentication | `internal/auth/`, `main.go` | ✅ Active on all API routes |
| Rate limiting | `internal/auth/rate_limiter.go`, `main.go` | ✅ Per-token rate limits with configurable policy |
| RBAC | `internal/rbac/` | ✅ Role-based access with RoleBinding resources |
| Encryption at rest | `internal/encryption/` | ✅ AES-256-GCM with key rotation |
| Field-level encryption | `internal/encryption/` | ✅ Per-field encryption for sensitive data |
| Audit logging | `internal/audit/` | ✅ Compliance-grade audit trail |
| CORS configuration | `main.go` | ✅ Origin whitelist with configurable allowed origins |
| Input sanitization (frontend) | `esc()`, `escHtml()`, `escapeHtml()` | ✅ Present in conductor, object-storage, admin, netintel dashboards |
| SQL injection prevention | `internal/quality/rules/sanitize.go` | ✅ Identifier validation for quality engine |
| TLS support | `internal/handlers/certificate_handler.go` | ✅ Certificate monitoring and renewal |
| Multi-tenant isolation | `internal/tenant/` | ✅ Tenant-scoped resource access |

---

## CRITICAL Findings

### SEC-01: Hardcoded Database Credentials in .env (CRITICAL)

**Location:** `.env`

All database passwords are default/weak values stored in plaintext:

| Database | Password |
|----------|----------|
| MySQL | `root` |
| MariaDB | `root` |
| PostgreSQL | `postgres` |
| MongoDB | `root` |
| Oracle | `oracle123` |
| Percona | `root` |

**Risk:** If `.env` is committed to version control, leaked via CI logs, or accessible on a compromised host, all databases are immediately compromised.

**Remediation:**
- Use a secrets manager (HashiCorp Vault, AWS Secrets Manager, or Kubernetes Secrets)
- Generate strong random passwords (32+ chars)
- Ensure `.env` is in `.gitignore` (it is) and never committed
- Use `IAM_RSA_PRIVATE_KEY_FILE` instead of inline `IAM_RSA_PRIVATE_KEY`

---

### SEC-02: RSA Private Key Embedded in .env (CRITICAL)

**Location:** `.env` — `IAM_RSA_PRIVATE_KEY` field

The full RSA private key used for JWT signing is embedded directly in the environment file. If this key is compromised, an attacker can forge valid JWT tokens for any user/role.

**Remediation:**
- Move to file-based key loading: `IAM_RSA_PRIVATE_KEY_FILE=/etc/axiomnizam/jwt.key`
- Set file permissions to `0600` (owner read only)
- Rotate the key and invalidate all existing tokens

---

### SEC-03: Hardcoded Demo Accounts with Admin Access (CRITICAL)

**Location:** `internal/handlers/auth_handler.go` (lines ~1032-1035)

Demo accounts with hardcoded passwords are embedded in source code:

```
admin/admin       → admin role
sysadmin/sysadmin → system-manager role
manager/manager   → manager role
user/user         → user role
```

**Risk:** These credentials are accessible to anyone with code access. If demo mode is not explicitly disabled in production, any user can authenticate as admin.

**Remediation:**
- Gate demo accounts behind `DEMO_MODE=true` environment variable (default: false)
- Log a startup warning when demo mode is enabled
- Never enable demo mode in production deployments
- Add a health check that flags demo mode as a security risk

---

## HIGH Findings

### SEC-04: XSS via innerHTML in Frontend Dashboards (HIGH)

**Location:** Multiple frontend JS files

The frontend extensively uses `.innerHTML = ...` to render dynamic content. While sanitization functions (`esc()`, `escHtml()`, `escapeHtml()`) exist and are used in many places, the pattern is error-prone — any missed call creates an XSS vector.

**Affected files (highest innerHTML count):**

| File | innerHTML assignments | Has sanitizer |
|------|---------------------|---------------|
| `admin.js` | ~50+ | `escapeHtml()` — used inconsistently |
| `object-storage.js` | ~25+ | `escHtml()` — used consistently |
| `conductor-dashboard.js` | ~15+ | `esc()` — used consistently |
| `iam-admin.js` | ~20+ | `escapeHtml()` — used in most places |
| `admin-dashboard.js` | ~10+ | Limited sanitization |
| `cdc-etl-dashboard.js` | ~10+ | Mixed |

**Risk:** If any API response contains user-controlled data that isn't sanitized before innerHTML assignment, an attacker can inject JavaScript that executes in the context of an authenticated admin session.

**Remediation:**
- Adopt `textContent` for plain text, `innerHTML` only for trusted HTML
- Implement a Content Security Policy (CSP) header that blocks inline scripts
- Audit every `innerHTML` assignment to verify sanitization
- Consider migrating to a framework with automatic escaping (React, Vue, Svelte)

---

### SEC-05: MD5 Hash Usage for Security-Adjacent Operations (HIGH)

**Location:** Multiple files

| File | Usage | Risk |
|------|-------|------|
| `internal/utils/hash/hash.go:40` | MD5 as supported hash algorithm | Collision attacks |
| `internal/cache/middleware.go:39` | MD5 for cache key generation | Cache poisoning |
| `internal/storage/native/native.go:601` | MD5 for ETag generation | ETag collision |
| `internal/utils/bot_protection.go:348` | MD5 for request fingerprinting | Fingerprint bypass |
| `internal/utils/uuid/uuid.go:213` | MD5 for config ID generation | ID collision |

**Remediation:** Replace MD5 with SHA-256 for all new code. For cache keys and ETags where collision resistance is less critical, MD5 can remain but should be documented as non-security usage.

---

### SEC-06: SHA1 Hash Usage (HIGH)

**Location:** `internal/utils/hash/hash.go:42,67,115`

SHA1 is supported as a hash algorithm option. While HMAC-SHA1 is still considered safe, bare SHA1 is vulnerable to collision attacks.

**Remediation:** Deprecate SHA1 option. Default to SHA-256. Log a warning if SHA1 is selected.

---

### SEC-07: Weak JWT Secret Fallback (HIGH)

**Location:** `internal/auth/auth.go` (lines ~265-280)

If the `JWT_SECRET` environment variable is not set, the code falls back to a process-ephemeral secret derived from `time.Now().UnixNano()`. This is predictable and not cryptographically random.

**Remediation:**
- Fail startup if `JWT_SECRET` is not set (no fallback)
- Or generate a cryptographically random fallback using `crypto/rand` and log a warning

---

### SEC-08: Command Injection Risk in Certificate Handler (HIGH)

**Location:** `internal/handlers/certificate_handler.go` (lines 252, 317)

The certificate renewal handler executes shell commands via `exec.Command()`. The command template comes from `CERT_RENEW_COMMAND` env var, and the target comes from user input (request body). While `buildRenewCommand()` likely sanitizes the target, the pattern of executing commands with user-influenced arguments is inherently risky.

**Remediation:**
- Validate the target against a strict allowlist of known certificate targets
- Never pass user input directly to shell commands
- Use a dedicated certificate renewal library instead of shell execution
- Log all command executions to the audit trail

---

### SEC-09: Custom SQL Execution in Quality Rules (HIGH)

**Location:** `internal/quality/rules/engine.go` (line ~214)

The `custom_sql` rule type allows users to define arbitrary SQL queries that are executed against datasources. While `ValidateRuleInputs()` checks for dangerous DML keywords, a determined attacker could bypass keyword filtering.

**Remediation:**
- Execute custom SQL with a read-only database connection
- Set a query timeout (already partially done via context)
- Add a `QUALITY_CUSTOM_SQL_ENABLED=false` feature flag (default: disabled)
- Log all custom SQL executions to the audit trail

---

## MEDIUM Findings

### SEC-10: Auth Tokens in localStorage (MEDIUM)

**Location:** `frontend/templates/auth.js` (lines 213-218, 244-246, 334-338)

JWT access tokens and refresh tokens are stored in `localStorage`, which is accessible to any JavaScript running on the page. If an XSS vulnerability exists (see SEC-04), tokens can be exfiltrated.

**Remediation:**
- Store tokens in HttpOnly cookies (not accessible to JavaScript)
- Use `SameSite=Strict` and `Secure` cookie flags
- Keep only a short-lived session identifier in localStorage if needed

---

### SEC-11: Missing CSRF Protection (MEDIUM)

**Location:** Frontend-wide

No CSRF tokens are used in form submissions or state-changing API requests. The platform relies on JWT Bearer tokens in the `Authorization` header, which provides some CSRF protection (browsers don't auto-send custom headers). However, any endpoint that also accepts cookies for auth is vulnerable.

**Remediation:**
- If cookie-based auth is used, implement CSRF tokens
- Ensure all state-changing requests require the `Authorization` header (not cookies)
- Add `SameSite=Strict` to any auth cookies

---

### SEC-12: SSRF Risk in OAuth Configuration (MEDIUM)

**Location:** `internal/handlers/auth_handler.go` (lines ~776, 853, 1728, 1798)

OAuth token URLs, userinfo URLs, and discovery endpoints are constructed from identity provider configuration. If an admin can configure arbitrary identity providers, they could point these URLs at internal services.

**Remediation:**
- Validate OAuth URLs against a scheme whitelist (https only)
- Block requests to private IP ranges (10.x, 172.16-31.x, 192.168.x, 127.x, ::1)
- Set a short timeout on OAuth HTTP requests (already partially done)

---

### SEC-13: Unbounded Request Body Reading (MEDIUM)

**Location:** Various handlers

Most handlers use Gin's `ShouldBindJSON` which has a default 32MB limit. However, some endpoints use `io.ReadAll` directly:

| File | Limit |
|------|-------|
| `internal/handlers/api_builder_handler.go:3074` | 2MB ✅ |
| `internal/apiscanner/openapi.go:115` | 8MB ✅ |
| `internal/handlers/auth_handler.go:789` | No explicit limit ⚠️ |

**Remediation:** Add `http.MaxBytesReader` wrapper to all `io.ReadAll` calls.

---

### SEC-14: Error Messages Leak Internal Details (MEDIUM)

**Location:** `main.go` (lines ~406, 413), various handlers

Error responses include full error messages that may reveal internal implementation details:

```go
c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
```

**Remediation:**
- Return generic error messages to clients: `"authentication failed"`
- Log detailed errors server-side only
- Never include stack traces, file paths, or internal error types in API responses

---

### SEC-15: Sensitive Fields in API Responses (MEDIUM)

**Location:** `internal/iam/admin/admin.go`, `internal/storage/admin/admin.go`

Client secrets and access keys are returned in API responses without masking.

**Remediation:**
- Mask client secrets in list/get responses (show only last 4 chars)
- Only return full secrets on creation (one-time display)
- Never log secrets

---

### SEC-16: Missing Authentication on Health Endpoints (MEDIUM)

**Location:** `main.go` (line ~526)

The `/health/reconcilers` endpoint exposes internal reconciler metrics without authentication. This reveals:
- Which reconcilers are active
- Error counts and states
- Internal module names

**Remediation:**
- Require authentication for detailed health endpoints
- Keep only a simple `/health` (returns 200 OK) as unauthenticated

---

### SEC-17: Privilege Escalation via Email Pattern Matching (MEDIUM)

**Location:** `internal/handlers/auth_handler.go` (lines ~1031-1037)

The `deriveOAuthRole()` function assigns roles based on email patterns. A federated user could claim an admin role if their email matches the pattern.

**Remediation:**
- Never derive roles from email patterns in production
- Use explicit role mappings from the identity provider
- Gate pattern-based role derivation behind `DEMO_MODE`

---

## LOW Findings

### SEC-18: math/rand for Security-Adjacent Operations (LOW)

**Location:** Multiple files

| File | Usage |
|------|-------|
| `internal/utils/backoff/backoff.go:54` | Jitter calculation |
| `internal/controllers/apiresource_controller.go:153` | Backoff jitter |
| `internal/anonymization/masker.go` | Masking noise/synthetic data |

`math/rand` is not cryptographically secure. For backoff jitter this is acceptable. For anonymization masking, it means the "random" noise is predictable.

**Remediation:**
- Use `crypto/rand` in `anonymization/masker.go` for security-sensitive masking
- `math/rand` is acceptable for backoff jitter (not security-sensitive)

---

### SEC-19: Missing Security Headers (LOW)

**Location:** `main.go`

The server does not set security headers on responses:

| Header | Status |
|--------|--------|
| `Content-Security-Policy` | ❌ Missing |
| `X-Frame-Options` | ❌ Missing |
| `X-Content-Type-Options` | ❌ Missing |
| `Strict-Transport-Security` | ❌ Missing |
| `Referrer-Policy` | ❌ Missing |
| `Permissions-Policy` | ❌ Missing |

Note: The `internal/apiscanner/` module checks for these headers on scanned APIs, but the platform itself doesn't set them.

**Remediation:** Add a middleware that sets all security headers:

```go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        if c.Request.TLS != nil {
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        c.Next()
    }
}
```

---

### SEC-20: Path Traversal Risk in File Upload (LOW)

**Location:** `internal/utils/examples.go:225`

`filepath.Join(uploadPath, safeFilename)` relies on `safeFilename` being properly sanitized. If the sanitization is bypassed, `../` sequences could escape the upload directory.

**Remediation:** Add `filepath.Clean()` and verify the result is still under `uploadPath`:

```go
fullPath := filepath.Join(uploadPath, filepath.Base(safeFilename))
if !strings.HasPrefix(fullPath, filepath.Clean(uploadPath)) {
    return fmt.Errorf("path traversal detected")
}
```

---

## Remediation Priority Matrix

### Immediate (Before Next Deploy)

| ID | Finding | Fix |
|----|---------|-----|
| SEC-01 | Hardcoded DB passwords | Rotate all passwords, use secrets manager |
| SEC-02 | RSA key in .env | Move to file-based key, rotate key |
| SEC-03 | Demo accounts | Gate behind DEMO_MODE env var |
| SEC-07 | Weak JWT fallback | Fail startup if JWT_SECRET not set |

### Short-Term (1-2 Weeks)

| ID | Finding | Fix |
|----|---------|-----|
| SEC-04 | XSS via innerHTML | Audit all innerHTML, add CSP header |
| SEC-05 | MD5 usage | Replace with SHA-256 |
| SEC-08 | Command injection | Validate targets against allowlist |
| SEC-09 | Custom SQL | Add feature flag, read-only connection |
| SEC-10 | localStorage tokens | Migrate to HttpOnly cookies |
| SEC-14 | Error info leaks | Generic client errors, detailed server logs |
| SEC-16 | Unauthenticated health | Add auth to detailed health endpoints |
| SEC-19 | Missing security headers | Add SecurityHeaders middleware |

### Medium-Term (1 Month)

| ID | Finding | Fix |
|----|---------|-----|
| SEC-06 | SHA1 usage | Deprecate, default to SHA-256 |
| SEC-11 | Missing CSRF | Implement if cookie auth is used |
| SEC-12 | SSRF in OAuth | URL validation, block private IPs |
| SEC-13 | Unbounded reads | Add MaxBytesReader |
| SEC-15 | Sensitive fields | Mask secrets in responses |
| SEC-17 | Email role derivation | Gate behind DEMO_MODE |
| SEC-18 | math/rand in masker | Use crypto/rand |
| SEC-20 | Path traversal | Add filepath.Clean + prefix check |

### Long-Term (Ongoing)

| Action | Timeline |
|--------|----------|
| Integrate SAST scanner (gosec, semgrep) into CI/CD | 2 weeks |
| Implement Content Security Policy (CSP) with nonces | 1 month |
| Conduct external penetration test | Quarterly |
| Security training for development team | Bi-annual |
| Dependency vulnerability scanning (govulncheck) | Continuous |

---

## Appendix: Tools Recommended

| Tool | Purpose | Integration |
|------|---------|-------------|
| `gosec` | Go static security analysis | CI/CD pipeline |
| `semgrep` | Multi-language SAST | CI/CD pipeline |
| `govulncheck` | Go dependency vulnerability scanning | CI/CD pipeline |
| `trivy` | Container image scanning | Already integrated |
| `eslint-plugin-security` | JavaScript security linting | Frontend CI |
| `helmet` (or equivalent) | Security headers middleware | Go middleware |

---

*Document maintained by Platform Architecture Team. Review quarterly or after any security incident.*
