# AxiomNizam — Security Architecture & Implementation Guide

**Last Updated:** 2026-05-02  
**Companion Document:** [`SECURITY_AUDIT.md`](./SECURITY_AUDIT.md) (38 findings, 2026-04-29)  
**Maintainer:** Platform Architecture Team

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current Security Inventory — Internal Modules](#current-security-inventory)
3. [Audit Finding ↔ Module Coverage Matrix](#audit-finding--module-coverage-matrix)
4. [Gaps Identified — What Exists But Is Not Wired](#gaps-identified)
5. [Additional Security Implementations Recommended](#additional-security-implementations-recommended)
6. [Implementation Roadmap](#implementation-roadmap)
7. [Security Configuration Checklist](#security-configuration-checklist)

---

## Executive Summary

AxiomNizam contains **19 security-relevant internal modules** spanning authentication, authorization, encryption, auditing, scanning, compliance, and threat detection. The April 2026 security audit (`SECURITY_AUDIT.md`) identified 38 findings across 9 categories.

**Key insight from internal scan:** Many audit findings (SEC-19 security headers, SEC-18 `math/rand`, SEC-05/06 weak hashes) already have remediation code inside `internal/` — the code exists but is **not wired into `main.go` or the handler layer**. Fixing these is mostly integration work, not new development.

This document inventories every security module, maps them to audit findings, and lists **additional hardening opportunities** beyond what the audit covered.

---

## Current Security Inventory

### Layer 1 — Authentication & Identity

| Module | Path | Status | Description |
|--------|------|--------|-------------|
| **JWT/OIDC Validator** | `internal/auth/auth.go` | ✅ Active | RSA JWKS validation with key refresh, demo HMAC fallback |
| **Auth Middleware** | `internal/auth/middleware.go` | ✅ Active | Bearer token extraction, role injection into Gin context |
| **Rate Limiter** | `internal/auth/rate_limit.go` | ✅ Active | Per-token call limits (500/token), expiry (10min), auto-cleanup |
| **Rate Limit Middleware** | `internal/auth/rate_limit_middleware.go` | ✅ Active | Gin middleware for rate enforcement |
| **IAM System** | `internal/iam/` | ✅ Active | Full IAM: authn, authz, identity, OAuth, token management, user store |
| **Bootstrap Secrets** | `internal/bootstrapsecrets/store.go` | ✅ Active | Atomic secret ensure via PostgreSQL (avoids hardcoded secrets) |
| **Password Validator** | `internal/utils/security_utils.go` | ⚠️ Exists, partially wired | 12-char min, upper/lower/digit/special, forbidden word check |

### Layer 2 — Authorization & Access Control

| Module | Path | Status | Description |
|--------|------|--------|-------------|
| **RBAC Engine** | `internal/rbac/` | ✅ Active | Role-based access with RoleBinding resources, 7 files |
| **Row-Level Security** | `internal/security/rls.go` | ⚠️ Exists, not fully wired | Per-table RLS policies, user-context, predicate evaluation, audit log |
| **Admission Chain** | `internal/admission/chain.go` | ✅ Active | K8s-style mutating + validating webhooks with per-plugin timeout |
| **Admission Policies** | `internal/policies/admission_policy.go` | ✅ Active | OPA-like rule engine: Deny/Warn/Mutate with condition evaluation |
| **Governance** | `internal/governance/` | ✅ Active | CompliancePolicy, RetentionPolicy, AccessRequest resources |

### Layer 3 — Encryption & Key Management

| Module | Path | Status | Description |
|--------|------|--------|-------------|
| **AES-256-GCM Encryption** | `internal/encryption/` | ✅ Active | Field-level encrypt/decrypt, key rotation with audit log |
| **Keyring** | `internal/keyring/keyring.go` | ✅ Active | Rotating-key manager, AES-GCM, key-ID prefixed ciphertexts |
| **Crypto Utilities** | `internal/utils/security_utils.go` | ✅ Available | `crypto/rand` secure random, SHA-256/512, HMAC generation |
| **TLS Config** | `internal/utils/security_utils.go` | ⚠️ Defined, not enforced | TLS 1.2 minimum, server cipher preference, cert validation |

### Layer 4 — Auditing & Compliance

| Module | Path | Status | Description |
|--------|------|--------|-------------|
| **Audit Compliance Manager** | `internal/audit/compliance.go` | ✅ Active | GDPR/HIPAA/SOC2/PCI-DSS frameworks, violation tracking, risk assessment |
| **Compliance Engine** | `internal/policies/compliance_engine.go` | ✅ Active | Rule-based compliance checks, remediation plans, audit trail |
| **Security Policy Engine** | `internal/policies/security/security.go` | ✅ Active | Auth/authz/encryption/audit/vuln policies, threat modeling, incident tracking |

### Layer 5 — Threat Detection & Scanning

| Module | Path | Status | Description |
|--------|------|--------|-------------|
| **File Scanner Orchestrator** | `internal/scanner/scanner.go` | ✅ Active | Multi-scanner orchestration with finding aggregation |
| **ClamAV Integration** | `internal/scanner/clamav.go` | ✅ Active | Antivirus via TCP INSTREAM protocol |
| **Macro/Script Scanner** | `internal/scanner/macro.go` | ✅ Active | VBA macros, PDF JavaScript, auto-exec, shell commands |
| **SVG XSS Scanner** | `internal/scanner/svg.go` | ✅ Active | Script tags, event handlers, javascript: URIs, foreignObject |
| **Archive Bomb Scanner** | `internal/scanner/archive.go` | ✅ Active | Zip bomb detection, path traversal, executable detection |
| **MIME Type Scanner** | `internal/scanner/mime.go` | ✅ Active | MIME validation and type detection |
| **Trivy Integration** | `internal/trivy/` | ✅ Active | Container vulnerability scanning with severity filtering |
| **Network Intelligence** | `internal/netintel/` | ✅ Active | Anomaly detection, threat alerts, rogue device detection |
| **API Security Scanner** | `internal/apiscanner/` | ✅ Active | Security header checks, vulnerability scanning for APIs |

### Layer 6 — Data Protection

| Module | Path | Status | Description |
|--------|------|--------|-------------|
| **Data Anonymization** | `internal/anonymization/masker.go` | ✅ Active | 8 techniques: hash, redact, partial, tokenize, noise, generalize, synthetic, shuffle |
| **Input Sanitizer** | `internal/utils/security_utils.go` | ✅ Available | URL scheme validation, string sanitization, length limiting |
| **Security Headers** | `internal/utils/security_utils.go` | ⚠️ Defined, not applied | CSP, X-Frame-Options, HSTS, etc. — struct exists, no middleware |
| **CORS Validator** | `internal/utils/security_utils.go` | ⚠️ Defined, partially used | Origin/method validation |

---

## Audit Finding ↔ Module Coverage Matrix

| Audit ID | Severity | Finding | Internal Module | Status |
|----------|----------|---------|-----------------|--------|
| SEC-01 | CRITICAL | Hardcoded DB passwords | `bootstrapsecrets/store.go` | **Module exists** — can replace .env secrets with PG-backed store |
| SEC-02 | CRITICAL | RSA key in .env | `bootstrapsecrets/store.go`, `keyring/` | **Module exists** — migrate to file-based key or keyring |
| SEC-03 | CRITICAL | Demo accounts | `auth/auth.go` (demo token logic) | **Partially gated** — `DEMO_JWT_SECRET` env var exists, needs `DEMO_MODE` gate |
| SEC-04 | HIGH | XSS via innerHTML | `scanner/svg.go` (server-side) | **Server-side only** — frontend needs CSP header from middleware |
| SEC-05 | HIGH | MD5 usage | `utils/security_utils.go` (SHA-256) | **Replacement exists** — `HashFunction()` defaults to SHA-256 |
| SEC-06 | HIGH | SHA1 usage | `utils/security_utils.go` | **SHA-256/512 available** — need to deprecate SHA1 option |
| SEC-07 | HIGH | Weak JWT fallback | `auth/auth.go:268-283` | **Already fixed** — uses `crypto/rand` with `time.Now` as last fallback |
| SEC-08 | HIGH | Command injection | — | ❌ No internal mitigation — needs allowlist validation |
| SEC-09 | HIGH | Custom SQL execution | — | ❌ No read-only connection enforced — needs feature flag |
| SEC-10 | MEDIUM | localStorage tokens | — | ❌ Frontend-only — needs HttpOnly cookie migration |
| SEC-11 | MEDIUM | Missing CSRF | — | ❌ No CSRF module — JWT Bearer provides partial protection |
| SEC-12 | MEDIUM | SSRF in OAuth | `utils/security_utils.go` (URL sanitizer) | **URL validator exists** — needs private IP blocking |
| SEC-13 | MEDIUM | Unbounded reads | — | ❌ Needs `MaxBytesReader` wrapper |
| SEC-14 | MEDIUM | Error info leaks | — | ❌ Needs error sanitization middleware |
| SEC-15 | MEDIUM | Sensitive fields exposed | `anonymization/masker.go` | **Masker exists** — can apply partial masking to API responses |
| SEC-16 | MEDIUM | Unauthenticated health | `auth/middleware.go` | **Middleware exists** — just need to apply to health endpoints |
| SEC-17 | MEDIUM | Email role derivation | — | ❌ Needs `DEMO_MODE` gate |
| SEC-18 | LOW | `math/rand` in masker | `anonymization/masker.go:15` | ❌ Confirmed — uses `math/rand` with seed `42` |
| SEC-19 | LOW | Missing security headers | `utils/security_utils.go:28-38` | **Struct defined** — needs Gin middleware to apply |
| SEC-20 | LOW | Path traversal | — | ❌ Needs `filepath.Clean` + prefix check |

**Summary:** 10 of 20 findings have existing internal code that can address them. 10 require new implementation.

---

## Gaps Identified

### Code Exists But Not Wired

These are the highest-ROI fixes — the code is already written, just not integrated:

1. **Security Headers Middleware** — `DefaultSecurityHeaders()` is defined in `utils/security_utils.go` but no Gin middleware applies it. Wire it in `main.go`.

2. **Row-Level Security** — `internal/security/rls.go` is a complete RLS engine with audit logging, but needs integration with the data access layer.

3. **Password Validator** — `NewPasswordValidator()` enforces 12-char passwords with complexity rules but isn't used during account creation.

4. **TLS Configuration** — `DefaultTLSConfig()` specifies TLS 1.2 minimum but isn't applied to the HTTP server.

5. **Input Sanitizer** — `NewInputSanitizer()` with URL/string sanitization exists but isn't used in handlers accepting user URLs.

6. **Bootstrap Secrets** — Can replace all `.env` hardcoded secrets with the PostgreSQL-backed atomic store.

---

## Additional Security Implementations Recommended

Beyond the audit findings, these are additional hardening opportunities identified from the internal scan:

### Priority 1 — Critical (Implement Before Next Release)

#### SEC-NEW-01: Security Headers Middleware
**Effort:** 1 hour  
**Impact:** Fixes SEC-19 + adds defense-in-depth against XSS

The `DefaultSecurityHeaders()` function already exists. Create and register a Gin middleware:

```go
// internal/auth/security_headers_middleware.go
func SecurityHeadersMiddleware() gin.HandlerFunc {
    headers := utils.DefaultSecurityHeaders()
    return func(c *gin.Context) {
        c.Header("Content-Security-Policy", headers.ContentSecurityPolicy)
        c.Header("X-Frame-Options", headers.XFrameOptions)
        c.Header("X-Content-Type-Options", headers.XContentTypeOptions)
        c.Header("Referrer-Policy", headers.ReferrerPolicy)
        c.Header("Permissions-Policy", headers.PermissionsPolicy)
        if c.Request.TLS != nil {
            c.Header("Strict-Transport-Security", headers.StrictTransportSecurity)
        }
        c.Next()
    }
}
```

#### SEC-NEW-02: Error Sanitization Middleware
**Effort:** 2 hours  
**Impact:** Fixes SEC-14, prevents internal detail leakage

```go
func ErrorSanitizationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        if c.Writer.Status() >= 400 {
            // Log detailed error server-side, return generic to client
            log.Printf("Error %d on %s: %v", c.Writer.Status(), c.Request.URL.Path, c.Errors)
        }
    }
}
```

#### SEC-NEW-03: Demo Mode Gate
**Effort:** 1 hour  
**Impact:** Fixes SEC-03, SEC-17

```go
var demoMode = os.Getenv("DEMO_MODE") == "true"

func init() {
    if demoMode {
        log.Println("⚠️  WARNING: DEMO_MODE is enabled. Do NOT use in production.")
    }
}
```

### Priority 2 — High (Implement Within 2 Weeks)

#### SEC-NEW-04: Request Signing / Webhook HMAC Validation
**Effort:** 4 hours  
**Impact:** Prevents webhook spoofing and request tampering

`HMACValidator` already exists in `security_utils.go`. Extend to verify incoming webhook signatures:

```go
validator := utils.NewHMACValidator(webhookSecret)
if !validator.VerifySignature(body, signatureHeader) {
    c.AbortWithStatus(403)
}
```

#### SEC-NEW-05: API Response Field Masking Middleware
**Effort:** 4 hours  
**Impact:** Fixes SEC-15 — auto-mask sensitive fields in responses

Use `anonymization/masker.go` to intercept and mask fields like `client_secret`, `access_key`, `password`:

```go
sensitiveFields := []string{"client_secret", "secret_key", "password", "access_key"}
masker := anonymization.NewMasker(hmacSecret)
// Apply partial masking to matching JSON keys in response body
```

#### SEC-NEW-06: SSRF Protection — Private IP Blocking
**Effort:** 3 hours  
**Impact:** Fixes SEC-12

Extend `InputSanitizer.SanitizeURL()` to block private/reserved IP ranges:

```go
func IsPrivateIP(ip net.IP) bool {
    privateRanges := []string{
        "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
        "127.0.0.0/8", "169.254.0.0/16", "::1/128", "fc00::/7",
    }
    // Check each range
}
```

#### SEC-NEW-07: Request Body Size Limiter
**Effort:** 1 hour  
**Impact:** Fixes SEC-13

```go
func MaxBodySizeMiddleware(maxBytes int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
        c.Next()
    }
}
```

#### SEC-NEW-08: Crypto/Rand in Anonymization Masker
**Effort:** 30 minutes  
**Impact:** Fixes SEC-18 — predictable masking noise

Replace `math/rand` with `crypto/rand` in `internal/anonymization/masker.go`:

```diff
- "math/rand"
+ "crypto/rand"
+ "encoding/binary"
```

### Priority 3 — Medium (Implement Within 1 Month)

#### SEC-NEW-09: Session Fixation Protection
**Effort:** 4 hours  
**Impact:** Prevents session hijacking after authentication

Regenerate session/token identifiers after successful login. The rate limiter already supports `RevokeToken()` — extend to force re-issue after auth events.

#### SEC-NEW-10: Brute Force Protection with Account Lockout
**Effort:** 6 hours  
**Impact:** Prevents credential stuffing attacks

`LockoutPolicy` is already defined in `policies/security/security.go`:
```go
type LockoutPolicy struct {
    MaxFailedAttempts int           // e.g., 5
    LockoutDuration   time.Duration // e.g., 15 min
    ResetAfter        time.Duration // e.g., 1 hour
}
```
Wire this into the auth handler login flow.

#### SEC-NEW-11: Mutual TLS (mTLS) for Internal Services
**Effort:** 8 hours  
**Impact:** Zero-trust internal communications

`TLSConfig` struct exists with cert validation and pinning flags. Implement:
- Client certificate validation for inter-service calls
- Certificate pinning for known internal services

#### SEC-NEW-12: Secret Rotation Automation
**Effort:** 8 hours  
**Impact:** Reduces exposure window for compromised credentials

The `keyring.Rotate()` and `encryption.RotateKey()` primitives exist. Build a reconciler that:
- Rotates encryption keys on a schedule (e.g., 90 days)
- Rotates JWT signing keys with graceful rollover
- Logs all rotations to the audit trail

#### SEC-NEW-13: Content-Type Validation Middleware
**Effort:** 2 hours  
**Impact:** Prevents content-type confusion attacks

```go
func RequireJSON() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method != "GET" && c.Request.Method != "DELETE" {
            ct := c.ContentType()
            if ct != "application/json" {
                c.AbortWithStatusJSON(415, gin.H{"error": "Content-Type must be application/json"})
                return
            }
        }
        c.Next()
    }
}
```

#### SEC-NEW-14: API Abuse Detection
**Effort:** 8 hours  
**Impact:** Detects and blocks automated attacks

Extend `netintel/analytics.go` anomaly detection to API layer:
- Detect unusual request patterns (rate, distribution, payload similarity)
- Flag credential stuffing (many failed logins from one IP)
- Alert on data exfiltration patterns (large response volumes)

#### SEC-NEW-15: Sensitive Data Discovery Scanner
**Effort:** 6 hours  
**Impact:** Proactive PII detection in stored data

Extend `anonymization/masker.go` patterns to scan database fields for:
- Social Security Numbers: `\d{3}-\d{2}-\d{4}`
- Credit card numbers: `\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}`
- Email addresses in non-email fields
- API keys / tokens in plain text fields

The `policies/admission_policy.go` already has a `CreatePIIProtectionPolicy()` — extend it to scan stored data, not just admission.

### Priority 4 — Long-Term (Ongoing)

#### SEC-NEW-16: CI/CD Security Pipeline
**Effort:** 16 hours  
**Impact:** Shift-left security

| Tool | Purpose | Status |
|------|---------|--------|
| `gosec` | Go SAST | ❌ Not integrated |
| `semgrep` | Multi-language SAST | ❌ Not integrated |
| `govulncheck` | Dependency vulns | ❌ Not integrated |
| `trivy` | Container scanning | ✅ Internal module exists |
| `eslint-plugin-security` | JS security | ❌ Not integrated |

#### SEC-NEW-17: Security Event Correlation
**Effort:** 16 hours  
**Impact:** Connects audit events across modules

Correlate events from:
- `audit/compliance.go` audit logs
- `netintel/analytics.go` anomalies
- `auth/rate_limit.go` rate limit violations
- `scanner/` file scan findings
- `policies/` admission denials

Into a unified security event stream for SIEM integration.

#### SEC-NEW-18: Data Loss Prevention (DLP)
**Effort:** 20 hours  
**Impact:** Prevents sensitive data exfiltration

Build on existing modules:
- `anonymization/masker.go` for detection patterns
- `policies/admission_policy.go` for enforcement
- `scanner/` for file content inspection
- `netintel/analytics.go` for network-level DLP

---

## Implementation Roadmap

### Phase 1 — Quick Wins (Week 1)
| Item | Effort | Fixes |
|------|--------|-------|
| Wire `SecurityHeadersMiddleware` in `main.go` | 1h | SEC-19 |
| Add `DEMO_MODE` env gate to demo accounts | 1h | SEC-03, SEC-17 |
| Replace `math/rand` with `crypto/rand` in masker | 30m | SEC-18 |
| Add `MaxBodySizeMiddleware` to API routes | 1h | SEC-13 |
| Apply error sanitization to auth error responses | 2h | SEC-14 |

### Phase 2 — Core Hardening (Weeks 2-3)
| Item | Effort | Fixes |
|------|--------|-------|
| Wire `PasswordValidator` into account creation | 2h | SEC-03 (related) |
| Add SSRF private IP blocking to OAuth URLs | 3h | SEC-12 |
| Implement API response field masking | 4h | SEC-15 |
| Wire `LockoutPolicy` into auth handler | 6h | SEC-NEW-10 |
| Add request body HMAC validation for webhooks | 4h | SEC-NEW-04 |

### Phase 3 — Advanced (Weeks 4-6)
| Item | Effort | Fixes |
|------|--------|-------|
| Implement secret rotation reconciler | 8h | SEC-01, SEC-02 |
| Add mTLS for internal services | 8h | SEC-NEW-11 |
| Build API abuse detection | 8h | SEC-NEW-14 |
| Implement sensitive data discovery | 6h | SEC-NEW-15 |
| CI/CD security pipeline | 16h | SEC-NEW-16 |

### Phase 4 — Enterprise (Ongoing)
| Item | Effort | Fixes |
|------|--------|-------|
| Security event correlation / SIEM | 16h | SEC-NEW-17 |
| DLP implementation | 20h | SEC-NEW-18 |
| External penetration test | — | Quarterly |
| Security training | — | Bi-annual |

---

## Security Configuration Checklist

### Production Deployment

```bash
# ── Critical: Must be set ──────────────────────────────────
DEMO_MODE=false                           # Never true in production
DEMO_JWT_SECRET=<random-64-char>          # Or unset (uses crypto/rand)
IAM_RSA_PRIVATE_KEY_FILE=/etc/axiomnizam/jwt.key  # File, not inline

# ── Database: Use strong passwords ─────────────────────────
DB_MYSQL_PASSWORD=<generated-32-char>
DB_POSTGRES_PASSWORD=<generated-32-char>
DB_MONGO_PASSWORD=<generated-32-char>

# ── Encryption ─────────────────────────────────────────────
ENCRYPTION_KEY_ROTATION_DAYS=90
TLS_MIN_VERSION=1.2
TLS_CERT_FILE=/etc/axiomnizam/tls.crt
TLS_KEY_FILE=/etc/axiomnizam/tls.key

# ── Rate Limiting ──────────────────────────────────────────
RATE_LIMIT_MAX_CALLS=500
RATE_LIMIT_TOKEN_VALIDITY_MINUTES=10

# ── Feature Flags ──────────────────────────────────────────
QUALITY_CUSTOM_SQL_ENABLED=false          # Disable custom SQL by default
CERT_RENEW_TARGETS_ALLOWLIST=*.example.com  # Restrict cert renewal targets

# ── Scanning ───────────────────────────────────────────────
CLAMAV_ADDRESS=clamav:3310
TRIVY_SEVERITY=HIGH,CRITICAL
```

### File Permissions

```bash
chmod 0600 /etc/axiomnizam/jwt.key
chmod 0600 /etc/axiomnizam/tls.key
chmod 0644 /etc/axiomnizam/tls.crt
```

### Health Check Validation

```bash
# These should require auth in production:
curl -H "Authorization: Bearer $TOKEN" http://localhost:7000/health/reconcilers
# This should be unauthenticated:
curl http://localhost:7000/health  # Returns 200 OK only
```

---

## Module Reference Quick Links

| Concern | Primary Module | Secondary |
|---------|---------------|-----------|
| Authentication | `internal/auth/` | `internal/iam/` |
| Authorization | `internal/rbac/` | `internal/admission/` |
| Encryption | `internal/encryption/` | `internal/keyring/` |
| Auditing | `internal/audit/` | `internal/policies/compliance_engine.go` |
| File Scanning | `internal/scanner/` | `internal/trivy/` |
| Data Masking | `internal/anonymization/` | — |
| Network Security | `internal/netintel/` | `internal/apiscanner/` |
| Compliance | `internal/governance/` | `internal/policies/security/` |
| Row-Level Security | `internal/security/` | — |
| Secret Management | `internal/bootstrapsecrets/` | `internal/keyring/` |
| Threat Modeling | `internal/policies/security/` | `internal/netintel/` |
| Utilities | `internal/utils/security_utils.go` | `internal/utils/crypto/` |

---

*This document is a living reference. Update after each security remediation sprint or audit cycle.*
