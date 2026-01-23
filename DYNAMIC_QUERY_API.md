# Dynamic Query API Documentation

## Overview
The AxiomNizam backend now supports **dynamic SQL queries** via Postman and HTTP clients. You can send SQL queries on-the-fly without pre-defined endpoints!

## Endpoints

### 1. **GET Dynamic Query** (SELECT only)
Execute SELECT queries with URL parameters.

```
GET /api/{db}/query?q=YOUR_QUERY&params=value1,value2
```

**Parameters:**
- `q` (required): SQL SELECT query
- `params` (optional): Comma-separated parameter values

**Example Requests:**

#### Get all users from MySQL:
```
GET http://localhost:8000/api/mysql/query?q=SELECT * FROM users
```

#### Get user by ID:
```
GET http://localhost:8000/api/mysql/query?q=SELECT * FROM users WHERE id = ?&params=1
```

#### Get users with age greater than 25:
```
GET http://localhost:8000/api/mysql/query?q=SELECT * FROM users WHERE age > ?&params=25
```

#### Complex query with multiple conditions:
```
GET http://localhost:8000/api/mysql/query?q=SELECT * FROM users WHERE age > ? AND name LIKE ?&params=25,%John%
```

---

### 2. **POST Dynamic Query** (All query types)
Execute any SQL query type: SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, etc.

```
POST /api/{db}/query
Content-Type: application/json

{
  "query": "SELECT * FROM users WHERE age > ? AND name LIKE ?",
  "params": [25, "%John%"]
}
```

**Example Requests:**

#### SELECT Query:
```json
{
  "query": "SELECT id, name, email FROM users WHERE age > ?",
  "params": [30]
}
```

#### INSERT Query:
```json
{
  "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
  "params": ["John Doe", "john@example.com", 28]
}
```

#### UPDATE Query:
```json
{
  "query": "UPDATE users SET age = ? WHERE id = ?",
  "params": [29, 1]
}
```

#### DELETE Query:
```json
{
  "query": "DELETE FROM users WHERE id = ?",
  "params": [1]
}
```

#### CREATE Table:
```json
{
  "query": "CREATE TABLE products (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(255), price DECIMAL(10,2))",
  "params": []
}
```

---

### 3. **POST Batch Queries** (Multiple queries)
Execute multiple queries in one request.

```
POST /api/{db}/query/batch
Content-Type: application/json

[
  {"query": "SELECT * FROM users", "params": []},
  {"query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "params": ["Jane", "jane@example.com", 26]}
]
```

**Full Example:**
```json
[
  {
    "query": "SELECT COUNT(*) as total FROM users",
    "params": []
  },
  {
    "query": "SELECT * FROM users WHERE age > ?",
    "params": [25]
  },
  {
    "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
    "params": ["Alice", "alice@example.com", 27]
  }
]
```

---

### 4. **GET Table Schema**
Get the schema/structure of a table.

```
GET /api/{db}/schema?table=table_name
```

**Example Requests:**

#### Get users table schema from MySQL:
```
GET http://localhost:8000/api/mysql/schema?table=users
```

**Response Example:**
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
    }
  ]
}
```

---

## Available Databases

- **mysql**: `GET/POST /api/mysql/query`
- **mariadb**: `GET/POST /api/mariadb/query`
- **postgres**: `GET/POST /api/postgres/query`
- **percona**: `GET/POST /api/percona/query`
- **oracle**: `GET/POST /api/oracle/query`

---

## Security Features

✅ **GET Requests**: Only SELECT queries allowed (Read-only)  
✅ **POST Requests**: All SQL queries allowed (CREATE, INSERT, UPDATE, DELETE, DROP, ALTER)  
✅ **Authentication**: All endpoints require valid JWT token in Bearer header  
✅ **Query Validation**: Prevents dangerous operations like `DROP DATABASE`  
✅ **Parameterized Queries**: Uses prepared statements to prevent SQL injection  

---

## Response Format

### Success Response:
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

### Error Response:
```json
{
  "status": "error",
  "error": "Query execution failed: column \"age_xyz\" does not exist"
}
```

---

## Postman Setup

### 1. Set Authorization Token
In Postman, set the `token` variable in your environment:
```
Bearer {{token}}
```

### 2. Example GET Request
```
GET http://localhost:8000/api/mysql/query?q=SELECT * FROM users WHERE age > ?&params=25
Authorization: Bearer YOUR_TOKEN
```

### 3. Example POST Request
```
POST http://localhost:8000/api/mysql/query
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "query": "SELECT * FROM users WHERE name LIKE ?",
  "params": ["%John%"]
}
```

---

## Use Cases

### List all records:
```
GET /api/mysql/query?q=SELECT * FROM users
```

### Filter by conditions:
```
GET /api/mysql/query?q=SELECT * FROM users WHERE age > ? AND status = ?&params=25,active
```

### Aggregate data:
```
POST /api/mysql/query
Body: {"query": "SELECT age, COUNT(*) as count FROM users GROUP BY age", "params": []}
```

### Complex joins:
```
POST /api/mysql/query
Body: {"query": "SELECT u.name, o.order_id FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > ?", "params": [25]}
```

### Batch operations:
```
POST /api/mysql/query/batch
Body: [
  {"query": "SELECT * FROM users", "params": []},
  {"query": "INSERT INTO audit_log (action) VALUES (?)", "params": ["fetched_users"]}
]
```

---

## Error Handling

| Status Code | Meaning |
|------------|---------|
| 200 | Query executed successfully |
| 400 | Bad request (missing query, invalid format) |
| 403 | Forbidden query type (e.g., DELETE on GET request) |
| 500 | Query execution error (SQL syntax error, database error) |
| 503 | Database not connected |

---

## Tips & Best Practices

1. **Always use parameters** for safety: `SELECT * FROM users WHERE id = ?` with `params: [1]`
2. **URL encode special characters** in GET parameters
3. **Use POST for write operations** (INSERT, UPDATE, DELETE, CREATE)
4. **Use batch queries** to execute multiple queries in transaction-like behavior
5. **Check table schema first** using the `/schema` endpoint before writing queries
6. **Monitor response time** for complex queries
7. **Use SELECT statements** for read operations to maintain consistency

---

## Example Workflows

### Workflow 1: Get all users and their count
```
Step 1: GET /api/mysql/query?q=SELECT COUNT(*) as total FROM users
Step 2: GET /api/mysql/query?q=SELECT * FROM users LIMIT 10
```

### Workflow 2: Insert and then verify
```
Step 1: POST /api/mysql/query with INSERT query
Step 2: GET /api/mysql/query?q=SELECT * FROM users WHERE email = ? with the inserted email
```

### Workflow 3: Get schema, then insert compatible data
```
Step 1: GET /api/mysql/schema?table=users (understand structure)
Step 2: POST /api/mysql/query with INSERT matching the schema
```

---

**Last Updated**: January 23, 2026  
**API Version**: 2.0
