# 🎯 RBAC Quick Reference Card

**Print this page or keep it nearby while testing!**

---

## 📌 Role Permissions

```
┌─────────────┬──────┬─────────┐
│ Operation   │Admin │ User    │
├─────────────┼──────┼─────────┤
│ GET         │ ✅   │ ✅      │
│ POST        │ ✅   │ ❌      │
│ PUT         │ ✅   │ ❌      │
│ DELETE      │ ✅   │ ❌      │
│ /health     │ ✅   │ ✅ *    │
│ /status     │ ✅   │ ✅ *    │
└─────────────┴──────┴─────────┘
* No token needed
```

---

## 🔑 Quick Commands

### Get Admin Token
```powershell
$adminToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body @{client_id="axiomnizam";client_secret="uzqxRJUEI44gpURiytWtCujKwQ1ESZrv";grant_type="password";username="admin";password="admin"}).access_token
Write-Host "✅ Token: $($adminToken.Substring(0,30))..."
```

### Get User Token
```powershell
$userToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" -Method POST -ContentType "application/x-www-form-urlencoded" -Body @{client_id="axiomnizam";client_secret="uzqxRJUEI44gpURiytWtCujKwQ1ESZrv";grant_type="password";username="testuser";password="password123"}).access_token
Write-Host "✅ Token: $($userToken.Substring(0,30))..."
```

---

## 🧪 Test Commands

### Admin Read ✅
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" `
  -Headers @{"Authorization"="Bearer $adminToken"}
```

### Admin Create ✅
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Headers @{"Authorization"="Bearer $adminToken";"Content-Type"="application/json"} `
  -Body '{"name":"Test","email":"test@test.com","age":25}'
```

### User Read ✅
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" `
  -Headers @{"Authorization"="Bearer $userToken"}
```

### User Create ❌
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Headers @{"Authorization"="Bearer $userToken";"Content-Type"="application/json"} `
  -Body '{"name":"Test","email":"test@test.com","age":25}'
# Expected: 403 Forbidden
```

### Admin Update ✅
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users/1" `
  -Method PUT `
  -Headers @{"Authorization"="Bearer $adminToken";"Content-Type"="application/json"} `
  -Body '{"name":"Updated","email":"updated@test.com","age":30}'
```

### User Update ❌
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users/1" `
  -Method PUT `
  -Headers @{"Authorization"="Bearer $userToken";"Content-Type"="application/json"} `
  -Body '{"name":"Updated","email":"updated@test.com","age":30}'
# Expected: 403 Forbidden
```

### Admin Delete ✅
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users/1" `
  -Method DELETE `
  -Headers @{"Authorization"="Bearer $adminToken"}
```

### User Delete ❌
```powershell
Invoke-RestMethod "http://localhost:8000/api/mysql/users/1" `
  -Method DELETE `
  -Headers @{"Authorization"="Bearer $userToken"}
# Expected: 403 Forbidden
```

---

## 🔐 Credentials

| Item | Value |
|------|-------|
| Keycloak URL | http://localhost:8080 |
| Admin Console | http://localhost:8080 |
| Username (admin) | admin |
| Password (admin) | admin |
| Username (user) | testuser |
| Password (user) | password123 |
| Client ID | axiomnizam |
| Client Secret | uzqxRJUEI44gpURiytWtCujKwQ1ESZrv |
| Realm | master |
| Token Endpoint | http://localhost:8080/realms/master/protocol/openid-connect/token |
| API Server | http://localhost:8000 |

---

## 📋 Keycloak Setup Checklist

- [ ] Access Keycloak admin console
- [ ] Create role: **admin**
- [ ] Create role: **user** (optional)
- [ ] Assign admin role to admin user
- [ ] Create test user: **testuser**
- [ ] Set password: **password123**
- [ ] Don't assign admin role to testuser
- [ ] Verify both users can login

---

## 🗄️ Database Endpoints

All follow same pattern (7 databases):

```
GET  /api/{db}/users         ✅ All users
GET  /api/{db}/users/:id     ✅ One user
POST /api/{db}/users         🔒 Admin only
PUT  /api/{db}/users/:id     🔒 Admin only
DELETE /api/{db}/users/:id   🔒 Admin only
```

**Databases**:
- mysql
- mariadb
- postgres
- percona
- mongodb
- firebase
- oracle

---

## 🧩 Code Locations

| File | What | Line |
|------|------|------|
| auth.go | Claims struct | ~24 |
| auth.go | RealmAccess struct | ~20 |
| auth.go | HasRole method | ~42 |
| middleware.go | RequireRole | ~50 |
| middleware.go | RequireAdmin | ~80 |
| main.go | authMiddleware setup | ~65 |
| main.go | adminMiddleware setup | ~73 |
| main.go | Route protection | ~90+ |

---

## ❌ Common Issues & Fixes

| Problem | Cause | Fix |
|---------|-------|-----|
| 401 Unauthorized | No token | Add Authorization header |
| 403 Forbidden | Not admin | Assign admin role in Keycloak |
| Token expired | >5 min old | Get new token |
| Role not found | Typo in role name | Check exact role name |
| No roles in token | Role not assigned | Assign role and get new token |

---

## ✅ Expected Responses

### Success (Admin Create)
```
Status: 201 Created
{
  "id": 123,
  "name": "Test",
  "email": "test@test.com",
  "age": 25,
  "created_at": "2026-01-22T10:30:00Z"
}
```

### Success (User Read)
```
Status: 200 OK
[
  {"id": 1, "name": "User1", "email": "user1@test.com", "age": 25},
  {"id": 2, "name": "User2", "email": "user2@test.com", "age": 30}
]
```

### Error (User Create)
```
Status: 403 Forbidden
{
  "error": "forbidden: user does not have 'admin' role",
  "user_roles": [],
  "required": "admin"
}
```

### Error (No Token)
```
Status: 401 Unauthorized
{
  "error": "missing authorization header"
}
```

---

## 🔄 Token Lifecycle

```
1. User logs in with username/password
2. Keycloak validates credentials
3. Keycloak checks user roles
4. Keycloak issues JWT with roles embedded
5. Client sends JWT in Authorization header
6. Backend validates JWT signature
7. Backend extracts roles from claims
8. Backend checks if operation allowed
9. If allowed: execute operation
10. If not: return 403 Forbidden
```

**Token expires in 300 seconds (5 minutes)**

---

## 📊 HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | GET request success |
| 201 | Created | POST request success |
| 204 | No Content | DELETE success |
| 400 | Bad Request | Invalid JSON |
| 401 | Unauthorized | Missing/invalid token |
| 403 | Forbidden | Lacks required role |
| 500 | Server Error | Backend error |

---

## 🎯 Test Workflow

```
1. Get admin token
   ↓
2. Get user token
   ↓
3. Admin can read ✅
   ↓
4. Admin can create ✅
   ↓
5. User can read ✅
   ↓
6. User cannot create ❌
   ↓
7. User cannot update ❌
   ↓
8. User cannot delete ❌
   ↓
✅ All tests passed!
```

---

## 📚 Documentation Map

| File | Purpose | Time |
|------|---------|------|
| RBAC_QUICK_START.md | Fast setup | 5 min |
| RBAC_SETUP_GUIDE.md | Complete guide | 30 min |
| RBAC_IMPLEMENTATION_DETAILS.md | Technical deep dive | 60 min |
| RBAC_SUMMARY.md | Overview | 10 min |
| **This file** | Quick reference | Always |

---

## 🚀 Ready?

1. ✅ Credentials configured
2. ✅ Code implemented
3. ✅ Tests provided
4. ✅ Docs complete

**Start with RBAC_QUICK_START.md or use commands above! 🎉**

---

**Last Updated**: January 22, 2026
**Status**: ✅ Ready for Production Testing
