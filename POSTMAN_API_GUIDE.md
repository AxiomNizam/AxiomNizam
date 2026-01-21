# AxiomNizam API Documentation for Postman

Complete API reference for testing all endpoints in Postman.

## Base URLs

- **API Backend**: `http://localhost:8000`
- **Frontend Dashboard**: `http://localhost:7000`
- **Keycloak Admin**: `http://localhost:8080/admin`

## Authentication

### Get Access Token from Keycloak

**Endpoint**: `POST` `http://localhost:8080/realms/master/protocol/openid-connect/token`

**Headers**:
```
Content-Type: application/x-www-form-urlencoded
```

**Body** (form-urlencoded):
```
client_id: axiomnizam
grant_type: password
username: admin
password: admin
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cC...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "token_type": "Bearer"
}
```

**Save the `access_token` and use in Authorization header for all protected endpoints**:
```
Authorization: Bearer <access_token>
```

---

## Public Endpoints (No Auth Required)

### 1. Health Check
**GET** `http://localhost:8000/health`

**Headers**: None

**Response**:
```json
{
    "status": "ok",
    "message": "AxiomNizam API is running"
}
```

---

### 2. System Status (All Databases)
**GET** `http://localhost:8000/status`

**Headers**: None

**Response**:
```json
{
    "status": "ok",
    "message": "System status",
    "data": {
        "elasticsearch": "connected",
        "etcd": "connected",
        "firebase": "connected",
        "mariadb": "disconnected",
        "mongodb": "connected",
        "mysql": "disconnected",
        "oracle": "connected",
        "percona": "connected",
        "postgres": "connected",
        "valkey": "connected"
    }
}
```

---

## Protected Endpoints (Require Bearer Token)

All CRUD operations require the `Authorization: Bearer <token>` header.

---

## MySQL CRUD Operations

### Create User
**POST** `http://localhost:8000/api/mysql/users`

**Headers**:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Body**:
```json
{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
}
```

**Response** (201 Created):
```json
{
    "status": "ok",
    "message": "User created successfully",
    "data": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30
    }
}
```

---

### Get All Users
**GET** `http://localhost:8000/api/mysql/users`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
    "status": "ok",
    "message": "Users retrieved successfully",
    "data": [
        {
            "id": 1,
            "name": "John Doe",
            "email": "john@example.com",
            "age": 30
        }
    ]
}
```

---

### Get User by ID
**GET** `http://localhost:8000/api/mysql/users/1`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
    "status": "ok",
    "message": "User retrieved successfully",
    "data": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30
    }
}
```

---

### Update User
**PUT** `http://localhost:8000/api/mysql/users/1`

**Headers**:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Body**:
```json
{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "age": 28
}
```

**Response** (200 OK):
```json
{
    "status": "ok",
    "message": "User updated successfully",
    "data": {
        "id": 1,
        "name": "Jane Doe",
        "email": "jane@example.com",
        "age": 28
    }
}
```

---

### Delete User
**DELETE** `http://localhost:8000/api/mysql/users/1`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
    "status": "ok",
    "message": "User deleted successfully"
}
```

---

## MariaDB CRUD Operations

Same as MySQL but use `/api/mariadb/users` endpoints:

- **POST** `http://localhost:8000/api/mariadb/users` - Create
- **GET** `http://localhost:8000/api/mariadb/users` - Get all
- **GET** `http://localhost:8000/api/mariadb/users/{id}` - Get by ID
- **PUT** `http://localhost:8000/api/mariadb/users/{id}` - Update
- **DELETE** `http://localhost:8000/api/mariadb/users/{id}` - Delete

---

## PostgreSQL CRUD Operations

Same structure as MySQL but use `/api/postgres/users` endpoints:

- **POST** `http://localhost:8000/api/postgres/users` - Create
- **GET** `http://localhost:8000/api/postgres/users` - Get all
- **GET** `http://localhost:8000/api/postgres/users/{id}` - Get by ID
- **PUT** `http://localhost:8000/api/postgres/users/{id}` - Update
- **DELETE** `http://localhost:8000/api/postgres/users/{id}` - Delete

---

## Percona CRUD Operations

Same structure but use `/api/percona/users` endpoints:

- **POST** `http://localhost:8000/api/percona/users` - Create
- **GET** `http://localhost:8000/api/percona/users` - Get all
- **GET** `http://localhost:8000/api/percona/users/{id}` - Get by ID
- **PUT** `http://localhost:8000/api/percona/users/{id}` - Update
- **DELETE** `http://localhost:8000/api/percona/users/{id}` - Delete

---

## MongoDB CRUD Operations

Same structure but use `/api/mongodb/users` endpoints:

- **POST** `http://localhost:8000/api/mongodb/users` - Create
- **GET** `http://localhost:8000/api/mongodb/users` - Get all
- **GET** `http://localhost:8000/api/mongodb/users/{id}` - Get by ID
- **PUT** `http://localhost:8000/api/mongodb/users/{id}` - Update
- **DELETE** `http://localhost:8000/api/mongodb/users/{id}` - Delete

---

## Firebase CRUD Operations

Same structure but use `/api/firebase/users` endpoints:

- **POST** `http://localhost:8000/api/firebase/users` - Create
- **GET** `http://localhost:8000/api/firebase/users` - Get all
- **GET** `http://localhost:8000/api/firebase/users/{id}` - Get by ID
- **PUT** `http://localhost:8000/api/firebase/users/{id}` - Update
- **DELETE** `http://localhost:8000/api/firebase/users/{id}` - Delete

---

## Oracle CRUD Operations

Same structure but use `/api/oracle/users` endpoints:

- **POST** `http://localhost:8000/api/oracle/users` - Create
- **GET** `http://localhost:8000/api/oracle/users` - Get all
- **GET** `http://localhost:8000/api/oracle/users/{id}` - Get by ID
- **PUT** `http://localhost:8000/api/oracle/users/{id}` - Update
- **DELETE** `http://localhost:8000/api/oracle/users/{id}` - Delete

---

## Error Responses

### Missing Authorization Header
**Status**: 401 Unauthorized

```json
{
    "error": "missing authorization header"
}
```

---

### Invalid Token
**Status**: 401 Unauthorized

```json
{
    "error": "invalid token: token has expired"
}
```

---

### Database Not Connected
**Status**: 503 Service Unavailable

```json
{
    "status": "error",
    "error": "Database not connected"
}
```

---

### User Not Found
**Status**: 404 Not Found

```json
{
    "status": "error",
    "error": "User not found"
}
```

---

### Database Error
**Status**: 500 Internal Server Error

```json
{
    "status": "error",
    "error": "Error message from database"
}
```

---

## Postman Collection Setup

### Step 1: Create Environment Variables

In Postman, create a new Environment with these variables:

```
Variable Name          | Initial Value              | Current Value
----------------------|----------------------------|---------------------------
base_url              | http://localhost:8000      | http://localhost:8000
keycloak_url          | http://localhost:8080      | http://localhost:8080
access_token          | (leave empty)              | (will be filled after auth)
client_id             | axiomnizam                 | axiomnizam
username              | admin                      | admin
password              | admin                      | admin
```

### Step 2: Create Get Token Request

**Name**: Get Access Token
**Method**: POST
**URL**: `{{keycloak_url}}/realms/master/protocol/openid-connect/token`

**Headers**:
```
Content-Type: application/x-www-form-urlencoded
```

**Body** (form-urlencoded):
```
client_id     = {{client_id}}
grant_type    = password
username      = {{username}}
password      = {{password}}
```

**Tests Tab** (Auto-save token):
```javascript
if (pm.response.code === 200) {
    var jsonData = pm.response.json();
    pm.environment.set("access_token", jsonData.access_token);
    console.log("Token saved: " + jsonData.access_token.substring(0, 20) + "...");
}
```

### Step 3: Create CRUD Request Examples

**Create User**
```
POST {{base_url}}/api/mysql/users
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
    "name": "Test User",
    "email": "test@example.com",
    "age": 25
}
```

**Get All Users**
```
GET {{base_url}}/api/mysql/users
Authorization: Bearer {{access_token}}
```

**Get User by ID**
```
GET {{base_url}}/api/mysql/users/1
Authorization: Bearer {{access_token}}
```

**Update User**
```
PUT {{base_url}}/api/mysql/users/1
Authorization: Bearer {{access_token}}
Content-Type: application/json

{
    "name": "Updated User",
    "email": "updated@example.com",
    "age": 26
}
```

**Delete User**
```
DELETE {{base_url}}/api/mysql/users/1
Authorization: Bearer {{access_token}}
```

---

## Testing Workflow in Postman

1. **Get Token First**
   - Run the "Get Access Token" request
   - The access token will be automatically saved to the environment

2. **Test Health Endpoint**
   - GET `{{base_url}}/health`
   - No authorization needed

3. **Test Status Endpoint**
   - GET `{{base_url}}/status`
   - See all database connection statuses

4. **Test CRUD Operations**
   - Create a user
   - Get all users
   - Get specific user
   - Update user
   - Delete user

5. **Test All Databases**
   - Repeat CRUD tests for each database by changing the path:
     - `/api/mysql/users`
     - `/api/mariadb/users`
     - `/api/postgres/users`
     - `/api/percona/users`
     - `/api/mongodb/users`
     - `/api/firebase/users`
     - `/api/oracle/users`

---

## Response Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Success GET, PUT, DELETE |
| 201 | Created | User successfully created |
| 400 | Bad Request | Invalid JSON body |
| 401 | Unauthorized | Missing or invalid token |
| 404 | Not Found | User ID doesn't exist |
| 500 | Server Error | Database error |
| 503 | Unavailable | Database not connected |

---

## Quick Test Commands (cURL Alternative)

### Get Token
```bash
TOKEN=$(curl -s -X POST \
  http://localhost:8080/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam&grant_type=password&username=admin&password=admin" \
  | jq -r '.access_token')

echo $TOKEN
```

### Create User
```bash
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }'
```

### Get All Users
```bash
curl -X GET http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN"
```

### Update User
```bash
curl -X PUT http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "age": 28
  }'
```

### Delete User
```bash
curl -X DELETE http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Troubleshooting

### Token Expires
- Get a new token by running the "Get Access Token" request again
- The environment variable will auto-update

### 401 Unauthorized
- Check if token is valid
- Get a new token
- Verify Authorization header format: `Bearer <token>`

### 503 Service Unavailable
- Check if database is running: `docker ps`
- Verify all services: `http://localhost:8000/status`

### Connection Refused
- Verify backend is running on port 8000
- Check firewall settings
- Restart containers: `docker-compose restart axiomnizam`

---

## Services Port Reference

| Service | Port | URL |
|---------|------|-----|
| API Backend | 8000 | http://localhost:8000 |
| Frontend Dashboard | 7000 | http://localhost:7000 |
| Keycloak | 8080 | http://localhost:8080 |
| MySQL | 3306 | localhost:3306 |
| MariaDB | 3307 | localhost:3307 |
| PostgreSQL | 5432 | localhost:5432 |
| Percona | 3308 | localhost:3308 |
| MongoDB | 27017 | localhost:27017 |
| Valkey | 6379 | localhost:6379 |
| Elasticsearch | 9200 | localhost:9200 |
| etcd | 2379 | localhost:2379 |
| Firebase | 4000, 5000, 8085, 9099 | localhost:4000 (UI) |

---

## User Model

All CRUD operations use this data structure:

```json
{
    "id": 1,
    "name": "string (required)",
    "email": "string (required, unique)",
    "age": "integer (optional)"
}
```

---

## Notes

- All endpoints return JSON responses
- Timestamps are in ISO 8601 format
- Database connections are persistent within Docker containers
- Each database has independent user tables
- Authentication is enforced at the application level
- CORS is enabled for frontend dashboard access
