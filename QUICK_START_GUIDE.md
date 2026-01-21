# AxiomNizam - Quick Start & Testing Guide

## 🚀 System Status: ALL SYSTEMS OPERATIONAL ✅

---

## 📋 Quick Reference

### Services & Ports
| Service | URL | Port | Purpose |
|---------|-----|------|---------|
| Backend API | http://localhost:8000 | 8000 | Main API server |
| Frontend | http://localhost:7000 | 7000 | Dashboard |
| Keycloak | http://localhost:8080 | 8080 | Authentication |
| MySQL | localhost:3306 | 3306 | Database |
| PostgreSQL | localhost:5432 | 5432 | Database |
| MongoDB | localhost:27017 | 27017 | Database |
| MariaDB | localhost:3307 | 3307 | Database |
| Percona | localhost:3308 | 3308 | Database |
| Oracle | localhost:1521 | 1521 | Database |
| Firebase | localhost:9000,8080 | Various | Emulator |
| Valkey | localhost:6379 | 6379 | Cache |
| Elasticsearch | localhost:9200 | 9200 | Search |
| etcd | localhost:2379 | 2379 | Config |

---

## 🔑 Authentication Setup

### 1. Get Token (FIRST STEP)

**PowerShell**:
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body "client_id=admin-cli&grant_type=password&username=admin&password=admin"

$token = $response.access_token
Write-Host "Token: $token"
```

**cURL**:
```bash
curl -X POST http://localhost:8080/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=admin-cli&grant_type=password&username=admin&password=admin"
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR...",
  "expires_in": 300,
  "token_type": "Bearer"
}
```

**Save the token** - you'll need it for all CRUD operations.

---

## 🧪 Test Public Endpoints (No Auth Needed)

### Health Check
```powershell
curl http://localhost:8000/health
```
**Response**: `{"status":"ok","message":"AxiomNizam API is running"}`

### System Status (All Databases)
```powershell
curl http://localhost:8000/status
```
**Response**: Shows connection status of all 10 services

---

## 📝 Test Protected Endpoints (Auth Required)

### Basic Pattern
```powershell
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

curl -Headers $headers http://localhost:8000/api/mysql/users
```

---

## 🗄️ Test Each Database

### MySQL Example (Full CRUD)

**CREATE**:
```powershell
$body = @{
    name = "John Doe"
    email = "john@example.com"
    age = 30
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Headers $headers `
  -Body $body
```

**READ ALL**:
```powershell
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" `
  -Headers $headers
```

**READ ONE**:
```powershell
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users/1" `
  -Headers $headers
```

**UPDATE**:
```powershell
$body = @{
    name = "Jane Doe"
    email = "jane@example.com"
    age = 31
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users/1" `
  -Method PUT `
  -Headers $headers `
  -Body $body
```

**DELETE**:
```powershell
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users/1" `
  -Method DELETE `
  -Headers $headers
```

---

## 🌐 Test All 7 Databases

Replace `/api/mysql/users` with:
- `/api/postgres/users` - PostgreSQL
- `/api/mongodb/users` - MongoDB
- `/api/mariadb/users` - MariaDB
- `/api/percona/users` - Percona
- `/api/firebase/users` - Firebase
- `/api/oracle/users` - Oracle

**All endpoints follow same CRUD pattern above.**

---

## 📮 Import Postman Collection

1. **Download**: `POSTMAN_COLLECTION.json`
2. **Open Postman** → Import → Choose file
3. **Select Environment Variables**:
   - `base_url`: http://localhost:8000
   - `keycloak_url`: http://localhost:8080
   - `admin_username`: admin
   - `admin_password`: admin
   - `token`: (auto-populated)
4. **First Request**: Run "Get Access Token" request (auto-saves token)
5. **Then**: Run any CRUD requests

---

## 🔧 Troubleshooting

### Keycloak Not Responding
```bash
# Check Keycloak is running
docker-compose ps keycloak

# Check logs
docker-compose logs keycloak | tail -50

# Wait 60 seconds, then verify
curl http://localhost:8080/realms/master/.well-known/openid-configuration
```

### Token Invalid
```powershell
# Get new token
$token = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body "client_id=admin-cli&grant_type=password&username=admin&password=admin").access_token

Write-Host $token
```

### 401 Unauthorized
- Token is missing or invalid
- Header format must be: `Authorization: Bearer <TOKEN>`
- Token may be expired (get new one)

### Database Connection Failed
```powershell
# Check status endpoint
curl http://localhost:8000/status

# Shows which databases are connected
```

---

## 📊 Testing Checklist

- [ ] Run: Get public health endpoint (no auth)
- [ ] Run: Get system status (no auth)
- [ ] Get token from Keycloak
- [ ] Test: Create user in MySQL
- [ ] Test: Read all users from MySQL
- [ ] Test: Read user by ID from MySQL
- [ ] Test: Update user in MySQL
- [ ] Test: Delete user from MySQL
- [ ] Repeat above for PostgreSQL
- [ ] Repeat above for MongoDB
- [ ] Repeat above for MariaDB
- [ ] Repeat above for Percona
- [ ] Repeat above for Firebase
- [ ] Repeat above for Oracle
- [ ] Test: 401 error without token
- [ ] Test: 401 error with invalid token

---

## 📖 Full Documentation

| Document | Purpose |
|----------|---------|
| [AUTH_GUIDE.md](AUTH_GUIDE.md) | Authentication & Keycloak setup |
| [API_GUIDE.md](API_GUIDE.md) | Complete API endpoints reference |
| [POSTMAN_API_GUIDE.md](POSTMAN_API_GUIDE.md) | Postman testing guide |
| [COMPLETE_SETUP_ANALYSIS.md](COMPLETE_SETUP_ANALYSIS.md) | Full architecture & verification |

---

## 🎯 Common Test Scenarios

### Scenario 1: Create 3 Users in Different Databases
```powershell
# Get token
$token_response = Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body "client_id=admin-cli&grant_type=password&username=admin&password=admin"

$token = $token_response.access_token
$headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }

# User 1 in MySQL
$user1 = @{ name="Alice"; email="alice@example.com"; age=25 } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Method POST -Headers $headers -Body $user1

# User 2 in PostgreSQL
$user2 = @{ name="Bob"; email="bob@example.com"; age=30 } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8000/api/postgres/users" -Method POST -Headers $headers -Body $user2

# User 3 in MongoDB
$user3 = @{ name="Charlie"; email="charlie@example.com"; age=35 } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8000/api/mongodb/users" -Method POST -Headers $headers -Body $user3

# Verify all were created
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers
Invoke-RestMethod -Uri "http://localhost:8000/api/postgres/users" -Headers $headers
Invoke-RestMethod -Uri "http://localhost:8000/api/mongodb/users" -Headers $headers
```

### Scenario 2: Load Testing with Multiple Requests
```powershell
# Create 10 users in MySQL
for ($i = 1; $i -le 10; $i++) {
    $user = @{ name="User$i"; email="user$i@example.com"; age=(25+$i) } | ConvertTo-Json
    Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" `
      -Method POST `
      -Headers $headers `
      -Body $user
    Write-Host "Created user $i"
}

# Get all users
$users = Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers
Write-Host "Total users: $($users.Count)"
```

---

## 🔐 Security Notes

- **Never hardcode tokens** - Get new token before each session
- **Use HTTPS in production** - Currently using HTTP for development
- **Token expires in 300 seconds** - Get new token if expired
- **All CRUD operations require token** - Public endpoints: /health, /status only
- **JWT validated on every request** - RSA signature verification

---

## 📞 Support Commands

### View Logs
```bash
docker-compose logs -f axiomnizam
docker-compose logs -f keycloak
docker-compose logs -f postgres
```

### Restart Everything
```bash
docker-compose down -v
docker-compose up -d
```

### Check Service Status
```bash
docker-compose ps
```

### View Database Contents (MySQL)
```bash
docker-compose exec mysql8 mysql -uroot -proot app_db -e "SELECT * FROM users;"
```

### View Database Contents (PostgreSQL)
```bash
docker-compose exec postgres psql -U postgres -d app_db -c "SELECT * FROM users;"
```

### View Database Contents (MongoDB)
```bash
docker-compose exec mongodb mongosh --username root --password root --authenticationDatabase admin
# Then: use app_db; db.users.find();
```

---

## ✅ Verification Checklist

- [x] All code read and verified
- [x] All documentation reviewed
- [x] Auth configuration checked
- [x] API endpoints documented
- [x] Database connectivity verified
- [x] Keycloak integration tested
- [x] Postman collection created
- [x] Environment variables configured
- [x] Docker services running
- [x] JWT validation implemented
- [x] Data persistence enabled
- [x] Ready for testing

---

## 🎉 Ready to Test!

1. **Access Keycloak**: http://localhost:8080/admin (admin/admin)
2. **Check Backend**: http://localhost:8000/health
3. **View Dashboard**: http://localhost:7000
4. **Get Token**: See "Authentication Setup" above
5. **Run Postman**: Import POSTMAN_COLLECTION.json
6. **Start Testing**: Follow "Test Protected Endpoints" section

**All 35 CRUD operations are ready for testing!**

