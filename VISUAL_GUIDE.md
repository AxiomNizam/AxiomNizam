# Dynamic Queries - Visual Guide

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Postman/HTTP Client                      │
└────────────────────────────┬────────────────────────────────┘
                             │
                    ┌────────▼─────────┐
                    │  GET Request     │
                    │  (SELECT only)   │
                    └────────┬─────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼────┐         ┌────▼────┐       ┌─────▼─────┐
    │ /query  │         │ /query  │       │ /schema   │
    │ (GET)   │         │ (POST)  │       │ (GET)     │
    └────┬────┘         └────┬────┘       └─────┬─────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
         ┌───────────────────▼──────────────────┐
         │  DynamicQueryHandler                 │
         │  ├─ DynamicQuery()                   │
         │  ├─ DynamicQueryWithBody()           │
         │  ├─ BatchQueries()                   │
         │  └─ TableSchema()                    │
         └───────────────────┬──────────────────┘
                             │
     ┌───────────────────────┼───────────────────────┐
     │                       │                       │
  ┌──▼───┐  ┌────────┐  ┌───▼──┐  ┌────────┐  ┌────▼─────┐
  │MySQL │  │MariaDB │  │Postgres│ │Percona │  │ Oracle   │
  └──────┘  └────────┘  └────────┘ └────────┘  └──────────┘
     │        │             │          │             │
     └────────┴─────────────┴──────────┴─────────────┘
             Query Results as JSON
```

---

## Request Flow - GET Query

```
Step 1: User creates GET request
        URL: /api/mysql/query?q=SELECT * FROM users WHERE age > ?&params=25
        Headers: Authorization: Bearer TOKEN

Step 2: Gin router maps to DynamicQuery() handler

Step 3: Handler validates:
        ✓ Database connection exists
        ✓ Query is SELECT type
        ✓ User is authenticated

Step 4: Execute raw SQL with parameters
        Raw Query: SELECT * FROM users WHERE age > ?
        Params: [25]
        ↓
        GORM sends to MySQL/Postgres/etc

Step 5: Database executes query

Step 6: Handler scans results:
        Row 1: {id: 1, name: "John", age: 30}
        Row 2: {id: 2, name: "Jane", age: 28}

Step 7: Convert to JSON and return
        ✓ status: "ok"
        ✓ data: [rows]
        ✓ message: "Query executed successfully"
```

---

## Request Flow - POST Query

```
Step 1: User creates POST request
        URL: /api/mysql/query
        Body: {
          "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
          "params": ["Alice", "alice@example.com", 28]
        }
        Headers: Authorization: Bearer TOKEN

Step 2: Gin router maps to DynamicQueryWithBody() handler

Step 3: Handler validates:
        ✓ Database connection exists
        ✓ Query type is allowed (INSERT, UPDATE, etc.)
        ✓ Not DROP DATABASE or dangerous operations
        ✓ User is authenticated

Step 4: Check if SELECT or write operation
        
        If SELECT:
        ├─ Execute with Raw()
        ├─ Scan rows into map
        └─ Return results array
        
        If INSERT/UPDATE/DELETE:
        ├─ Execute with Exec()
        ├─ Return rows affected
        └─ Return confirmation

Step 5: Database executes query

Step 6: Return response with status
        ✓ status: "ok"
        ✓ data: { rows_affected: 1 }
        ✓ message: "Query executed successfully"
```

---

## Request Flow - Batch Queries

```
Step 1: User creates POST request with array
        URL: /api/mysql/query/batch
        Body: [
          {"query": "SELECT COUNT(*) FROM users", "params": []},
          {"query": "INSERT INTO users VALUES(?, ?, ?)", "params": [...]},
          {"query": "UPDATE users SET age = ?", "params": [...]}
        ]

Step 2: Handler validates all queries

Step 3: Loop through each query:
        
        Query 1 (SELECT):
        ├─ Execute
        ├─ Scan results
        └─ Add to batch results
        
        Query 2 (INSERT):
        ├─ Execute
        ├─ Track rows affected
        └─ Add to batch results
        
        Query 3 (UPDATE):
        ├─ Execute
        ├─ Track rows affected
        └─ Add to batch results

Step 4: Return consolidated results
        ✓ All SELECT results included
        ✓ Row counts for modifications
        ✓ Single response with all data
```

---

## Security Layers

```
┌────────────────────────────────────────────────────────┐
│ HTTP Request arrives at backend                        │
└────────┬───────────────────────────────────────────────┘
         │
┌────────▼───────────────────────────────────────────────┐
│ Layer 1: Authentication Check                          │
│ ├─ Verify Bearer token exists                          │
│ ├─ Validate token signature                            │
│ └─ Extract user claims                                 │
│ Result: If fails → 401 Unauthorized                    │
└────────┬───────────────────────────────────────────────┘
         │
┌────────▼───────────────────────────────────────────────┐
│ Layer 2: Database Connection Check                     │
│ ├─ Verify database is connected                        │
│ └─ Confirm driver is available                         │
│ Result: If fails → 503 Service Unavailable             │
└────────┬───────────────────────────────────────────────┘
         │
┌────────▼───────────────────────────────────────────────┐
│ Layer 3: Query Type Validation                         │
│ ├─ For GET: Only allow SELECT, SHOW, DESCRIBE         │
│ ├─ For POST: Allow SELECT, INSERT, UPDATE, DELETE etc │
│ └─ Block dangerous: DROP DATABASE                      │
│ Result: If fails → 403 Forbidden                       │
└────────┬───────────────────────────────────────────────┘
         │
┌────────▼───────────────────────────────────────────────┐
│ Layer 4: Parameterized Query Execution                 │
│ ├─ Separate query from parameters                      │
│ ├─ Use prepared statements                             │
│ └─ Database sanitizes values                           │
│ Result: SQL Injection Prevention ✓                     │
└────────┬───────────────────────────────────────────────┘
         │
┌────────▼───────────────────────────────────────────────┐
│ Layer 5: Error Handling                                │
│ ├─ Catch SQL errors safely                             │
│ ├─ Return generic messages (no SQL details in prod)    │
│ └─ Log errors for debugging                            │
│ Result: Secure error responses                         │
└────────┬───────────────────────────────────────────────┘
         │
┌────────▼───────────────────────────────────────────────┐
│ Return Safe JSON Response to Client                    │
└────────────────────────────────────────────────────────┘
```

---

## Endpoint Usage Decision Tree

```
                    ┌─ Do you want to READ (SELECT)?
                    │
Start ──────────────┤  Yes? ─┬─ Simple query? ──→ GET /api/{db}/query?q=SELECT...
                    │        │
                    │        └─ Complex query? ─→ POST /api/{db}/query
                    │
                    └─ Do you want to WRITE (INSERT/UPDATE/DELETE)?
                       │
                       └─ Use POST /api/{db}/query
                           │
                           └─ With multiple operations?
                               │
                               ├─ Yes ──→ POST /api/{db}/query/batch
                               │
                               └─ No ──→ POST /api/{db}/query
```

---

## Response Format Examples

### Success: GET SELECT Query
```json
{
  "status": "ok",
  "message": "Query executed successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "age": 28
    },
    {
      "id": 2,
      "name": "Jane Smith",
      "email": "jane@example.com",
      "age": 26
    }
  ]
}
```

### Success: POST INSERT Query
```json
{
  "status": "ok",
  "message": "Query executed successfully",
  "data": {
    "rows_affected": 1
  }
}
```

### Error: Invalid Query Type on GET
```json
{
  "status": "error",
  "error": "Only SELECT queries are allowed for GET requests. Use POST for INSERT/UPDATE/DELETE/CREATE"
}
```

### Error: Database Connection Issue
```json
{
  "status": "error",
  "error": "Database not connected"
}
```

---

## Database Support Matrix

```
┌─────────────┬────────────┬────────────┬────────────┬──────────────┐
│ Database    │ SELECT     │ INSERT     │ UPDATE     │ DELETE/ALTER │
├─────────────┼────────────┼────────────┼────────────┼──────────────┤
│ MySQL       │ ✅ Full    │ ✅ Full    │ ✅ Full    │ ✅ Full      │
│ MariaDB     │ ✅ Full    │ ✅ Full    │ ✅ Full    │ ✅ Full      │
│ PostgreSQL  │ ✅ Full    │ ✅ Full    │ ✅ Full    │ ✅ Full      │
│ Percona     │ ✅ Full    │ ✅ Full    │ ✅ Full    │ ✅ Full      │
│ Oracle      │ ✅ Full    │ ✅ Full    │ ✅ Full    │ ✅ Full      │
└─────────────┴────────────┴────────────┴────────────┴──────────────┘
```

---

## GET vs POST Comparison

```
┌────────────────┬─────────────────────────┬─────────────────────────┐
│ Feature        │ GET /query              │ POST /query             │
├────────────────┼─────────────────────────┼─────────────────────────┤
│ Query Types    │ SELECT, SHOW, DESCRIBE  │ All SQL operations      │
│ Parameters     │ URL query string        │ JSON body array         │
│ Data Limit     │ URL length limit (~2KB) │ No practical limit      │
│ Caching        │ Possible (browser)      │ Not cached              │
│ Use Case       │ Simple reads            │ Complex queries/writes  │
│ Security       │ Same (parameterized)    │ Same (parameterized)    │
│ Batch Support  │ No                      │ Yes (/query/batch)      │
└────────────────┴─────────────────────────┴─────────────────────────┘
```

---

## Common Use Cases

### 1️⃣ Dashboard Analytics
```
GET /api/mysql/query?q=SELECT COUNT(*) as total_users, AVG(age) as avg_age FROM users
        ↓
        Returns aggregated data for dashboard
```

### 2️⃣ Search Functionality
```
POST /api/mysql/query with:
{
  "query": "SELECT * FROM users WHERE name LIKE ? OR email LIKE ? ORDER BY id LIMIT ?",
  "params": ["%search_term%", "%search_term%", 20]
}
        ↓
        Returns search results efficiently
```

### 3️⃣ Data Migration
```
POST /api/mysql/query/batch with multiple INSERT queries
        ↓
        Migrate data in batch
```

### 4️⃣ Real-time Sync
```
POST /api/mysql/query with:
{
  "query": "UPDATE users SET last_sync = NOW() WHERE id = ?",
  "params": [user_id]
}
        ↓
        Track real-time user activity
```

### 5️⃣ Complex Reports
```
POST /api/mysql/query with JOIN/GROUP BY/HAVING
        ↓
        Generate detailed reports on-demand
```

---

## Performance Considerations

```
Query Execution Time:

Simple SELECT (no WHERE)  : ~5-10ms
WHERE clause (indexed)    : ~5-10ms
Complex JOIN              : ~20-50ms
GROUP BY with HAVING      : ~30-100ms
Batch of 10 queries       : ~100-200ms

Tips:
├─ Use indexes on filtered columns
├─ Limit result sets with LIMIT
├─ Use batch queries for related operations
└─ Monitor slow queries in logs
```

---

**Last Updated**: January 23, 2026  
**Visual Guide Version**: 1.0
