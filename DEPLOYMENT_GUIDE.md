# Dynamic Queries - Configuration & Deployment Guide

## ✅ Deployment Checklist

- [x] Handler implementation complete (`dynamic_query_handler.go`)
- [x] Routes registered in `main.go`
- [x] Authentication integrated
- [x] Error handling implemented
- [x] Documentation created
- [x] Examples provided
- [x] Postman collection ready
- [ ] Tested in your environment
- [ ] Deployed to production

---

## 🔧 Configuration

### No Additional Configuration Required!

The dynamic query handler uses your existing setup:

```
✓ Database connections from: internal/database/connections.go
✓ Authentication from: internal/auth/middleware.go
✓ Environment variables from: .env file
✓ CORS headers already configured
✓ Error handling already in place
```

### Verify Your .env Contains

```env
# Database URLs
MYSQL_URL=mysql://user:password@localhost:3306/database
MARIADB_URL=mysql://user:password@localhost:3307/database
POSTGRES_URL=postgres://user:password@localhost:5432/database
PERCONA_URL=mysql://user:password@localhost:3308/database
ORACLE_URL=oracle://user:password@localhost:1521/ORCLCDB

# API
API_PORT=8000
API_HOST=0.0.0.0

# Authentication (if using Keycloak)
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=master
KEYCLOAK_CLIENT_ID=backend
```

---

## 🚀 Deployment Steps

### 1. Rebuild the Application
```bash
# Navigate to project root
cd /path/to/AxiomNizam

# Download dependencies (if needed)
go mod download

# Build the application
go build -o axiomnizam main.go

# Or run directly
go run main.go
```

### 2. Verify Routes are Registered
When you start the app, you should see in console:

```
Dynamic Query endpoints (authenticated users):
  GET  /api/{db}/query            - Execute SELECT queries with parameters
  POST /api/{db}/query            - Execute any query
  POST /api/{db}/query/batch      - Execute multiple queries at once
  GET  /api/{db}/schema           - Get table schema
  Available databases: mysql, mariadb, postgres, percona, oracle
```

### 3. Test Connection
```bash
# Test health check
curl http://localhost:8000/health

# Test status
curl http://localhost:8000/status
```

### 4. Get Authentication Token
```bash
# From Keycloak or your auth provider
curl -X POST http://localhost:8080/auth/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=backend&client_secret=secret&grant_type=client_credentials"

# Save the access_token from response
TOKEN="your_access_token_here"
```

### 5. Test Dynamic Query Endpoint
```bash
# Test GET endpoint
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%20*%20FROM%20users%20LIMIT%205"

# Should return JSON with user data
```

---

## 📊 Environment Setup

### Development Environment
```bash
# Docker compose will handle all databases
docker-compose up

# Verify all services are running
docker-compose ps

# Check logs
docker-compose logs axiomnizam
```

### Production Environment
```bash
# Build container
docker build -t axiom-nizam-prod .

# Run with environment variables
docker run -d \
  --env-file .env \
  --network app-network \
  -p 8000:8000 \
  axiom-nizam-prod

# Monitor health
curl http://localhost:8000/health
```

---

## 🔒 Security Configuration

### For Production:

1. **Enable HTTPS**
   - Use reverse proxy (nginx, Traefik)
   - Install SSL certificates
   - Force HTTPS only

2. **Configure CORS Carefully**
   - In production, replace `*` with specific domains
   - Current in main.go: `Access-Control-Allow-Origin: *`
   - Change to: `Access-Control-Allow-Origin: https://yourdomain.com`

3. **API Rate Limiting**
   - Implement rate limiting per user/IP
   - Recommended: 100 queries per minute per user
   - Use middleware or reverse proxy

4. **Query Timeout**
   - Set database query timeout (5-30 seconds)
   - Prevent long-running queries from blocking

5. **Audit Logging**
   - Log all queries executed
   - Track user actions
   - Monitor for suspicious patterns

---

## 📝 Suggested Production Code Addition

Add this to `main.go` for production security:

```go
// Before router.Run()

// Add rate limiting middleware
rateLimiter := func(c *gin.Context) {
    // Implement your rate limiting logic
    // Example using gin-ratelimit
    c.Next()
}
router.Use(rateLimiter)

// Add query logging middleware
queryLogger := func(c *gin.Context) {
    if c.Request.URL.Path == "/api/mysql/query" || 
       c.Request.URL.Path == "/api/postgres/query" {
        // Log the query being executed
        log.Printf("User: %v, Path: %v, Method: %v", 
            c.GetString("user_id"), 
            c.Request.URL.Path,
            c.Request.Method)
    }
    c.Next()
}
router.Use(queryLogger)

// Restrict CORS for production
if os.Getenv("ENV") == "production" {
    router.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
        c.Next()
    })
}
```

---

## 🧪 Testing Procedures

### Unit Test Example
```go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestDynamicQuery(t *testing.T) {
    // Create test database
    db := setupTestDB()
    handler := NewDynamicQueryHandler(db)
    
    // Test execution
    // Assert results
    assert.NotNil(t, handler)
}
```

### Integration Test Checklist
- [ ] GET SELECT query works
- [ ] GET with parameters works
- [ ] GET rejects INSERT query
- [ ] POST SELECT query works
- [ ] POST INSERT query works
- [ ] POST UPDATE query works
- [ ] POST DELETE query works
- [ ] POST BATCH queries work
- [ ] Schema endpoint returns correct structure
- [ ] Authentication blocks unauthorized requests
- [ ] Error messages are clear

### Manual Testing in Postman
1. Import `DYNAMIC_QUERIES_POSTMAN.json`
2. Set `token` variable with valid JWT
3. Run all requests in the collection
4. Verify responses match documentation

---

## 📊 Monitoring & Logs

### What to Monitor

```bash
# Watch application logs
docker-compose logs -f axiomnizam

# Monitor database connections
docker-compose logs -f mysql
docker-compose logs -f postgres

# Check resource usage
docker stats
```

### Key Metrics to Track

1. **Query Execution Time**
   - Log duration of each query
   - Alert if exceeds threshold

2. **Error Rate**
   - Track failed queries
   - Monitor SQL syntax errors

3. **Connection Pool**
   - Active connections count
   - Connection pool exhaustion

4. **User Activity**
   - Queries per user
   - Most used endpoints

---

## 🐛 Troubleshooting Deployment

### Issue: "Database not connected"
```bash
# Check database service is running
docker-compose ps

# Check connection string
echo $MYSQL_URL

# Test connection manually
mysql -h localhost -u root -proot

# Check logs
docker-compose logs mysql
```

### Issue: "Unauthorized" error
```bash
# Verify token is valid
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/health

# Check Keycloak is running
curl http://localhost:8080/health/ready

# Verify token expiry
# Token TTL is in response: "expires_in": 3600
```

### Issue: "Query execution failed"
```bash
# Test query directly in database client
mysql -h localhost -u root -proot -e "SELECT * FROM users"

# Check GORM logs (enable in code)
db.Logger = logger.Default.LogMode(logger.Info)

# Verify table exists
SHOW TABLES;
```

### Issue: Slow Queries
```bash
# Enable slow query log
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 2;

# Check slow query log
SHOW VARIABLES LIKE 'slow_query_log_file';
```

---

## 🔄 Updating the Code

If you need to modify the handler:

1. **Edit `dynamic_query_handler.go`**
2. **Rebuild**: `go build`
3. **Test**: Run in Postman
4. **Deploy**: Push to your server

Changes to routes require editing `main.go`:
```go
// Example: Adding rate-limited dynamic query endpoint
router.POST("/api/mysql/query/admin", 
    authMiddleware,
    rateLimitMiddleware,  // Add this
    mysqlDynamicHandler.DynamicQueryWithBody)
```

---

## 📈 Scaling Considerations

### For High Volume

1. **Database Connection Pooling**
   - GORM handles this automatically
   - Adjust pool size in connection string:
   ```go
   db.DB().SetMaxOpenConns(100)
   db.DB().SetMaxIdleConns(10)
   ```

2. **Query Caching**
   - Implement Redis caching layer
   - Cache SELECT query results

3. **Load Balancing**
   - Run multiple instances of API
   - Use nginx/HAProxy for load balancing

4. **Database Replication**
   - Set up master-slave replication
   - Read from slaves, write to master

---

## 📚 Additional Resources

- [GORM Documentation](https://gorm.io)
- [Gin Framework Docs](https://gin-gonic.com)
- [Go Database/SQL Docs](https://golang.org/pkg/database/sql)
- [SQL Injection Prevention](https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html)

---

## 🎯 Next Phase Improvements

Potential enhancements:

1. **Query Caching**
   - Cache SELECT results in Valkey
   - Auto-invalidate on INSERT/UPDATE

2. **Query Analytics**
   - Track query patterns
   - Suggest indexes

3. **Advanced Filtering**
   - GraphQL support
   - Complex filter DSL

4. **Transaction Support**
   - Execute batch queries in transaction
   - Rollback on error

5. **Webhook Notifications**
   - Trigger webhooks on query results
   - Event-driven architecture

---

## ✅ Final Checklist Before Production

- [ ] All environment variables configured
- [ ] Database connections tested
- [ ] Authentication configured and tested
- [ ] HTTPS/TLS enabled
- [ ] Rate limiting configured
- [ ] Logging enabled
- [ ] Monitoring alerts set up
- [ ] Backup procedures documented
- [ ] Disaster recovery plan ready
- [ ] Team trained on new endpoints
- [ ] Documentation deployed with code
- [ ] Rollback plan prepared

---

## 🎓 Training for Your Team

Share these files with your team:

1. **DYNAMIC_QUERIES_QUICK_START.md** - For quick reference
2. **DYNAMIC_QUERY_API.md** - For detailed API docs
3. **VISUAL_GUIDE.md** - For understanding architecture
4. **DYNAMIC_QUERIES_POSTMAN.json** - For testing
5. **This file** - For deployment/operations

---

**Deployment Guide Version**: 1.0  
**Last Updated**: January 23, 2026  
**Status**: Ready for Deployment ✅
