# Dynamic Query System - Quick Start Guide

## What's New?

Your AxiomNizam backend now supports **dynamic SQL queries** through simple HTTP requests! No need to create new endpoints for each query - just send your SQL directly to Postman.

---

## Quick Examples

### Example 1: GET all users from MySQL

**Postman Request:**
```
Method: GET
URL: http://localhost:8000/api/mysql/query?q=SELECT * FROM users
Headers: Authorization: Bearer YOUR_TOKEN
```

**Result:**
```json
{
  "status": "ok",
  "message": "Query executed successfully",
  "data": [
    {"id": 1, "name": "John", "email": "john@example.com", "age": 28},
    {"id": 2, "name": "Jane", "email": "jane@example.com", "age": 26}
  ]
}
```

---

### Example 2: Filter users by age

**Postman Request:**
```
Method: GET
URL: http://localhost:8000/api/mysql/query?q=SELECT * FROM users WHERE age > ?&params=25
Headers: Authorization: Bearer YOUR_TOKEN
```

---

### Example 3: Complex query with POST (recommended for complex queries)

**Postman Request:**
```
Method: POST
URL: http://localhost:8000/api/mysql/query
Headers: 
  - Authorization: Bearer YOUR_TOKEN
  - Content-Type: application/json

Body:
{
  "query": "SELECT * FROM users WHERE age > ? AND name LIKE ?",
  "params": [25, "%John%"]
}
```

---

### Example 4: INSERT data

**Postman Request:**
```
Method: POST
URL: http://localhost:8000/api/mysql/query
Headers: 
  - Authorization: Bearer YOUR_TOKEN
  - Content-Type: application/json

Body:
{
  "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
  "params": ["Alice Smith", "alice@example.com", 30]
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Query executed successfully",
  "data": {
    "rows_affected": 1
  }
}
```

---

### Example 5: UPDATE data

**Postman Request:**
```
Method: POST
URL: http://localhost:8000/api/mysql/query
Headers: 
  - Authorization: Bearer YOUR_TOKEN
  - Content-Type: application/json

Body:
{
  "query": "UPDATE users SET age = ? WHERE id = ?",
  "params": [31, 1]
}
```

---

### Example 6: DELETE data

**Postman Request:**
```
Method: POST
URL: http://localhost:8000/api/mysql/query
Headers: 
  - Authorization: Bearer YOUR_TOKEN
  - Content-Type: application/json

Body:
{
  "query": "DELETE FROM users WHERE id = ?",
  "params": [1]
}
```

---

### Example 7: Get table schema

**Postman Request:**
```
Method: GET
URL: http://localhost:8000/api/mysql/schema?table=users
Headers: Authorization: Bearer YOUR_TOKEN
```

**Result:**
```json
{
  "status": "ok",
  "message": "Table schema retrieved successfully",
  "data": [
    {
      "Field": "id",
      "Type": "bigint(20) unsigned",
      "Null": "NO",
      "Key": "PRI",
      "Default": null,
      "Extra": "auto_increment"
    },
    {
      "Field": "name",
      "Type": "varchar(255)",
      "Null": "YES",
      "Key": "",
      "Default": null,
      "Extra": ""
    },
    {
      "Field": "email",
      "Type": "varchar(255)",
      "Null": "YES",
      "Key": "UNI",
      "Default": null,
      "Extra": ""
    },
    {
      "Field": "age",
      "Type": "int",
      "Null": "YES",
      "Key": "",
      "Default": null,
      "Extra": ""
    }
  ]
}
```

---

### Example 8: Execute multiple queries at once (Batch)

**Postman Request:**
```
Method: POST
URL: http://localhost:8000/api/mysql/query/batch
Headers: 
  - Authorization: Bearer YOUR_TOKEN
  - Content-Type: application/json

Body:
[
  {
    "query": "SELECT COUNT(*) as total FROM users",
    "params": []
  },
  {
    "query": "SELECT * FROM users WHERE age > ? ORDER BY age DESC LIMIT ?",
    "params": [25, 5]
  },
  {
    "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
    "params": ["Bob Johnson", "bob@example.com", 27]
  }
]
```

---

## API Endpoints Reference

### GET Request (SELECT only)
```
GET /api/{db}/query?q=YOUR_QUERY&params=value1,value2
```
- **db**: mysql, mariadb, postgres, percona, oracle
- **q**: URL-encoded SQL SELECT query
- **params**: Comma-separated parameter values (optional)

### POST Request (All query types)
```
POST /api/{db}/query
Content-Type: application/json

{
  "query": "SQL_QUERY",
  "params": ["value1", "value2"]
}
```

### Batch Queries
```
POST /api/{db}/query/batch
Content-Type: application/json

[
  {"query": "SQL_QUERY_1", "params": []},
  {"query": "SQL_QUERY_2", "params": ["value1"]}
]
```

### Table Schema
```
GET /api/{db}/schema?table=table_name
```

---

## Supported Databases

- **MySQL**: `/api/mysql/query`
- **MariaDB**: `/api/mariadb/query`
- **PostgreSQL**: `/api/postgres/query`
- **Percona**: `/api/percona/query`
- **Oracle**: `/api/oracle/query`

---

## Key Features

✅ **No pre-defined endpoints** - Send any valid SQL query
✅ **GET requests** - Read-only (SELECT queries only)
✅ **POST requests** - Full SQL support (INSERT, UPDATE, DELETE, CREATE, DROP, ALTER)
✅ **Parameterized queries** - Prevents SQL injection
✅ **Batch operations** - Execute multiple queries at once
✅ **Schema inspection** - View table structure before writing queries
✅ **Full authentication** - All endpoints require valid JWT token
✅ **Error handling** - Clear error messages with debugging info

---

## Common Query Patterns

### Count records:
```json
{
  "query": "SELECT COUNT(*) as total FROM users",
  "params": []
}
```

### Join tables:
```json
{
  "query": "SELECT u.name, o.order_id FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > ?",
  "params": [25]
}
```

### Group and aggregate:
```json
{
  "query": "SELECT age, COUNT(*) as count FROM users GROUP BY age HAVING COUNT(*) > ?",
  "params": [5]
}
```

### Pagination:
```json
{
  "query": "SELECT * FROM users ORDER BY id LIMIT ? OFFSET ?",
  "params": [10, 0]
}
```

### Search with wildcards:
```json
{
  "query": "SELECT * FROM users WHERE name LIKE ? OR email LIKE ?",
  "params": ["%john%", "%@example.com"]
}
```

### Transaction-like batch:
```json
[
  {"query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "params": ["New User", "new@example.com", 25]},
  {"query": "INSERT INTO audit_log (action, user_email) VALUES (?, ?)", "params": ["user_created", "new@example.com"]}
]
```

---

## Error Handling

### Error: "Only SELECT queries are allowed for GET requests"
**Cause**: Tried INSERT/UPDATE/DELETE on GET endpoint
**Solution**: Use POST endpoint instead

### Error: "Query execution failed: column 'xyz' does not exist"
**Cause**: Referenced wrong column name
**Solution**: Check table schema using `/schema?table=your_table`

### Error: "Missing query parameter"
**Cause**: Didn't include `q` parameter in GET or `query` in POST body
**Solution**: Add the `q=YOUR_QUERY` or send proper JSON body

### Error: "Database not connected"
**Cause**: Database service is down
**Solution**: Check if the database container is running

---

## Security Notes

🔒 **Parameterized Queries**: Always use the `params` array to prevent SQL injection
🔒 **Authentication**: All endpoints require Bearer token authentication
🔒 **Read vs Write**: GET endpoints only allow SELECT queries
🔒 **Dangerous Operations**: Some operations like `DROP DATABASE` are blocked

---

## Postman Collection Tips

1. **Set Authorization Token Variable**:
   - Go to Environment
   - Create/Edit variable: `token` = your_jwt_token
   - Use `{{token}}` in Authorization header

2. **Save Frequently Used Queries**:
   - Create Postman "Requests" folder
   - Save your common queries for quick access

3. **Use Pre-request Scripts**:
   - Execute test data setup before running queries
   - Useful for testing INSERT/UPDATE/DELETE

4. **Use Tests**:
   - Validate response format
   - Extract data for use in other requests
   - Example: `pm.environment.set("user_id", pm.response.json().data[0].id)`

---

## Migration Path

### Old Way (Pre-defined endpoints):
```
GET /api/mysql/users/:id  → Only gets one user
GET /api/mysql/users      → Gets all users (no filter)
```

### New Way (Dynamic queries):
```
GET /api/mysql/query?q=SELECT * FROM users WHERE id = ?&params=1
GET /api/mysql/query?q=SELECT * FROM users WHERE age > ? AND status = ?&params=25,active
GET /api/mysql/query?q=SELECT * FROM users ORDER BY created_at DESC LIMIT 10
```

**The old endpoints still work!** Both old and new ways are supported side-by-side.

---

## Performance Tips

1. **Use LIMIT for large tables**: `SELECT * FROM users LIMIT 100`
2. **Add indexes**: Create indexes on frequently filtered columns
3. **Use batch queries**: Reduce round-trips with `/query/batch`
4. **Check execution plans**: Use `EXPLAIN` for slow queries
5. **Monitor response times**: Log and optimize slow queries

---

## Next Steps

1. Copy your JWT token from the login endpoint
2. Open Postman
3. Try the quick examples above
4. Build your own queries based on your needs
5. Use `/schema?table=your_table` to understand table structure
6. Combine queries for complex workflows

---

**Enjoy the flexibility of dynamic queries!** 🚀
