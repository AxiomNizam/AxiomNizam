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

### Health & Status
```bash
# Health check
curl http://localhost:8000/health

# Check all connections
curl http://localhost:8000/status
```

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

### Create User
```powershell
$body = @{
    name = "John Doe"
    email = "john@example.com"
    age = 30
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Body $body `
  -ContentType "application/json"
```

### Get All Users
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users" -Method GET
```

### Update User
```powershell
$body = @{
    name = "Jane Doe"
    email = "jane@example.com"
    age = 28
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users/1" `
  -Method PUT `
  -Body $body `
  -ContentType "application/json"
```

### Delete User
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/mysql/users/1" -Method DELETE
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
