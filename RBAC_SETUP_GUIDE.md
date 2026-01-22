# 🔐 Role-Based Access Control (RBAC) Setup Guide

**Last Updated**: January 22, 2026  
**Status**: ✅ Ready for Configuration

---

## 📋 Overview

Your AxiomNizam API now implements Role-Based Access Control (RBAC) using Keycloak:

| Role | Permissions |
|------|------------|
| **admin** | ✅ CREATE (POST), READ (GET), UPDATE (PUT), DELETE (DELETE) |
| **user** / no role | ✅ READ (GET) only |

---

## 🎯 Quick Summary: What Changed

### Code Changes Made

1. **`internal/auth/auth.go`** - Added role support to JWT claims
   ```go
   type RealmAccess struct {
       Roles []string `json:"roles"`
   }
   
   type Claims struct {
       // ... existing fields ...
       RealmAccess RealmAccess `json:"realm_access"`
   }
   
   // HasRole method checks if user has specific role
   func (c *Claims) HasRole(role string) bool { ... }
   ```

2. **`internal/auth/middleware.go`** - Added role-based middleware
   ```go
   RequireRole(role string)  // Check for specific role
   RequireAdmin()            // Check for admin role specifically
   ```

3. **`main.go`** - Routes now use RBAC
   ```go
   // Read endpoints (all authenticated users)
   router.GET("/api/mysql/users", authMiddleware, userHandler.GetAllUsers)
   
   // Write endpoints (admin only)
   router.POST("/api/mysql/users", adminMiddleware, userHandler.CreateUser)
   router.PUT("/api/mysql/users/:id", adminMiddleware, userHandler.UpdateUser)
   router.DELETE("/api/mysql/users/:id", adminMiddleware, userHandler.DeleteUser)
   ```

---

## ⚙️ Step 1: Access Keycloak Admin Console

1. Open browser and go to: **http://localhost:8080**
2. Click "Administration Console"
3. Login with:
   - **Username**: `admin`
   - **Password**: `admin`

---

## 🛠️ Step 2: Create Roles in Master Realm

### Add 'admin' Role

1. In Keycloak Admin Console:
   - Left sidebar → **Realm Roles**
   - Click **Create role**
   
2. Fill in:
   - **Role name**: `admin`
   - **Display name**: `Administrator`
   - **Description**: `Admin users can perform all CRUD operations`
   - Click **Save**

3. Verify the role appears in the Realm Roles list

### Add 'user' Role (Optional)

1. **Create role**
   - **Role name**: `user`
   - **Display name**: `Regular User`
   - **Description**: `Regular users can only read data`
   - Click **Save**

---

## 👥 Step 3: Assign Roles to Users

### Assign Admin Role to Test User

1. Left sidebar → **Users**
2. Click on **admin** user (or create a new test user)
3. Go to **Role mapping** tab
4. Under "Realm roles":
   - Find **admin** in Available Roles
   - Click **Add selected**
5. Verify **admin** appears in Assigned Roles
6. Click **Save**

### Create Non-Admin Test User

1. **Users** → **Add user**
2. Fill in:
   - **Username**: `testuser`
   - **Email**: `testuser@example.com`
   - **First Name**: `Test`
   - **Last Name**: `User`
   - **Enabled**: ON
   - Click **Create**

3. Set password:
   - **Credentials** tab
   - **Set password**: `password123`
   - **Temporary**: OFF
   - Click **Set password**

4. Assign role:
   - Go to **Role mapping** tab
   - Add **user** role (or leave without admin role)
   - Click **Save**

---

## 🧪 Step 4: Get Tokens for Both Users

### Admin User Token

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

Write-Host "✅ Admin Token: $($adminToken.Substring(0, 50))..."
```

### Non-Admin User Token

```powershell
$userToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "password"
    username      = "testuser"
    password      = "password123"
  }).access_token

Write-Host "✅ User Token: $($userToken.Substring(0, 50))..."
```

---

## 🔍 Step 5: Verify Token Claims (Optional)

Decode the JWT token to see roles:

```powershell
# Install if needed: Install-Module -Name JwtTokens -Force

function Decode-JwtToken {
    param($Token)
    
    $parts = $Token.Split('.')
    $payload = $parts[1]
    
    # Add padding if needed
    $padding = 4 - ($payload.Length % 4)
    if ($padding -ne 4) {
        $payload += '=' * $padding
    }
    
    $decoded = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($payload))
    $decoded | ConvertFrom-Json
}

# Check admin token
Write-Host "=== Admin Token Claims ===" -ForegroundColor Green
Decode-JwtToken $adminToken | Select-Object preferred_username, realm_access

# Check user token  
Write-Host "`n=== User Token Claims ===" -ForegroundColor Cyan
Decode-JwtToken $userToken | Select-Object preferred_username, realm_access
```

**Expected Output**:
```
=== Admin Token Claims ===
preferred_username realm_access
------------------ -----------
admin              @{roles=System.Object[]}  # Contains "admin"

=== User Token Claims ===
preferred_username realm_access
------------------ -----------
testuser           @{roles=System.Object[]}  # May be empty or contain "user"
```

---

## ✅ Step 6: Test RBAC Implementation

### Test 1: Admin Can Read ✅

```powershell
$headers = @{ "Authorization" = "Bearer $adminToken" }
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers
```

**Expected**: Returns list of users ✅

### Test 2: Admin Can Create ✅

```powershell
$headers = @{ "Authorization" = "Bearer $adminToken" }
$body = @{
    name = "New User"
    email = "newuser@example.com"
    age = 25
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Headers $headers `
  -ContentType "application/json" `
  -Body $body
```

**Expected**: User created successfully ✅

### Test 3: User Can Read ✅

```powershell
$headers = @{ "Authorization" = "Bearer $userToken" }
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers
```

**Expected**: Returns list of users ✅

### Test 4: User Cannot Create ❌

```powershell
$headers = @{ "Authorization" = "Bearer $userToken" }
$body = @{
    name = "New User"
    email = "newuser@example.com"
    age = 25
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Headers $headers `
  -ContentType "application/json" `
  -Body $body
```

**Expected Response** (403 Forbidden):
```json
{
  "error": "forbidden: user does not have 'admin' role",
  "user_roles": [],
  "required": "admin"
}
```

### Test 5: User Cannot Update ❌

```powershell
$headers = @{ "Authorization" = "Bearer $userToken" }
$body = @{
    name = "Updated User"
    email = "updated@example.com"
    age = 30
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users/1" `
  -Method PUT `
  -Headers $headers `
  -ContentType "application/json" `
  -Body $body
```

**Expected**: 403 Forbidden error ❌

### Test 6: User Cannot Delete ❌

```powershell
$headers = @{ "Authorization" = "Bearer $userToken" }

Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users/1" `
  -Method DELETE `
  -Headers $headers
```

**Expected**: 403 Forbidden error ❌

---

## 📊 Complete Test Workflow

Run all tests together:

```powershell
function Test-RBAC {
    param(
        [string]$AdminToken,
        [string]$UserToken
    )
    
    $apiUrl = "http://localhost:8000/api/mysql/users"
    
    Write-Host "`n🔐 RBAC Testing" -ForegroundColor Yellow
    Write-Host "=" * 50
    
    # Test 1: Admin Read
    Write-Host "`n[1/6] Admin - READ (GET)" -ForegroundColor Green
    try {
        $response = Invoke-RestMethod -Uri $apiUrl `
          -Headers @{"Authorization" = "Bearer $AdminToken"}
        Write-Host "✅ Success: Retrieved $(($response | Measure-Object).Count) users"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 2: Admin Create
    Write-Host "`n[2/6] Admin - CREATE (POST)" -ForegroundColor Green
    try {
        $body = @{name="TestAdmin"; email="admin@test.com"; age=30} | ConvertTo-Json
        $response = Invoke-RestMethod -Uri $apiUrl `
          -Method POST `
          -Headers @{"Authorization" = "Bearer $AdminToken"; "Content-Type" = "application/json"} `
          -Body $body
        Write-Host "✅ Success: Created user ID $($response.id)"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 3: User Read
    Write-Host "`n[3/6] User - READ (GET)" -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri $apiUrl `
          -Headers @{"Authorization" = "Bearer $UserToken"}
        Write-Host "✅ Success: Retrieved $(($response | Measure-Object).Count) users"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 4: User Create (Should Fail)
    Write-Host "`n[4/6] User - CREATE (POST) - Expected to FAIL" -ForegroundColor Cyan
    try {
        $body = @{name="TestUser"; email="user@test.com"; age=25} | ConvertTo-Json
        $response = Invoke-RestMethod -Uri $apiUrl `
          -Method POST `
          -Headers @{"Authorization" = "Bearer $UserToken"; "Content-Type" = "application/json"} `
          -Body $body
        Write-Host "❌ ERROR: Should have been forbidden!"
    } catch {
        if ($_ -match "403") {
            Write-Host "✅ Correctly forbidden: $($_.Exception.Response.StatusCode)"
        } else {
            Write-Host "❓ Unexpected error: $($_.Exception.Message)"
        }
    }
    
    # Test 5: User Update (Should Fail)
    Write-Host "`n[5/6] User - UPDATE (PUT) - Expected to FAIL" -ForegroundColor Cyan
    try {
        $body = @{name="Updated"; email="updated@test.com"; age=40} | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$apiUrl/1" `
          -Method PUT `
          -Headers @{"Authorization" = "Bearer $UserToken"; "Content-Type" = "application/json"} `
          -Body $body
        Write-Host "❌ ERROR: Should have been forbidden!"
    } catch {
        if ($_ -match "403") {
            Write-Host "✅ Correctly forbidden: $($_.Exception.Response.StatusCode)"
        } else {
            Write-Host "❓ Unexpected error: $($_.Exception.Message)"
        }
    }
    
    # Test 6: User Delete (Should Fail)
    Write-Host "`n[6/6] User - DELETE (DELETE) - Expected to FAIL" -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "$apiUrl/1" `
          -Method DELETE `
          -Headers @{"Authorization" = "Bearer $UserToken"}
        Write-Host "❌ ERROR: Should have been forbidden!"
    } catch {
        if ($_ -match "403") {
            Write-Host "✅ Correctly forbidden: $($_.Exception.Response.StatusCode)"
        } else {
            Write-Host "❓ Unexpected error: $($_.Exception.Message)"
        }
    }
    
    Write-Host "`n" + "=" * 50
    Write-Host "🎉 RBAC Testing Complete!" -ForegroundColor Yellow
}

# Run tests
Test-RBAC -AdminToken $adminToken -UserToken $userToken
```

---

## 🔄 How RBAC Works

### JWT Token Contains Roles

When user logs in, Keycloak includes roles in JWT token:

```json
{
  "sub": "user-id",
  "preferred_username": "admin",
  "email": "admin@example.com",
  "realm_access": {
    "roles": ["admin", "user"]  // ← Roles here
  },
  "exp": 1674329400,
  "iat": 1674329100
}
```

### Backend Validates Roles

1. Backend receives request with token
2. Validates JWT signature with Keycloak public key
3. Extracts `realm_access.roles` from claims
4. Checks if user has required role:
   - **GET requests**: Any authenticated user ✅
   - **POST/PUT/DELETE requests**: Must have "admin" role ✅

### Response on Authorization Failure

```json
{
  "error": "forbidden: user does not have 'admin' role",
  "user_roles": ["user"],
  "required": "admin"
}
```

---

## 🎛️ Middleware Flow

```
Request with Token
    ↓
[Middleware] Extract Bearer Token
    ↓
[auth.Middleware] Validate JWT signature
    ↓
[auth.Middleware] Store claims in context
    ↓
For POST/PUT/DELETE:
    ↓
[RequireAdmin()] Check if "admin" in roles
    ↓
If admin role exists → Continue to handler ✅
If admin role missing → Return 403 Forbidden ❌
    ↓
If GET request → Skip role check, allow all authenticated users ✅
    ↓
Handler processes request
    ↓
Return response
```

---

## 🚀 Using with Postman

### Setup Postman Collection for RBAC

1. Import POSTMAN_COLLECTION.json
2. Update environment variables:
   ```
   client_id: axiomnizam
   client_secret: uzqxRJUEI44gpURiytWtCujKwQ1ESZrv
   username: admin (or testuser)
   password: admin (or password123)
   base_url: http://localhost:8000
   ```

3. Collection includes:
   - "Get Admin Token" - for admin user
   - "Get User Token" - for non-admin user
   - All 35 CRUD endpoints with proper auth headers

4. Run requests:
   - First run "Get Admin Token"
   - Then run any CRUD endpoint
   - Token automatically added to Authorization header

---

## 🔒 Security Best Practices

1. **Strong Passwords**
   - Don't use "admin/admin" in production
   - Use strong passwords for all users

2. **Token Expiration**
   - Current: 300 seconds (5 minutes)
   - Increase for production as needed
   - Use refresh tokens for extended sessions

3. **HTTPS in Production**
   - Use HTTPS instead of HTTP
   - Protects tokens during transmission

4. **Rotate Credentials**
   - Periodically rotate client secrets
   - Update users' passwords regularly

5. **Audit Logging**
   - Enable Keycloak audit logs
   - Monitor unauthorized access attempts

6. **Role Hierarchy** (Optional)
   - Create more granular roles if needed
   - Example: admin, moderator, user, viewer
   - Assign multiple roles to users

---

## 🐛 Troubleshooting

### Issue: "Token has expired"
**Solution**: Get a new token (expires in 5 minutes)

### Issue: "user does not have 'admin' role"
**Solution**: Verify user has admin role assigned in Keycloak:
1. Users → Select user → Role mapping → Check for "admin" role

### Issue: "Keycloak initialization failed"
**Solution**: Ensure Keycloak is running on port 8080

### Issue: Token endpoint returns 401
**Solution**: Check credentials:
- Client ID: `axiomnizam`
- Client Secret: `uzqxRJUEI44gpURiytWtCujKwQ1ESZrv`
- Realm: `master`

### Issue: No roles in token
**Solution**: 
1. Check user has roles assigned in Keycloak
2. Verify role assignment was saved
3. Get a new token after assigning role

---

## 📚 Related Documentation

- [KEYCLOAK_SETUP_GUIDE.md](KEYCLOAK_SETUP_GUIDE.md) - Keycloak authentication setup
- [KEYCLOAK_CREDENTIALS_INTEGRATION.md](KEYCLOAK_CREDENTIALS_INTEGRATION.md) - Token acquisition
- [API_GUIDE.md](API_GUIDE.md) - All API endpoints
- [POSTMAN_API_GUIDE.md](POSTMAN_API_GUIDE.md) - Postman testing guide

---

## 🎯 What's Next?

1. ✅ Setup admin and user roles in Keycloak
2. ✅ Assign roles to test users
3. ✅ Get tokens for both admin and user
4. ✅ Test all RBAC scenarios
5. ✅ Test with all 7 databases
6. 🚀 Deploy to production with proper security

---

**Configuration Complete! Ready for RBAC Testing! 🔐**
