# 🏗️ AxiomNizam Authentication Architecture

## Complete Authentication Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      KEYCLOAK REALM: axiomnizam                             │
│                         http://localhost:8080                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────────────────┐      ┌──────────────────────────────┐   │
│  │  PUBLIC CLIENT               │      │  CONFIDENTIAL CLIENT         │   │
│  │  axiomnizam-frontend         │      │  axiomnizam-backend          │   │
│  │                              │      │                              │   │
│  │  Type: Web Application       │      │  Type: Service Account       │   │
│  │  Grant: Password             │      │  Grant: Client Credentials   │   │
│  │  Redirect: localhost:7000    │      │  Secret: [CONFIGURED]        │   │
│  └──────────────────────────────┘      └──────────────────────────────┘   │
│           ▲                                      ▲                         │
│           │                                      │                         │
└───────────┼──────────────────────────────────────┼─────────────────────────┘
            │                                      │
            │ /token (password grant)              │ /token (client credentials)
            │ /userinfo (validate token)           │
            │ /logout (end session)                │
            │                                      │
    ┌───────▼────────┐                    ┌───────▼────────┐
    │  FRONTEND      │                    │   BACKEND      │
    │  PORT: 7000    │                    │   PORT: 8000   │
    │                │                    │                │
    │ - Login Form   │                    │ - API Handler  │
    │ - Dashboard    │                    │ - Database     │
    │ - User Token   │                    │ - Service Auth │
    │ - Refresh Tok  │                    │ - RBAC         │
    └────────────────┘                    └────────────────┘
             │                                      │
             │ GET /api/users (with token)         │
             │ POST /api/users (with token)        │
             └──────────────────────────────────────┘

```

---

## User Authentication Flow (Step by Step)

```
USER
  │
  ├─ 1. Opens http://localhost:7000
  │  │
  │  └─ Sees Login Form
  │
  ├─ 2. Enters username (admin) & password (admin)
  │  │
  │  └─ Clicks "Login"
  │
  ├─ 3. Frontend sends to Keycloak:
  │  │  POST /realms/axiomnizam/protocol/openid-connect/token
  │  │  client_id: axiomnizam-frontend
  │  │  grant_type: password
  │  │  username: admin
  │  │  password: admin
  │  │
  │  ├─ Keycloak validates credentials ✓
  │  │
  │  └─ Returns:
  │     {
  │       "access_token": "eyJhbGc...",
  │       "refresh_token": "eyJhbGc...",
  │       "expires_in": 300,
  │       "token_type": "Bearer"
  │     }
  │
  ├─ 4. Frontend stores tokens:
  │  │  user_token = access_token
  │  │  refresh_token = refresh_token
  │  │
  │  └─ Redirects to Dashboard
  │
  ├─ 5. Frontend validates session:
  │  │  GET /realms/axiomnizam/protocol/openid-connect/userinfo
  │  │  Header: Authorization: Bearer {access_token}
  │  │
  │  ├─ Keycloak validates token ✓
  │  │
  │  └─ Returns user info:
  │     {
  │       "preferred_username": "admin",
  │       "email": "admin@domain.com",
  │       "given_name": "Admin",
  │       "family_name": "User",
  │       "roles": ["user", "admin"]
  │     }
  │
  ├─ 6. User makes API calls:
  │  │  GET /api/mysql/users
  │  │  Header: Authorization: Bearer {access_token}
  │  │
  │  ├─ Backend receives request
  │  ├─ Backend validates token with Keycloak
  │  ├─ Checks RBAC (roles & permissions)
  │  └─ Returns data ✓
  │
  ├─ 7. Token expires (after 5 minutes)
  │  │
  │  ├─ Frontend detects expired token
  │  │
  │  └─ Sends refresh request:
  │     POST /realms/axiomnizam/protocol/openid-connect/token
  │     grant_type: refresh_token
  │     refresh_token: {refresh_token}
  │
  │  ├─ Keycloak validates refresh token ✓
  │  │
  │  └─ Returns new access_token
  │
  ├─ 8. User logs out
  │  │
  │  └─ Frontend sends:
  │     POST /realms/axiomnizam/protocol/openid-connect/logout
  │     refresh_token: {refresh_token}
  │
  │  ├─ Keycloak invalidates session ✓
  │  │
  │  └─ Returns to Login Form
```

---

## Backend Service Flow (Client Credentials)

```
BACKEND SERVICE (http://localhost:8000)
  │
  ├─ Reads .env:
  │  KEYCLOAK_CLIENT_ID=axiomnizam-backend
  │  KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
  │
  ├─ 1. Initialization:
  │  │  POST /realms/axiomnizam/protocol/openid-connect/token
  │  │  client_id: axiomnizam-backend
  │  │  client_secret: 6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
  │  │  grant_type: client_credentials
  │  │
  │  ├─ Keycloak validates client ✓
  │  │
  │  └─ Returns service token
  │
  ├─ 2. Store token for API calls
  │  │  backend_token = service_token
  │  │  token_expires_at = now + expires_in
  │  │
  │  └─ Continue...
  │
  ├─ 3. When calling other services:
  │  │  GET/POST /api/other-service
  │  │  Header: Authorization: Bearer {backend_token}
  │  │
  │  └─ Service validates token
  │
  ├─ 4. When token expired:
  │  │
  │  └─ Get new token (repeat step 1)
```

---

## Postman Testing Flow

```
POSTMAN COLLECTION: AxiomNizam API

├─ Authentication & Login (NEW FOLDER)
│  │
│  ├─ 1️⃣ Backend: Get Client Credentials Token
│  │  POST /realms/axiomnizam/protocol/openid-connect/token
│  │  └─ Saves: {{backend_token}}
│  │
│  ├─ 2️⃣ User: Login with Password
│  │  POST /realms/axiomnizam/protocol/openid-connect/token
│  │  └─ Saves: {{user_token}}, {{refresh_token}}
│  │
│  ├─ 3️⃣ ✅ Check User Session (Validate Token)
│  │  GET /realms/axiomnizam/protocol/openid-connect/userinfo
│  │  └─ Returns: User info (confirms login)
│  │
│  ├─ 4️⃣ Refresh User Token
│  │  POST /realms/axiomnizam/protocol/openid-connect/token
│  │  └─ Saves: {{user_token}} (new)
│  │
│  └─ 5️⃣ Logout User
│     POST /realms/axiomnizam/protocol/openid-connect/logout
│     └─ Invalidates: Session
│
├─ Health & Status
│  └─ Uses: {{user_token}} in Bearer header
│
├─ MySQL
│  └─ Uses: {{user_token}} in Bearer header
│
├─ PostgreSQL
│  └─ Uses: {{user_token}} in Bearer header
│
└─ ... (All other endpoints use {{user_token}})
```

---

## Configuration Files Structure

```
AxiomNizam/
├── .env (Backend Config)
│  ├─ KEYCLOAK_URL=http://localhost:8080
│  ├─ KEYCLOAK_REALM=axiomnizam
│  ├─ KEYCLOAK_CLIENT_ID=axiomnizam-backend
│  ├─ KEYCLOAK_CLIENT_SECRET=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72
│  └─ KEYCLOAK_GRANT_TYPE=client_credentials
│
├── frontend/
│  └─ .env (Frontend Config)
│     ├─ KEYCLOAK_URL=http://localhost:8080
│     ├─ KEYCLOAK_REALM=axiomnizam
│     ├─ KEYCLOAK_CLIENT_ID=axiomnizam-frontend
│     └─ KEYCLOAK_REDIRECT_URI=http://localhost:7000/callback
│
├── POSTMAN_COLLECTION.json (Updated)
│  └─ Authentication & Login folder
│     ├─ 5 new endpoints
│     └─ Variables updated
│
├── KEYCLOAK_AUTH_SETUP.md (Documentation)
├── AUTH_QUICK_REFERENCE.md (Quick Guide)
├── AUTH_CONFIGURATION_COMPLETE.md (Summary)
└── KEYCLOAK_ARCHITECTURE.md (This file)
```

---

## Token Lifecycle

```
ACCESS TOKEN (Valid for 5 minutes)
├─ Format: JWT (JSON Web Token)
├─ Contains: sub, email, name, roles, permissions
├─ Used in: Authorization: Bearer {token}
├─ Expires: 300 seconds (5 minutes)
└─ Renewal: Use refresh_token to get new one

REFRESH TOKEN (Valid for 24+ hours)
├─ Format: Opaque string
├─ Contains: Session reference only
├─ Used in: grant_type=refresh_token flow
├─ Expires: 86400+ seconds
└─ Purpose: Get new access_token without re-login

SESSION
├─ Created: On login with password grant
├─ Maintained: By Keycloak server
├─ Invalidated: On logout
└─ Cleared: Refresh token deleted
```

---

## Security Model

```
┌─────────────────────────────────────────┐
│         KEYCLOAK REALM                  │
│        (axiomnizam)                     │
├─────────────────────────────────────────┤
│                                         │
│  OAUTH 2.0 / OPENID CONNECT             │
│  ├─ Client Credentials Flow             │
│  │  └─ Backend ↔ Backend                │
│  │                                      │
│  └─ Resource Owner Password Flow        │
│     └─ User → Frontend → Backend        │
│                                         │
│  BEARER TOKENS                          │
│  ├─ JWT Access Tokens                   │
│  │  └─ Short lived (5 min)              │
│  │                                      │
│  └─ Refresh Tokens                      │
│     └─ Long lived (24+ hours)           │
│                                         │
│  ROLES & PERMISSIONS                    │
│  ├─ Role Mapping                        │
│  ├─ Scope Management                    │
│  └─ Client Roles                        │
│                                         │
│  TOKEN VALIDATION                       │
│  ├─ Signature verification              │
│  ├─ Expiration check                    │
│  └─ Claim validation                    │
│                                         │
└─────────────────────────────────────────┘
```

---

## Implementation Checklist

```
BACKEND Implementation:
☐ Read KEYCLOAK_* from .env using os.Getenv()
☐ Implement token validation middleware
☐ Add token refresh logic
☐ Implement RBAC checks
☐ Add audit logging
☐ Cache token validation (optional)
☐ Implement logout revocation
☐ Add token expiry handling

FRONTEND Implementation:
☐ Add Keycloak client library
☐ Implement login form
☐ Implement OAuth2 flow
☐ Store tokens securely (httpOnly cookies)
☐ Implement auto-refresh
☐ Add logout functionality
☐ Add session check on app load
☐ Implement error handling

TESTING:
☐ Test all Postman endpoints
☐ Test token validation
☐ Test token refresh
☐ Test logout
☐ Test RBAC enforcement
☐ Test concurrent logins
☐ Test token expiry scenarios
☐ Load testing
```

---

**Created**: January 22, 2026
**Architecture Version**: 1.0
**Status**: ✅ COMPLETE
