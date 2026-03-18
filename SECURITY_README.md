# AxiomNizam Security README

Date: 2026-03-18

## Scope and Method

This document is rewritten from a code-backed scan of the internal directory and current runtime wiring.

Scan coverage:
- internal Go files scanned: 330
- Focus areas: authentication, authorization, SQL execution paths, file scanning, logging, encryption, rate limiting, and configuration defaults

## Executive Summary

Current posture:
- Strong baseline controls exist (JWT validation, role middleware, admin/system-manager gates on privileged routes, SQL read-only policy checks for API Builder templates, SafeGate file scanning pipeline).
- Critical credential and token-hardening gaps remain (hardcoded CLI auth key/admin, insecure config defaults, secrets handling hygiene).
- Startup guardrails are implemented and can enforce in production, but rollout is currently designed to begin in audit mode.

Top risks to address first:
1. Hardcoded CLI auth secret and default admin credentials in internal handlers.
2. Insecure default credentials in configuration fallbacks.
3. Token validation helper endpoint that returns success for any presented bearer token.
4. Runtime custom API auth_required flag is stored but not enforced at invocation.
5. Query logging stores raw SQL and params, which may include sensitive data.

## Security Controls Implemented

### 1) Authentication and Authorization

Implemented:
- Keycloak-backed JWT validation with JWKS refresh and role extraction.
  - internal/auth/auth.go
- Role middleware helpers for single and multi-role checks.
  - internal/auth/middleware.go
- Main runtime route protection uses auth, admin, and admin-or-system-manager middleware.
  - main.go

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

### 5) Audit and Security Framework Components

Implemented as modules/framework:
- Audit log handlers and report/query/delete support.
  - internal/audit/handlers.go
- RLS manager and policy model.
  - internal/security/rls.go
- Encryption key and encrypt/decrypt APIs.
  - internal/encryption/handlers.go
  - internal/encryption/models.go

Note:
- Some framework endpoints are model-complete but have partial implementation details (see findings).

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
- Why this is medium:
  - Risk is configuration-dependent; if relaxed in production, local credentials/token modes can weaken identity controls.

7. Encryption policy endpoints include placeholder behavior
- Evidence:
  - internal/encryption/handlers.go
  - Comments: policy storage/list retrieval marked as to be implemented
- Why this is medium:
  - Partial policy management can create mismatched expectations about enforcement depth.

## Already Addressed Improvements

These are present and should be retained:
- CORS moved to explicit allowlist logic with origin checks.
  - main.go
- Dynamic SQL write endpoints restricted to admin/system-manager middleware.
  - main.go
- API Builder SQL policy supports compat and strict modes with statement classification and blocklists.
  - internal/handlers/api_builder_handler.go
- Security guardrails support off, audit, enforce with production-aware blocking in enforce mode.
  - main.go

## Guardrail Rollout (Staging First)

Recommended staged rollout:
1. Set AXIOMNIZAM_ENV=production in staging.
2. Set SECURITY_GUARDRAILS_MODE=audit in staging first.
3. Observe logs and fix all guardrail issues.
4. Switch SECURITY_GUARDRAILS_MODE=enforce only when clean.

Guardrails currently check for:
- KEYCLOAK_CLIENT_SECRET quality
- DEMO_JWT_SECRET presence
- CORS_ALLOWED_ORIGINS presence
- Default-like DB passwords as warnings

Environment note:
- .env now includes staged rollout comments and audit-first defaults for this flow.

## Security Test Coverage Snapshot

Validated tests relevant to current security controls:
- internal/handlers/api_builder_sql_policy_test.go
- internal/rbac/handlers_access_requests_test.go
- main_rbac_access_requests_integration_test.go

Coverage gaps to add next:
- auth_required enforcement behavior tests for InvokeCustomAPI
- auth/validate endpoint cryptographic validation tests
- log redaction tests for query logger
- guardrail enforce-mode startup behavior tests

## Remediation Plan

### Phase 0 (Immediate)

1. Replace hardcoded CLI JWT key with environment/secret-manager sourced value.
2. Remove default admin credential bootstrap in CLI auth or force explicit initialization.
3. Implement real token validation in auth validate endpoint using existing validator.
4. Add log redaction for SQL parameters and sensitive fields.
5. Rotate any exposed secrets and verify they are not committed in tracked files.

### Phase 1 (Short-term)

1. Enforce auth_required in custom API runtime invocation.
2. Tighten production defaults in configuration loading.
3. Complete encryption policy persistence and retrieval paths.
4. Add integration tests for role boundary behavior on high-impact endpoints.

### Phase 2 (Hardening)

1. Run staging with AXIOMNIZAM_ENV=production and SECURITY_GUARDRAILS_MODE=audit until clean.
2. Move to SECURITY_GUARDRAILS_MODE=enforce in staging, then production.
3. Add periodic security regression checks to CI.

## Verification Checklist

- Dynamic SQL write endpoints require privileged roles.
- API Builder runtime rejects non-read-only SQL templates.
- auth/validate performs actual token verification.
- No hardcoded signing keys or default admin credentials remain.
- Query logs are redacted for sensitive fields.
- Guardrails are clean in audit mode before enforce rollout.

## Ownership Suggestion

- Platform team: route protections, guardrails, runtime policy enforcement.
- Security team: secret lifecycle, log redaction policy, review cadence.
- Data/API team: SQL policy maintenance and runtime API behavior tests.
