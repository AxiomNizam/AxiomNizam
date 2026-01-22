# 🔐 KEYCLOAK CREDENTIALS - REFERENCE CARD

## Your Configuration Values

### Keycloak Server
```
URL:    http://localhost:8080
Realm:  axiomnizam
```

---

## Backend Client Configuration

### File: `.env`

```dotenv
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-backend
KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
KEYCLOAK_GRANT_TYPE=client_credentials
```

### Backend Uses:
```
Client ID:     axiomnizam-backend
Client Secret: 6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
Grant Type:    client_credentials (Server-to-Server)
Purpose:       Backend APIs authentication
```

---

## Frontend Client Configuration

### File: `frontend/.env`

```dotenv
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-frontend
KEYCLOAK_REDIRECT_URI=http://localhost:7000/callback
```

### Frontend Uses:
```
Client ID:       axiomnizam-frontend
Client Secret:   (none - public client)
Grant Type:      password (User Login)
Redirect URI:    http://localhost:7000/callback
Purpose:         User authentication
```

---

## Postman Configuration

### Variables to Set:

```json
{
  "realm": "axiomnizam",
  "keycloak_url": "http://localhost:8080",
  "backend_client_id": "axiomnizam-backend",
  "backend_client_secret": "6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72",
  "frontend_client_id": "axiomnizam-frontend",
  "username": "admin",
  "password": "admin",
  "base_url": "http://localhost:8000"
}
```

### Auto-Generated Variables:

```json
{
  "backend_token": "(filled after running Backend endpoint)",
  "user_token": "(filled after running User Login endpoint)",
  "refresh_token": "(filled after running User Login endpoint)",
  "token_expires_in": "(filled after running Backend endpoint)",
  "user_token_expires_in": "(filled after running User Login endpoint)"
}
```

---

## Quick Copy-Paste Values

### Backend .env
```
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-backend
KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
KEYCLOAK_GRANT_TYPE=client_credentials
```

### Frontend .env
```
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-frontend
KEYCLOAK_REDIRECT_URI=http://localhost:7000/callback
```

### Postman Test Credentials
```
Username: admin
Password: admin
Realm:    axiomnizam
```

---

## HTTP Endpoints

### Token Endpoint
```
POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token
```

### UserInfo Endpoint
```
GET http://localhost:8080/realms/axiomnizam/protocol/openid-connect/userinfo
```

### Logout Endpoint
```
POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/logout
```

---

## How Each Component Uses It

### Backend Code
```go
keycloakURL := os.Getenv("KEYCLOAK_URL")
realm := os.Getenv("KEYCLOAK_REALM")
clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
clientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")

// POST to: keycloakURL/realms/{realm}/protocol/openid-connect/token
// body: client_id={clientID}&client_secret={clientSecret}&grant_type=client_credentials
```

### Frontend Code
```javascript
const keycloakURL = process.env.REACT_APP_KEYCLOAK_URL;
const realm = process.env.REACT_APP_KEYCLOAK_REALM;
const clientID = process.env.REACT_APP_KEYCLOAK_CLIENT_ID;
const redirectUri = process.env.REACT_APP_KEYCLOAK_REDIRECT_URI;

// Login form → POST to keycloakURL/realms/{realm}/protocol/openid-connect/token
// body: client_id={clientID}&grant_type=password&username=...&password=...
```

### Postman Collection
```json
{
  "request": {
    "url": "{{keycloak_url}}/realms/{{realm}}/protocol/openid-connect/token",
    "body": {
      "client_id": "{{backend_client_id}}",
      "client_secret": "{{backend_client_secret}}",
      "grant_type": "client_credentials"
    }
  }
}
```

---

## Verification Checklist

- [ ] Backend .env has all 5 KEYCLOAK_* variables
- [ ] Frontend .env has all 4 KEYCLOAK_* variables
- [ ] Postman collection has "Authentication & Login" folder
- [ ] Postman has 5 endpoints (Get Token, Login, Check Session, Refresh, Logout)
- [ ] Postman variables include all realm/client IDs and secrets
- [ ] Secret value matches exactly: `6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72`
- [ ] Realm value matches exactly: `axiomnizam`
- [ ] All URLs use `http://localhost:8080`

---

## Files to Check

```
✅ .env                          - Backend configuration
✅ frontend/.env                 - Frontend configuration
✅ POSTMAN_COLLECTION.json      - Postman endpoints and variables
✅ KEYCLOAK_AUTH_SETUP.md       - Complete documentation
✅ AUTH_QUICK_REFERENCE.md      - Quick reference
✅ KEYCLOAK_ARCHITECTURE.md     - Architecture diagrams
✅ AUTH_CONFIGURATION_COMPLETE.md - Completion summary
```

---

## Test Sequence

### In Postman:

1. **Get Backend Token**
   ```
   Endpoint: Backend: Get Client Credentials Token
   Result: {{backend_token}} populated
   ```

2. **Login User**
   ```
   Endpoint: User: Login with Password
   Result: {{user_token}} and {{refresh_token}} populated
   ```

3. **Validate Login** ✅
   ```
   Endpoint: Check User Session (Validate Token)
   Result: User info displayed (proves you're logged in)
   ```

4. **Refresh Token**
   ```
   Endpoint: Refresh User Token
   Result: {{user_token}} refreshed with new value
   ```

5. **Logout**
   ```
   Endpoint: Logout User
   Result: Session invalidated, tokens revoked
   ```

---

**Reference Card Created**: January 22, 2026
**Status**: ✅ Ready to use
