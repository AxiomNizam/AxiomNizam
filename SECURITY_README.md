# AxiomNizam Security Assessment README

Date: 2026-03-16

## Scope

This document captures the current security posture of the API platform based on code inspection of authentication, authorization, CORS, dynamic SQL, API Builder runtime execution, file scanning, and operational endpoints.

## Executive Summary

The platform has strong foundational controls (JWT auth, route-level RBAC, token rate limiting, and file scanning pipeline), but there are several high-risk exposures that should be remediated first:

1. Authenticated users can execute destructive SQL through dynamic query endpoints.
2. Authenticated users can execute arbitrary SQL through runtime custom API invocation.
3. Stored API security metadata (auth_required and rate_limit) is not enforced during runtime execution.
4. Multiple secrets/default credentials are present in repository configuration and fallback code.
5. CORS policy reflects arbitrary Origin while credentials are enabled.

## Priority Findings

## 1) Critical: Authenticated users can run destructive SQL

Severity: Critical

Evidence:
- Route protection is auth-only for write query endpoints:
  - [main.go](main.go#L454)
  - [main.go](main.go#L460)
  - [main.go](main.go#L466)
  - [main.go](main.go#L472)
  - [main.go](main.go#L478)
- SQL allowlist includes write/DDL operations:
  - [internal/handlers/dynamic_query_handler.go](internal/handlers/dynamic_query_handler.go#L459)

Risk:
- Any authenticated user can alter schema/data (CREATE, DROP, ALTER, TRUNCATE, REPLACE).

Mitigation:
- Split read and write endpoints.
- Require adminOrSys middleware on all write/DDL endpoints.
- Disable DDL by default with an explicit environment flag.
- Validate SQL with parser-backed policy per dialect.

## 2) Critical: Custom API runtime can execute arbitrary SQL

Severity: Critical

Evidence:
- Runtime endpoint is auth-only:
  - [main.go](main.go#L1012)
  - [main.go](main.go#L1013)
- Query accepted from caller body and executed:
  - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L851)
  - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L725)
  - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L780)

Risk:
- Arbitrary SQL execution by authenticated users against configured source DBs.

Mitigation:
- Replace free-form SQL with pre-approved parameterized templates.
- Restrict runtime invocation by role and per-API policy.
- Block write SQL unless explicitly approved per custom API.

## 3) High: API security metadata is stored but not enforced

Severity: High

Evidence:
- Metadata exists and is persisted:
  - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L51)
  - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L52)
  - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L403)
  - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L404)
- Runtime invoke path does not enforce these controls.

Risk:
- Endpoint-level auth/rate intent can drift from runtime behavior.

Mitigation:
- Enforce auth_required and rate_limit in runtime invocation path.
- Add negative/positive tests for both flags.

## 4) High: Secrets and defaults are exposed

Severity: High

Evidence:
- Secrets/defaults in env file:
  - [.env](.env#L84)
  - [.env](.env#L91)
  - [.env](.env#L115)
  - [.env](.env#L33)
  - [.env](.env#L54)
  - [.env](.env#L101)
- Hardcoded CLI auth key and default admin credentials:
  - [internal/handlers/cli_auth_handler.go](internal/handlers/cli_auth_handler.go#L38)
  - [internal/handlers/cli_auth_handler.go](internal/handlers/cli_auth_handler.go#L47)
- Fallback secret in auth handler:
  - [internal/handlers/auth_handler.go](internal/handlers/auth_handler.go#L42)

Risk:
- Credential leakage, unauthorized access, and weak production hygiene.

Mitigation:
- Rotate exposed secrets immediately.
- Move secrets to secret manager and remove from repo.
- Fail startup in production when defaults are detected.

## 5) High: CORS origin reflection with credentials enabled

Severity: High

Evidence:
- Reflected origin + credentials true:
  - [main.go](main.go#L129)
  - [main.go](main.go#L133)
  - [main.go](main.go#L134)

Risk:
- Increased CSRF/cross-origin abuse risk if trusted origin boundaries are weak.

Mitigation:
- Use strict allowlist of origins from environment.
- Disable credentials unless strictly required.

## 6) Medium: Token validate endpoint does not validate token

Severity: Medium

Evidence:
- Endpoint route:
  - [main.go](main.go#L325)
- Handler currently accepts presence of token header:
  - [internal/handlers/auth_handler.go](internal/handlers/auth_handler.go#L439)
  - [internal/handlers/auth_handler.go](internal/handlers/auth_handler.go#L452)
  - [internal/handlers/auth_handler.go](internal/handlers/auth_handler.go#L456)

Risk:
- False confidence for clients relying on this endpoint.

Mitigation:
- Reuse full validator logic and return verified claims/expiry only.

## 7) Medium: CLI auth verification is in-memory token map based

Severity: Medium

Evidence:
- In-memory token lookup behavior:
  - [internal/handlers/cli_auth_handler.go](internal/handlers/cli_auth_handler.go#L94)
  - [internal/handlers/cli_auth_handler.go](internal/handlers/cli_auth_handler.go#L108)
  - [internal/handlers/cli_auth_handler.go](internal/handlers/cli_auth_handler.go#L139)

Risk:
- Tokens are not independently verified for signature/expiry at verification time.

Mitigation:
- Validate JWT signature and exp on each request.
- Add jti-based revocation and remove default admin account.

## 8) Medium: Sensitive log content retention

Severity: Medium

Evidence:
- Keycloak response body log:
  - [internal/handlers/auth_handler.go](internal/handlers/auth_handler.go#L288)
- Query logs include raw query and params:
  - [internal/handlers/query_logger.go](internal/handlers/query_logger.go#L18)
  - [internal/handlers/query_logger.go](internal/handlers/query_logger.go#L19)
  - [internal/handlers/query_logger.go](internal/handlers/query_logger.go#L147)

Risk:
- Potential leakage of secrets/PII in logs and persisted stores.

Mitigation:
- Redact tokens/passwords and sensitive fields.
- Add PII-aware query logging policy.
- Tighten file permissions and retention controls.

## 9) Low-Medium: Some status/list endpoints may be overexposed

Severity: Low-Medium

Evidence:
- Notification status publicly exposed:
  - [main.go](main.go#L569)
- Builder scanner/csv listing broadly available to authenticated users:
  - [main.go](main.go#L990)
  - [main.go](main.go#L1004)
  - [main.go](main.go#L1005)

Risk:
- Operational metadata exposure.

Mitigation:
- Reclassify sensitivity and tighten with adminOrSys where needed.

## Existing Security Controls (Positive Coverage)

1. Central token validation and per-token rate checks:
   - [main.go](main.go#L229)
   - [main.go](main.go#L254)
   - [main.go](main.go#L278)

2. RBAC role enforcement for admin and system-manager classes:
   - [main.go](main.go#L340)
   - [main.go](main.go#L361)

3. SafeGate scanner pipeline (metadata, MIME, SVG, macro, archive, ClamAV):
   - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L168)
   - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L184)
   - [internal/handlers/api_builder_handler.go](internal/handlers/api_builder_handler.go#L975)
   - [internal/scanner/scanner.go](internal/scanner/scanner.go#L100)

## Recommended Remediation Plan

## Phase 0 (Immediate: 24-48 hours)

1. [x] Lock write/DDL query endpoints behind adminOrSys.
2. [x] Disable arbitrary SQL execution in runtime custom API path.
3. [~] Rotate exposed credentials and webhook/token secrets.
4. [x] Replace permissive CORS reflection with fixed allowlist.

### Phase 0 Implementation Notes

- Dynamic SQL write/batch endpoints now require `admin` or `system-manager` roles.
- API Builder runtime no longer accepts request-provided SQL text; it executes only stored `sql_template` values.
- Runtime SQL templates are restricted to read-only statements and must use parameter placeholders (`?`).
- Runtime calls now accept parameter values only (query/body params), with type validation from API Builder `query_params`.
- CORS now uses a fixed allowlist from `CORS_ALLOWED_ORIGINS` (no dynamic reflection).
- Hardcoded auth secret fallback was removed; demo JWT secret is now env-based or ephemeral.
- Public Discord webhook token was removed from `.env`.

### Remaining Operator Action

- Rotate and synchronize `KEYCLOAK_CLIENT_SECRET` in both Keycloak client settings and runtime environment.

## Phase 1 (Short-term: 1-2 weeks)

1. Enforce auth_required and rate_limit at custom API runtime.
2. Implement strict SQL policy parser and query class restrictions.
3. Harden CLI auth verification with full JWT validation.
4. Implement log redaction and sensitive-field filtering.

## Phase 2 (Mid-term: 2-4 weeks)

1. Introduce environment profile guardrails (production hard-fail on defaults).
2. Add security-focused integration tests for route protections and SQL controls.
3. Add endpoint exposure review and least-privilege policy matrix.

## Verification Checklist

- Write SQL routes return forbidden for non-admin users.
- Runtime custom API rejects unapproved SQL.
- CORS allows only configured origins.
- Token validation endpoint cryptographically validates tokens.
- No secrets/default credentials remain in tracked files.
- Logs are redacted for tokens/passwords/PII.

## Ownership Suggestion

- Platform team: route/middleware hardening, CORS, endpoint exposure review.
- Data team: dynamic SQL policy and query execution guardrails.
- Security team: secret rotation, log policy, verification controls.
