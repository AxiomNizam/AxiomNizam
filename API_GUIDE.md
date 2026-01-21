# AxiomNizam
AxiomNizam is an open-source, cloud-native platform designed to standardize how APIs, data flows, and system integrations are defined, managed, and executed.

# AxiomNizam API - CRUD Operations Guide

## System is Running! 🚀

All services are up and running on their respective ports:
- API: http://localhost:8000
- MySQL: localhost:3306
- MariaDB: localhost:3307  
- Percona: localhost:3308
- PostgreSQL: localhost:5432
- MongoDB: localhost:27017
- Valkey/Redis: localhost:6379
- Elasticsearch: localhost:9200
- etcd: localhost:2379

---

## Quick Test Endpoints

### Health & Status Check

#### Health Endpoint
```bash
curl http://localhost:8000/health
```

**Response:**
```json
{
  "status": "ok",
  "message": "AxiomNizam API is running"
}
```

#### Status Endpoint - Check All Database Connections
```bash
curl http://localhost:8000/status
```

**Response:**
```json
{
  "status": "ok",
  "message": "System status",
  "data": {
    "elasticsearch": "connected",
    "etcd": "connected",
    "mariadb": "connected",
    "mongodb": "connected",
    "mysql": "connected",
    "postgres": "connected",
    "valkey": "connected"
  }
}
```

## All Available API Endpoints

### System Health & Status
- `GET /health` - Check API is running
- `GET /status` - Check all database connections

### MySQL CRUD Operations
- `POST /api/mysql/users` - Create new user
- `GET /api/mysql/users` - Get all users
- `GET /api/mysql/users/:id` - Get user by ID
- `PUT /api/mysql/users/:id` - Update user
- `DELETE /api/mysql/users/:id` - Delete user

### MariaDB CRUD Operations
- `POST /api/mariadb/users` - Create new user
- `GET /api/mariadb/users` - Get all users
- `GET /api/mariadb/users/:id` - Get user by ID
- `PUT /api/mariadb/users/:id` - Update user
- `DELETE /api/mariadb/users/:id` - Delete user

### PostgreSQL CRUD Operations
- `POST /api/postgres/users` - Create new user
- `GET /api/postgres/users` - Get all users
- `GET /api/postgres/users/:id` - Get user by ID
- `PUT /api/postgres/users/:id` - Update user
- `DELETE /api/postgres/users/:id` - Delete user

### Firebase CRUD Operations
- `POST /api/firebase/users` - Create new user
- `GET /api/firebase/users` - Get all users
- `GET /api/firebase/users/:id` - Get user by ID
- `PUT /api/firebase/users/:id` - Update user
- `DELETE /api/firebase/users/:id` - Delete user

---

## CRUD Operations

### MySQL Operations

#### Create User (POST)
```bash
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","age":30}'
```

#### Get All Users (GET)
```bash
curl http://localhost:8000/api/mysql/users
```

#### Get User by ID (GET)
```bash
curl http://localhost:8000/api/mysql/users/1
```

#### Update User (PUT)
```bash
curl -X PUT http://localhost:8000/api/mysql/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","age":28}'
```

#### Delete User (DELETE)
```bash
curl -X DELETE http://localhost:8000/api/mysql/users/1
```

---

### MariaDB Operations

Same endpoints as MySQL, replace `/mysql/` with `/mariadb/`:

```bash
# Create
curl -X POST http://localhost:8000/api/mariadb/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","age":25}'

# Get All
curl http://localhost:8000/api/mariadb/users

# Get By ID
curl http://localhost:8000/api/mariadb/users/1

# Update
curl -X PUT http://localhost:8000/api/mariadb/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.new@example.com","age":26}'

# Delete
curl -X DELETE http://localhost:8000/api/mariadb/users/1
```

---

### PostgreSQL Operations

Same endpoints as MySQL, replace `/mysql/` with `/postgres/`:

```bash
# Create
curl -X POST http://localhost:8000/api/postgres/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob Smith","email":"bob@example.com","age":35}'

# Get All
curl http://localhost:8000/api/postgres/users

# Get By ID
curl http://localhost:8000/api/postgres/users/1

# Update
curl -X PUT http://localhost:8000/api/postgres/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob Updated","email":"bob.new@example.com","age":36}'

# Delete
curl -X DELETE http://localhost:8000/api/postgres/users/1
```

---

## Testing with Powershell

### Health Check
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/health" -Method GET | Select-Object -ExpandProperty Content
```

### Status Check - All Database Connections
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/status" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MySQL - Create User
```powershell
$body = @{
    name = "John Doe"
    email = "john@example.com"
    age = 30
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MySQL - Get All Users
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MySQL - Get User by ID
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users/1" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MySQL - Update User
```powershell
$body = @{
    name = "Jane Doe"
    email = "jane@example.com"
    age = 28
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users/1" `
  -Method PUT `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MySQL - Delete User
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users/1" -Method DELETE | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MariaDB - Create User
```powershell
$body = @{
    name = "Alice Smith"
    email = "alice@example.com"
    age = 25
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/mariadb/users" `
  -Method POST `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MariaDB - Get All Users
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mariadb/users" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MariaDB - Get User by ID
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mariadb/users/1" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MariaDB - Update User
```powershell
$body = @{
    name = "Alice Updated"
    email = "alice.new@example.com"
    age = 26
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/mariadb/users/1" `
  -Method PUT `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### MariaDB - Delete User
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mariadb/users/1" -Method DELETE | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### PostgreSQL - Create User
```powershell
$body = @{
    name = "Bob Smith"
    email = "bob@example.com"
    age = 35
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/postgres/users" `
  -Method POST `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### PostgreSQL - Get All Users
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/postgres/users" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### PostgreSQL - Get User by ID
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/postgres/users/1" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### PostgreSQL - Update User
```powershell
$body = @{
    name = "Bob Updated"
    email = "bob.new@example.com"
    age = 36
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/postgres/users/1" `
  -Method PUT `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### PostgreSQL - Delete User
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/postgres/users/1" -Method DELETE | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### Firebase - Create User
```powershell
$body = @{
    name = "Firebase User"
    email = "firebase@example.com"
    age = 40
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/firebase/users" `
  -Method POST `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### Firebase - Get All Users
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/firebase/users" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### Firebase - Get User by ID
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/firebase/users/1" -Method GET | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### Firebase - Update User
```powershell
$body = @{
    name = "Firebase User Updated"
    email = "firebase.updated@example.com"
    age = 41
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/firebase/users/1" `
  -Method PUT `
  -Body $body `
  -ContentType "application/json" | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

### Firebase - Delete User
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/firebase/users/1" -Method DELETE | Select-Object -ExpandProperty Content | ConvertFrom-Json | ConvertTo-Json
```

---

## Database Features

✅ **Auto-Migration**: Tables are created automatically on startup
✅ **GORM ORM**: Using GORM for type-safe database operations
✅ **Environment Configuration**: All credentials in .env file
✅ **Multi-Database Support**: MySQL, MariaDB, PostgreSQL
✅ **RESTful API**: Standard HTTP methods for CRUD operations

---

## API Response Format

### Success Response
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

### Error Response
```json
{
  "status": "error",
  "error": "User not found"
}
```

---

## Oracle Database CRUD Operations

### Create User
```bash
curl -X POST http://localhost:8000/api/oracle/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }'
```

**Response (201 Created):**
```json
{
  "status": "ok",
  "message": "User created successfully in Oracle",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }
}
```

### Get All Users
```bash
curl http://localhost:8000/api/oracle/users
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "message": "Users retrieved successfully from Oracle",
  "data": [
    {
      "id": 1,
      "name": "Oracle User",
      "email": "oracle@example.com",
      "age": 30
    }
  ]
}
```

### Get User by ID
```bash
curl http://localhost:8000/api/oracle/users/1
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "message": "User retrieved successfully from Oracle",
  "data": {
    "id": 1,
    "name": "Oracle User",
    "email": "oracle@example.com",
    "age": 30
  }
}
```

### Update User
```bash
curl -X PUT http://localhost:8000/api/oracle/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "age": 28
  }'
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "message": "User updated successfully in Oracle",
  "data": {
    "id": 1,
    "name": "Jane Doe",
    "email": "jane@example.com",
    "age": 28
  }
}
```

### Delete User
```bash
curl -X DELETE http://localhost:8000/api/oracle/users/1
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "message": "User deleted successfully from Oracle"
}
```

---

## User Model

```go
type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name"`
    Email string `json:"email" gorm:"uniqueIndex"`
    Age   int    `json:"age"`
}
```

---

## Configuration

Edit `.env` to customize:
- Database hosts and ports
- Credentials (user/password)
- API port and host
- Elasticsearch and etcd settings

```env
MYSQL_HOST=mysql8
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=root
MYSQL_DATABASE=app_db

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=app_db

# ... more settings
```

---

## Logs & Debugging

View service logs:
```bash
docker-compose logs -f axiomnizam
```

View specific database logs:
```bash
docker-compose logs -f mysql8
docker-compose logs -f postgres
docker-compose logs -f mariadb
```

---

## Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

---

Enjoy your AxiomNizam API! 🎉
