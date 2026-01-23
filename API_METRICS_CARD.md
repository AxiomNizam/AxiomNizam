# 📊 API Metrics - Quick Reference Card

## 🎯 Your 3 New Admin Endpoints

```
┌─────────────────────────────────────────────┐
│  GET /api/admin/metrics/count               │
│  ─────────────────────────────────────────  │
│  Returns: Total APIs & Usage Summary        │
│  Use For: "How many APIs?" "Total calls?"   │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  GET /api/admin/metrics/all                 │
│  ─────────────────────────────────────────  │
│  Returns: Detailed Metrics Per Endpoint     │
│  Use For: "Which is slowest?" "Most used?"  │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  GET /api/admin/metrics/stats               │
│  ─────────────────────────────────────────  │
│  Returns: Aggregated Statistics             │
│  Use For: "Overall stats?" "Filter by API?" │
└─────────────────────────────────────────────┘
```

---

## 🚀 5-Minute Setup

```bash
# 1. Get admin token
TOKEN=$(curl -s -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | \
  jq -r '.data.access_token')

# 2. Check how many APIs exist
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/count" | jq '.data.total_unique_endpoints'

# 3. See endpoint usage
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/count" | jq '.data.endpoint_usage'

# Done! ✅
```

---

## 📋 What Each Endpoint Returns

### #1 Count Endpoint
```
Endpoint: /api/admin/metrics/count

Returns:
{
  "total_unique_endpoints": 45,      ← How many unique APIs
  "total_api_calls": 8234,            ← Total calls to ALL APIs
  "endpoint_usage": {
    "/api/mysql/users": 1234,
    "/health": 3456,
    ...
  }
}
```

### #2 All Metrics Endpoint
```
Endpoint: /api/admin/metrics/all

Returns (for EACH endpoint):
{
  "endpoint": "/api/mysql/users",
  "method": "GET",
  "total_calls": 156,
  "success_calls": 150,
  "error_calls": 6,
  "average_duration_ms": 45,
  "max_duration_ms": 120,
  "min_duration_ms": 15,
  "last_called": "2024-01-01T12:34:56Z",
  "status_codes": {"200": 150, "400": 4, ...}
}
```

### #3 Stats Endpoint
```
Endpoint: /api/admin/metrics/stats?endpoint=/path/to/api

Returns:
{
  "total_calls": 245,
  "success_calls": 238,
  "error_calls": 7,
  "average_duration_ms": 52,
  "metrics": [ ... endpoint details ... ]
}
```

---

## 🎯 Quick Answers

| Question | Endpoint | What to Look At |
|----------|----------|-----------------|
| "How many APIs do I have?" | `/metrics/count` | `total_unique_endpoints` |
| "How many total calls?" | `/metrics/count` | `total_api_calls` |
| "Which API called most?" | `/metrics/count` | `endpoint_usage` (max value) |
| "Which API is slowest?" | `/metrics/all` | `average_duration_ms` (max) |
| "Which API has errors?" | `/metrics/all` | `error_calls` (max) |
| "Stats for API X?" | `/metrics/stats?endpoint=X` | Full response |
| "Overall health?" | `/metrics/stats` | Success/error ratio |

---

## 🔐 Authentication

```bash
# Required header for ALL 3 endpoints:
Authorization: Bearer <YOUR_ADMIN_TOKEN>

# Get token:
curl -X POST http://localhost:8000/auth/login \
  -d '{"username":"admin","password":"password"}'

# Use token:
curl -H "Authorization: Bearer <TOKEN>" \
  "http://localhost:8000/api/admin/metrics/count"
```

---

## 📊 Sample Responses

### Most Used APIs (From /metrics/count)
```json
"endpoint_usage": {
  "/health": 3456,           ← Most used
  "/api/mysql/users": 1234,
  "/api/mysql/query": 890,
  "/api/postgres/users": 567,
  "/api/admin/metrics/count": 87
}
```

### Slowest APIs (From /metrics/all, sorted)
```
1. /api/oracle/query:     max_duration_ms: 5432
2. /api/postgres/query:   max_duration_ms: 3456
3. /api/mysql/users:      max_duration_ms: 1234
```

### Error-Prone APIs (From /metrics/all)
```
1. /api/percona/query:  error_calls: 45
2. /api/oracle/query:   error_calls: 23
3. /api/mysql/users:    error_calls: 6
```

---

## 💻 Postman Setup

```
1. Import: API_METRICS_POSTMAN.json
2. Set Variable: admin_token = your_token
3. Click Send on any request
4. See results in response body
```

---

## 🎯 Common Tasks

### Task 1: List all APIs with call count
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/count" | \
  jq '.data.endpoint_usage'

# Shows: {"/api/mysql/users": 1234, "/health": 3456, ...}
```

### Task 2: Find slowest API
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/all" | \
  jq '.data.endpoints | max_by(.average_duration_ms) | 
      {endpoint: .endpoint, avg_ms: .average_duration_ms}'

# Shows: {"endpoint": "/api/oracle/query", "avg_ms": 287}
```

### Task 3: Find API with most errors
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/all" | \
  jq '.data.endpoints | max_by(.error_calls) | 
      {endpoint: .endpoint, errors: .error_calls}'

# Shows: {"endpoint": "/api/percona/query", "errors": 45}
```

### Task 4: Total API count
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/count" | \
  jq '.data.total_unique_endpoints'

# Shows: 45
```

---

## 🔧 Key Metrics Legend

```
total_calls         = How many times this API was called
success_calls       = Calls that returned 2xx (success)
error_calls         = Calls that returned 4xx or 5xx (errors)
average_duration_ms = Average execution time in milliseconds
max_duration_ms     = Slowest execution time
min_duration_ms     = Fastest execution time
last_called         = When this API was last used (ISO 8601)
status_codes        = Distribution of HTTP response codes
```

---

## ✅ Verification Checklist

- [ ] Can login and get admin token
- [ ] Can hit `/api/admin/metrics/count` successfully
- [ ] Can see `total_unique_endpoints` in response
- [ ] Can see `endpoint_usage` with API names
- [ ] Can hit `/api/admin/metrics/all` successfully
- [ ] Can see individual endpoint metrics
- [ ] Can hit `/api/admin/metrics/stats` successfully
- [ ] Postman collection works
- [ ] All endpoints require admin auth

---

## 📚 Documentation Map

| Doc | Purpose | Best For |
|-----|---------|----------|
| This card | Quick reference | Quick lookup |
| `API_METRICS_QUICK_START.md` | Getting started | First-time users |
| `API_METRICS.md` | Complete guide | Detailed learning |
| `API_METRICS_POSTMAN.json` | Hands-on testing | Postman users |
| `API_METRICS_SUMMARY.md` | Overview | Project overview |

---

## 🚀 You're Ready!

```
✅ 3 new endpoints created
✅ Automatic tracking enabled
✅ Admin authentication required
✅ Postman collection provided
✅ Documentation complete

→ Import Postman collection
→ Get admin token
→ Hit endpoints
→ View metrics

That's it! 🎉
```

---

## 🆘 Quick Troubleshooting

| Problem | Solution |
|---------|----------|
| 401 Unauthorized | Use admin token, not regular user token |
| 403 Forbidden | User doesn't have admin role |
| 500 Error | Check if Valkey/Redis is running |
| Empty response | Make some API calls first, then check metrics |
| No endpoint data | Metrics appear after requests are made |

---

**Start tracking now!** Use any of the 3 endpoints above.

Questions? Check the full docs → `API_METRICS.md`
