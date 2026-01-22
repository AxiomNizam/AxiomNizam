# Keycloak Configuration Guide for AxiomNizam

**Client Credentials**:
- Client ID: `axiomnizam`
- Client Secret: `uzqxRJUEI44gpURiytWtCujKwQ1ESZrv`
- Realm: `master`

---

## 🔐 Current Configuration

Your Keycloak is now configured with:

```
Realm:          master
Client ID:      axiomnizam
Client Secret:  uzqxRJUEI44gpURiytWtCujKwQ1ESZrv
Type:           Confidential Client (has secret)
Database:       PostgreSQL (keycloak database)
```

This is stored in `.env`:
```dotenv
KEYCLOAK_HOST=keycloak
KEYCLOAK_PORT=8080
KEYCLOAK_REALM=master
KEYCLOAK_CLIENT_ID=axiomnizam
KEYCLOAK_CLIENT_SECRET=uzqxRJUEI44gpURiytWtCujKwQ1ESZrv
```

---

## 🚀 Two Token Acquisition Methods

### Method 1: Resource Owner Password Grant (For Users)

**When to use**: User login with username/password

**Request**:
```bash
POST http://localhost:8080/realms/master/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

client_id=axiomnizam
client_secret=uzqxRJUEI44gpURiytWtCujKwQ1ESZrv
grant_type=password
username=admin
password=admin
```

**PowerShell**:
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "password"
    username      = "admin"
    password      = "admin"
  } | ConvertTo-Json

$token = $response.access_token
Write-Host "Token: $token"
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjEifQ...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjEifQ...",
  "token_type": "Bearer",
  "not-before-policy": 0,
  "session_state": "abc123...",
  "scope": "openid profile email"
}
```

---

### Method 2: Client Credentials Grant (For Services)

**When to use**: Service-to-service authentication (no user involved)

**Request**:
```bash
POST http://localhost:8080/realms/master/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

client_id=axiomnizam
client_secret=uzqxRJUEI44gpURiytWtCujKwQ1ESZrv
grant_type=client_credentials
```

**PowerShell**:
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "client_credentials"
  }

$token = $response.access_token
Write-Host "Token: $token"
```

---

## 📖 Using Token in API Requests

Once you have the token, use it in all CRUD requests:

```powershell
# Get token
$token = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "password"
    username      = "admin"
    password      = "admin"
  }).access_token

# Create headers with token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type"  = "application/json"
}

# Create user in MySQL
$user = @{
    name  = "John Doe"
    email = "john@example.com"
    age   = 30
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Headers $headers `
  -Body $user
```

---

## 🔑 Token Validation in Backend

Your backend automatically validates tokens using:

1. **Get Public Keys** from Keycloak:
   ```
   GET http://keycloak:8080/realms/master/protocol/openid-connect/certs
   ```

2. **Validate Signature** using RSA public key

3. **Extract Claims**:
   ```json
   {
     "sub": "user-id",
     "preferred_username": "admin",
     "email": "admin@keycloak.local",
     "name": "Admin User",
     "exp": 1234567890,
     "iat": 1234567800
   }
   ```

4. **Store in Context** for use in handlers

---

## 🔄 Token Refresh

Tokens expire in 300 seconds. Use refresh_token to get new access_token:

```powershell
$refreshResponse = Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "refresh_token"
    refresh_token = $refreshToken
  }

$newAccessToken = $refreshResponse.access_token
```

---

## 🔧 Keycloak Admin Console

### Access Admin Console
```
URL: http://localhost:8080/admin
Username: admin
Password: admin
Realm: master
```

### Verify Client Configuration
1. Open http://localhost:8080/admin
2. Login with admin/admin
3. Go to Clients → axiomnizam
4. You should see:
   - Client ID: axiomnizam
   - Client Authentication: ON (Confidential)
   - Client Secret: uzqxRJUEI44gpURiytWtCujKwQ1ESZrv

### Create Additional Users (Optional)
1. Go to Users
2. Click "Create new user"
3. Fill in Username, Email, etc.
4. Click "Create"
5. Go to Credentials tab
6. Set password
7. Use this user for testing

---

## 🧪 Test Token Acquisition

### Quick PowerShell Test

```powershell
# Test with Resource Owner Password Grant
$token_response = Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "password"
    username      = "admin"
    password      = "admin"
  }

Write-Host "Access Token:" $token_response.access_token
Write-Host "Expires In:" $token_response.expires_in
Write-Host "Token Type:" $token_response.token_type

# Save token for API requests
$token = $token_response.access_token
```

### Test API Request with Token

```powershell
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type"  = "application/json"
}

# Test public endpoint (should work without token)
Invoke-RestMethod -Uri "http://localhost:8000/health" | ConvertTo-Json

# Test protected endpoint (requires token)
Invoke-RestMethod -Uri "http://localhost:8000/api/mysql/users" -Headers $headers | ConvertTo-Json
```

---

## 📋 Common Endpoints

### Token Acquisition
```
POST /realms/master/protocol/openid-connect/token
```

### Public Key Set (JWKS)
```
GET /realms/master/protocol/openid-connect/certs
```

### OpenID Configuration
```
GET /realms/master/.well-known/openid-configuration
```

### User Info
```
GET /realms/master/protocol/openid-connect/userinfo
Header: Authorization: Bearer <token>
```

### Logout
```
POST /realms/master/protocol/openid-connect/logout
Body: client_id=axiomnizam, refresh_token=<token>
```

---

## 🔒 Security Best Practices

1. **Never Hardcode Secrets**
   - Use environment variables (.env file)
   - ✅ You're doing this correctly

2. **Use HTTPS in Production**
   - Current: HTTP (dev only)
   - Production: Use HTTPS with SSL cert
   - Update keycloak_url to https://...

3. **Rotate Client Secrets**
   - Change secret every 90 days
   - Go to Keycloak Admin → Clients → axiomnizam → Credentials → Regenerate

4. **Token Expiration**
   - Current: 300 seconds (5 minutes)
   - For production: Consider 1-2 hours
   - Configure in Keycloak Admin Console

5. **Refresh Token Rotation**
   - Enable token rotation in Keycloak
   - Reduces token reuse vulnerability

---

## 🚀 Postman Setup with Client Secret

### Create New Environment Variable

```json
{
  "client_id": "axiomnizam",
  "client_secret": "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv",
  "username": "admin",
  "password": "admin"
}
```

### Update Token Request in Postman

**Body** (form-urlencoded):
```
client_id: {{client_id}}
client_secret: {{client_secret}}
grant_type: password
username: {{username}}
password: {{password}}
```

### Auto-Save Token Script

Add to **Tests** tab:
```javascript
var jsonData = pm.response.json();
pm.environment.set("token", jsonData.access_token);
console.log("Token saved: " + jsonData.access_token.substring(0, 20) + "...");
console.log("Expires in: " + jsonData.expires_in + " seconds");
```

---

## 🔍 Troubleshooting

### Error: "Invalid client credentials"
**Cause**: Wrong client_id or client_secret  
**Solution**: Verify credentials in .env match Keycloak

### Error: "invalid_grant"
**Cause**: Wrong username/password or client not configured  
**Solution**: Check user exists in Keycloak

### Error: "401 Unauthorized"
**Cause**: Token missing or expired  
**Solution**: Get new token using above methods

### Token Not Working in API
**Check**:
1. Token format: `Authorization: Bearer <token>`
2. Token not expired (expires_in)
3. Backend logs for validation errors
4. Keycloak JWKS endpoint accessible

---

## 📊 Token Claims Structure

Tokens issued contain these claims:

```json
{
  "sub": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "preferred_username": "admin",
  "email": "admin@keycloak.local",
  "name": "Admin User",
  "exp": 1642857600,
  "iat": 1642857300,
  "auth_time": 1642857300,
  "jti": "1234567890abcdef",
  "typ": "Bearer",
  "azp": "axiomnizam",
  "allowed-origins": ["http://localhost:8000"],
  "resource_access": {
    "axiomnizam": {
      "roles": ["user", "admin"]
    }
  }
}
```

**Use in Backend**:
- `sub`: User ID
- `preferred_username`: Username
- `email`: User email
- `exp`: Expiration timestamp
- Access via `c.Get("user")` in handlers

---

## ✅ Configuration Checklist

- [x] Client ID: axiomnizam
- [x] Client Secret: uzqxRJUEI44gpURiytWtCujKwQ1ESZrv
- [x] Added to .env file
- [x] Backend config updated
- [x] Token endpoint: /realms/master/protocol/openid-connect/token
- [x] Grant type: password (Resource Owner)
- [x] Default credentials: admin/admin
- [x] Realm: master
- [x] Database: PostgreSQL

---

## 🎯 Quick Reference Commands

```powershell
# Get token (Resource Owner)
$t=(Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body "client_id=axiomnizam&client_secret=uzqxRJUEI44gpURiytWtCujKwQ1ESZrv&grant_type=password&username=admin&password=admin").access_token

# Test with token
curl -H "Authorization: Bearer $t" http://localhost:8000/api/mysql/users

# Get token (Client Credentials)
$t=(Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body "client_id=axiomnizam&client_secret=uzqxRJUEI44gpURiytWtCujKwQ1ESZrv&grant_type=client_credentials").access_token
```

---

## 📚 Additional Resources

- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [OAuth 2.0 Grant Types](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-grant-types)
- [JWT Introduction](https://jwt.io/introduction)

---

**Ready to use your Keycloak credentials!**
✅ Configuration complete with client_id and client_secret
✅ Two token acquisition methods available
✅ Backend validates tokens automatically
✅ Postman collection supports client secret

