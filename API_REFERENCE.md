# API Reference - AxiomNizam

Complete reference guide for all 35+ API endpoints across 7 databases.

---

## Overview

AxiomNizam provides RESTful CRUD operations for:
- MySQL
- MariaDB  
- PostgreSQL
- Percona
- MongoDB
- Firebase
- Oracle

**Total Endpoints**: 35+ for database operations + 5 admin/health endpoints = 40+ total

---

## Base URL

```
http://localhost:8000
```

---

## Authentication

All endpoints except `/health` and `/status` require:

```
Authorization: Bearer YOUR_TOKEN
```

Get a token (see [AUTHENTICATION.md](AUTHENTICATION.md)):

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72&grant_type=client_credentials" \
  | jq -r '.access_token')
```

---

## Response Format

### Success (2xx)

```json
{
  "status": "success",
  "message": "Operation successful",
  "data": { ... }
}
```

### Error (4xx, 5xx)

```json
{
  "status": "error",
  "message": "Error description",
  "error": "error_code"
}
```

---

## Public Endpoints (No Auth)

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "message": "AxiomNizam API is running"
}
```

### System Status

```http
GET /status
```

**Response:**
```json
{
  "status": "ok",
  "message": "System status",
  "data": {
    "mysql": "connected",
    "postgres": "connected",
    "mongodb": "disconnected",
    "mariadb": "connected",
    "percona": "connected",
    "firebase": "connected",
    "oracle": "connected"
  }
}
```

---

## Database CRUD Endpoints

Each database has 5 endpoints:

### 1. Get All Users

```http
GET /api/{database}/users
Authorization: Bearer {token}
```

**Parameters:**
- `limit` (optional): Number of records to return
- `offset` (optional): Pagination offset
- `sort` (optional): Sort field
- `order` (optional): asc or desc

**Example:**
```bash
curl "http://localhost:8000/api/mysql/users?limit=10&offset=0&sort=name&order=asc" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "status": "success",
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30,
      "created_at": "2024-01-22T10:00:00Z"
    }
  ]
}
```

### 2. Get User by ID

```http
GET /api/{database}/users/{id}
Authorization: Bearer {token}
```

**Example:**
```bash
curl http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "status": "success",
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "created_at": "2024-01-22T10:00:00Z"
  }
}
```

### 3. Create User (Admin Only)

```http
POST /api/{database}/users
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "age": 28
}
```

**Example:**
```bash
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Smith",
    "email": "jane@example.com",
    "age": 28
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": 2,
    "name": "Jane Smith",
    "email": "jane@example.com",
    "age": 28,
    "created_at": "2024-01-22T10:30:00Z"
  }
}
```

### 4. Update User (Admin Only)

```http
PUT /api/{database}/users/{id}
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "name": "Jane Doe",
  "age": 29
}
```

**Example:**
```bash
curl -X PUT http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "age": 29
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "User updated successfully",
  "data": {
    "id": 1,
    "name": "Jane Doe",
    "email": "john@example.com",
    "age": 29,
    "updated_at": "2024-01-22T10:45:00Z"
  }
}
```

### 5. Delete User (Admin Only)

```http
DELETE /api/{database}/users/{id}
Authorization: Bearer {admin_token}
```

**Example:**
```bash
curl -X DELETE http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Response:**
```json
{
  "status": "success",
  "message": "User deleted successfully"
}
```

---

## Supported Databases

### MySQL
```
GET    /api/mysql/users
POST   /api/mysql/users         (Admin)
GET    /api/mysql/users/{id}
PUT    /api/mysql/users/{id}    (Admin)
DELETE /api/mysql/users/{id}    (Admin)
```

### MariaDB
```
GET    /api/mariadb/users
POST   /api/mariadb/users       (Admin)
GET    /api/mariadb/users/{id}
PUT    /api/mariadb/users/{id}  (Admin)
DELETE /api/mariadb/users/{id}  (Admin)
```

### PostgreSQL
```
GET    /api/postgres/users
POST   /api/postgres/users      (Admin)
GET    /api/postgres/users/{id}
PUT    /api/postgres/users/{id} (Admin)
DELETE /api/postgres/users/{id} (Admin)
```

### Percona
```
GET    /api/percona/users
POST   /api/percona/users       (Admin)
GET    /api/percona/users/{id}
PUT    /api/percona/users/{id}  (Admin)
DELETE /api/percona/users/{id}  (Admin)
```

### MongoDB
```
GET    /api/mongodb/users
POST   /api/mongodb/users       (Admin)
GET    /api/mongodb/users/{id}
PUT    /api/mongodb/users/{id}  (Admin)
DELETE /api/mongodb/users/{id}  (Admin)
```

### Firebase
```
GET    /api/firebase/users
POST   /api/firebase/users      (Admin)
GET    /api/firebase/users/{id}
PUT    /api/firebase/users/{id} (Admin)
DELETE /api/firebase/users/{id} (Admin)
```

### Oracle
```
GET    /api/oracle/users
POST   /api/oracle/users        (Admin)
GET    /api/oracle/users/{id}
PUT    /api/oracle/users/{id}   (Admin)
DELETE /api/oracle/users/{id}   (Admin)
```

---

## Admin Endpoints (Admin Only)

### Create Database

```http
POST /api/admin/database/create
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "database_name": "new_database",
  "database_type": "mysql"
}
```

**Example:**
```bash
curl -X POST http://localhost:8000/api/admin/database/create \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "database_name": "customers",
    "database_type": "mysql"
  }'
```

### List Databases

```http
GET /api/admin/database/list
Authorization: Bearer {admin_token}
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "name": "axiomnizam",
      "type": "mysql",
      "size": "10MB",
      "tables": 5
    }
  ]
}
```

### Create Table

```http
POST /api/admin/table/create
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "database_name": "axiomnizam",
  "table_name": "products",
  "columns": [
    {"name": "id", "type": "INT", "primary_key": true},
    {"name": "title", "type": "VARCHAR(255)"},
    {"name": "price", "type": "DECIMAL(10,2)"}
  ]
}
```

### List Tables

```http
GET /api/admin/table/list
Authorization: Bearer {admin_token}

{
  "database_name": "axiomnizam"
}
```

---

## Notification Endpoints

### Send Custom Notification

```http
POST /api/notifications/send
Authorization: Bearer {token}
Content-Type: application/json

{
  "message": "Database backup completed",
  "level": "info"
}
```

### Send Health Notification

```http
POST /api/notifications/health
Authorization: Bearer {token}
```

### Send Status Notification

```http
POST /api/notifications/status
Authorization: Bearer {token}
```

### Get Status

```http
GET /api/notifications/status
```

---

## Error Codes

| Code | Meaning | Solution |
|------|---------|----------|
| 200 | OK | Request successful |
| 201 | Created | Resource created |
| 204 | No Content | Deleted successfully |
| 400 | Bad Request | Check request format |
| 401 | Unauthorized | Get a valid token |
| 403 | Forbidden | Use admin token |
| 404 | Not Found | Check ID/endpoint |
| 500 | Server Error | Check server logs |

---

## Common Examples

### PowerShell

```powershell
# Get token
$token = (Invoke-RestMethod -Uri "http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token" `
  -Method Post `
  -Body @{client_id="axiomnizam-backend"; client_secret="6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72"; grant_type="client_credentials"} `
  -ContentType "application/x-www-form-urlencoded").access_token

# Get users
$headers = @{Authorization="Bearer $token"}
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers

# Create user
$body = @{name="Alice"; email="alice@example.com"; age=25} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" `
  -Method Post `
  -Headers $headers `
  -Body $body `
  -ContentType "application/json"
```

### cURL

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72&grant_type=client_credentials" \
  | jq -r '.access_token')

# Get users
curl http://localhost:8000/api/mysql/users -H "Authorization: Bearer $TOKEN"

# Create user
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","age":25}'
```

### JavaScript/Fetch

```javascript
// Get token
const response = await fetch('http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token', {
  method: 'POST',
  headers: {'Content-Type': 'application/x-www-form-urlencoded'},
  body: 'client_id=axiomnizam-backend&client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72&grant_type=client_credentials'
});
const {access_token} = await response.json();

// Get users
const users = await fetch('http://localhost:8000/api/mysql/users', {
  headers: {'Authorization': `Bearer ${access_token}`}
});
const data = await users.json();
```

---

## Data Models

### User Model

```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "age": 30,
  "created_at": "2024-01-22T10:00:00Z",
  "updated_at": "2024-01-22T10:00:00Z"
}
```

### Request Body Example

```json
{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "age": 28
}
```

---

## Testing with Postman

Import the Postman collection:
```
POSTMAN_COLLECTION.json
```

The collection includes:
- ✅ Pre-configured requests for all endpoints
- ✅ Auto-token scripts
- ✅ Environment variables
- ✅ Response examples
- ✅ Test scripts

---

## Rate Limiting

Currently: **No rate limiting** (can be configured in production)

**Recommended for production:**
- 100 requests/minute per token
- 1000 requests/hour per IP

---

## Best Practices

1. **Always use HTTPS** in production
2. **Store tokens securely** - don't hardcode them
3. **Validate data** before sending
4. **Handle errors** gracefully
5. **Use pagination** for large datasets
6. **Cache responses** when possible
7. **Log API calls** for debugging
8. **Set timeouts** on requests

---

## Troubleshooting

### 401 Unauthorized
- Get a new token
- Check token hasn't expired
- Verify Authorization header format

### 403 Forbidden
- Use admin token for POST/PUT/DELETE
- Check user role in Keycloak

### 404 Not Found
- Verify endpoint URL is correct
- Check resource ID exists

### 500 Server Error
- Check backend logs: `docker-compose logs backend`
- Verify database connections: `curl http://localhost:8000/status`

---

## Learn More

- **[AUTHENTICATION.md](AUTHENTICATION.md)** - Token & RBAC details
- **[SETUP_GUIDE.md](SETUP_GUIDE.md)** - Configuration
- **[QUICK_START.md](QUICK_START.md)** - Get started quickly
- **[POSTMAN_API_GUIDE.md](POSTMAN_API_GUIDE.md)** - Postman setup

---

**Ready to test? Import [POSTMAN_COLLECTION.json](POSTMAN_COLLECTION.json) to get started!**
