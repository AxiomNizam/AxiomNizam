# Authentication & Authorization Guide

Complete guide to AxiomNizam authentication, JWT tokens, and role-based access control (RBAC).

---

## Overview

AxiomNizam uses:
- **Keycloak** for authentication and user management
- **JWT Tokens** for stateless API authentication
- **Role-Based Access Control (RBAC)** for authorization
- **OpenID Connect** for modern OAuth 2.0 flows

---

## Authentication Flow

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

---

## Token Acquisition

### Method 1: Client Credentials (Service-to-Service)

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
  "token_type": "Bearer",
  "not-before-policy": 0,
  "scope": "profile email"
}
```

### Method 2: Resource Owner Password Grant (User Login)

**Best for**: End-user login, personal access

```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=password" \
  -d "username=admin_user" \
  -d "password=password123"
```

### Method 3: Authorization Code (Browser-based)

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

---

## Using Tokens in API Requests

### Standard Authorization Header

```bash
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Common Patterns

**JavaScript/Fetch:**
```javascript
fetch('http://localhost:8000/api/mysql/users', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
```

**PowerShell:**
```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers
```

**cURL:**
```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/mysql/users
```

---

## JWT Token Structure

A JWT token has 3 parts separated by dots: `header.payload.signature`

### Payload Example

```json
{
  "exp": 1642771200,
  "iat": 1642767600,
  "auth_time": 1642767594,
  "jti": "abc123...",
  "iss": "http://localhost:8080/realms/axiomnizam",
  "aud": "account",
  "sub": "abc123def456",
  "typ": "Bearer",
  "azp": "axiomnizam-backend",
  "session_state": "abc123...",
  "acr": "1",
  "allowed-origins": ["http://localhost:8000"],
  "realm_access": {
    "roles": ["admin"]
  },
  "resource_access": {
    "account": {
      "roles": ["manage-account"]
    }
  },
  "name": "Admin User",
  "preferred_username": "admin_user",
  "given_name": "Admin",
  "family_name": "User",
  "email": "admin@example.com"
}
```

### Key Fields

| Field | Meaning |
|-------|---------|
| `exp` | Token expiration time (Unix timestamp) |
| `iat` | Token issued at time |
| `sub` | Subject (user ID) |
| `preferred_username` | Username |
| `realm_access.roles` | User's roles (admin, user, etc.) |
| `email` | User's email |

---

## Role-Based Access Control (RBAC)

### Permission Matrix

| Operation | Admin | Non-Admin | Anonymous |
|-----------|-------|-----------|-----------|
| GET (Read) | ✅ | ✅ | ❌ |
| POST (Create) | ✅ | ❌ | ❌ |
| PUT (Update) | ✅ | ❌ | ❌ |
| DELETE (Delete) | ✅ | ❌ | ❌ |
| /health | ✅ | ✅ | ✅ |
| /status | ✅ | ✅ | ✅ |

### Checking Roles in Code

Backend (Go):
```go
// Middleware to require admin role
func RequireAdmin(c *gin.Context) {
    claims := c.MustGet("claims").(*auth.Claims)
    
    if !claims.HasRole("admin") {
        c.JSON(403, gin.H{"error": "Forbidden"})
        c.Abort()
        return
    }
    
    c.Next()
}

// In route
router.POST("/api/mysql/users", authMiddleware, RequireAdmin, createUser)
```

---

## Public Endpoints (No Auth Required)

```bash
# Health check
curl http://localhost:8000/health

# System status
curl http://localhost:8000/status

# Keycloak well-known config
curl http://localhost:8080/realms/axiomnizam/.well-known/openid-configuration
```

---

## Protected Endpoints (Auth Required)

### Admin Only

```bash
# Create database
curl -X POST http://localhost:8000/api/admin/database/create \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Create table
curl -X POST http://localhost:8000/api/admin/table/create \
  -H "Authorization: Bearer $ADMIN_TOKEN"

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

### Any Authenticated User

```bash
# Read operations (GET)
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $USER_TOKEN"

curl http://localhost:8000/api/mysql/users/1 \
  -H "Authorization: Bearer $USER_TOKEN"

curl http://localhost:8000/api/postgres/users \
  -H "Authorization: Bearer $USER_TOKEN"

# And for all other databases...
```

---

## Error Responses

### 401 Unauthorized (No/Invalid Token)

```json
{
  "error": "unauthorized",
  "message": "Invalid or missing token"
}
```

**Solutions:**
- Get a new token
- Verify token hasn't expired
- Check token format: `Bearer YOUR_TOKEN`

### 403 Forbidden (Insufficient Role)

```json
{
  "error": "forbidden",
  "message": "You don't have permission for this operation"
}
```

**Solutions:**
- Use admin token for write operations
- Check user role in Keycloak
- Assign proper role to user

### 400 Bad Request (Invalid Data)

```json
{
  "error": "bad_request",
  "message": "Invalid request body"
}
```

**Solutions:**
- Check JSON format
- Verify required fields
- Check content-type header

---

## Token Refresh

### Get Refresh Token

Include `scope=offline_access` in token request:

```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=client_credentials" \
  -d "scope=offline_access"
```

Response includes:
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 3600
}
```

### Use Refresh Token

```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend" \
  -d "client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN"
```

---

## Keycloak User Management

### Create User

```bash
# Access Keycloak admin console
# http://localhost:8080 → Admin Console
# Users → Create user
# Fill in: Username, Email, First/Last Name
# Set password
# Assign roles: admin or user
```

### Assign Roles

1. Go to Keycloak admin console
2. Users → Select user
3. Role Mappings
4. Select role → Add Selected

### Change Password

```bash
# Via admin console
Users → Select user → Credentials → Set Password

# Or via API
curl -X PUT http://localhost:8080/admin/realms/axiomnizam/users/{userId}/reset-password \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"password","value":"newpassword","temporary":false}'
```

---

## Security Best Practices

### Token Handling
- ✅ Store tokens securely (not in localStorage if sensitive)
- ✅ Use HTTPS in production
- ✅ Never commit tokens to version control
- ✅ Rotate tokens regularly
- ✅ Implement token expiration

### API Security
- ✅ Always validate token signature
- ✅ Check token expiration time
- ✅ Verify token issuer matches Keycloak
- ✅ Use HTTPS for API calls
- ✅ Implement rate limiting
- ✅ Log authentication attempts

### User Management
- ✅ Use strong passwords
- ✅ Implement MFA if possible
- ✅ Regular security audits
- ✅ Remove inactive users
- ✅ Monitor role assignments

---

## Troubleshooting

### Token Expired Error

```json
{
  "error": "unauthorized",
  "message": "Token expired"
}
```

**Solution:** Get a new token

```bash
# Get fresh token
TOKEN=$(curl -s -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72&grant_type=client_credentials" \
  | jq -r '.access_token')
```

### Invalid Signature Error

**Solutions:**
- Refresh the public key
- Verify realm name is correct
- Check PUBLIC_KEY in backend config

### Role Not Recognized

**Solutions:**
- Verify role exists in Keycloak
- Check role assignment to user
- Restart backend to refresh role cache

---

## Integration Examples

### Python

```python
import requests

# Get token
response = requests.post(
    'http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token',
    data={
        'client_id': 'axiomnizam-backend',
        'client_secret': 'YOUR_SECRET',
        'grant_type': 'client_credentials'
    }
)
token = response.json()['access_token']

# Use token
headers = {'Authorization': f'Bearer {token}'}
response = requests.get('http://localhost:8000/api/mysql/users', headers=headers)
```

### Node.js

```javascript
const axios = require('axios');

// Get token
const response = await axios.post(
  'http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token',
  new URLSearchParams({
    client_id: 'axiomnizam-backend',
    client_secret: 'YOUR_SECRET',
    grant_type: 'client_credentials'
  })
);
const token = response.data.access_token;

// Use token
const apiResponse = await axios.get(
  'http://localhost:8000/api/mysql/users',
  { headers: { Authorization: `Bearer ${token}` } }
);
```

---

## Learn More

- **Keycloak Documentation**: https://www.keycloak.org/documentation.html
- **OpenID Connect**: https://openid.net/connect/
- **JWT Info**: https://jwt.io/
- **OAuth 2.0**: https://oauth.net/2/

---

**Next: Read [API_REFERENCE.md](API_REFERENCE.md) for all available endpoints**
