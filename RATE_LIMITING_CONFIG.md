# Rate Limiting Configuration

## Environment Variables

Add these to your `.env` file to configure rate limiting:

```bash
# Maximum API calls allowed per token (default: 500)
RATE_LIMIT_MAX_CALLS=500

# Token validity in minutes (default: 10)
RATE_LIMIT_VALIDITY_MINUTES=10
```

## Configuration Examples

### Example 1: Strict Limits (Testing)
```bash
RATE_LIMIT_MAX_CALLS=10
RATE_LIMIT_VALIDITY_MINUTES=1
```
- 10 calls per token
- Token expires in 1 minute

### Example 2: Standard Limits (Default)
```bash
RATE_LIMIT_MAX_CALLS=500
RATE_LIMIT_VALIDITY_MINUTES=10
```
- 500 calls per token
- Token expires in 10 minutes

### Example 3: Generous Limits (Production)
```bash
RATE_LIMIT_MAX_CALLS=5000
RATE_LIMIT_VALIDITY_MINUTES=60
```
- 5000 calls per token
- Token expires in 60 minutes (1 hour)

### Example 4: Very Strict (High Security)
```bash
RATE_LIMIT_MAX_CALLS=100
RATE_LIMIT_VALIDITY_MINUTES=5
```
- 100 calls per token
- Token expires in 5 minutes

---

## Accessing Configuration in Code

The rate limiting values are loaded into the config on startup:

```go
cfg := config.LoadConfig()

maxCalls := cfg.RateLimiting.MaxCallsPerToken        // e.g., 500
validityMinutes := cfg.RateLimiting.TokenValidityMinutes  // e.g., 10
```

---

## Default Values

If environment variables are NOT set, these defaults are used:

| Variable | Default |
|----------|---------|
| `RATE_LIMIT_MAX_CALLS` | 500 |
| `RATE_LIMIT_VALIDITY_MINUTES` | 10 |

---

## Valid Values

- **RATE_LIMIT_MAX_CALLS**: Any positive integer (1-999999)
- **RATE_LIMIT_VALIDITY_MINUTES**: Any positive integer (1-1440)

---

## Error Handling

If an invalid value is provided (not a number):

```
⚠️  Invalid integer for RATE_LIMIT_MAX_CALLS: strconv.ParseInt: parsing "abc": invalid syntax, using default: 500
```

The system will:
1. Log a warning message
2. Use the default value
3. Continue operation normally

---

## Dynamic Behavior

The rate limiting configuration is read once when the application starts:

```
Application Startup
  ↓
Load .env file
  ↓
Parse config variables
  ↓
Initialize rate limiter with config values
  ↓
Start serving requests
```

To change rate limiting values, you must **restart the application**.

---

## Example .env File

Add these lines to your `.env`:

```bash
# Database connections...
MYSQL_HOST=localhost
MYSQL_PORT=3306
# ... etc ...

# Rate Limiting Configuration
RATE_LIMIT_MAX_CALLS=500
RATE_LIMIT_VALIDITY_MINUTES=10
```

---

## Checking Current Configuration

### Via Code
```go
fmt.Printf("Max calls per token: %d\n", cfg.RateLimiting.MaxCallsPerToken)
fmt.Printf("Token validity: %d minutes\n", cfg.RateLimiting.TokenValidityMinutes)
```

### Via Logs
Check application logs when it starts:
```
✅ Configuration loaded
  Max API calls per token: 500
  Token validity: 10 minutes
```

---

## Production Recommendations

- **Standard APIs**: Use default (500 calls, 10 min)
- **Public APIs**: Consider higher limits (1000+ calls)
- **Internal APIs**: Can use stricter limits (100 calls, 5 min)
- **High-traffic APIs**: Increase max calls (2000+ calls)
- **Security-critical APIs**: Decrease validity (5-10 min)

---

## Testing Rate Limits

```bash
# Set strict limits for testing
RATE_LIMIT_MAX_CALLS=5
RATE_LIMIT_VALIDITY_MINUTES=1
```

With these settings:
- Get token: `POST /auth/login`
- Make 5 calls to any protected endpoint
- 6th call → 401 Unauthorized
- Wait 1 minute → Token expires
- Login again for fresh token

---

## Monitoring

After starting with rate limit config:

```bash
# Check rate limit status
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/auth/token-status | jq '.data'

# Check all tokens (admin only)
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8000/auth/admin/tokens-status | jq '.data'
```

Both endpoints respect the configured limits.

---

## Summary

Rate limiting is now **fully configurable via environment variables**:

| What | Where |
|------|-------|
| Max calls per token | `RATE_LIMIT_MAX_CALLS` in `.env` |
| Token validity | `RATE_LIMIT_VALIDITY_MINUTES` in `.env` |
| Default (max calls) | 500 |
| Default (validity) | 10 minutes |
| How it's used | Loaded at startup via config |

No code changes needed - just update `.env` and restart!
