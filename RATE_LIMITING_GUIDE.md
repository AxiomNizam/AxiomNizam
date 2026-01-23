# 🔐 API Rate Limiting & Token Validity System

**Status**: ✅ Complete & Production Ready  
**Date**: January 23, 2026  
**Version**: 1.0

---

## Overview

Your API now has a **rate limiting and token validity system**:

- **500 API calls per token**: Each token can make exactly 500 authenticated API requests
- **10-minute token validity**: Tokens expire after 10 minutes of issuance
- **Public endpoints**: 2 endpoints remain public (no token needed):
  - `GET /health`
  - `GET /status`
- **Protected endpoints**: All other APIs require valid tokens with remaining calls

---

## How It Works

### Token Lifecycle

```
1. User calls /auth/login
   ↓
2. Token issued with:
   - 500 API calls available
   - 10 minutes validity
   - Timestamp recorded
   ↓
3. User makes API calls with token
   ↓
4. Each call:
   - Checks if token is still valid (< 10 min)
   - Checks if calls remaining > 0
   - Decrements call count by 1
   - Returns X-RateLimit headers
   ↓
5. When expired OR calls = 0:
   - Returns 401 Unauthorized
   - User must login again
```

### Call Flow

```
Request with Authorization: Bearer <TOKEN>
  ↓
Check 1: Is token registered?
  ├─ NO → 401 Unauthorized
  └─ YES → Continue
  ↓
Check 2: Is token expired? (> 10 minutes)
  ├─ YES → 401 Unauthorized ("token expired")
  └─ NO → Continue
  ↓
Check 3: Are calls remaining? (> 0)
  ├─ NO → 401 Unauthorized ("api call limit exceeded")
  └─ YES → Continue
  ↓
Check 4: Is JWT valid?
  ├─ NO → 401 Unauthorized
  └─ YES → Continue
  ↓
Increment call count
  ↓
Execute handler
  ↓
Return response with:
├─ X-RateLimit-Limit: 500
├─ X-RateLimit-Remaining: <remaining>
└─ X-Token-Expires-At: <timestamp>
```

---

## Login Response

When you login successfully:

```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

You get:

```json
{
  "status": "ok",
  "access_token": "eyJ0eXAiOiJKV1QiLCJhbGc...",
  "expires_in": 300,
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5...",
  "token_type": "Bearer",
  "username": "admin",
  "rate_limit": {
    "max_calls": 500,
    "validity_min": 10,
    "expires_at": "2024-01-23 12:25:30",
    "message": "You have 500 API calls available with this token. Token expires in 10 minutes."
  }
}
```

**Key Info**:
- `access_token`: Use this in Authorization header
- `rate_limit.max_calls`: 500 calls available
- `rate_limit.validity_min`: Expires in 10 minutes
- `rate_limit.expires_at`: Exact expiration time

---

## Using Your Token

### Make API Call

```bash
TOKEN="eyJ0eXAiOiJKV1QiLCJhbGc..."

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users" | jq
```

### Response Headers

Every successful API response includes:

```
X-RateLimit-Limit: 500
X-RateLimit-Remaining: 499
X-Token-Expires-At: 2024-01-23 12:25:30
```

**What these mean**:
- `X-RateLimit-Limit`: Maximum calls per token (always 500)
- `X-RateLimit-Remaining`: Calls still available
- `X-Token-Expires-At`: When token expires

---

## Checking Token Status

### Your Token Status

```bash
TOKEN="your_token_here"

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq
```

**Response**:

```json
{
  "status": "ok",
  "data": {
    "username": "admin",
    "calls_made": 47,
    "max_calls": 500,
    "calls_remaining": 453,
    "issued_at": "2024-01-23T12:15:30Z",
    "expires_at": "2024-01-23T12:25:30Z",
    "is_expired": false,
    "time_remaining": "9m45s",
    "last_used": "2024-01-23T12:24:15Z"
  }
}
```

**What it shows**:
- `calls_made`: How many API calls you've made
- `calls_remaining`: How many you can still make
- `is_expired`: Whether token has expired
- `time_remaining`: How long until token expires
- `last_used`: When this token was last used

### All Active Tokens (Admin Only)

```bash
TOKEN="admin_token_here"

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/admin/tokens-status" | jq
```

**Response**:

```json
{
  "status": "ok",
  "data": {
    "active_tokens": 5,
    "total_api_calls": 847,
    "tokens": [
      {
        "username": "admin",
        "calls_made": 47,
        "calls_remaining": 453,
        "expires_at": "2024-01-23T12:25:30Z",
        "time_remaining": "9m45s"
      },
      {
        "username": "user1",
        "calls_made": 156,
        "calls_remaining": 344,
        "expires_at": "2024-01-23T12:28:15Z",
        "time_remaining": "12m30s"
      }
    ]
  }
}
```

---

## Error Scenarios

### Error 1: Token Expired

```
Status: 401 Unauthorized

{
  "error": "token expired",
  "message": "your token is no longer valid. please login again to get a new token",
  "expired_at": "2024-01-23 12:25:30"
}
```

**Solution**: Login again to get a new token

```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Error 2: API Call Limit Exceeded

```
Status: 401 Unauthorized

{
  "error": "api call limit exceeded",
  "message": "you have used all 500 api calls allowed per token",
  "calls_limit": 500,
  "expires_at": "2024-01-23 12:25:30",
  "action_required": "login again to get a fresh token with new 500 calls",
  "action_endpoint": "/auth/login"
}
```

**Solution**: Login again to get a fresh token with 500 new calls

```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Error 3: Missing Token

```
Status: 401 Unauthorized

{
  "error": "missing authorization header"
}
```

**Solution**: Include Authorization header

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8000/api/mysql/users"
```

### Error 4: Invalid Token

```
Status: 401 Unauthorized

{
  "error": "invalid or unregistered token"
}
```

**Solution**: Use a valid token from `/auth/login`

---

## Public Endpoints (No Token Required)

These 2 endpoints don't need authentication:

```bash
# Health check
GET /health

# Status info
GET /status
```

**Examples**:

```bash
curl http://localhost:8000/health
curl http://localhost:8000/status
```

---

## Protected Endpoints (Token Required)

All other endpoints require a valid token:

```bash
# All user endpoints
GET /api/mysql/users
GET /api/mysql/users/:id
POST /api/mysql/users
PUT /api/mysql/users/:id
DELETE /api/mysql/users/:id

# Similar for mariadb, postgres, percona, mongodb, firebase, oracle

# Dynamic queries
POST /api/mysql/query
POST /api/mariadb/query
# etc.

# Admin operations
POST /api/admin/database/create
POST /api/admin/table/create
# etc.

# Metrics endpoints
GET /api/admin/metrics/count
GET /api/admin/metrics/all
GET /api/admin/metrics/stats

# Query logs
GET /api/query-logs
GET /api/query-logs/:database
# etc.
```

---

## Implementation Details

### Key Components

**1. RateLimiter** (`internal/auth/rate_limit.go`)
- Tracks API call count per token
- Manages token validity (10 minutes)
- Thread-safe with sync.RWMutex
- Auto-cleanup of expired tokens every 5 minutes

**2. Rate Limit Middleware** (`internal/auth/rate_limit_middleware.go`)
- `RateLimitMiddleware()`: Only checks rate limits
- `CombinedAuthMiddleware()`: JWT + Rate limit validation

**3. Auth Handler Updates** (`internal/handlers/auth_handler.go`)
- `Login()`: Now registers token in rate limiter
- `GetTokenStatus()`: Check your token status
- `GetAllTokensStatus()`: Admin view of all tokens

**4. Main.go Integration**
- Initialize: `rateLimiter := auth.NewRateLimiter(500, 10)`
- Register middleware: Used for all protected endpoints

---

## Token Registration

When user logs in:

```go
// Internal process
1. Keycloak validates credentials
2. JWT token created
3. Token registered in rate limiter:
   - Username: admin
   - Call count: 0
   - Issued at: now
   - Last used: now
   - Max calls: 500
   - Validity: 10 minutes
4. Token returned to client
```

---

## Rate Limit Checking

On each protected API call:

```go
// Internal process
1. Extract token from Authorization header
2. Check: Token exists in rate limiter?
   - If not → 401 Unauthorized
3. Check: Token not expired? (< 10 min old)
   - If expired → 401 Unauthorized
4. Check: Calls remaining > 0?
   - If 0 → 401 Unauthorized
5. Validate JWT signature
   - If invalid → 401 Unauthorized
6. Increment call count by 1
7. Add response headers:
   - X-RateLimit-Limit: 500
   - X-RateLimit-Remaining: <new count>
   - X-Token-Expires-At: <expiration>
8. Execute handler
9. Return response
```

---

## Real-World Examples

### Example 1: Make 5 API Calls

```bash
TOKEN=$(curl -s -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.access_token')

# Call 1
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users" | jq '.data | {calls: .[].id}'

# Call 2
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users" | jq '.data | length'

# Call 3
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.calls_remaining'
# Output: 497

# Call 4 (check status again)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.calls_remaining'
# Output: 496

# Call 5
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users/:1" | jq '.data'

# Summary: 495 calls remaining
```

### Example 2: Exhaust 500 Calls

```bash
TOKEN="your_token_here"

# Make 500 calls
for i in {1..500}; do
  curl -s -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8000/health" > /dev/null
done

# Try call 501
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users"

# Response:
# {
#   "error": "api call limit exceeded",
#   "message": "you have used all 500 api calls allowed per token",
#   "calls_limit": 500,
#   "action_required": "login again to get a fresh token with new 500 calls",
#   "action_endpoint": "/auth/login"
# }
```

### Example 3: Wait for Token to Expire

```bash
TOKEN="your_token_here"

# Check status
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.time_remaining'
# Output: 9m30s

# Wait 10 minutes...

# Try to make request
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users"

# Response:
# {
#   "error": "token expired",
#   "message": "your token is no longer valid. please login again to get a new token",
#   "expired_at": "2024-01-23 12:25:30"
# }
```

---

## Configuration

### Default Settings

- **Max calls per token**: 500
- **Token validity**: 10 minutes
- **Cleanup interval**: 5 minutes (auto-cleanup of expired tokens)

### Change Configuration

To change these settings, modify in main.go:

```go
// Currently:
rateLimiter := auth.NewRateLimiter(500, 10)

// To change:
rateLimiter := auth.NewRateLimiter(1000, 30) // 1000 calls, 30 minutes
```

---

## Response Headers

Every authenticated API response includes rate limit headers:

```
X-RateLimit-Limit: 500
X-RateLimit-Remaining: 495
X-Token-Expires-At: 2024-01-23 12:25:30
```

Use these to:
- Track your usage in your client
- Warn user when approaching limit
- Calculate remaining time before expiration

---

## API Endpoints Summary

| Endpoint | Method | Auth? | Purpose |
|----------|--------|-------|---------|
| `/health` | GET | NO | Health check |
| `/status` | GET | NO | System status |
| `/auth/login` | POST | NO | Get token |
| `/auth/refresh` | POST | NO | Refresh token |
| `/auth/validate` | GET | NO | Validate token |
| `/auth/token-status` | GET | YES | Your token status |
| `/auth/admin/tokens-status` | GET | YES (admin) | All tokens status |
| `/api/*/users` | GET | YES | Read users |
| `/api/*/users` | POST | YES (admin) | Create user |
| `/api/*/users/:id` | PUT | YES (admin) | Update user |
| `/api/*/users/:id` | DELETE | YES (admin) | Delete user |
| `/api/*/query` | POST | YES | Execute query |
| `/api/admin/*` | * | YES (admin) | Admin operations |

---

## Troubleshooting

### Q: How do I know how many calls I have left?

A: Check token status endpoint:
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.calls_remaining'
```

Or look at response headers:
```bash
curl -i -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users" | grep "X-RateLimit"
```

### Q: My token expired, what do I do?

A: Login again:
```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Q: Can I extend my token?

A: No, but you can refresh it. Use the `/auth/refresh` endpoint:
```bash
curl -X POST http://localhost:8000/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"your_refresh_token"}'
```

### Q: I used all 500 calls, how do I continue?

A: Login again to get a new token with 500 fresh calls:
```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Q: Can I make more than 500 calls?

A: You need to manage multiple tokens. After token 1 hits 500 calls, login again to get token 2 (another 500 calls). You can use multiple tokens simultaneously.

### Q: Are there any endpoints that don't count against the limit?

A: Only `/health` and `/status` don't require authentication, so they don't count. All other endpoints require auth and consume calls.

---

## Security Notes

✅ **Tokens are registered at login**: Enables call tracking  
✅ **Expiration enforced**: 10-minute limit prevents token reuse  
✅ **Call limits enforced**: 500 calls per token prevents abuse  
✅ **JWT validation still required**: Signature and claims checked  
✅ **Admin-only endpoints protected**: Require admin role  
✅ **Thread-safe operations**: RWMutex protects concurrent access  
✅ **Auto-cleanup**: Expired tokens removed every 5 minutes  

---

## Monitoring

### Check Active Tokens (Admin)

```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "http://localhost:8000/auth/admin/tokens-status" | jq
```

### Monitor Your Usage

```bash
# Check how many calls you've made
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.calls_made'

# Check your success rate
TOKEN_STATUS=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data')
echo "Calls made: $(echo $TOKEN_STATUS | jq '.calls_made')"
echo "Calls remaining: $(echo $TOKEN_STATUS | jq '.calls_remaining')"
echo "Time remaining: $(echo $TOKEN_STATUS | jq '.time_remaining')"
```

---

## Summary

| Feature | Details |
|---------|---------|
| **Calls per token** | 500 |
| **Token validity** | 10 minutes |
| **Public endpoints** | /health, /status (2 total) |
| **Protected endpoints** | All others (require valid token with remaining calls) |
| **Auto cleanup** | Every 5 minutes |
| **Thread safe** | Yes (RWMutex) |
| **Admin monitoring** | Yes (/auth/admin/tokens-status) |
| **User self-service** | Yes (/auth/token-status) |

---

## Getting Started

1. **Login**: `POST /auth/login` → Get token
2. **Use token**: Add `Authorization: Bearer TOKEN` to requests
3. **Check status**: `GET /auth/token-status` → See calls remaining
4. **Manage calls**: Make API calls, track usage via response headers
5. **When done**: Login again for fresh 500 calls

---

**Status**: ✅ Complete & Production Ready  

🔐 Your API is now secured with rate limiting and token validity!
