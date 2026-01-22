# ✅ Keycloak Authentication Configuration - Complete Setup

## Summary
All authentication configurations have been set up for AxiomNizam with Keycloak integration across backend, frontend, and Postman collection.

---

## 1. Backend Configuration (.env)

### Location
`c:\Users\office\Documents\AxiomNizam\AxiomNizam\.env`

### Backend Keycloak Settings
```dotenv
# Keycloak Configuration - Backend
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-backend
KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
KEYCLOAK_GRANT_TYPE=client_credentials
```

### Backend Configuration Details
- **Realm**: `axiomnizam`
- **Client ID**: `axiomnizam-backend` (Server-to-Server authentication)
- **Client Secret**: `6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72`
- **Grant Type**: `client_credentials` (for backend API calls)
- **Use Case**: Backend service gets token using credentials for calling other services

---

## 2. Frontend Configuration (.env)

### Location
`c:\Users\office\Documents\AxiomNizam\AxiomNizam\frontend\.env`

### Frontend Keycloak Settings
```dotenv
# Frontend Environment Variables
FRONTEND_PORT=7000
BACKEND_PORT=8000
BACKEND_URL=http://localhost:8000

# Keycloak Configuration - Frontend
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-frontend
KEYCLOAK_REDIRECT_URI=http://localhost:7000/callback
```

### Frontend Configuration Details
- **Realm**: `axiomnizam`
- **Client ID**: `axiomnizam-frontend` (User-facing application)
- **Redirect URI**: `http://localhost:7000/callback` (OAuth2 redirect endpoint)
- **Use Case**: Users login through frontend, get redirected after authentication

---

## 3. Postman Collection Updates

### Location
`c:\Users\office\Documents\AxiomNizam\AxiomNizam\POSTMAN_COLLECTION.json`

### New Authentication Folder with 5 Endpoints

#### 1️⃣ Backend: Get Client Credentials Token
- **Method**: POST
- **Endpoint**: `{{keycloak_url}}/realms/{{realm}}/protocol/openid-connect/token`
- **Auth Type**: Client Credentials Flow
- **Parameters**:
  - `client_id`: `axiomnizam-backend`
  - `client_secret`: `6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72`
  - `grant_type`: `client_credentials`
- **Response**: Saves token to `backend_token` variable
- **Use Case**: Backend service authenticates with Keycloak

#### 2️⃣ User: Login with Password
- **Method**: POST
- **Endpoint**: `{{keycloak_url}}/realms/{{realm}}/protocol/openid-connect/token`
- **Auth Type**: Resource Owner Password Credentials Flow
- **Parameters**:
  - `client_id`: `axiomnizam-frontend`
  - `grant_type`: `password`
  - `username`: `{{username}}` (default: admin)
  - `password`: `{{password}}` (default: admin)
- **Response**: Saves tokens to `user_token`, `refresh_token` variables
- **Use Case**: User login via frontend or Postman

#### 3️⃣ Check User Session (Validate Token) ✅ LOGIN CHECK
- **Method**: GET
- **Endpoint**: `{{keycloak_url}}/realms/{{realm}}/protocol/openid-connect/userinfo`
- **Auth**: Bearer {{user_token}}
- **Response**: Returns user information (username, email, roles, etc.)
- **Use Case**: Verify user is logged in and get user details
- **Return Data**:
  ```json
  {
    "sub": "user-id",
    "email_verified": true,
    "name": "Full Name",
    "preferred_username": "username",
    "given_name": "First",
    "family_name": "Last",
    "email": "user@example.com"
  }
  ```

#### 4️⃣ Refresh User Token
- **Method**: POST
- **Endpoint**: `{{keycloak_url}}/realms/{{realm}}/protocol/openid-connect/token`
- **Parameters**:
  - `client_id`: `axiomnizam-frontend`
  - `grant_type`: `refresh_token`
  - `refresh_token`: `{{refresh_token}}`
- **Response**: Returns new access token
- **Use Case**: Get new token when current one expires

#### 5️⃣ Logout User
- **Method**: POST
- **Endpoint**: `{{keycloak_url}}/realms/{{realm}}/protocol/openid-connect/logout`
- **Parameters**:
  - `client_id`: `axiomnizam-frontend`
  - `refresh_token`: `{{refresh_token}}`
- **Use Case**: Invalidate user session

---

## 4. Postman Variables (Updated)

```json
{
  "base_url": "http://localhost:8000",
  "keycloak_url": "http://localhost:8080",
  "realm": "axiomnizam",
  "backend_client_id": "axiomnizam-backend",
  "backend_client_secret": "6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72",
  "frontend_client_id": "axiomnizam-frontend",
  "username": "admin",
  "password": "admin",
  "backend_token": "",
  "user_token": "",
  "refresh_token": "",
  "token": "",
  "token_expires_in": "",
  "user_token_expires_in": ""
}
```

---

## 5. How to Use in Postman

### Step 1: Get Backend Token
1. Open Postman
2. Go to: `Authentication & Login` → `Backend: Get Client Credentials Token`
3. Click **Send**
4. Token saved to `{{backend_token}}` variable

### Step 2: User Login
1. Go to: `Authentication & Login` → `User: Login with Password`
2. Click **Send**
3. Tokens saved to `{{user_token}}` and `{{refresh_token}}`

### Step 3: Check if User is Logged In ✅
1. Go to: `Authentication & Login` → `Check User Session (Validate Token)`
2. Click **Send**
3. Response shows current user info (confirms login status)

### Step 4: Use Token in API Requests
All API requests will use `{{user_token}}` via Bearer token in Authorization header

### Step 5: Refresh Token (When Expired)
1. Go to: `Authentication & Login` → `Refresh User Token`
2. Click **Send**
3. New token saved to `{{user_token}}`

### Step 6: Logout
1. Go to: `Authentication & Login` → `Logout User`
2. Click **Send**
3. Session invalidated

---

## 6. Auth Flow Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     AxiomNizam Auth Flow                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  BACKEND SERVICE                                                │
│  ├─ Client Credentials Flow                                    │
│  ├─ Gets token: axiomnizam-backend + secret                    │
│  └─ Uses token for service-to-service calls                    │
│                                                                 │
│  FRONTEND/USER AUTHENTICATION                                   │
│  ├─ Resource Owner Password Flow                               │
│  ├─ User enters username/password                              │
│  ├─ Gets access_token + refresh_token                          │
│  ├─ Access token used for API calls (expires in 5min)          │
│  ├─ Refresh token used to get new access token                 │
│  └─ Token validation via userinfo endpoint                     │
│                                                                 │
│  USER LOGIN CHECK ENDPOINTS                                    │
│  ├─ GET /userinfo - Validates token + returns user info        │
│  ├─ POST /token (with refresh) - Gets new token                │
│  └─ POST /logout - Invalidates session                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 7. Environment Variables Summary

### Backend (.env)
| Variable | Value |
|----------|-------|
| KEYCLOAK_URL | http://localhost:8080 |
| KEYCLOAK_REALM | axiomnizam |
| KEYCLOAK_CLIENT_ID | axiomnizam-backend |
| KEYCLOAK_CLIENT_SECRET | 6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72 |

### Frontend (.env)
| Variable | Value |
|----------|-------|
| KEYCLOAK_URL | http://localhost:8080 |
| KEYCLOAK_REALM | axiomnizam |
| KEYCLOAK_CLIENT_ID | axiomnizam-frontend |
| KEYCLOAK_REDIRECT_URI | http://localhost:7000/callback |

---

## 8. Next Steps

1. **Update Backend Code** to read from .env:
   - Use `os.Getenv()` to read Keycloak config
   - Implement middleware to validate tokens
   - Add role-based access control (RBAC)

2. **Update Frontend Code** to integrate Keycloak:
   - Add login form
   - Implement OAuth2 redirect flow
   - Store tokens in secure storage (httpOnly cookies)
   - Refresh token before expiry

3. **Test in Postman**:
   - Follow the 6-step guide above
   - Verify user login check endpoint works
   - Test API calls with token

---

## 9. Keycloak Credentials Reference

```
Realm: axiomnizam
Master User: admin
Master Password: admin (default)

Backend Client:
  - ID: axiomnizam-backend
  - Type: Confidential/Service Account
  - Secret: 6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72

Frontend Client:
  - ID: axiomnizam-frontend
  - Type: Public
  - Redirect URI: http://localhost:7000/callback
```

---

**Configuration Date**: January 22, 2026
**Status**: ✅ Complete - Ready for backend/frontend implementation
