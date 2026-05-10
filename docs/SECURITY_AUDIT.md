# 🔒 AxiomNizam — Consolidated Security Audit

**Audit Period:** 2026-04-29 → 2026-05-07 (merged)  
**Scope:** All Go backend (`internal/`, `main.go`), Frontend (`frontend/`), CLI (`cmd/`), Infrastructure (`Dockerfile`, `docker-compose.yml`, `.env`)  
**Codebase:** ~200K+ lines Go, ~30 JS files, 100 internal packages  
**Severity Scale:** 🔴 Critical · 🟠 High · 🟡 Medium · 🟢 Low · ℹ️ Informational

---

## Executive Summary

This document consolidates findings from three separate security reviews into a single authoritative reference. **38 unique findings** were identified across the platform's backend, frontend, CLI, and infrastructure layers.

| Severity | Count | Status |
|----------|-------|--------|
| 🔴 Critical | 5 | Requires immediate action |
| 🟠 High | 12 | Requires action before production |
| 🟡 Medium | 13 | Should be addressed |
| 🟢 Low / ℹ️ Info | 8 | Nice-to-have improvements |

> [!CAUTION]
> The most urgent action is fixing `.gitignore` to exclude `.env`, rotating all committed secrets, and purging them from Git history. Until this is done, the platform's entire authentication infrastructure should be considered compromised.

---

## Positive Security Controls (Already Present)

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
| Row-level security | `internal/security/rls.go` | ✅ RLS policy engine with audit log |
| Security guardrails | `main.go:applySecurityGuardrails()` | ✅ Startup checks for default credentials |

---

## 🔴 CRITICAL Findings

### SEC-01 · RSA Private Key Committed to `.env` (and Git)

**Location:** `.env:132` (`IAM_RSA_PRIVATE_KEY`)  
**Layer:** Internal / Infrastructure

The full RSA private key used to sign all JWT tokens is hardcoded inline in `.env`. The `.gitignore` only excludes `env` (no dot prefix), meaning `.env` **is tracked by Git**.

**Risk:** Every collaborator, fork, and potentially the public has access to the signing key. All tokens ever signed with this key must be considered compromised.

**Remediation:**
1. Rotate the RSA key immediately
2. Fix `.gitignore` to exclude `.env` and `.env.*`
3. Run `git filter-repo` to purge the key from history
4. Use `IAM_RSA_PRIVATE_KEY_FILE` with a mounted secret (Docker/K8s secret, Vault)

---

### SEC-02 · Keycloak Admin Credentials & Client Secret in `.env`

**Location:** `.env:124,136-137`  
**Layer:** Infrastructure

```
KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
KEYCLOAK_ADMIN_USERNAME=admin
KEYCLOAK_ADMIN_PASSWORD=admin
```

Default master-realm admin credentials and the client secret are committed. Combined with SEC-01, the entire identity federation chain is compromised.

**Remediation:** Use Docker/Kubernetes secrets; never store admin credentials in source control. Rotate immediately.

---

### SEC-03 · DEMO_JWT_SECRET in Source Control

**Location:** `.env:128`  
**Layer:** Internal

```
DEMO_JWT_SECRET=smw7flNLvrFeIQNKH7X8u7h_T5TXXtDrSnaz0GoSIP-7cyIITQAwdbZJFPDI3zsa
```

Static, committed, grants token-forging capabilities for the demo auth path.

**Remediation:** Rotate secret, store via secret manager, add `.env` to `.gitignore`.

---

### SEC-04 · Hardcoded Demo Accounts with Admin Access

**Location:** `internal/handlers/auth_handler.go` (~lines 1032-1035)  
**Layer:** Internal

Demo accounts with hardcoded passwords are embedded in source code:

| Username | Password | Role |
|----------|----------|------|
| admin | admin | admin |
| sysadmin | sysadmin | system-manager |
| manager | manager | manager |
| user | user | user |

**Risk:** If demo mode is not explicitly disabled in production (`ENABLE_DEMO_ACCOUNTS=false`), anyone can authenticate as admin.

**Remediation:**
- Gate demo accounts behind `ENABLE_DEMO_ACCOUNTS=true` (default: false)
- Log a startup warning when demo mode is enabled
- Never enable in production

---

### SEC-05 · SQL Injection Protection Blocks Its Own Legitimate Queries

**Location:** `internal/utils/sql_injection_protection.go:21-28`  
**Layer:** Internal

The `forbiddenKeywords` list includes `SELECT`, `INSERT`, `UPDATE`, `DELETE`, and `UNION` — which means `ValidateSQLInput()` rejects **all SQL queries**, including legitimate SELECT statements used by `DynamicQueryHandler`.

```go
forbiddenKeywords: []string{
    "DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE", "ALTER",
    "EXEC", "EXECUTE", "UNION", "SELECT", "REPLACE",
    "--", "/*", "*/", ";", "xp_", "sp_", "|", "&",
},
```

The protection is either broken (rejecting all queries) or bypassed (negating the protection). Either scenario is critical.

**Remediation:** Validate user-supplied *parameters* only (not the full query text). Use parameterized queries exclusively.

---

## 🟠 HIGH Findings

### SEC-06 · XSS via innerHTML in Frontend Dashboards

**Location:** Multiple frontend JS files  
**Layer:** Frontend

The frontend extensively uses `.innerHTML = ...` to render dynamic content. Sanitization functions exist but are used inconsistently:

| File | innerHTML count | Sanitizer |
|------|----------------|-----------|
| `admin.js` | ~50+ | `escapeHtml()` — inconsistent |
| `object-storage.js` | ~25+ | `escHtml()` — consistent |
| `conductor-dashboard.js` | ~15+ | `esc()` — consistent |
| `iam-admin.js` | ~20+ | `escapeHtml()` — mostly |
| `admin-dashboard.js` | ~10+ | Limited |

**Remediation:** Adopt `textContent` for plain text; implement CSP header blocking inline scripts; audit every `innerHTML` assignment.

---

### SEC-07 · `safeHTML` Template Function Enables XSS

**Location:** `frontend/main.go:133-135`  
**Layer:** Frontend

```go
"safeHTML": func(html string) template.HTML {
    return template.HTML(html)  // bypasses all HTML escaping
},
```

If any user-controlled data passes through `{{safeHTML .someField}}` in templates, it enables stored/reflected XSS.

**Remediation:** Audit all template usages; restrict to static HTML only. Consider `bluemonday` sanitizer.

---

### SEC-08 · Frontend Authentication is Client-Side Only

**Location:** `frontend/main.go:77-107`  
**Layer:** Frontend

The `requireFrontendRoles` middleware reads `authToken` and `userRole` from client-side cookies with **no server-side token validation**. Any user can set `userRole=system-manager` cookie to access `/iam-admin`, `/governance`, etc.

**Remediation:** Validate the JWT server-side by calling `/iam/auth/whoami` before granting access.

---

### SEC-09 · OAuth Access Token Passed in URL Fragment

**Location:** `internal/handlers/auth_handler.go:540-551`  
**Layer:** Internal

The access token is placed in the URL fragment on OAuth callback redirect. It appears in browser history, is accessible to all JS on the page, and may be logged by extensions/proxies.

**Remediation:** Use a short-lived authorization code exchange, or set the token as an `HttpOnly`, `Secure`, `SameSite=Lax` cookie.

---

### SEC-10 · Command Injection Risk in Certificate Handler

**Location:** `internal/handlers/certificate_handler.go` (lines 252, 317)  
**Layer:** Internal

The certificate renewal handler executes shell commands via `exec.Command()`. The command template comes from `CERT_RENEW_COMMAND` env var, and the target comes from user input.

**Remediation:** Validate target against a strict allowlist; never pass user input directly to shell commands; use a dedicated cert renewal library.

---

### SEC-11 · Custom SQL Execution in Quality Rules

**Location:** `internal/quality/rules/engine.go` (~line 214)  
**Layer:** Internal

The `custom_sql` rule type allows users to define arbitrary SQL queries executed against datasources. Keyword filtering can be bypassed.

**Remediation:** Execute with read-only database connection; add `QUALITY_CUSTOM_SQL_ENABLED=false` feature flag (default: disabled).

---

### SEC-12 · MD5 / SHA1 Hash Usage for Security Operations

**Location:** Multiple files  
**Layer:** Internal

| File | Usage | Risk |
|------|-------|------|
| `internal/utils/hash/hash.go:40` | MD5 as hash algorithm | Collision attacks |
| `internal/cache/middleware.go:39` | MD5 for cache keys | Cache poisoning |
| `internal/storage/native/native.go:601` | MD5 for ETag | ETag collision |
| `internal/utils/bot_protection.go:348` | MD5 for fingerprinting | Bypass |
| `internal/utils/hash/hash.go:42,67,115` | SHA1 support | Collision attacks |

**Remediation:** Replace MD5/SHA1 with SHA-256. Deprecate SHA1 option; log a warning if selected.

---

### SEC-13 · Default Database Passwords & Discord Webhook Committed

**Location:** `.env:63-94,164`, `docker-compose.yml:233-234`  
**Layer:** Infrastructure

| Service | Password |
|---------|----------|
| MySQL | `root` |
| MariaDB | `root` |
| PostgreSQL | `postgres` |
| MongoDB | `root` |
| Oracle | `oracle123` |
| RabbitMQ | `axiom` |
| Percona | `root` |
| Discord Webhook | Full URL with token |

**Remediation:** Auto-generate passwords at first start; store in secret manager; rotate Discord webhook.

---

### SEC-14 · No CSRF Protection

**Location:** `main.go:305-327`  
**Layer:** Internal

CORS middleware sets `Access-Control-Allow-Credentials: true` but no CSRF token mechanism exists. `X-CSRF-Token` is listed in `Allow-Headers` but never validated.

**Remediation:** Implement double-submit cookie or synchronizer token pattern.

---

### SEC-15 · PostgreSQL SSL Disabled & Sysadmin Password in Frontend `.env`

**Location:** `.env:87`, `frontend/.env:11`  
**Layer:** Infrastructure

`POSTGRES_SSLMODE=disable` — all DB traffic unencrypted. `IAM_SYSADMIN_PASSWORD=changeme-on-first-login` exposed in frontend `.env` — the frontend has no business knowing the sysadmin password.

**Remediation:** Set `POSTGRES_SSLMODE=require` in production. Remove `IAM_SYSADMIN_PASSWORD` from `frontend/.env`.

---

### SEC-16 · Weak JWT Secret Fallback

**Location:** `internal/auth/auth.go` (~lines 265-280)  
**Layer:** Internal

If `DEMO_JWT_SECRET` is not set, the code falls back to a process-ephemeral secret derived from `time.Now().UnixNano()` — predictable and not cryptographically random.

**Remediation:** Fail startup if secret is not set, or generate using `crypto/rand` and log a warning.

---

### SEC-17 · Auth Tokens in localStorage

**Location:** `frontend/templates/auth.js` (lines 213-218, 244-246, 334-338)  
**Layer:** Frontend

JWT tokens stored in `localStorage` are accessible to any JS on the page. Combined with XSS (SEC-06/07), tokens can be exfiltrated.

**Remediation:** Store tokens in `HttpOnly`, `SameSite=Strict`, `Secure` cookies.

---

## 🟡 MEDIUM Findings

### SEC-18 · SSRF Risk in OAuth Configuration

**Location:** `internal/handlers/auth_handler.go` (~lines 776, 853, 1728, 1798)

Admin-configurable OAuth URLs (token, userinfo, discovery) could point at internal services.

**Remediation:** Validate URLs against scheme whitelist (https only); block private IP ranges; set short timeouts.

---

### SEC-19 · Unbounded Request Body Reading

**Location:** Various handlers (e.g., `auth_handler.go:789` — no explicit limit)

Some endpoints use `io.ReadAll` without `MaxBytesReader`. `gin.Default()` used without `MaxMultipartMemory`.

**Remediation:** Add `http.MaxBytesReader` to all `io.ReadAll` calls; set `router.MaxMultipartMemory = 32 << 20`.

---

### SEC-20 · Error Messages Leak Internal Details

**Location:** `main.go` (~lines 406, 413), various handlers

```go
c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
```

**Remediation:** Return generic error messages to clients; log detailed errors server-side only.

---

### SEC-21 · Sensitive Fields in API Responses

**Location:** `internal/iam/admin/admin.go`, `internal/storage/admin/admin.go`

Client secrets and access keys returned unmasked in API responses.

**Remediation:** Mask secrets in list/get responses (last 4 chars only); return full secrets only on creation.

---

### SEC-22 · Privilege Escalation via Email Pattern Matching

**Location:** `internal/handlers/auth_handler.go` (~lines 1031-1037), `main.go:443-451`

`deriveOAuthRole()` assigns roles based on email patterns. Additionally, if a JWT token has no roles but its email matches `IAM_SYSADMIN_EMAIL`, sysadmin is auto-granted.

**Remediation:** Never derive roles from email patterns in production; require explicit role mappings from IdP; remove sysadmin email fallback or require issuer verification.

---

### SEC-23 · `ValidateQuerySafety` is Dead Code

**Location:** `internal/handlers/dynamic_query_handler.go:581-594`

Exported but unused function that only checks `DROP DATABASE` and `DROP SCHEMA` — trivially bypassed. Provides false sense of security.

**Remediation:** Remove dead code or replace with proper SQL safety validation.

---

### SEC-24 · Token Accepted from Query Parameter (Unscoped)

**Location:** `main.go:419-421`

Tokens in query parameters appear in server logs, proxy logs, browser history. Intended for WebSocket but not scoped to upgrade requests.

**Remediation:** Only accept query-parameter tokens for `Upgrade: websocket` requests.

---

### SEC-25 · etcd Has No Authentication

**Location:** `docker-compose.yml:157-174`

etcd exposed on port `2379` with no TLS, no auth. Anyone with network access can read/write all KV data.

**Remediation:** Enable etcd TLS and client certificate authentication.

---

### SEC-26 · OAuth State Cookie Shares JWT Key

**Location:** `internal/handlers/auth_handler.go:256-261`

`OAUTH_STATE_SECRET` falls back to `DemoJWTSecret()` — key compromise in one domain affects the other.

**Remediation:** Set a dedicated `OAUTH_STATE_SECRET`.

---

### SEC-27 · Demo Token Path Always Active

**Location:** `internal/auth/auth.go:356-398`

Despite `ENABLE_DEMO_ACCOUNTS=false`, `ValidateToken()` always falls back to `ValidateDemoToken()` when RSA validation fails.

**Remediation:** Guard demo token fallback behind the `ENABLE_DEMO_ACCOUNTS` flag.

---

### SEC-28 · Kafka/RabbitMQ Communication Is Plaintext

**Location:** `docker-compose.yml:398-423`

All Kafka listeners use `PLAINTEXT`. RabbitMQ uses `amqp://` (no TLS).

**Remediation:** Enable TLS for all message bus connections in production.

---

### SEC-29 · Missing Security Headers

**Location:** `main.go`, `frontend/main.go`

Neither backend nor frontend set security headers:

| Header | Status |
|--------|--------|
| `Content-Security-Policy` | ❌ Missing |
| `X-Frame-Options` | ❌ Missing |
| `X-Content-Type-Options` | ❌ Missing |
| `Strict-Transport-Security` | ❌ Missing |
| `Referrer-Policy` | ❌ Missing |
| `Permissions-Policy` | ❌ Missing |

**Remediation:** Add a `SecurityHeaders()` middleware to both routers.

---

### SEC-30 · Missing Auth on Detailed Health Endpoints

**Location:** `main.go` (~line 526)

`/health/reconcilers` exposes internal reconciler metrics, error counts, and module names without authentication.

**Remediation:** Require auth for detailed health; keep only simple `/health` (200 OK) as unauthenticated.

---

## 🟢 LOW / ℹ️ INFORMATIONAL Findings

### SEC-31 · math/rand for Security-Adjacent Operations

**Location:** `internal/anonymization/masker.go`, `internal/utils/backoff/backoff.go:54`

`math/rand` (not cryptographically secure) used for anonymization masking. Acceptable for backoff jitter.

**Remediation:** Use `crypto/rand` in `anonymization/masker.go`.

---

### SEC-32 · Path Traversal Risk in File Upload

**Location:** `internal/utils/examples.go:225`

`filepath.Join(uploadPath, safeFilename)` relies on proper sanitization. If bypassed, `../` sequences could escape the upload directory.

**Remediation:** Add `filepath.Clean()` + prefix check to verify result is under `uploadPath`.

---

### SEC-33 · CLI `--password` Flag Visible in Process List

**Location:** `cmd/axiomnizamctl/auth.go:316`

`axiomnizamctl login --password secret` leaks password in `ps aux` and shell history.

**Remediation:** Remove `--password` flag; always use interactive prompt or `stdin` pipe.

---

### SEC-34 · `--insecure-skip-tls-verify` Available by Default

**Location:** `cmd/axiomnizamctl/auth.go:321`

Allows MITM attacks on CLI-to-server connections without visible warning.

**Remediation:** Emit a visible warning when `--insecure-skip-tls-verify` is used.

---

### SEC-35 · Health/Status Endpoints Expose Backend Topology

**Location:** `main.go:545-548`

`/health`, `/status`, `/distributed` are unauthenticated and reveal internal service topology and version information.

---

### SEC-36 · Docker Runtime Uses Root User

**Location:** `Dockerfile:48`

The runtime container runs as root. Container escape vulnerabilities are amplified.

**Remediation:** Add a non-root user: `RUN useradd -r appuser && USER appuser`.

---

### SEC-37 · CORS `Access-Control-Max-Age: 86400` Is Aggressive

**Location:** `main.go:311`

24-hour preflight cache means CORS policy changes won't take effect for existing users for up to a day.

---

### SEC-38 · Firebase Credentials Placeholder in `.env`

**Location:** `.env:155-162`

Firebase config contains placeholder values (`fake-private-key`). Not a current vulnerability but could be overlooked when real values are substituted.

---

## Remediation Priority Matrix

### P0 — Immediate (Before Next Deploy)

| ID | Finding | Fix |
|----|---------|-----|
| SEC-01 | RSA key in `.env` / Git | Rotate key, fix `.gitignore`, purge Git history |
| SEC-02 | Keycloak admin creds committed | Rotate, use K8s/Docker secrets |
| SEC-03 | DEMO_JWT_SECRET committed | Rotate, secret manager |
| SEC-04 | Hardcoded demo accounts | Gate behind `ENABLE_DEMO_ACCOUNTS` flag |
| SEC-05 | SQL injection filter broken | Redesign validation to target params only |
| SEC-16 | Weak JWT fallback | Fail startup if not set |

### P1 — This Sprint (1–2 Weeks)

| ID | Finding | Fix |
|----|---------|-----|
| SEC-06 | XSS via innerHTML | Audit all innerHTML, add CSP header |
| SEC-07 | `safeHTML` XSS | Remove or restrict to static HTML |
| SEC-08 | Frontend auth client-side only | Server-side JWT validation |
| SEC-09 | OAuth token in URL fragment | Use auth code exchange |
| SEC-10 | Command injection in cert handler | Allowlist targets |
| SEC-11 | Custom SQL in quality rules | Feature flag + read-only connection |
| SEC-12 | MD5/SHA1 usage | Replace with SHA-256 |
| SEC-13 | Default DB passwords + Discord webhook | Rotate all, use secret manager |
| SEC-14 | No CSRF protection | Implement double-submit cookie |
| SEC-15 | PG SSL disabled + sysadmin pw in frontend | Fix both configs |
| SEC-17 | localStorage tokens | Migrate to HttpOnly cookies |
| SEC-29 | Missing security headers | Add SecurityHeaders middleware |

### P2 — Medium-Term (1 Month)

| ID | Finding | Fix |
|----|---------|-----|
| SEC-18 | SSRF in OAuth | URL validation, block private IPs |
| SEC-19 | Unbounded reads | Add MaxBytesReader |
| SEC-20 | Error info leaks | Generic client errors |
| SEC-21 | Sensitive fields unmasked | Mask secrets in responses |
| SEC-22 | Email-based role escalation | Gate behind DEMO_MODE |
| SEC-23 | Dead ValidateQuerySafety code | Remove |
| SEC-24 | Token in query param unscoped | Scope to WebSocket only |
| SEC-25 | etcd no auth | Enable TLS + client certs |
| SEC-26 | OAuth state shares JWT key | Set dedicated OAUTH_STATE_SECRET |
| SEC-27 | Demo token always active | Guard behind feature flag |
| SEC-28 | Message bus plaintext | Enable TLS |
| SEC-30 | Unauthenticated detailed health | Add auth |

### P3 — Long-Term (Ongoing)

| ID | Finding | Fix |
|----|---------|-----|
| SEC-31–38 | Low/informational items | See individual entries |

| Action | Timeline |
|--------|----------|
| Integrate SAST scanner (gosec, semgrep) into CI/CD | 2 weeks |
| Implement Content Security Policy (CSP) with nonces | 1 month |
| Conduct external penetration test | Quarterly |
| Security training for development team | Bi-annual |
| Dependency vulnerability scanning (govulncheck) | Continuous |

---

## Appendix: Recommended Tools

| Tool | Purpose | Integration |
|------|---------|-------------|
| `gosec` | Go static security analysis | CI/CD pipeline |
| `semgrep` | Multi-language SAST | CI/CD pipeline |
| `govulncheck` | Go dependency vulnerability scanning | CI/CD pipeline |
| `trivy` | Container image scanning | Already integrated |
| `eslint-plugin-security` | JavaScript security linting | Frontend CI |
| `bluemonday` | HTML sanitization for Go templates | Go middleware |
| `helmet` (or equivalent) | Security headers middleware | Go middleware |

---

*Consolidated from audits dated 2026-04-29 and 2026-05-07. Review quarterly or after any security incident.*
