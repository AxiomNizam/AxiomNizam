# 🔐 Rate Limiting - Quick Reference

**What**: 500 calls per token, 10 minutes validity  
**Public APIs**: 2 endpoints don't need auth  
**Protected APIs**: Everything else needs valid token with calls remaining  

---

## Quick Start

### 1. Login
```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq
```

**Get**: `access_token`, `rate_limit` info

### 2. Use Token
```bash
TOKEN="your_token_here"

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users"
```

**You get**: `X-RateLimit-Remaining: 499` (after 1 call)

### 3. Check Status
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.calls_remaining'
```

---

## The 2 Public Endpoints

No token needed:

```bash
GET /health
GET /status
```

---

## All Other Endpoints

Require valid token with remaining calls:

```bash
GET /api/mysql/users               ✅ Protected
POST /api/mysql/users              ✅ Protected (admin only)
POST /api/mysql/query              ✅ Protected
GET /api/admin/metrics/count       ✅ Protected (admin only)
# ... and all other endpoints
```

---

## When Token Expires/Hits Limit

You get `401 Unauthorized`:

```json
{
  "error": "token expired",
  "message": "please login again to get a new token"
}
```

Or:

```json
{
  "error": "api call limit exceeded",
  "message": "you have used all 500 api calls. login again for fresh 500 calls"
}
```

**Solution**: Login again

```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

---

## Response Headers

Every protected API response:

```
X-RateLimit-Limit: 500              (always 500)
X-RateLimit-Remaining: 495          (calls left)
X-Token-Expires-At: <timestamp>     (when token expires)
```

---

## Rate Limit Logic

```
Request with Token
  ↓
✅ Token registered? (was in /auth/login)
✅ Token expired? (< 10 min)
✅ Calls remaining? (> 0)
✅ JWT valid?
  ↓
Decrement calls by 1
  ↓
Execute API
  ↓
Return response
```

If any check fails → **401 Unauthorized**

---

## Example: 5 Calls

```bash
TOKEN=$(curl -s -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.access_token')

# After each curl, calls_remaining decreases by 1

curl -H "Authorization: Bearer $TOKEN" "http://localhost:8000/api/mysql/users"
# Remaining: 499

curl -H "Authorization: Bearer $TOKEN" "http://localhost:8000/api/mysql/users"
# Remaining: 498

curl -H "Authorization: Bearer $TOKEN" "http://localhost:8000/api/mysql/users"
# Remaining: 497

curl -H "Authorization: Bearer $TOKEN" "http://localhost:8000/api/mysql/users"
# Remaining: 496

curl -H "Authorization: Bearer $TOKEN" "http://localhost:8000/api/mysql/users"
# Remaining: 495
```

---

## Monitor Usage

```bash
# Your token status
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data'

# Output:
{
  "username": "admin",
  "calls_made": 47,
  "max_calls": 500,
  "calls_remaining": 453,
  "issued_at": "2024-01-23T12:15:30Z",
  "expires_at": "2024-01-23T12:25:30Z",
  "is_expired": false,
  "time_remaining": "8m45s",
  "last_used": "2024-01-23T12:24:15Z"
}
```

---

## Admin: View All Tokens

```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "http://localhost:8000/auth/admin/tokens-status" | jq '.data'

# Output:
{
  "active_tokens": 3,
  "total_api_calls": 847,
  "tokens": [
    {
      "username": "admin",
      "calls_made": 47,
      "calls_remaining": 453,
      "expires_at": "2024-01-23T12:25:30Z",
      "time_remaining": "8m45s"
    },
    {
      "username": "user1",
      "calls_made": 156,
      "calls_remaining": 344,
      "expires_at": "2024-01-23T12:28:15Z",
      "time_remaining": "11m30s"
    }
  ]
}
```

---

## Configuration

**In main.go**:

```go
// Current:
rateLimiter := auth.NewRateLimiter(500, 10)

// 500 = max calls per token
// 10 = validity in minutes
```

---

## Key Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/health` | GET | NO | Public health check |
| `/status` | GET | NO | Public status |
| `/auth/login` | POST | NO | Get token |
| `/auth/token-status` | GET | YES | Your usage |
| `/auth/admin/tokens-status` | GET | YES | All tokens (admin) |
| All others | * | YES | Protected APIs |

---

## Limits

- **Max calls**: 500 per token
- **Validity**: 10 minutes
- **Auto-cleanup**: Every 5 minutes (expired tokens removed)

---

## Common Tasks

### Check how many calls left
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.calls_remaining'
```

### Check when token expires
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq '.data.expires_at'
```

### Use all 500 calls then get more
```bash
# Make 500 API calls...
# Then:
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
# Get fresh token with 500 new calls
```

### Monitor real-time usage
```bash
watch -n 1 'curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/auth/token-status | jq ".data.calls_remaining"'
```

---

**Read**: `RATE_LIMITING_GUIDE.md` for complete documentation

✅ Ready to use!
