# AxiomNizam - Multi-Database API Platform

> A comprehensive, production-ready API platform with Keycloak authentication, Role-Based Access Control (RBAC), and support for 7 major databases.

**Status**: ✅ **FULLY OPERATIONAL** | Latest Update: January 22, 2026

---

## 📑 Table of Contents

1. [Overview](#-what-is-axiomnizam)
2. [Quick Start (5 min)](#-quick-start-5-minutes)
3. [Full Setup Guide](#-full-setup-guide-30-minutes)
4. [Keycloak Configuration](#-keycloak-configuration)
5. [Authentication & RBAC](#-authentication--authorization)
6. [API Reference](#-api-reference)
7. [Frontend Dashboard](#-frontend-dashboard)
8. [Testing with Postman](#-testing-with-postman)
9. [Architecture](#-system-architecture)
10. [Troubleshooting](#-troubleshooting)

---

## 🎯 What is AxiomNizam?

AxiomNizam is a unified API platform that allows you to manage data across multiple database systems through a single, authenticated REST API interface.

### Supported Databases
- ✅ MySQL
- ✅ MariaDB
- ✅ PostgreSQL
- ✅ Percona (MySQL fork)
- ✅ MongoDB
- ✅ Firebase
- ✅ Oracle

### Key Features

#### ✅ Authentication & Authorization
- **Keycloak Integration** - OpenID Connect authentication
- **JWT Tokens** - Secure token-based access
- **Role-Based Access Control** - Admin/User role separation
- **Multi-Realm Support** - Flexible authentication configuration

#### ✅ API Capabilities
- **35+ API Endpoints** - Full CRUD operations across all databases
- **Request/Response Validation** - Comprehensive error handling
- **CORS Enabled** - Frontend integration ready
- **Rate Limiting Ready** - Scalable architecture

#### ✅ Frontend Dashboard
- **Real-time Health Monitoring** - Live database connection status
- **API Documentation** - Built-in API reference
- **Dark/Light/Default Themes** - Customizable UI
- **Admin Panel** - Database and table management
- **System Manager** - Advanced operations

#### ✅ Notification System
- **Discord Integration** - Real-time alerts
- **Health Notifications** - Automatic status updates
- **Custom Notifications** - Send custom messages

---

## 🚀 Quick Start (5 Minutes)

### Prerequisites
- Docker & Docker Compose
- Go 1.18+
- Postman (optional, for testing)

### Step 1: Start All Services (1 min)

```bash
cd AxiomNizam
docker-compose up -d
```

This starts:
- ✅ Keycloak (http://localhost:8080)
- ✅ Backend API (http://localhost:8000)
- ✅ Frontend Dashboard (http://localhost:7000)
- ✅ All databases (PostgreSQL, MySQL, MongoDB, etc.)

**Wait for services to be ready** (check: `docker-compose ps`)

```bash
# Expected output:
# CONTAINER ID   NAME         STATUS
# ...            keycloak     Up 2 minutes
# ...            backend      Up 1 minute
# ...            frontend     Up 1 minute
# ...            postgres     Up 2 minutes
# ...            mysql        Up 2 minutes
```

### Step 2: Get Your Auth Token (2 min)

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72&grant_type=client_credentials" \
  | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

echo "Your token: $TOKEN"
```

Save the token value - you'll need it for the next step.

### Step 3: Test the API (1 min)

Replace `YOUR_TOKEN` with the token from Step 2:

```bash
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected response:**
```json
{
  "status": "success",
  "message": "Users retrieved successfully",
  "data": []
}
```

### Step 4: View the Dashboard (1 min)

Open your browser:
```
http://localhost:7000
```

You'll see:
- ✅ Health status
- ✅ Database connections
- ✅ Available APIs
- ✅ API documentation

### 🎉 You're Done!

Your AxiomNizam API is now running. What's next?

**Try some commands:**

```bash
# Create a user (Admin only)
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "age": 28
  }'

# Get all users
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN"

# Check health (no token needed)
curl http://localhost:8000/health

# Check database connections (no token needed)
curl http://localhost:8000/status
```

---

## 🔧 Full Setup Guide (30 Minutes)

### Prerequisites

Before starting, ensure you have:
- ✅ Docker & Docker Compose installed
- ✅ Go 1.18 or later (for local development)
- ✅ 8GB RAM minimum
- ✅ Ports available: 7000, 8000, 8080, 3306, 5432, 27017

### Installation Methods

#### Method 1: Docker Compose (Recommended)

##### Step 1: Clone or Download Project
```bash
cd AxiomNizam
```

##### Step 2: Start All Services
```bash
docker-compose up -d
```

This starts:
- **Keycloak** (localhost:8080) - Authentication server
- **Backend API** (localhost:8000) - Go API server  
- **Frontend** (localhost:7000) - Dashboard
- **PostgreSQL** (localhost:5432) - Database
- **MySQL** (localhost:3306) - Database
- **MongoDB** (localhost:27017) - Database

##### Step 3: Verify Services
```bash
docker-compose ps

# Should show all containers as "Up"
```

##### Step 4: Wait for Startup
- Keycloak: ~30-40 seconds
- Databases: ~10-15 seconds
- Backend/Frontend: ~5 seconds

Check logs if needed:
```bash
docker-compose logs -f keycloak    # Watch Keycloak startup
docker-compose logs -f axiomnizam  # Watch Backend
docker-compose logs -f frontend    # Watch Frontend
```

#### Method 2: Local Development

##### Backend Setup

1. **Install dependencies:**
```bash
cd AxiomNizam
go mod download
```

2. **Configure environment** (edit `.env`):
```bash
BACKEND_PORT=8000
FRONTEND_PORT=7000
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
DATABASE_TYPE=postgres
DATABASE_HOST=localhost
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DISCORD_WEBHOOK_URL=https://discordapp.com/api/webhooks/...
```

3. **Run backend:**
```bash
go run main.go
# Server runs on http://localhost:8000
```

##### Frontend Setup

1. **Navigate to frontend:**
```bash
cd AxiomNizam/frontend
go mod download
```

2. **Configure environment** (edit `.env`):
```bash
FRONTEND_PORT=7000
BACKEND_URL=http://localhost:8000
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-frontend
```

3. **Run frontend:**
```bash
go run main.go
# Server runs on http://localhost:7000
```

### Configuration Files

#### Backend `.env`

```dotenv
# Server Configuration
BACKEND_PORT=8000
FRONTEND_PORT=7000

# Keycloak Configuration
KEYCLOAK_URL=http://keycloak:8080  # Use service name in Docker
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT=axiomnizam-backend
KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72

# Database Connections
# PostgreSQL
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=axiomnizam

# MySQL
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=root
MYSQL_DATABASE=axiomnizam

# MongoDB
MONGODB_URI=mongodb://mongodb:27017

# Oracle (optional)
ORACLE_HOST=localhost
ORACLE_PORT=1521
ORACLE_USER=admin
ORACLE_PASSWORD=admin
ORACLE_SID=xe

# Discord Integration (optional)
DISCORD_WEBHOOK_URL=https://discordapp.com/api/webhooks/...

# JWT Configuration
JWT_EXPIRATION=3600
JWT_REFRESH_EXPIRATION=86400
```

#### Frontend `.env`

```dotenv
FRONTEND_PORT=7000
BACKEND_URL=http://localhost:8000
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-frontend
KEYCLOAK_REDIRECT_URI=http://localhost:7000/callback
```

### Services & Ports

| Service | Port | URL |
|---------|------|-----|
| Frontend | 7000 | http://localhost:7000 |
| Backend API | 8000 | http://localhost:8000 |
| Keycloak | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | localhost |
| MySQL | 3306 | localhost |
| MongoDB | 27017 | localhost |

### Verification Checklist

#### Services Running
```bash
# Check all services
docker-compose ps

# Expected output:
# ✅ keycloak     Up
# ✅ axiomnizam   Up
# ✅ frontend     Up
# ✅ postgres     Up
# ✅ mysql        Up
# ✅ mongodb      Up
```

#### Health Checks
```bash
# Backend health
curl http://localhost:8000/health
# Expected: {"status":"ok","message":"..."}

# Backend status
curl http://localhost:8000/status
# Expected: Database connectivity status

# Frontend
curl http://localhost:7000
# Expected: HTML dashboard

# Keycloak
curl http://localhost:8080/health
# Expected: Keycloak health info
```

---

## 🔐 Keycloak Configuration

### Create a New Realm

#### Step 1: Login to Keycloak Admin Console

```
URL: http://localhost:8080
Username: admin
Password: admin
```

#### Step 2: Create Realm

1. Click **Realm dropdown** (top-left corner)
2. Click **Create Realm** button
3. Fill in the form:
   - **Realm name**: `axiomnizam`
   - **Enabled**: ON
4. Click **Create**

✅ **Realm `axiomnizam` created successfully**

### Create Frontend Client

#### Purpose
The frontend client is used by the web browser to authenticate users via OpenID Connect.

#### Configuration

1. Go to **Realm: axiomnizam** (top-left dropdown)
2. Left sidebar → **Clients** → **Create client**
3. Fill in:
   - **Client ID**: `axiomnizam-frontend`
   - **Client type**: `OpenID Connect`
   - **Name**: `AxiomNizam Frontend`
4. Click **Next**
5. Capability Configuration:
   - **Client authentication**: OFF
   - **Standard flow enabled**: ON
   - Click **Next** → **Save**
6. Configure URLs:
   - **Root URL**: http://localhost:7000
   - **Home URL**: http://localhost:7000
   - **Valid redirect URIs**: http://localhost:7000/*
   - **Valid post logout redirect URIs**: http://localhost:7000/*
   - **Web origins**: http://localhost:7000
   - Click **Save**

✅ **Frontend client configured**

### Create Backend Client

#### Purpose
The backend client is used for:
- Direct user authentication (resource owner password grant)
- Server-to-server communication
- Keeping the client secret secure (never exposed to browser)

#### Configuration

1. **Clients** → **Create client**
2. Fill in:
   - **Client ID**: `axiomnizam-backend`
   - **Client type**: `OpenID Connect`
   - **Name**: `AxiomNizam Backend`
3. Click **Next**
4. Capability Configuration:
   - **Client authentication**: ON
   - **Service account roles**: ON
   - **Direct access grants**: ON
   - Click **Next** → **Save**
5. Go to **Credentials** tab
6. Copy the **Client secret** value
7. Update `.env` file with the secret

**Important**: Store the client secret securely in `.env`. Never commit to git!

### Create Roles

1. Go to **Realm: axiomnizam**
2. Left sidebar → **Roles** → **Create role**
3. Create the following roles:

| Role Name      | Description                 |
| -------------- | --------------------------- |
| system-manager | Full system admin access    |
| admin          | Admin panel access          |
| user           | Standard user access        |

✅ **All roles created**

### Create Users

#### Admin User

1. Go to **Users** → **Add user**
2. Fill in:
   - **Username**: `admin`
   - **Email**: `admin@example.com`
   - **First name**: `Admin`
   - **Last name**: `User`
   - **Email verified**: ON
   - **Enabled**: ON
3. Click **Create**
4. Go to **Credentials** tab → **Set password**
   - Enter: `admin`
   - **Temporary**: OFF
   - Click **Set Password**
5. Go to **Role mapping** tab → **Assign a role**
   - Select: `admin`
   - Click **Assign**

✅ **Admin user created**

#### System Manager User

1. **Users** → **Add user**
2. Fill in:
   - **Username**: `sysadmin`
   - **Email**: `sysadmin@example.com`
   - **First name**: `System`
   - **Last name**: `Admin`
3. Click **Create**
4. **Credentials** → **Set password** → `sysadmin123` → **Set Password**
5. **Role mapping** → **Assign a role** → `system-manager` → **Assign**

✅ **System manager user created**

#### Regular User

1. **Users** → **Add user**
2. Fill in:
   - **Username**: `viewer`
   - **Email**: `viewer@example.com`
3. Click **Create**
4. **Credentials** → **Set password** → `viewer123` → **Set Password**
5. **Role mapping** → **Assign a role** → `user` → **Assign**

✅ **Regular user created**

### Test Login Credentials

| Username | Password      | Role           | Access Level |
| -------- | ------------- | -------------- | ------------ |
| admin    | admin         | admin          | Full access  |
| sysadmin | sysadmin123   | system-manager | System admin |
| viewer   | viewer123     | user           | Read-only    |

---

## 🔐 Authentication & Authorization

### Overview

AxiomNizam uses:
- **Keycloak** for authentication and user management
- **JWT Tokens** for stateless API authentication
- **Role-Based Access Control (RBAC)** for authorization
- **OpenID Connect** for modern OAuth 2.0 flows

### Authentication Flow

```
┌──────────────┐
│   Client     │
│   (Browser)  │
└──────┬───────┘
       │
       │ 1. Request Token
       ↓
┌──────────────────────┐
│  Keycloak Server     │
│  (localhost:8080)    │
└──────┬───────────────┘
       │
       │ 2. Return JWT
       ↓
┌──────────────────────┐
│  API Request + JWT   │
│  Backend (8000)      │
└──────┬───────────────┘
       │
       │ 3. Validate JWT
       │ 4. Check Role
       ↓
┌──────────────────────┐
│  Execute Operation   │
│  Access Database     │
└──────────────────────┘
```

### Token Acquisition

#### Method 1: Client Credentials (Service-to-Service)

**Best for**: Backend services, automated tasks

```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=client_credentials"
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

#### Method 2: Resource Owner Password Grant (User Login)

**Best for**: End-user login, personal access

```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=password" \
  -d "username=admin" \
  -d "password=admin"
```

#### Method 3: Authorization Code (Browser-based)

**Best for**: Web applications, user-interactive flows

1. Redirect user to:
```
http://localhost:8080/realms/axiomnizam/protocol/openid-connect/auth?
  client_id=axiomnizam-frontend&
  response_type=code&
  scope=openid%20profile%20email&
  redirect_uri=http://localhost:7000/callback
```

2. Exchange code for token:
```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=authorization_code" \
  -d "code=YOUR_AUTH_CODE" \
  -d "redirect_uri=http://localhost:7000/callback"
```

### Using Tokens in API Requests

#### Standard Authorization Header

```bash
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### JavaScript/Fetch
```javascript
fetch('http://localhost:8000/api/mysql/users', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
```

#### PowerShell
```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers
```

### JWT Token Structure

A JWT token has 3 parts: `header.payload.signature`

#### Payload Example
```json
{
  "exp": 1642771200,
  "iat": 1642767600,
  "iss": "http://localhost:8080/realms/axiomnizam",
  "sub": "abc123def456",
  "preferred_username": "admin",
  "realm_access": {
    "roles": ["admin", "default-roles-axiomnizam"]
  },
  "email": "admin@example.com"
}
```

#### Key Fields

| Field | Meaning |
|-------|---------|
| `exp` | Token expiration time (Unix timestamp) |
| `iat` | Token issued at time |
| `sub` | Subject (user ID) |
| `preferred_username` | Username |
| `realm_access.roles` | User's roles (admin, user, etc.) |
| `email` | User's email |

### Role-Based Access Control (RBAC)

#### Permission Matrix

| Operation | Admin | Non-Admin | Anonymous |
|-----------|-------|-----------|-----------|
| GET (Read) | ✅ | ✅ | ❌ |
| POST (Create) | ✅ | ❌ | ❌ |
| PUT (Update) | ✅ | ❌ | ❌ |
| DELETE (Delete) | ✅ | ❌ | ❌ |
| /health | ✅ | ✅ | ✅ |
| /status | ✅ | ✅ | ✅ |

### Public Endpoints (No Auth Required)

```bash
# Health check
curl http://localhost:8000/health

# System status
curl http://localhost:8000/status

# Keycloak well-known config
curl http://localhost:8080/realms/axiomnizam/.well-known/openid-configuration
```

### Protected Endpoints (Auth Required)

#### Admin Only

```bash
# Create user (any database)
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# Update user
curl -X PUT http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane"}'

# Delete user
curl -X DELETE http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

#### Any Authenticated User

```bash
# Read operations (GET)
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $USER_TOKEN"

curl http://localhost:8000/api/postgres/users \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Error Responses

#### 401 Unauthorized (No/Invalid Token)
```json
{
  "error": "unauthorized",
  "message": "Invalid or missing token"
}
```

#### 403 Forbidden (Insufficient Role)
```json
{
  "error": "forbidden",
  "message": "You don't have permission for this operation"
}
```

### Token Refresh

Include `scope=offline_access` in token request to get refresh token:

```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=client_credentials" \
  -d "scope=offline_access"
```

Use refresh token:
```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN"
```

---

## 📡 API Reference

### Overview

AxiomNizam provides RESTful CRUD operations for 7 databases with **35+ API endpoints** + 5 admin/health endpoints = **40+ total endpoints**.

### Base URL

```
http://localhost:8000
```

### Response Format

#### Success (2xx)

```json
{
  "status": "success",
  "message": "Operation successful",
  "data": { ... }
}
```

#### Error (4xx, 5xx)

```json
{
  "status": "error",
  "message": "Error description",
  "error": "error_code"
}
```

### Database CRUD Endpoints

Each database has 5 endpoints:

#### 1. Get All Users

```http
GET /api/{database}/users
Authorization: Bearer {token}
```

**Parameters:**
- `limit` (optional): Number of records
- `offset` (optional): Pagination offset
- `sort` (optional): Sort field
- `order` (optional): asc or desc

**Example:**
```bash
curl "http://localhost:8000/api/mysql/users?limit=10&offset=0" \
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
      "age": 30
    }
  ]
}
```

#### 2. Get User by ID

```http
GET /api/{database}/users/{id}
Authorization: Bearer {token}
```

#### 3. Create User (Admin Only)

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

#### 4. Update User (Admin Only)

```http
PUT /api/{database}/users/{id}
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "name": "Jane Doe",
  "age": 29
}
```

#### 5. Delete User (Admin Only)

```http
DELETE /api/{database}/users/{id}
Authorization: Bearer {admin_token}
```

### Supported Databases

| Database   | Endpoints |
|-----------|-----------|
| MySQL | GET, POST, PUT, DELETE on `/api/mysql/users` |
| MariaDB | GET, POST, PUT, DELETE on `/api/mariadb/users` |
| PostgreSQL | GET, POST, PUT, DELETE on `/api/postgres/users` |
| Percona | GET, POST, PUT, DELETE on `/api/percona/users` |
| MongoDB | GET, POST, PUT, DELETE on `/api/mongodb/users` |
| Firebase | GET, POST, PUT, DELETE on `/api/firebase/users` |
| Oracle | GET, POST, PUT, DELETE on `/api/oracle/users` |

### Admin Endpoints (Admin Only)

#### Create Database
```http
POST /api/admin/database/create
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "database_name": "new_database",
  "database_type": "mysql"
}
```

#### List Databases
```http
GET /api/admin/database/list
Authorization: Bearer {admin_token}
```

#### Create Table
```http
POST /api/admin/table/create
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "database_name": "axiomnizam",
  "table_name": "products",
  "columns": [
    {"name": "id", "type": "INT", "primary_key": true},
    {"name": "title", "type": "VARCHAR(255)"}
  ]
}
```

#### List Tables
```http
GET /api/admin/table/list
Authorization: Bearer {admin_token}

{
  "database_name": "axiomnizam"
}
```

### Notification Endpoints

#### Send Custom Notification
```http
POST /api/notifications/send
Authorization: Bearer {token}
Content-Type: application/json

{
  "message": "Database backup completed",
  "level": "info"
}
```

### Error Codes

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

### Common Examples

#### PowerShell

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

#### cURL

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

#### JavaScript/Fetch

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
```

---

## 🎨 Frontend Dashboard

### Features

- 🎨 Beautiful, responsive web interface
- 💚 Real-time health status monitoring
- 🗄️ Database connection status display
- 🔄 Auto-refresh capability (default: every 5 seconds)
- 📱 Mobile-friendly design
- 🚀 Lightweight and fast

### Running the Frontend

#### Prerequisites
- Go 1.21 or higher
- The AxiomNizam backend running on `http://localhost:8000`

#### Development Mode

```bash
cd frontend

# Download dependencies
go mod download

# Run the frontend
go run main.go
```

The dashboard will be available at `http://localhost:7000`

#### With Custom Backend URL

```bash
BACKEND_URL=http://your-backend-host:8000 go run main.go
```

#### Using Docker

Build the Docker image:
```bash
docker build -t axiomnizam-frontend:latest .
```

Run the container:
```bash
docker run -d \
  -p 7000:7000 \
  -e BACKEND_URL=http://axiomnizam:8000 \
  --name axiomnizam-frontend \
  axiomnizam-frontend:latest
```

#### With Docker Compose

The frontend service is included in the main `docker-compose.yml`:

```bash
docker-compose up -d frontend
```

### Environment Variables

- `BACKEND_URL`: The URL of the AxiomNizam backend (default: `http://localhost:8000`)
- `FRONTEND_PORT`: The port to run the dashboard on (default: `7000`)

### API Endpoints

The frontend exposes:
- `GET /` - Main dashboard page
- `GET /api/health` - Fetch backend health status (JSON)
- `GET /api/status` - Fetch database connection status (JSON)

### Dashboard Features

#### Health Status
Displays the overall health of the backend API:
- Status indicator (OK/Error)
- Status message

#### Database Connections
Shows connection status for all configured databases:
- MySQL, MariaDB
- PostgreSQL, Percona
- MongoDB
- Firebase
- Oracle
- Keycloak

#### Auto-Refresh
- Automatically refreshes data every 5 seconds (configurable)
- Manual refresh button
- Last update timestamp

### Customization

Edit `templates/` files to customize:
- Colors and styling in `style.css`
- Layout in `*.html` files
- Refresh interval in JavaScript files

---

## 📮 Testing with Postman

### Import the Collection

1. Download `POSTMAN_COLLECTION.json` from the project root
2. Open Postman
3. Click **Import** → Select the JSON file
4. All endpoints are pre-configured and ready to use!

### Setup Environment Variables

Create a new Environment with:

```
Variable Name     | Value
-----------------|------------------------
base_url          | http://localhost:8000
keycloak_url      | http://localhost:8080
access_token      | (leave empty, auto-filled)
client_id         | axiomnizam-backend
client_secret     | 6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
username          | admin
password          | admin
```

### Get Access Token

**Name**: Get Access Token
**Method**: POST
**URL**: `{{keycloak_url}}/realms/axiomnizam/protocol/openid-connect/token`

**Headers**:
```
Content-Type: application/x-www-form-urlencoded
```

**Body** (form-urlencoded):
```
client_id = {{client_id}}
client_secret = {{client_secret}}
grant_type = client_credentials
```

**Tests** (Auto-save token):
```javascript
var jsonData = pm.response.json();
pm.environment.set("access_token", jsonData.access_token);
```

### Pre-configured Requests

The collection includes ready-to-use requests for:
- ✅ Health check (no auth)
- ✅ System status (no auth)
- ✅ Get all users (all databases)
- ✅ Get user by ID
- ✅ Create user (admin)
- ✅ Update user (admin)
- ✅ Delete user (admin)
- ✅ Admin database operations

### Auto-Token Script

Each authenticated request includes:
```javascript
Authorization: Bearer {{access_token}}
```

Run "Get Access Token" request first to populate the token!

---

## 🏗️ System Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Browser / Client                      │
└─────────────────────────┬────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ↓                 ↓                 ↓
    Frontend          Health Check      Status Check
    :7000             :8000/health       :8000/status
    
        ↓─────────────────┴─────────────────↓
        
┌──────────────────────────────────────────────────────────┐
│                    API Gateway (Go)                      │
│                    :8000                                 │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Authentication & RBAC Middleware                │   │
│  │  - Validate JWT tokens                           │   │
│  │  - Check role permissions                        │   │
│  │  - Rate limiting                                 │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────┘
                      │
    ┌─────────────────┼────────────────┬──────────────┬──────────────┐
    ↓                 ↓                ↓              ↓              ↓
  MySQL           PostgreSQL       MongoDB       Firebase        Oracle
  :3306           :5432            :27017        (Cloud)         :1521
```

### Data Flow

1. **Client (Browser)** → Requests API endpoint with JWT token
2. **Frontend Server** → Serves dashboard, redirects based on role
3. **API Gateway** → Validates JWT, checks RBAC, routes to database
4. **Database** → Executes query, returns results
5. **Response** → Returns to client with status and data

### Key Components

| Component | Purpose | Technology |
|-----------|---------|-----------|
| **Frontend** | Web dashboard, user interface | Go + HTML/CSS/JS |
| **Backend** | API gateway, business logic | Go + Gin framework |
| **Authentication** | User auth, JWT validation | Keycloak + OpenID Connect |
| **Databases** | Data storage | Multiple: MySQL, PostgreSQL, MongoDB, etc. |
| **Authorization** | Role-based access control | JWT claims + custom middleware |

---

## 🐛 Troubleshooting

### Services won't start

```bash
# Check Docker status
docker-compose ps

# View logs
docker-compose logs keycloak
docker-compose logs axiomnizam
docker-compose logs frontend

# Stop and restart
docker-compose down
docker-compose up -d
```

### Port conflicts

```bash
# Check what's using port 8000
netstat -tuln | grep 8000

# Or use lsof (macOS/Linux)
lsof -i :8000
```

### Can't get token?

```bash
# Verify Keycloak is running
curl http://localhost:8080/health

# Check realm configuration
curl http://localhost:8080/realms/axiomnizam/.well-known/openid-configuration
```

### API returns 401 Unauthorized

- Token expired? Get a new one
- Wrong token? Verify client credentials
- Role insufficient? Use admin token for write operations

### API returns 403 Forbidden

- Your user role doesn't have permission
- Use admin token for POST/PUT/DELETE operations
- Check role assignment in Keycloak

### Frontend can't connect to backend

```bash
# Verify backend is running
curl http://localhost:8000/health

# Check BACKEND_URL in frontend .env
echo $BACKEND_URL

# Check browser console for CORS errors
```

### Database connection errors

```bash
# Check database is running
docker-compose ps | grep postgres

# View logs
docker-compose logs postgres

# Test connection
psql -h localhost -U postgres -d axiomnizam
```

### Token validation fails

- Verify PUBLIC_KEY in backend `.env`
- Get it from: `http://localhost:8080/realms/axiomnizam/protocol/openid-connect/certs`
- Ensure token matches realm

### Keycloak taking too long

- Keycloak can take 30-60 seconds to start
- Check logs: `docker-compose logs keycloak`
- Wait longer and try again

### "Invalid Client" Error

- Client secret in `.env` is wrong
- Client credentials are for wrong client
- Client was deleted
- **Fix**: Go to Keycloak → Clients → axiomnizam-backend → Credentials → Copy secret → Update `.env`

### CORS Error from Frontend

- Frontend is calling Keycloak directly instead of backend proxy
- **Fix**: Ensure frontend calls `http://localhost:8000/auth/login` NOT `http://localhost:8080/...`

---

## 📊 Key Credentials Reference

### Keycloak
- **URL**: http://localhost:8080
- **Realm**: `axiomnizam`
- **Admin User**: `admin` / `admin`

### Backend Service
- **Client ID**: `axiomnizam-backend`
- **Client Secret**: `6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72`
- **Grant Type**: Client Credentials

### Frontend Service
- **Client ID**: `axiomnizam-frontend`
- **Realm**: `axiomnizam`

### Default Test Users
| Username | Password      | Role           |
|----------|---------------|----------------|
| admin    | admin         | admin          |
| sysadmin | sysadmin123   | system-manager |
| viewer   | viewer123     | user           |

---

## ✅ Verification Checklist

- [x] Keycloak running and configured
- [x] Backend API operational
- [x] Frontend dashboard accessible
- [x] All 7 databases connected
- [x] RBAC implemented and tested
- [x] JWT authentication working
- [x] Postman collection ready
- [x] Documentation complete

---

## 📦 Project Structure

```
AxiomNizam/
├── README.md                          # Main documentation
├── QUICK_START.md                     # 5-minute guide (deprecated)
├── SETUP_GUIDE.md                     # Setup guide (deprecated)
├── API_REFERENCE.md                   # API docs (deprecated)
├── AUTHENTICATION.md                  # Auth docs (deprecated)
├── POSTMAN_API_GUIDE.md              # Postman guide (deprecated)
├── KEYCLOAK_SETUP.md                 # Keycloak details (deprecated)
├── POSTMAN_COLLECTION.json           # Ready-to-import
│
├── main.go                            # Backend entry point
├── go.mod                             # Dependencies
├── docker-compose.yml                 # Service orchestration
├── init-postgres.sql                  # Database initialization
│
├── frontend/
│   ├── main.go                        # Frontend server
│   ├── go.mod
│   └── templates/
│       ├── layout.html                # Base layout
│       ├── public-dashboard.html      # Public view
│       ├── admin.html                 # Admin panel
│       ├── system-manager.html        # System manager
│       ├── auth.js                    # Auth module
│       ├── dashboard.js               # Dashboard logic
│       ├── admin.js                   # Admin logic
│       ├── style.css                  # Styles
│       └── responsive.css             # Mobile styles
│
└── internal/
    ├── auth/
    │   ├── auth.go                    # JWT validation
    │   └── middleware.go              # Auth middleware
    ├── config/
    │   └── config.go                  # Configuration
    ├── database/
    │   └── connections.go             # DB connections
    ├── handlers/
    │   ├── handlers.go                # Route handlers
    │   ├── mysql.go                   # MySQL handlers
    │   ├── postgres.go                # PostgreSQL handlers
    │   ├── mongodb.go                 # MongoDB handlers
    │   ├── firebase.go                # Firebase handlers
    │   ├── oracle.go                  # Oracle handlers
    │   ├── admin_handler.go           # Admin handlers
    │   └── notification_handler.go    # Notification handlers
    ├── models/
    │   └── models.go                  # Data models
    └── utils/
        └── ...                        # Utility functions
```

---

## 🚀 Next Steps

1. **Get Started**: Follow Quick Start section (5 min)
2. **Configure**: Use Full Setup Guide section (30 min)
3. **Test APIs**: Import POSTMAN_COLLECTION.json
4. **Learn More**: Read individual sections as needed
5. **Deploy**: Use docker-compose or Kubernetes

---

## 🎓 Learning Paths

### Path 1: Just Make It Work (5 min)
1. Start services: `docker-compose up -d`
2. Get token (see Quick Start)
3. Make API call (see Quick Start)
4. Done! ✅

### Path 2: Understand the System (30 min)
1. Read Quick Start section
2. Read Full Setup Guide section
3. Read Authentication & Authorization section
4. Import POSTMAN_COLLECTION.json
5. Test all endpoints
6. Done! ✅

### Path 3: Master Everything (60+ min)
1. Read all sections
2. Review system architecture
3. Study RBAC implementation
4. Explore database schemas
5. Deploy to production
6. Implement monitoring
7. Done! ✅

---

## 📞 Support & Resources

### Questions?
1. Check **Quick Start** section for fast answers
2. Read **Full Setup Guide** for configuration
3. Review **API Reference** for endpoints
4. See **Authentication & Authorization** for auth details

### Still stuck?
- Check `.env` files for correct configuration
- Review container logs: `docker-compose logs`
- Verify all services are running: `docker-compose ps`

### External Resources
- **Keycloak Docs**: https://www.keycloak.org/documentation.html
- **Go Docs**: https://golang.org/doc
- **Docker Docs**: https://docs.docker.com
- **OpenID Connect**: https://openid.net/connect/
- **JWT Info**: https://jwt.io/
- **OAuth 2.0**: https://oauth.net/2/

---

## 📄 License

MIT License - See LICENSE file for details

---

## 🏆 Achievements

✅ **FULLY OPERATIONAL** - All systems tested and working
✅ **Production-Ready** - Enterprise-grade implementation
✅ **Well-Documented** - Comprehensive guides and examples
✅ **Secure** - JWT authentication + RBAC
✅ **Scalable** - Supports multiple databases
✅ **Developer-Friendly** - Easy to test and integrate

---

**Made with ❤️ for seamless multi-database management**

**Last Updated**: January 22, 2026 | **Status**: ✅ Fully Operational

