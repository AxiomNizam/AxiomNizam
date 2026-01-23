# ✅ Rate Limiting Configuration Made Dynamic

**Status**: ✅ Complete  
**Date**: January 23, 2026

---

## What Changed

### Before (Hardcoded)
```go
// main.go line 71 (BEFORE)
rateLimiter := auth.NewRateLimiter(500, 10)  // Hardcoded values
```

### After (Configurable)
```go
// main.go line 71 (AFTER)
rateLimiter := auth.NewRateLimiter(
    cfg.RateLimiting.MaxCallsPerToken,
    cfg.RateLimiting.TokenValidityMinutes,
)
```

---

## Implementation

### 1. Added Config Struct
**File**: `internal/config/config.go`

```go
type RateLimitingConfig struct {
    MaxCallsPerToken    int64
    TokenValidityMinutes int
}
```

### 2. Updated Config Loader
**File**: `internal/config/config.go`

```go
RateLimiting: RateLimitingConfig{
    MaxCallsPerToken:    getEnvInt("RATE_LIMIT_MAX_CALLS", 500),
    TokenValidityMinutes: getEnvInt("RATE_LIMIT_VALIDITY_MINUTES", 10),
},
```

### 3. Added Helper Function
**File**: `internal/config/config.go`

```go
func getEnvInt(key string, defaultValue int64) int64 {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    intVal, err := strconv.ParseInt(value, 10, 64)
    if err != nil {
        log.Printf("⚠️  Invalid integer for %s: %v, using default: %d", 
            key, err, defaultValue)
        return defaultValue
    }
    return intVal
}
```

### 4. Updated Imports
**File**: `internal/config/config.go`

Added: `"strconv"`

---

## Usage

### Set in .env File

```bash
# Maximum API calls allowed per token
RATE_LIMIT_MAX_CALLS=500

# Token validity in minutes
RATE_LIMIT_VALIDITY_MINUTES=10
```

### If Not Set

Defaults are used automatically:
- `RATE_LIMIT_MAX_CALLS` → 500
- `RATE_LIMIT_VALIDITY_MINUTES` → 10

### Examples

**Testing (Strict)**:
```bash
RATE_LIMIT_MAX_CALLS=10
RATE_LIMIT_VALIDITY_MINUTES=1
```

**Production (Generous)**:
```bash
RATE_LIMIT_MAX_CALLS=5000
RATE_LIMIT_VALIDITY_MINUTES=60
```

**High Security**:
```bash
RATE_LIMIT_MAX_CALLS=100
RATE_LIMIT_VALIDITY_MINUTES=5
```

---

## How It Works

```
1. Application Startup
   ↓
2. LoadConfig() reads .env file
   ↓
3. getEnvInt("RATE_LIMIT_MAX_CALLS", 500)
   ├─ If env var exists → parse and use
   ├─ If empty → use default 500
   └─ If invalid → log warning, use default
   ↓
4. RateLimiter initialized with loaded values
   ↓
5. All API calls use configured limits
```

---

## Files Modified

1. **internal/config/config.go**
   - Added `RateLimitingConfig` struct
   - Updated `Config` struct to include `RateLimiting` field
   - Added `getEnvInt()` helper function
   - Updated `LoadConfig()` to load rate limit settings
   - Added `import "strconv"`

2. **main.go**
   - Updated RateLimiter initialization to use config values instead of hardcoded 500, 10

---

## Files Created

1. **RATE_LIMITING_CONFIG.md**
   - Complete configuration guide
   - Examples and use cases
   - Best practices

---

## Error Handling

If someone sets an invalid value:

```bash
RATE_LIMIT_MAX_CALLS=abc  # Invalid!
```

Output:
```
⚠️  Invalid integer for RATE_LIMIT_MAX_CALLS: strconv.ParseInt: parsing "abc": invalid syntax, using default: 500
```

The system continues with the default value.

---

## Testing

```bash
# 1. Update .env
echo "RATE_LIMIT_MAX_CALLS=50" >> .env
echo "RATE_LIMIT_VALIDITY_MINUTES=5" >> .env

# 2. Rebuild
go build

# 3. Start and login
./axiomnizam

# 4. Test in another terminal
TOKEN=$(curl -s -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.access_token')

# 5. Check token status
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/auth/token-status | jq '.data'

# Should show 50 calls available, expires in ~5 minutes
```

---

## Dynamic vs Static

- **Static**: Change only at startup
- **Dynamic**: Would require hot-reload (not implemented)

To change rate limits:
1. Update `.env`
2. Restart application
3. New limits apply immediately

---

## Benefits

✅ **No code changes needed** - Just update `.env`  
✅ **Different environments** - Dev/test/prod can have different limits  
✅ **Easy management** - Change limits without recompiling  
✅ **Error handling** - Invalid values logged and defaults used  
✅ **Backwards compatible** - Defaults match original hardcoded values  

---

## Environment Variables Reference

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `RATE_LIMIT_MAX_CALLS` | int | 500 | API calls per token |
| `RATE_LIMIT_VALIDITY_MINUTES` | int | 10 | Token validity in minutes |

---

## Example .env File

```bash
# API Configuration
API_PORT=8000
API_HOST=0.0.0.0

# Database Configuration
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=root
MYSQL_DATABASE=app_db

# Keycloak Configuration
KEYCLOAK_HOST=localhost
KEYCLOAK_PORT=8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-backend
KEYCLOAK_CLIENT_SECRET=your-secret

# Rate Limiting Configuration
RATE_LIMIT_MAX_CALLS=500
RATE_LIMIT_VALIDITY_MINUTES=10

# ... other configs ...
```

---

## Accessing in Code

If you need to access the rate limit config anywhere:

```go
import "example.com/axiomnizam/internal/config"

cfg := config.LoadConfig()

maxCalls := cfg.RateLimiting.MaxCallsPerToken          // 500
validity := cfg.RateLimiting.TokenValidityMinutes      // 10

fmt.Printf("Rate limit: %d calls per %d minutes\n", 
    maxCalls, validity)
```

---

## Summary

✅ **Rate limiting is now configurable**  
✅ **Via environment variables (.env)**  
✅ **Defaults match original values**  
✅ **No code changes needed**  
✅ **Error handling included**  
✅ **Documentation complete**  

**To use**: Add `RATE_LIMIT_MAX_CALLS` and `RATE_LIMIT_VALIDITY_MINUTES` to your `.env` file!

---

**Read**: `RATE_LIMITING_CONFIG.md` for detailed configuration guide
