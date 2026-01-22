# ⚡ RBAC Quick Start (5 Minutes)

**TL;DR**: Admin can do CRUD. Non-admin can only READ.

---

## 🚀 Fastest Path: Setup & Test

### Step 1: Assign Admin Role (2 minutes)

1. Go to: **http://localhost:8080**
2. Login: `admin` / `admin`
3. Left menu → **Realm Roles** → Create role
   - Name: `admin`
   - Save
4. Left menu → **Users** → Select **admin** user
5. **Role mapping** tab → Add **admin** role → Save

### Step 2: Create Non-Admin User (2 minutes)

1. **Users** → **Add user**
   - Username: `testuser`
   - Email: `testuser@example.com`
   - Enabled: ON
   - Create
2. **Credentials** tab → Set password: `password123` (Temporary: OFF)
3. **Role mapping** tab → Don't assign admin role → Save

### Step 3: Get Tokens & Test (1 minute)

**Get Admin Token:**
```powershell
$adminToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body @{client_id="axiomnizam";client_secret="uzqxRJUEI44gpURiytWtCujKwQ1ESZrv";grant_type="password";username="admin";password="admin"}).access_token
```

**Get User Token:**
```powershell
$userToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body @{client_id="axiomnizam";client_secret="uzqxRJUEI44gpURiytWtCujKwQ1ESZrv";grant_type="password";username="testuser";password="password123"}).access_token
```

---

## ✅ Test Results Expected

| Action | Admin | User |
|--------|-------|------|
| GET (Read) | ✅ 200 OK | ✅ 200 OK |
| POST (Create) | ✅ 201 Created | ❌ 403 Forbidden |
| PUT (Update) | ✅ 200 OK | ❌ 403 Forbidden |
| DELETE | ✅ 204 No Content | ❌ 403 Forbidden |

---

## 🧪 One-Line Tests

**Admin READ:**
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Headers @{"Authorization"="Bearer $adminToken"}
```

**Admin CREATE:**
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Method POST -Headers @{"Authorization"="Bearer $adminToken";"Content-Type"="application/json"} -Body '{"name":"Test","email":"test@test.com","age":25}'
```

**User READ (Works):**
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Headers @{"Authorization"="Bearer $userToken"}
```

**User CREATE (Fails - Expected):**
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Method POST -Headers @{"Authorization"="Bearer $userToken";"Content-Type"="application/json"} -Body '{"name":"Test","email":"test@test.com","age":25}'
```

---

## 📊 RBAC Architecture

```
┌─────────────────────────────────┐
│   Client Request with JWT       │
│  (GET/POST/PUT/DELETE + Token)  │
└────────────┬────────────────────┘
             ↓
   ┌────────────────────────┐
   │ Validate JWT Signature │
   │  (Keycloak Public Key) │
   └────────────┬───────────┘
                ↓
   ┌────────────────────────┐
   │  Extract Claims        │
   │  - username            │
   │  - realm_access.roles  │
   └────────────┬───────────┘
                ↓
        ┌───────────────┐
        │ Is GET? (Read)│
        └───────────────┘
         │              │
        YES             NO
         │              │
         ↓              ↓
    ✅ ALLOW        Check Role
    (All users)       ↓
                 Has admin?
                  │      │
                 YES     NO
                  │      │
                  ↓      ↓
              ✅ ALLOW  ❌ 403 Forbidden
            (admin)     (non-admin)
```

---

## 🔍 View Token Claims

```powershell
function Decode-JWT {
    param([string]$Token)
    
    $parts = $Token.Split('.')
    $payload = $parts[1] + '=' * (4 - ($parts[1].Length % 4))
    [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($payload)) | ConvertFrom-Json
}

# Check admin roles
Decode-JWT $adminToken | Select preferred_username, realm_access

# Check user roles
Decode-JWT $userToken | Select preferred_username, realm_access
```

---

## 🎯 For All 7 Databases

Apply same role-based access to all databases:

```powershell
$databases = @("mysql", "mariadb", "postgres", "percona", "mongodb", "firebase", "oracle")

foreach ($db in $databases) {
    Write-Host "`n📊 $db" -ForegroundColor Cyan
    
    # Admin can READ
    Write-Host "  [Admin READ]" -ForegroundColor Green
    Invoke-RestMethod "http://localhost:8000/api/$db/users" -Headers @{"Authorization"="Bearer $adminToken"} | Measure-Object | Select-Object Count
    
    # Admin can CREATE
    Write-Host "  [Admin CREATE]" -ForegroundColor Green
    Invoke-RestMethod "http://localhost:8000/api/$db/users" -Method POST -Headers @{"Authorization"="Bearer $adminToken";"Content-Type"="application/json"} -Body '{"name":"Test","email":"test@test.com","age":25}' | Select-Object id
    
    # User can READ
    Write-Host "  [User READ]" -ForegroundColor Cyan
    Invoke-RestMethod "http://localhost:8000/api/$db/users" -Headers @{"Authorization"="Bearer $userToken"} | Measure-Object | Select-Object Count
    
    # User CANNOT CREATE
    Write-Host "  [User CREATE - Should Fail]" -ForegroundColor Yellow
    try {
        Invoke-RestMethod "http://localhost:8000/api/$db/users" -Method POST -Headers @{"Authorization"="Bearer $userToken";"Content-Type"="application/json"} -Body '{"name":"Test","email":"test@test.com","age":25}'
        Write-Host "    ERROR: Should have been forbidden!" -ForegroundColor Red
    } catch {
        Write-Host "    ✅ Correctly forbidden (403)" -ForegroundColor Green
    }
}
```

---

## 🎛️ Code Changes Summary

### What's Protected Now?

| Endpoint Type | Method | Who Can Access |
|---------------|--------|-----------------|
| Health/Status | GET | Everyone (no token needed) |
| Read Users | GET | All authenticated users |
| Create User | POST | Admin role only |
| Update User | PUT | Admin role only |
| Delete User | DELETE | Admin role only |

### Where Roles Come From

1. **Keycloak stores roles** for each user
2. **JWT token includes roles** when user authenticates
3. **Backend checks roles** in middleware
4. **Allows/denies** based on role

### Key Code Changes

**auth.go**: Added role support
```go
type Claims struct {
    RealmAccess RealmAccess `json:"realm_access"`
}

func (c *Claims) HasRole(role string) bool
```

**middleware.go**: Added role checking
```go
RequireAdmin()           // Check for admin role
RequireRole(role string) // Check for any role
```

**main.go**: Applied role checks
```go
router.GET("/api/mysql/users", authMiddleware, ...)        // All users
router.POST("/api/mysql/users", adminMiddleware, ...)      // Admin only
```

---

## 🧐 Error Examples

**Successful Admin CREATE:**
```json
{
  "id": 123,
  "name": "Test",
  "email": "test@test.com",
  "age": 25,
  "created_at": "2026-01-22T10:30:00Z"
}
```

**Failed User CREATE (403):**
```json
{
  "error": "forbidden: user does not have 'admin' role",
  "user_roles": [],
  "required": "admin"
}
```

---

## ⚡ Instant Commands

Copy-paste these directly into PowerShell:

```powershell
# 1. Get admin token
$adminToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body @{client_id="axiomnizam";client_secret="uzqxRJUEI44gpURiytWtCujKwQ1ESZrv";grant_type="password";username="admin";password="admin"}).access_token; Write-Host "✅ Admin Token: $($adminToken.Substring(0,30))..."

# 2. Get user token
$userToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body @{client_id="axiomnizam";client_secret="uzqxRJUEI44gpURiytWtCujKwQ1ESZrv";grant_type="password";username="testuser";password="password123"}).access_token; Write-Host "✅ User Token: $($userToken.Substring(0,30))..."

# 3. Admin READ (✅ Works)
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Headers @{"Authorization"="Bearer $adminToken"}

# 4. User READ (✅ Works)
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Headers @{"Authorization"="Bearer $userToken"}

# 5. Admin CREATE (✅ Works)
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Method POST -Headers @{"Authorization"="Bearer $adminToken";"Content-Type"="application/json"} -Body (@{name="Admin Test";email="admintest@test.com";age=30} | ConvertTo-Json)

# 6. User CREATE (❌ Fails - 403 Forbidden) 
Invoke-RestMethod "http://localhost:8000/api/mysql/users" -Method POST -Headers @{"Authorization"="Bearer $userToken";"Content-Type"="application/json"} -Body (@{name="User Test";email="usertest@test.com";age=25} | ConvertTo-Json)
```

---

## 🎓 Learning Path

1. **5 min**: Setup admin/user roles in Keycloak
2. **5 min**: Create test users
3. **5 min**: Get tokens
4. **5 min**: Test with curl/Postman
5. **✨ 20 minutes total to full RBAC understanding!**

---

## 📚 Next Steps

- [Full RBAC Setup Guide](RBAC_SETUP_GUIDE.md) - Complete details
- [Keycloak Setup](KEYCLOAK_SETUP_GUIDE.md) - Advanced Keycloak config
- [API Guide](API_GUIDE.md) - All endpoints
- [Postman Collection](POSTMAN_COLLECTION.json) - Ready-to-use requests

---

**Ready to test? Run the instant commands above! 🚀**
