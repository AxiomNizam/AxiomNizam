# 🚀 Quick Start: Testing Query Persistence

## 1. Start the Application

```bash
cd /path/to/AxiomNizam
docker-compose up -d
```

Wait for all services to start (~30 seconds):
```bash
docker-compose logs -f axiomnizam | grep "listening"
```

## 2. Get Authentication Token

```bash
# Login to get JWT token
TOKEN=$(curl -s -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.data.access_token')

echo $TOKEN
```

Or use your existing Keycloak token.

## 3. Send Some Test Queries

### Test 1: Simple SELECT
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%201"
```

### Test 2: INSERT (creates a log entry)
```bash
curl -X POST http://localhost:8000/api/mysql/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
    "params": ["Test User", "test@example.com", 25]
  }'
```

### Test 3: SELECT with Filter
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%20*%20FROM%20users%20LIMIT%205"
```

## 4. Retrieve Query Logs

### Get All Logs
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/logs?limit=10" | jq
```

**Expected Response**:
```json
{
  "status": "ok",
  "data": [
    {
      "id": "axiomnizam-1704067200000-123",
      "query": "SELECT 1",
      "database": "mysql",
      "user_id": "admin",
      "status": "success",
      "duration": 15,
      "timestamp": "2024-01-01T12:00:00Z",
      "hostname": "axiomnizam"
    }
  ],
  "pagination": {"total": 3, "limit": 10, "offset": 0}
}
```

### Get Query Statistics
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/stats" | jq
```

**Expected Response**:
```json
{
  "status": "ok",
  "data": {
    "total_queries": 3,
    "success_count": 3,
    "error_count": 0,
    "average_duration_ms": 25,
    "min_duration_ms": 10,
    "max_duration_ms": 50
  }
}
```

## 5. Verify Persistence

### Check Disk Logs
```bash
# View log directory
docker exec axiomnizam ls -lah /data/query_logs/

# View actual log content (JSONL format)
docker exec axiomnizam cat /data/query_logs/2024-01-01.jsonl | head -5
```

Example output:
```json
{"id":"axiomnizam-1704067200000-123","query":"SELECT 1","params":[],"database":"mysql","user_id":"admin","status":"success","duration":15,"timestamp":"2024-01-01T12:00:00Z","hostname":"axiomnizam"}
```

### Check Valkey/Redis Logs
```bash
# Connect to Valkey
docker exec valkey redis-cli

# View all query log keys
> KEYS query_logs:*

# Get a specific log entry
> GET query_logs:mysql:1704067200:axiomnizam-123

# Count logs by database
> KEYS query_logs:mysql:*
> KEYS query_logs:postgres:*

# Exit
> EXIT
```

## 6. Test Multi-Pod Scalability (Optional)

### Scale to 3 Pods
```bash
docker-compose up -d --scale axiomnizam=3
```

### Send Queries to Each Pod
```bash
# Pod 1 (default)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%201" 

# All logs unified in Valkey and disk
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/logs" | jq '.data | length'
```

Each pod writes to:
- Shared disk volume: `/data/query_logs`
- Shared Valkey instance: `6379`

### Verify Pod Names in Logs
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/logs?limit=20" | jq '.data[].hostname' | sort | uniq
```

Output should show multiple pod hostnames:
```
"axiomnizam"
"axiomnizam_2"
"axiomnizam_3"
```

## 7. Test Data Persistence

### Restart Container
```bash
docker-compose down
docker-compose up -d
```

### Verify Logs Still Exist
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/logs" | jq '.data | length'
```

Logs should still be there (persisted in volume).

## 8. Monitor Real-time Logs

### Watch Disk Logs
```bash
docker exec axiomnizam tail -f /data/query_logs/*.jsonl
```

### Watch Valkey Updates
```bash
docker exec valkey redis-cli
> MONITOR

# In another terminal, send queries
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%201"

# You'll see Valkey operations in real-time
```

## Troubleshooting

### Logs Not Appearing
```bash
# Check if container is running
docker ps | grep axiomnizam

# Check logs for errors
docker-compose logs axiomnizam | tail -20

# Check if volume is mounted
docker inspect axiomnizam | grep -A 5 "Mounts"

# Check permissions
docker exec axiomnizam ls -la /data/query_logs
```

### Valkey Connection Error
```bash
# Check if Valkey is running
docker ps | grep valkey

# Test connection
docker exec valkey redis-cli ping
# Should return: PONG
```

### Token Error (401)
```bash
# Get a fresh token
TOKEN=$(curl -s -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r '.data.access_token')

# Verify token is valid
echo $TOKEN | jq .
```

## Performance Testing

### Send 100 Queries
```bash
for i in {1..100}; do
  curl -s -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8000/api/mysql/query?q=SELECT%20$i" > /dev/null
done

# Check total logged
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/stats" | jq '.data.total_queries'
```

### Check Performance Impact
```bash
# Measure query execution time
time curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%201"

# Compare to baseline (should be <50ms with logging)
```

## Complete API Reference

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/mysql/query` | GET/POST | Execute query |
| `/api/mysql/logs` | GET | Get query audit trail |
| `/api/mysql/stats` | GET | Get query statistics |
| `/api/mariadb/logs` | GET | MariaDB audit trail |
| `/api/postgres/logs` | GET | PostgreSQL audit trail |
| `/api/percona/logs` | GET | Percona audit trail |
| `/api/oracle/logs` | GET | Oracle audit trail |

## Success Indicators ✅

You've successfully implemented persistence when:

- [x] `/api/mysql/logs` returns your executed queries
- [x] `/api/mysql/stats` shows total_queries > 0
- [x] Logs exist in `/data/query_logs/` directory
- [x] Logs appear in Valkey: `redis-cli KEYS query_logs:*`
- [x] Logs persist after `docker-compose down` and up
- [x] Multiple queries show different `id` and `duration`
- [x] All logs have a `hostname` field
- [x] Scaling to multiple pods consolidates logs

---

**Next**: Check [PERSISTENCE_IMPLEMENTATION.md](PERSISTENCE_IMPLEMENTATION.md) for complete documentation.
