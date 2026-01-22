# 🛠️ Admin API - Database & Table Management

**Date**: January 22, 2026  
**Status**: ✅ Ready for Testing  
**Security**: Admin role required

---

## 📋 Overview

Two new **admin-only** API endpoints for database and table management:

1. **POST /api/admin/database/create** - Create a new database
2. **POST /api/admin/table/create** - Create a new table

Plus utility endpoints:
- **GET /api/admin/database/list** - List databases
- **GET /api/admin/table/list** - List tables

---

## 🔐 Security

✅ **All endpoints require**:
- Valid JWT token
- Token must have "admin" role
- Non-admin users get 403 Forbidden

---

## 🚀 Endpoint Details

### 1️⃣ Create Database

**Endpoint**: `POST /api/admin/database/create`  
**Authentication**: ✅ Required (admin role)

**Request**:
```json
{
  "database_name": "new_database",
  "db_type": "mysql"
}
```

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| database_name | string | Yes | Name of the database to create |
| db_type | string | Yes | Database type: mysql, mariadb, postgres, percona, oracle |

**Supported Databases**:
- ✅ mysql
- ✅ mariadb
- ✅ percona
- ✅ postgres
- ✅ oracle
- ❌ mongodb (native, no SQL)
- ❌ firebase (use Firestore)

**Response** (201 Created):
```json
{
  "status": "success",
  "message": "Database 'new_database' created successfully",
  "database": "new_database",
  "db_type": "mysql"
}
```

**Error Responses**:
```json
// 400 Bad Request - Invalid request
{
  "error": "Invalid request: database_name required"
}

// 400 Bad Request - Unsupported database type
{
  "error": "Database type 'mongodb' not supported"
}

// 503 Service Unavailable - Database not connected
{
  "error": "Database 'mysql' is not connected"
}

// 500 Internal Server Error - Query failed
{
  "error": "Failed to create database: error details..."
}
```

---

### 2️⃣ Create Table

**Endpoint**: `POST /api/admin/table/create`  
**Authentication**: ✅ Required (admin role)

**Request**:
```json
{
  "table_name": "products",
  "db_type": "mysql",
  "columns": [
    {
      "name": "id",
      "type": "INT",
      "nullable": false,
      "primary": true
    },
    {
      "name": "name",
      "type": "VARCHAR",
      "size": 255,
      "nullable": false
    },
    {
      "name": "price",
      "type": "DECIMAL",
      "size": 10,
      "nullable": true
    },
    {
      "name": "description",
      "type": "TEXT",
      "nullable": true
    }
  ]
}
```

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| table_name | string | Yes | Name of the table |
| db_type | string | Yes | Database type |
| columns | array | Yes | Column definitions (min 1) |

**Column Definition**:
| Field | Type | Description |
|-------|------|-------------|
| name | string | Column name |
| type | string | Data type (VARCHAR, INT, TEXT, DECIMAL, etc.) |
| size | int | Size (for VARCHAR, DECIMAL) |
| nullable | bool | Allow NULL values |
| primary | bool | Primary key column |

**Response** (201 Created):
```json
{
  "status": "success",
  "message": "Table 'products' created successfully",
  "table": "products",
  "db_type": "mysql",
  "columns": 4
}
```

---

### 3️⃣ List Databases

**Endpoint**: `GET /api/admin/database/list?db_type=mysql`  
**Authentication**: ✅ Required (admin role)

**Query Parameters**:
| Name | Required | Description |
|------|----------|-------------|
| db_type | Yes | Database type (mysql, postgres, etc.) |

**Response** (200 OK):
```json
{
  "status": "success",
  "db_type": "mysql",
  "databases": [
    "information_schema",
    "mysql",
    "performance_schema",
    "sys",
    "app_db",
    "new_database"
  ],
  "count": 6
}
```

---

### 4️⃣ List Tables

**Endpoint**: `GET /api/admin/table/list?db_type=mysql`  
**Authentication**: ✅ Required (admin role)

**Query Parameters**:
| Name | Required | Description |
|------|----------|-------------|
| db_type | Yes | Database type |

**Response** (200 OK):
```json
{
  "status": "success",
  "db_type": "mysql",
  "tables": [
    "users",
    "products",
    "orders"
  ],
  "count": 3
}
```

---

## 💻 PowerShell Examples

### Get Admin Token

```powershell
$adminToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "password"
    username      = "admin"
    password      = "admin"
  }).access_token

Write-Host "✅ Token: $($adminToken.Substring(0, 50))..."
```

### Create Database

```powershell
$headers = @{
    "Authorization" = "Bearer $adminToken"
    "Content-Type"  = "application/json"
}

$body = @{
    database_name = "my_app_db"
    db_type       = "mysql"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/database/create" `
    -Method POST `
    -Headers $headers `
    -Body $body

Write-Host "✅ Database created: $($response.database)"
```

### Create Table

```powershell
$headers = @{
    "Authorization" = "Bearer $adminToken"
    "Content-Type"  = "application/json"
}

$body = @{
    table_name = "customers"
    db_type    = "mysql"
    columns    = @(
        @{
            name     = "id"
            type     = "INT"
            nullable = $false
            primary  = $true
        },
        @{
            name     = "name"
            type     = "VARCHAR"
            size     = 255
            nullable = $false
        },
        @{
            name     = "email"
            type     = "VARCHAR"
            size     = 255
            nullable = $false
        },
        @{
            name     = "phone"
            type     = "VARCHAR"
            size     = 20
            nullable = $true
        }
    )
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/table/create" `
    -Method POST `
    -Headers $headers `
    -Body $body

Write-Host "✅ Table created with $($response.columns) columns"
```

### List Databases

```powershell
$headers = @{
    "Authorization" = "Bearer $adminToken"
}

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/database/list?db_type=mysql" `
    -Headers $headers

Write-Host "📊 Databases: $($response.databases -join ', ')"
Write-Host "Total: $($response.count)"
```

### List Tables

```powershell
$headers = @{
    "Authorization" = "Bearer $adminToken"
}

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/table/list?db_type=mysql" `
    -Headers $headers

Write-Host "📊 Tables: $($response.tables -join ', ')"
Write-Host "Total: $($response.count)"
```

---

## 🧪 Test Scenarios

### Test 1: Admin Can Create Database ✅

```powershell
# Admin creates database
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/database/create" `
    -Method POST `
    -Headers @{"Authorization"="Bearer $adminToken";"Content-Type"="application/json"} `
    -Body '{"database_name":"test_db","db_type":"mysql"}'

# Expected: 201 Created
# Response: {status: "success", database: "test_db", ...}
```

### Test 2: User Cannot Create Database ❌

```powershell
# Non-admin user tries to create database
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/database/create" `
        -Method POST `
        -Headers @{"Authorization"="Bearer $userToken";"Content-Type"="application/json"} `
        -Body '{"database_name":"test_db","db_type":"mysql"}'
} catch {
    # Expected: 403 Forbidden
    # Error: "forbidden: user does not have 'admin' role"
    Write-Host "✅ Correctly forbidden: $($_.Exception.Response.StatusCode)"
}
```

### Test 3: Admin Can Create Table ✅

```powershell
$body = @{
    table_name = "orders"
    db_type    = "mysql"
    columns    = @(
        @{name="id";type="INT";nullable=$false;primary=$true},
        @{name="amount";type="DECIMAL";size=10;nullable=$false}
    )
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/table/create" `
    -Method POST `
    -Headers @{"Authorization"="Bearer $adminToken";"Content-Type"="application/json"} `
    -Body $body

# Expected: 201 Created
# Response: {status: "success", table: "orders", columns: 2, ...}
```

### Test 4: Admin Can List Databases ✅

```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/database/list?db_type=mysql" `
    -Headers @{"Authorization"="Bearer $adminToken"}

# Expected: 200 OK
# Response: {status: "success", databases: [...], count: 6}
```

### Test 5: Admin Can List Tables ✅

```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/admin/table/list?db_type=mysql" `
    -Headers @{"Authorization"="Bearer $adminToken"}

# Expected: 200 OK
# Response: {status: "success", tables: [...], count: 3}
```

---

## 📊 Complete Test Workflow

```powershell
function Test-AdminAPIs {
    param([string]$AdminToken, [string]$UserToken)
    
    $baseUrl = "http://localhost:8000"
    $headers = @{"Authorization" = "Bearer $AdminToken"; "Content-Type" = "application/json"}
    
    Write-Host "`n🔧 Admin API Testing" -ForegroundColor Yellow
    Write-Host "=" * 60
    
    # Test 1: Create Database
    Write-Host "`n[1/5] Admin - CREATE DATABASE" -ForegroundColor Green
    try {
        $body = @{database_name="admin_test_db"; db_type="mysql"} | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/admin/database/create" `
            -Method POST -Headers $headers -Body $body
        Write-Host "✅ Success: Database '$($response.database)' created"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 2: Create Table
    Write-Host "`n[2/5] Admin - CREATE TABLE" -ForegroundColor Green
    try {
        $cols = @(
            @{name="id"; type="INT"; nullable=$false; primary=$true},
            @{name="email"; type="VARCHAR"; size=255; nullable=$false}
        )
        $body = @{table_name="admin_users"; db_type="mysql"; columns=$cols} | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/admin/table/create" `
            -Method POST -Headers $headers -Body $body
        Write-Host "✅ Success: Table '$($response.table)' with $($response.columns) columns created"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 3: List Databases
    Write-Host "`n[3/5] Admin - LIST DATABASES" -ForegroundColor Green
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/admin/database/list?db_type=mysql" `
            -Headers $headers
        Write-Host "✅ Success: Found $($response.count) databases"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 4: List Tables
    Write-Host "`n[4/5] Admin - LIST TABLES" -ForegroundColor Green
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/admin/table/list?db_type=mysql" `
            -Headers $headers
        Write-Host "✅ Success: Found $($response.count) tables"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 5: User Cannot Create Database
    Write-Host "`n[5/5] User - CREATE DATABASE (Should FAIL)" -ForegroundColor Cyan
    try {
        $userHeaders = @{"Authorization" = "Bearer $UserToken"; "Content-Type" = "application/json"}
        $body = @{database_name="fail_db"; db_type="mysql"} | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/admin/database/create" `
            -Method POST -Headers $userHeaders -Body $body
        Write-Host "❌ ERROR: Should have been forbidden!"
    } catch {
        if ($_ -match "403") {
            Write-Host "✅ Correctly forbidden: User cannot create database"
        }
    }
    
    Write-Host "`n" + "=" * 60
    Write-Host "🎉 Admin API Testing Complete!" -ForegroundColor Yellow
}

# Run tests
Test-AdminAPIs -AdminToken $adminToken -UserToken $userToken
```

---

## 🗄️ Supported Database Types

### SQL Databases (Full Support)

| DB Type | Create DB | Create Table | List DB | List Tables |
|---------|-----------|--------------|---------|-------------|
| MySQL | ✅ | ✅ | ✅ | ✅ |
| MariaDB | ✅ | ✅ | ✅ | ✅ |
| Percona | ✅ | ✅ | ✅ | ✅ |
| PostgreSQL | ✅ | ✅ | ✅ | ✅ |
| Oracle | ⚠️ | ✅ | ⚠️ | ✅ |

### NoSQL Databases (Limited)

| DB Type | Create DB | Create Table | Notes |
|---------|-----------|--------------|-------|
| MongoDB | ❌ | ❌ | Use MongoDB commands |
| Firebase | ❌ | ❌ | Use Firestore console |

---

## 🔍 Data Types Supported

### MySQL/MariaDB/Percona
- VARCHAR(size)
- INT
- BIGINT
- DECIMAL(size)
- TEXT
- LONGTEXT
- DATETIME
- DATE
- TIMESTAMP
- BOOLEAN
- JSON

### PostgreSQL
- varchar(size)
- integer
- bigint
- numeric(size)
- text
- timestamp
- date
- boolean
- json
- uuid

---

## 🛡️ Security & Best Practices

✅ **Implemented**:
- Admin role required for all operations
- JWT token validation
- Proper error responses (400, 403, 500, 503)
- Query logging
- Request validation

✅ **Recommendations**:
- Use meaningful database/table names
- Always validate column definitions
- Test in staging before production
- Monitor database growth
- Backup before creating structures
- Use transaction support where available

---

## 🐛 Troubleshooting

### Error: "Database type 'xxx' not supported"
**Cause**: Using unsupported database type  
**Solution**: Use one of: mysql, mariadb, postgres, percona, oracle

### Error: "Database 'xxx' is not connected"
**Cause**: Database connection failed  
**Solution**: Check RBAC_SETUP_GUIDE.md for connection troubleshooting

### Error: "user does not have 'admin' role"
**Cause**: Non-admin user trying to create resources  
**Solution**: Use admin token or assign admin role to user

### Error: "Invalid request: database_name required"
**Cause**: Missing required field in request  
**Solution**: Include all required fields in JSON

### Error: "database already exists"
**Cause**: Database already exists  
**Solution**: Use different name or DROP existing database first

---

## 📚 Related Documentation

- [RBAC_SETUP_GUIDE.md](RBAC_SETUP_GUIDE.md) - Role configuration
- [RBAC_QUICK_REFERENCE.md](RBAC_QUICK_REFERENCE.md) - Token commands
- [API_GUIDE.md](API_GUIDE.md) - All API endpoints

---

## 🎯 Summary

| Feature | Status | Details |
|---------|--------|---------|
| Create Database | ✅ | MySQL, MariaDB, Percona, PostgreSQL, Oracle |
| Create Table | ✅ | All SQL databases |
| List Databases | ✅ | All SQL databases |
| List Tables | ✅ | All SQL databases |
| Admin Only | ✅ | Role-based access control |
| Error Handling | ✅ | Comprehensive error messages |
| Logging | ✅ | Query success/failure logged |

---

**Implementation Date**: January 22, 2026  
**Status**: ✅ Complete & Ready  
**Testing**: Use Test Workflow above  
**Production**: Ready to deploy! 🚀
