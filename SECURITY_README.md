# AxiomNizam Security README

Date: 2026-05-03

## Scope and Method

This document is the authoritative security posture reference for the
AxiomNizam platform, rewritten from a code-backed scan of the internal
directory and current runtime wiring.

Scan coverage:
- Internal Go files scanned: 619 (across 100 packages)
- Focus areas: authentication, authorization, SQL execution paths, file scanning, logging, encryption, rate limiting, configuration defaults, reconciler security, and startup guardrails

Companion documents:
- [SECURITY_AUDIT.md](docs/SECURITY_AUDIT.md) — 38 code-level findings (2026-04-29)
- [CODING_PRACTICES.md](docs/CODING_PRACTICES.md) — enforced security-related coding standards

## Executive Summary

Current posture:
- Strong baseline controls exist (JWT validation, role middleware, admin/system-manager gates on privileged routes, SQL read-only policy checks for API Builder templates, SafeGate file scanning pipeline, AES-256-GCM field encryption, RBAC with reconciled RoleBinding resources).
- 29 reconciler controllers run in shadow mode by default, providing declarative state management with dual-write migration path.
- IAM system provides full OIDC/OAuth with admin, authn, authz, identity, and token management.
- Critical credential and token-hardening gaps remain (hardcoded CLI auth key/admin, insecure config defaults, secrets handling hygiene).
- Startup guardrails are implemented and can enforce in production, but rollout is currently designed to begin in audit mode.

Top risks to address first:
1. Hardcoded CLI auth secret and default admin credentials in internal handlers.
2. Insecure default credentials in configuration fallbacks.
3. Token validation helper endpoint that returns success for any presented bearer token.
4. Runtime custom API auth_required flag is stored but not enforced at invocation.
5. Query logging stores raw SQL and params, which may include sensitive data.

## Security Module Inventory

### Layer 1 — Authentication & Identity

| Module | Path | Status |
|--------|------|--------|
| JWT/OIDC Validator | `internal/auth/auth.go` | ✅ Active |
| Auth Middleware | `internal/auth/middleware.go` | ✅ Active |
| Rate Limiter | `internal/auth/rate_limit.go` | ✅ Active |
| IAM System (OIDC) | `internal/iam/` (15 files) | ✅ Active |
| Bootstrap Secrets | `internal/bootstrapsecrets/store.go` | ✅ Active |

### Layer 2 — Authorization & Access Control

| Module | Path | Status |
|--------|------|--------|
| RBAC Engine | `internal/rbac/` (7 files) | ✅ Active — reconciled via GenericController |
| Admission Chain | `internal/admission/chain.go` | ✅ Active |
| Policy Engine (CEL/Rego/DSL) | `internal/policies/` (16 files) | ✅ Active |
| Row-Level Security | `internal/security/rls.go` | ⚠️ Exists, not fully wired |

### Layer 3 — Encryption & Key Management

| Module | Path | Status |
|--------|------|--------|
| AES-256-GCM Encryption | `internal/encryption/` | ✅ Active — reconciled via GenericController |
| Keyring | `internal/keyring/keyring.go` | ✅ Active |
| Crypto Utilities | `internal/utils/security_utils.go` | ✅ Available |

### Layer 4 — Auditing & Compliance

| Module | Path | Status |
|--------|------|--------|
| Audit Compliance Manager | `internal/audit/` (9 files) | ✅ Active — reconciled via GenericController |
| Compliance Engine | `internal/policies/compliance_engine.go` | ✅ Active |
| Security Policy Engine | `internal/policies/security/` | ✅ Active |

### Layer 5 — Scanning & Threat Detection

| Module | Path | Status |
|--------|------|--------|
| SafeGate File Scanner | `internal/scanner/` (7 files) | ✅ Active |
| Trivy Integration | `internal/trivy/` (7 files) | ✅ Active |
| API Security Scanner | `internal/apiscanner/` (10 files) | ✅ Active |
| Network Intelligence | `internal/netintel/` (5 files) | ✅ Active |

### Layer 6 — Data Protection

| Module | Path | Status |
|--------|------|--------|
| Data Anonymization | `internal/anonymization/masker.go` | ✅ Active |
| Input Sanitizer | `internal/utils/security_utils.go` | ✅ Available |
| Security Headers | `internal/utils/security_utils.go` | ⚠️ Defined, not applied as middleware |

## Security Controls Implemented

### 1) Authentication and Authorization

Implemented:
- Keycloak-backed JWT validation with JWKS refresh and role extraction.
  - internal/auth/auth.go
- Role middleware helpers for single and multi-role checks.
  - internal/auth/middleware.go
- Main runtime route protection uses auth, admin, and admin-or-system-manager middleware.
  - main.go
- Full IAM system with OIDC discovery, OAuth flows, admin operations, user/client/role management.
  - internal/iam/

Notes:
- API route groups in many internal modules expose plain Register...Routes functions and rely on caller wiring to add auth middleware.

### 2) SQL Execution Controls

Implemented:
- API Builder SQL templates are validated as read-only with policy modes compat and strict.
  - internal/handlers/api_builder_handler.go
- Runtime custom API execution uses stored templates and parameter extraction, then validates placeholder count.
  - internal/handlers/api_builder_handler.go
- Dynamic SQL GET path is read-only.
  - internal/handlers/dynamic_query_handler.go
- Dynamic SQL POST and batch routes are privileged at router layer.
  - main.go
- SQL identifier validation (ValidateIdentifier) prevents injection in quality rules engine.
  - internal/quality/rules/sanitize.go

### 3) File Upload and Malware Scanning

Implemented:
- SafeGate pipeline for uploads includes metadata, MIME, SVG, macro, archive, and ClamAV scanners.
  - internal/handlers/api_builder_handler.go
  - internal/scanner/scanner.go

### 4) Rate Limiting

Implemented:
- Token usage tracking and expiry window enforcement.
  - internal/auth/rate_limit.go
- Combined token validation and rate-limit middleware exists for auth flows.
  - internal/auth/rate_limit_middleware.go
- Per-custom-API runtime rate limiting is enforced in API Builder invocation path.
  - internal/handlers/api_builder_handler.go

### 5) Encryption at Rest

Implemented:
- AES-256-GCM field-level encryption with key rotation and audit log.
  - internal/encryption/
- Encryption key and policy management via reconciled resources.
  - EncryptionKeyResource, EncryptionPolicyResource via GenericController
- Rotating keyring with key-ID prefixed ciphertexts.
  - internal/keyring/keyring.go

### 6) Audit and Compliance

Implemented:
- Audit log handlers with report/query/delete support (GDPR/HIPAA/SOC2/PCI-DSS frameworks).
  - internal/audit/
- Compliance engine with rule-based checks and remediation plans.
  - internal/policies/compliance_engine.go
- Security policy engine with threat modeling and incident tracking.
  - internal/policies/security/

## Priority Findings

### Critical

1. Hardcoded CLI JWT key and default admin account
- Evidence:
  - internal/handlers/cli_auth_handler.go
- Why this is critical:
  - A static signing secret and default admin credentials allow token forgery and unauthorized access if exposed.

2. Insecure default credential fallbacks in runtime config
- Evidence:
  - internal/config/config.go
- Why this is critical:
  - Default root/postgres/oracle-style credentials increase risk of insecure deployment if env overrides are missing.

### High

3. Validate token endpoint is not doing cryptographic validation
- Evidence:
  - internal/handlers/auth_handler.go
  - Function: ValidateToken
- Why this is high:
  - Endpoint currently responds success when token is present, creating false assurance for clients.

4. Custom API auth_required flag is not enforced in runtime invocation path
- Evidence:
  - Model field present in internal/handlers/api_builder_handler.go
  - Runtime InvokeCustomAPI enforces status and rate limit but does not evaluate auth_required per API policy.
- Why this is high:
  - Security metadata intent can diverge from effective behavior.

5. Query logs can capture sensitive SQL and parameter values without redaction
- Evidence:
  - internal/handlers/dynamic_query_handler.go
  - internal/handlers/query_logger.go
- Why this is high:
  - Persisted logs may contain secrets or personal data.

### Medium

6. Demo/platform local login branches exist when Keycloak-only mode is relaxed
- Evidence:
  - internal/handlers/auth_handler.go
  - internal/auth/auth.go (demo secret behavior)

7. Security headers middleware exists but is not applied
- Evidence:
  - internal/utils/security_utils.go (DefaultSecurityHeaders defined)
  - No Gin middleware applies them in main.go

8. math/rand used in anonymization masker (predictable noise)
- Evidence:
  - internal/anonymization/masker.go

## Startup Security Guardrails

The platform includes configurable security guardrails that run at startup.

Environment variables:
- `AXIOMNIZAM_ENV` / `APP_ENV` / `ENVIRONMENT` / `GO_ENV` — resolved in priority order
- `SECURITY_GUARDRAILS_MODE` — `off` (default dev), `audit` (default prod), `enforce`

Guardrails currently check for:
- `IAM_SYSADMIN_PASSWORD` quality (rejects empty/default-like values)
- `DEMO_JWT_SECRET` presence
- `CORS_ALLOWED_ORIGINS` presence
- Default-like DB passwords (MySQL, PostgreSQL, Oracle) as warnings

Recommended staged rollout:
1. Set AXIOMNIZAM_ENV=production in staging.
2. Set SECURITY_GUARDRAILS_MODE=audit in staging first.
3. Observe logs and fix all guardrail issues.
4. Switch SECURITY_GUARDRAILS_MODE=enforce only when clean.

## Security Test Coverage

Validated tests:
- internal/handlers/api_builder_sql_policy_test.go
- internal/rbac/handlers_access_requests_test.go
- main_rbac_access_requests_integration_test.go

Coverage gaps to add:
- auth_required enforcement behavior tests for InvokeCustomAPI
- auth/validate endpoint cryptographic validation tests
- Log redaction tests for query logger
- Guardrail enforce-mode startup behavior tests

## Remediation Plan

### Phase 0 (Immediate)

1. Replace hardcoded CLI JWT key with environment/secret-manager sourced value.
2. Remove default admin credential bootstrap in CLI auth or force explicit initialization.
3. Implement real token validation in auth validate endpoint using existing validator.
4. Add log redaction for SQL parameters and sensitive fields.
5. Wire SecurityHeadersMiddleware in main.go (code exists in utils).

### Phase 1 (Short-term)

1. Enforce auth_required in custom API runtime invocation.
2. Tighten production defaults in configuration loading.
3. Replace math/rand with crypto/rand in anonymization masker.
4. Add MaxBodySizeMiddleware to API routes.
5. Wire LockoutPolicy into auth handler login flow.
6. Add SSRF private IP blocking to OAuth URLs.

### Phase 2 (Hardening)

1. Run staging with AXIOMNIZAM_ENV=production and SECURITY_GUARDRAILS_MODE=audit until clean.
2. Move to SECURITY_GUARDRAILS_MODE=enforce in staging, then production.
3. Add periodic security regression checks to CI (gosec, semgrep, govulncheck).
4. Implement secret rotation reconciler for encryption keys and JWT signing keys.

## Verification Checklist

- Dynamic SQL write endpoints require privileged roles.
- API Builder runtime rejects non-read-only SQL templates.
- auth/validate performs actual token verification.
- No hardcoded signing keys or default admin credentials remain.
- Query logs are redacted for sensitive fields.
- Guardrails are clean in audit mode before enforce rollout.
- Security headers (CSP, X-Frame-Options, HSTS, etc.) are set on all responses.

## Ownership

- Platform team: route protections, guardrails, runtime policy enforcement.
- Security team: secret lifecycle, log redaction policy, review cadence.
- Data/API team: SQL policy maintenance and runtime API behavior tests.
