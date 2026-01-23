# AxiomNizam - Comprehensive Documentation

> A comprehensive, production-ready API platform with Keycloak authentication, Role-Based Access Control (RBAC), and support for 7 major databases.

**Status**: ✅ **FULLY OPERATIONAL** | Latest Update: January 23, 2026

---

## 📑 Table of Contents

1. [System Overview](#system-overview)
2. [Quick Start (5 min)](#quick-start-5-minutes)
3. [Full Setup Guide (30 min)](#full-setup-guide-30-minutes)
4. [Keycloak Configuration](#keycloak-configuration)
5. [Authentication & RBAC](#authentication--rbac)
6. [Dynamic Query System](#dynamic-query-system)
7. [Query Persistence & Logging](#query-persistence--logging)
8. [API Metrics & Analytics](#api-metrics--analytics)
9. [Rate Limiting & Token Management](#rate-limiting--token-management)
10. [Frontend Dashboard](#frontend-dashboard)
11. [Testing with Postman](#testing-with-postman)
12. [System Architecture](#system-architecture)
13. [Docker & Deployment](#docker--deployment)
14. [Troubleshooting](#troubleshooting)

---

## System Overview

### What is AxiomNizam?

AxiomNizam is a unified API platform that allows you to manage data across multiple database systems through a single, authenticated REST API interface with advanced features like query logging, rate limiting, metrics tracking, and horizontal scaling support.

### Supported Databases

- ✅ MySQL
- ✅ MariaDB
- ✅ PostgreSQL
- ✅ Percona (MySQL fork)
- ✅ MongoDB
- ✅ Firebase
- ✅ Oracle
- ✅ Elasticsearch
- ✅ Valkey/Redis
- ✅ etcd

### Key Features

#### ✅ Authentication & Authorization
- **Keycloak Integration** - OpenID Connect authentication
- **JWT Tokens** - Secure token-based access
- **Role-Based Access Control** - Admin/User role separation
- **Multi-Realm Support** - Flexible authentication configuration
- **Rate Limiting** - 500 API calls per token, 10 minute validity
- **Token Management** - Self-service and admin monitoring endpoints

#### ✅ Query System
- **Dynamic SQL Queries** - Send any SQL query without creating endpoints
- **Multi-Database Support** - Same query endpoints for all databases
- **Query Logging** - Automatic audit trail of all executed queries
- **Query Persistence** - Disk and Valkey storage with multi-pod support
- **Query Statistics** - Performance metrics and usage analytics
- **Batch Operations** - Execute multiple queries in one request

#### ✅ API Capabilities
- **45+ API Endpoints** - Full CRUD operations across all databases
- **Request/Response Validation** - Comprehensive error handling
- **CORS Enabled** - Frontend integration ready
- **API Metrics** - Track endpoint usage and performance
- **Query Logs** - Retrieve execution history per database
- **Schema Inspection** - Inspect table structures dynamically

#### ✅ Frontend Dashboard
- **Real-time Health Monitoring** - Live database connection status
- **API Documentation** - Built-in API reference
- **Dark/Light/Default Themes** - Customizable UI
- **Admin Panel** - Database and table management
- **System Manager** - Advanced operations
- **Query Explorer** - Test queries directly from UI

#### ✅ Notification System
- **Discord Integration** - Real-time alerts
- **Health Notifications** - Automatic status updates
- **Custom Notifications** - Send custom messages
- **Query Alerts** - Notify on slow queries

---

## Quick Start (5 Minutes)

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
- ✅ All databases (PostgreSQL, MySQL, MongoDB, Valkey, etc.)

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

Your AxiomNizam API is now running.

---

## Full Setup Guide (30 Minutes)

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

This starts all services with persistent volumes for data storage.

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

2. **Configure environment** (create `.env`):
```bash
BACKEND_PORT=8000
FRONTEND_PORT=7000
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
DATABASE_TYPE=postgres
DATABASE_HOST=localhost
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
RATE_LIMIT_MAX_CALLS=500
RATE_LIMIT_VALIDITY_MINUTES=10
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

2. **Configure environment** (create `.env`):
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
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT=axiomnizam-backend
KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72

# Database Connections
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=root
MYSQL_DATABASE=axiomnizam

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=axiomnizam

MONGODB_URI=mongodb://mongodb:27017

# Rate Limiting Configuration
RATE_LIMIT_MAX_CALLS=500
RATE_LIMIT_VALIDITY_MINUTES=10

# Discord Integration (optional)
DISCORD_WEBHOOK_URL=https://discordapp.com/api/webhooks/...
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

---

## Keycloak Configuration

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

### Create Clients

#### Frontend Client
1. Go to **Realm: axiomnizam**
2. Left sidebar → **Clients** → **Create client**
3. Fill in:
   - **Client ID**: `axiomnizam-frontend`
   - **Client type**: `OpenID Connect`
4. Click **Next**
5. Capability Configuration:
   - **Client authentication**: OFF
   - **Standard flow enabled**: ON
6. Configure URLs:
   - **Valid redirect URIs**: http://localhost:7000/*
   - **Web origins**: http://localhost:7000

#### Backend Client
1. **Clients** → **Create client**
2. Fill in:
   - **Client ID**: `axiomnizam-backend`
   - **Client type**: `OpenID Connect`
3. Capability Configuration:
   - **Client authentication**: ON
   - **Service account roles**: ON

### Create Roles and Users

1. **Roles**: Create `admin`, `user`, and `system-manager` roles
2. **Users**: Create users and assign roles

---

## Authentication & RBAC

### How It Works

Your AxiomNizam backend uses a multi-layer authentication system:

```
User Request
  ↓
Step 1: Check Authorization Header
  ├─ Missing → 401 Unauthorized
  └─ Present → Continue
  ↓
Step 2: Validate JWT Signature
  ├─ Invalid → 401 Unauthorized
  └─ Valid → Continue
  ↓
Step 3: Check Token Validity (10 minutes)
  ├─ Expired → 401 Unauthorized
  └─ Valid → Continue
  ↓
Step 4: Check API Call Rate Limit (500 calls)
  ├─ Exceeded → 401 Unauthorized
  └─ Available → Continue
  ↓
Step 5: Check User Role (for admin operations)
  ├─ User without admin role → 403 Forbidden
  └─ Admin → Continue
  ↓
Execute Handler
```

### Getting Your Token

```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=YOUR_SECRET&grant_type=client_credentials"
```

### Using Your Token

```bash
TOKEN="your_token_here"

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/users"
```

### Rate Limiting Details

- **500 API calls per token**: Each token allows 500 authenticated requests
- **10-minute validity**: Tokens expire after 10 minutes
- **Automatic decrement**: Each call uses 1 of your 500 calls
- **New token required**: After expiration or limit reached

### Check Your Token Status

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/auth/token-status" | jq
```

Response includes:
- `calls_remaining`: How many calls you have left
- `is_expired`: Whether token has expired
- `time_remaining`: Time until expiration
- `expires_at`: Exact expiration timestamp

---

## Dynamic Query System

### Overview

The Dynamic Query System allows you to send SQL queries directly to any supported database without creating new endpoints.

### Endpoints

#### GET Request (Read-only)
```
GET /api/{db}/query?q=YOUR_QUERY&params=value1,value2
Authorization: Bearer TOKEN
```

**Restrictions**: Only SELECT, SHOW, DESCRIBE, EXPLAIN queries allowed

**Example**:
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%20*%20FROM%20users%20WHERE%20age%20%3E%20?&params=25"
```

#### POST Request (All operations)
```
POST /api/{db}/query
Authorization: Bearer TOKEN
Content-Type: application/json

{
  "query": "SQL_QUERY",
  "params": ["value1", "value2"]
}
```

**Supported**: SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, TRUNCATE

**Example**:
```json
{
  "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
  "params": ["Alice", "alice@example.com", 28]
}
```

#### Batch Queries
```
POST /api/{db}/query/batch
Authorization: Bearer TOKEN
Content-Type: application/json

[
  {"query": "QUERY_1", "params": []},
  {"query": "QUERY_2", "params": ["value1"]}
]
```

#### Table Schema
```
GET /api/{db}/schema?table=table_name
Authorization: Bearer TOKEN
```

### Supported Databases

| Database | Prefix |
|----------|--------|
| MySQL | `/api/mysql/` |
| MariaDB | `/api/mariadb/` |
| PostgreSQL | `/api/postgres/` |
| Percona | `/api/percona/` |
| Oracle | `/api/oracle/` |

### Response Format

**Success**:
```json
{
  "status": "ok",
  "message": "Query executed successfully",
  "data": [
    {"id": 1, "name": "John", "email": "john@example.com"}
  ]
}
```

**Write Operation**:
```json
{
  "status": "ok",
  "message": "Query executed successfully",
  "data": {
    "rows_affected": 1
  }
}
```

**Error**:
```json
{
  "status": "error",
  "error": "Query execution failed: column does not exist"
}
```

### Usage Examples

#### Example 1: Get All Users
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%20*%20FROM%20users"
```

#### Example 2: Insert Data
```bash
curl -X POST http://localhost:8000/api/mysql/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
    "params": ["Alice Smith", "alice@example.com", 28]
  }'
```

#### Example 3: Update Data
```bash
curl -X POST http://localhost:8000/api/mysql/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "UPDATE users SET age = ? WHERE id = ?",
    "params": [30, 1]
  }'
```

#### Example 4: Delete Data
```bash
curl -X POST http://localhost:8000/api/mysql/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "DELETE FROM users WHERE id = ?",
    "params": [1]
  }'
```

#### Example 5: Batch Operations
```bash
curl -X POST http://localhost:8000/api/mysql/query/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '[
    {"query": "SELECT COUNT(*) as total FROM users", "params": []},
    {"query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "params": ["Bob", "bob@example.com", 27]}
  ]'
```

---

## Query Persistence & Logging

### Overview

Every API query executed is automatically logged with complete metadata. These logs are stored in two places for redundancy and performance.

### What Gets Logged

```json
{
  "id": "pod1-1704067200000-123",
  "query": "SELECT * FROM users",
  "params": ["value1"],
  "database": "mysql",
  "user_id": "user123",
  "status": "success",
  "error": null,
  "duration": 45,
  "timestamp": "2024-01-01T12:00:00Z",
  "hostname": "axiomnizam-pod-1"
}
```

### Storage

#### Disk-Based Storage
- **Location**: `/data/query_logs` (Docker volume)
- **Format**: JSONL (one JSON per line)
- **Retention**: Permanent (never auto-deleted)
- **Rotation**: Daily by date
- **Survives**: Pod crashes and restarts

#### Valkey/Redis Cache
- **Real-time**: Instantly accessible
- **Distributed**: All pods can access
- **TTL**: 30-day automatic cleanup
- **Format**: Sorted sets indexed by database
- **Cross-pod**: Single source of truth

### New Endpoints

#### Query Logs (Audit Trail)
```bash
GET /api/{db}/logs?limit=100&offset=0
Authorization: Bearer TOKEN
```

**Response**:
```json
{
  "status": "ok",
  "data": [
    {
      "id": "hostname-1704067200000-123",
      "query": "SELECT * FROM users WHERE id = ?",
      "params": ["1"],
      "database": "mysql",
      "user_id": "user123",
      "status": "success",
      "error": null,
      "duration": 45,
      "timestamp": "2024-01-01T12:00:00Z",
      "hostname": "axiomnizam-pod-1"
    }
  ],
  "pagination": {
    "total": 1250,
    "limit": 100,
    "offset": 0
  }
}
```

#### Query Statistics
```bash
GET /api/{db}/stats
Authorization: Bearer TOKEN
```

**Response**:
```json
{
  "status": "ok",
  "data": {
    "total_queries": 2500,
    "success_count": 2480,
    "error_count": 20,
    "average_duration_ms": 87,
    "min_duration_ms": 5,
    "max_duration_ms": 5000,
    "queries_by_status": {
      "success": 2480,
      "error": 20
    },
    "queries_by_type": {
      "SELECT": 1500,
      "INSERT": 400,
      "UPDATE": 350,
      "DELETE": 200,
      "OTHER": 50
    }
  }
}
```

### Testing Logging

```bash
# 1. Send a test query
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/query?q=SELECT%201"

# 2. Retrieve logs
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/logs" | jq

# 3. Check statistics
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/stats" | jq

# 4. Verify disk storage
docker exec axiomnizam ls -la /data/query_logs/
docker exec axiomnizam cat /data/query_logs/2024-01-01.jsonl

# 5. Verify Valkey
docker exec valkey redis-cli KEYS "query_logs:*"
```

### Multi-Pod Scaling

When you scale to multiple pods:

```bash
docker-compose up -d --scale axiomnizam=3
```

All 3 pods:
- Write to same persistent volume
- Write to same Valkey instance
- Generate unique query IDs (hostname + timestamp)
- Contribute to unified audit trail

Verify:
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/logs" | jq '.data[].hostname' | sort | uniq
```

---

## API Metrics & Analytics

### Overview

Your API automatically tracks all endpoint usage with comprehensive metrics including success rates, performance, and error tracking.

### Endpoints

#### Get API Count & Usage Summary
```bash
GET /api/admin/metrics/count
Authorization: Bearer ADMIN_TOKEN
```

**Response**:
```json
{
  "status": "ok",
  "data": {
    "total_unique_endpoints": 45,
    "total_api_calls": 8234,
    "endpoint_usage": {
      "/health": 3456,
      "/api/mysql/users": 1234,
      "/api/mysql/query": 890
    }
  }
}
```

#### Get All Detailed Metrics
```bash
GET /api/admin/metrics/all
Authorization: Bearer ADMIN_TOKEN
```

**Response** (detailed metrics for each endpoint):
```json
{
  "status": "ok",
  "data": {
    "total_unique_endpoints": 45,
    "total_calls": 8234,
    "endpoints": [
      {
        "endpoint": "/api/mysql/users",
        "method": "GET",
        "total_calls": 156,
        "success_calls": 150,
        "error_calls": 6,
        "average_duration_ms": 45,
        "max_duration_ms": 120,
        "min_duration_ms": 15,
        "last_called": "2024-01-01T12:34:56Z",
        "status_codes": {
          "200": 150,
          "400": 4,
          "401": 2
        }
      }
    ]
  }
}
```

#### Get API Statistics
```bash
GET /api/admin/metrics/stats?endpoint=/optional/path
Authorization: Bearer ADMIN_TOKEN
```

### Real-World Examples

#### Find Total Number of APIs
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/count" | jq '.data.total_unique_endpoints'
```

#### Find Most-Used API
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/count" | \
  jq '.data.endpoint_usage | to_entries | max_by(.value)'
```

#### Find Slowest API
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/all" | \
  jq '.data.endpoints | max_by(.average_duration_ms) | {endpoint: .endpoint, avg_ms: .average_duration_ms}'
```

#### Find API with Most Errors
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/all" | \
  jq '.data.endpoints | max_by(.error_calls) | {endpoint: .endpoint, errors: .error_calls}'
```

#### Get Success Rate (All APIs)
```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/admin/metrics/stats" | \
  jq '.data | {success_rate: (.success_calls / .total_calls * 100 | round)}'
```

---

## Rate Limiting & Token Management

### How Rate Limiting Works

```
Login → Get Token with 500 API calls
  ↓
Make API Call → Uses 1 call
  ↓
X-RateLimit-Remaining: 499
  ↓
After 500 calls OR 10 minutes → Token Expires
  ↓
Login Again → Get Fresh Token with 500 calls
```

### Configuration

Rate limiting is configured via environment variables:

```bash
RATE_LIMIT_MAX_CALLS=500           # Max calls per token
RATE_LIMIT_VALIDITY_MINUTES=10     # Token validity in minutes
```

#### Configuration Examples

**Testing (Strict)**:
```bash
RATE_LIMIT_MAX_CALLS=10
RATE_LIMIT_VALIDITY_MINUTES=1
```

**Standard (Default)**:
```bash
RATE_LIMIT_MAX_CALLS=500
RATE_LIMIT_VALIDITY_MINUTES=10
```

**Production (Generous)**:
```bash
RATE_LIMIT_MAX_CALLS=5000
RATE_LIMIT_VALIDITY_MINUTES=60
```

### Response Headers

Every protected API response includes rate limit info:

```
X-RateLimit-Limit: 500              (max calls)
X-RateLimit-Remaining: 495          (calls left)
X-Token-Expires-At: 2024-01-23 12:25:30
```

### Token Status Endpoints

#### Check Your Token
```bash
GET /auth/token-status
Authorization: Bearer TOKEN
```

**Response**:
```json
{
  "status": "ok",
  "data": {
    "username": "admin",
    "calls_made": 47,
    "max_calls": 500,
    "calls_remaining": 453,
    "issued_at": "2024-01-23T12:15:30Z",
    "expires_at": "2024-01-23T12:25:30Z",
    "is_expired": false,
    "time_remaining": "9m45s",
    "last_used": "2024-01-23T12:24:15Z"
  }
}
```

#### Admin: View All Tokens
```bash
GET /auth/admin/tokens-status
Authorization: Bearer ADMIN_TOKEN
```

**Response**:
```json
{
  "status": "ok",
  "data": {
    "active_tokens": 5,
    "total_api_calls": 847,
    "tokens": [
      {
        "username": "admin",
        "calls_made": 47,
        "calls_remaining": 453,
        "expires_at": "2024-01-23T12:25:30Z",
        "time_remaining": "9m45s"
      }
    ]
  }
}
```

### Public Endpoints (No Token Required)

These endpoints don't count against your rate limit:

```bash
GET /health     # Health check
GET /status     # System status
```

### Error Scenarios

#### Token Expired
```json
{
  "error": "token expired",
  "message": "your token is no longer valid. please login again to get a new token"
}
```

#### API Call Limit Exceeded
```json
{
  "error": "api call limit exceeded",
  "message": "you have used all 500 api calls allowed per token",
  "action_required": "login again to get a fresh token with new 500 calls"
}
```

#### Missing Token
```json
{
  "error": "missing authorization header"
}
```

---

## Frontend Dashboard

### Features

- **Real-time Health Monitoring** - Live database connection status
- **API Documentation** - Built-in API reference
- **Dark/Light/Default Themes** - Customizable UI
- **Admin Panel** - Database and table management
- **System Manager** - Advanced operations
- **Query Explorer** - Test queries directly
- **Responsive Design** - Works on desktop and mobile

### Accessing the Dashboard

```
http://localhost:7000
```

### Default Login

```
Username: admin
Password: admin
```

### Dashboard Features

1. **Health Status Tab**
   - View all database connections
   - See which databases are online/offline
   - Real-time connection indicators

2. **API Documentation Tab**
   - Browse all available endpoints
   - Read endpoint descriptions
   - See example requests and responses

3. **Admin Dashboard Tab** (Admin only)
   - Manage databases
   - Create/drop tables
   - View user management
   - System configuration

4. **Query Explorer Tab**
   - Write and test SQL queries
   - Execute against any database
   - View results in real-time
   - Save frequently used queries

---

## Testing with Postman

### Import Postman Collection

1. Download available Postman collections:
   - `POSTMAN_COLLECTION.json` - Main API collection
   - `API_METRICS_POSTMAN.json` - Metrics endpoints
   - `DYNAMIC_QUERIES_POSTMAN.json` - Query system

2. Open Postman
3. Click **Import** button
4. Select the JSON file
5. Click **Import**

### Set Variables in Postman

1. Click **Variables** tab (top left)
2. Find `admin_token` variable
3. Get token:
   ```bash
   curl -X POST http://localhost:8000/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"password"}'
   ```
4. Paste token value
5. Click **Save**

### Run Collections

- Click any request in the collection
- Click **Send** button
- View response

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (React/HTML)                    │
│                    http://localhost:7000                    │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│              Backend API (Go + Gin)                         │
│              http://localhost:8000                          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Rate Limiting Middleware (500 calls, 10 min)        │  │
│  │  JWT Authentication Middleware                       │  │
│  │  Query Logging Middleware                            │  │
│  │  Metrics Tracking Middleware                         │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
         │                    │                   │
    ┌────▼────┐        ┌─────▼──────┐      ┌─────▼──────┐
    │ Keycloak │        │  Database  │      │   Valkey/  │
    │ Auth     │        │   Layer    │      │   Redis    │
    │ (JWT)    │        │  (GORM)    │      │   (Cache)  │
    └──────────┘        └─────┬──────┘      └────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            │                 │                 │
       ┌────▼────┐       ┌────▼────┐     ┌────▼────┐
       │  MySQL  │       │ Postgres │     │ MongoDB  │
       └────┬────┘       └────┬────┘     └────┬────┘
            │                 │                │
       ┌────▼─────────────────▼────────────────▼────┐
       │  Query Logs & Persistence                  │
       │  /data/query_logs (JSONL + Valkey)         │
       └─────────────────────────────────────────────┘
```

### Request Flow

```
1. HTTP Request arrives at Backend
   ↓
2. Rate Limiting Middleware
   ├─ Check token exists
   ├─ Check token not expired
   ├─ Check calls remaining > 0
   └─ Decrement call count
   ↓
3. JWT Authentication Middleware
   ├─ Extract JWT from Authorization header
   ├─ Validate signature
   └─ Extract claims (user ID, roles)
   ↓
4. Query Logging Middleware (if applicable)
   ├─ Record request start time
   └─ Prepare logging context
   ↓
5. Metrics Tracking Middleware
   ├─ Record endpoint path
   └─ Record HTTP method
   ↓
6. Route Handler
   ├─ Execute business logic
   ├─ Query appropriate database
   └─ Return results
   ↓
7. Query Logging (post-execution)
   ├─ Record duration
   ├─ Record status code
   ├─ Log to disk (/data/query_logs)
   └─ Log to Valkey
   ↓
8. Metrics Update
   ├─ Record endpoint execution
   ├─ Record performance metrics
   └─ Store in Valkey
   ↓
9. Response Sent
   ├─ Include rate limit headers
   ├─ Include standard headers
   └─ Return JSON response
```

### Database Abstraction Layer

All databases use GORM (Go ORM) for consistency:

```
Handler
  ↓
GORM Interface
  ├─ MySQL (via mysql driver)
  ├─ PostgreSQL (via postgres driver)
  ├─ MariaDB (via mysql driver)
  ├─ Percona (via mysql driver)
  └─ Oracle (via postgres driver)
  ↓
Actual Database
```

---

## Docker & Deployment

### Docker Compose Services

The `docker-compose.yml` includes:

```yaml
Services:
  - keycloak (port 8080)      - Authentication server
  - axiomnizam (port 8000)    - Backend API
  - frontend (port 7000)      - Frontend dashboard
  - postgres (port 5432)      - PostgreSQL database
  - mysql (port 3306)         - MySQL database
  - mongodb (port 27017)      - MongoDB database
  - valkey (port 6379)        - Redis/Valkey cache
  - etcd (port 2379)          - Configuration service

Volumes:
  - postgres_data             - PostgreSQL persistence
  - mysql_data                - MySQL persistence
  - mongodb_data              - MongoDB persistence
  - query_logs                - Query logging persistence
  - keycloak_data             - Keycloak persistence
```

### Docker Build

#### Build Backend
```bash
docker build -t axiomnizam:latest .
```

#### Build Frontend
```bash
docker build -f frontend/Dockerfile -t axiomnizam-frontend:latest frontend/
```

### Docker Run (Single Container)

```bash
docker run -d \
  -p 8000:8000 \
  -e BACKEND_PORT=8000 \
  -e MYSQL_HOST=host.docker.internal \
  -e POSTGRES_HOST=host.docker.internal \
  -e KEYCLOAK_URL=http://host.docker.internal:8080 \
  -e RATE_LIMIT_MAX_CALLS=500 \
  -e RATE_LIMIT_VALIDITY_MINUTES=10 \
  axiomnizam:latest
```

### Kubernetes Deployment (Optional)

For production Kubernetes deployments:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: axiomnizam
spec:
  replicas: 3
  selector:
    matchLabels:
      app: axiomnizam
  template:
    metadata:
      labels:
        app: axiomnizam
    spec:
      containers:
      - name: axiomnizam
        image: axiomnizam:latest
        ports:
        - containerPort: 8000
        env:
        - name: BACKEND_PORT
          value: "8000"
        - name: RATE_LIMIT_MAX_CALLS
          value: "500"
        - name: RATE_LIMIT_VALIDITY_MINUTES
          value: "10"
        volumeMounts:
        - name: query-logs
          mountPath: /data/query_logs
      volumes:
      - name: query-logs
        persistentVolumeClaim:
          claimName: query-logs-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: query-logs-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 100Gi
```

---

## Troubleshooting

### Issue: Keycloak Not Starting

**Symptoms**: Docker shows Keycloak container exiting

```bash
# Check logs
docker-compose logs keycloak

# Check port availability
netstat -an | grep 8080

# Kill process on port 8080 if needed
lsof -i :8080 | grep LISTEN | awk '{print $2}' | xargs kill -9
```

### Issue: Database Connection Failed

**Symptoms**: API returns database connection errors

```bash
# Check database containers running
docker-compose ps | grep postgres
docker-compose ps | grep mysql

# Check logs
docker-compose logs postgres
docker-compose logs mysql

# Check network connectivity
docker network inspect axiomnizam_default
```

### Issue: Token Expired (401 Unauthorized)

**Solution**: Get a fresh token

```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Issue: API Call Limit Exceeded

**Symptoms**: Getting 401 after making 500 calls

**Solution**: Get a fresh token (logs out current session)

```bash
# Same as above
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Issue: Query Logs Not Appearing

**Symptoms**: `/api/mysql/logs` returns empty results

```bash
# Check query_logs volume exists
docker volume ls | grep query_logs

# Check volume mounted correctly
docker inspect axiomnizam | grep -A 3 "Mounts"

# Check logs directory
docker exec axiomnizam ls -la /data/query_logs/

# Check Valkey running
docker exec valkey redis-cli ping
# Should return: PONG
```

### Issue: Port Already in Use

**Example**: Port 8000 already in use

```bash
# Find what's using the port
lsof -i :8000

# Kill the process
kill -9 <PID>

# Or use a different port in docker-compose.yml
# Change: 8000:8000 to 8001:8000
```

### Issue: Slow Query Execution

**Symptoms**: Queries taking longer than expected

```bash
# Check query statistics
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/stats" | jq '.data.average_duration_ms'

# Check database performance
docker exec mysql mysql -u root -p<password> -e "SHOW PROCESSLIST;"

# Check Valkey performance
docker exec valkey redis-cli --stat
```

### Issue: High Memory Usage

**Symptoms**: Docker containers consuming lots of RAM

```bash
# Check container memory usage
docker stats

# Increase Docker memory limit
# In Docker Desktop: Settings → Resources → Memory

# Or in docker-compose.yml add:
# services:
#   mysql:
#     deploy:
#       resources:
#         limits:
#           memory: 2G
```

### Issue: Container Keep Crashing

**Symptoms**: `docker-compose ps` shows containers restarting

```bash
# Check logs with timestamps
docker-compose logs --timestamps

# Increase startup timeout
docker-compose up --wait

# Check disk space
df -h

# Check if /data/query_logs needs permissions
docker exec axiomnizam chmod -R 755 /data/query_logs
```

---

## Advanced Usage

### Scaling to Multiple Pods

```bash
# Scale to 3 instances
docker-compose up -d --scale axiomnizam=3

# Verify all running
docker-compose ps | grep axiomnizam

# Test unified logs
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/logs" | jq '.data[].hostname' | sort | uniq
```

### Custom Rate Limit Configuration

Edit `.env`:

```bash
RATE_LIMIT_MAX_CALLS=1000      # Increase to 1000
RATE_LIMIT_VALIDITY_MINUTES=30  # Increase to 30 minutes
```

Restart application:

```bash
docker-compose restart axiomnizam
```

### Elasticsearch Integration (Optional)

Uncomment Elasticsearch in `connections.go` and add to `.env`:

```bash
ELASTICSEARCH_URL=http://localhost:9200
```

### Custom Notification Webhooks

Set Discord webhook URL in `.env`:

```bash
DISCORD_WEBHOOK_URL=https://discordapp.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_TOKEN
```

### Enable Query Caching

Valkey is already configured for caching. To increase cache size:

```yaml
# In docker-compose.yml, modify valkey service
valkey:
  command: redis-server --maxmemory 2gb --maxmemory-policy allkeys-lru
```

---

## Monitoring & Maintenance

### Daily Monitoring Tasks

```bash
# Check all services running
docker-compose ps

# Check rate limiting activity
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "http://localhost:8000/auth/admin/tokens-status"

# Check API metrics
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "http://localhost:8000/api/admin/metrics/count"

# Check query logs
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/mysql/stats"
```

### Weekly Maintenance

```bash
# Clean up old query logs (>90 days)
docker exec axiomnizam find /data/query_logs -name "*.jsonl" -mtime +90 -delete

# Check disk usage
docker exec axiomnizam du -sh /data/query_logs

# Verify Valkey memory usage
docker exec valkey redis-cli INFO memory
```

### Monthly Reviews

- Review API metrics trends
- Check error rates per endpoint
- Analyze slow query patterns
- Update rate limiting if needed

---

## Summary

AxiomNizam is a **production-ready, fully-featured API platform** with:

✅ **Multi-Database Support** - MySQL, PostgreSQL, MongoDB, and more  
✅ **Automatic Query Logging** - Complete audit trail with dual persistence  
✅ **Rate Limiting** - 500 calls per token, 10 minute validity  
✅ **API Metrics** - Track all endpoints and their performance  
✅ **Dynamic Queries** - Send any SQL without creating endpoints  
✅ **Authentication** - Keycloak + JWT + RBAC  
✅ **Frontend Dashboard** - Admin interface for management  
✅ **Horizontal Scaling** - Multi-pod support with unified logs  
✅ **Production Ready** - Docker deployment included  

**Get started now**: See [Quick Start (5 Minutes)](#quick-start-5-minutes) section above!

---

**For detailed information on specific topics, refer to individual sections above.**

**Status**: ✅ Complete & Production Ready  
**Last Updated**: January 23, 2026  
**Version**: 1.0
