# Quick Start - AxiomNizam (5 minutes)

Get up and running in 5 minutes with just 4 commands.

---

## Step 1: Start Services (1 min)

```bash
docker-compose up -d
```

This starts:
- ✅ Keycloak (http://localhost:8080)
- ✅ Backend API (http://localhost:8000)
- ✅ Frontend Dashboard (http://localhost:7000)
- ✅ All databases

**Wait for services to be ready** (check: `docker-compose ps`)

---

## Step 2: Get Your Token (2 min)

Copy and run this command to get an authentication token:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72&grant_type=client_credentials" \
  | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

echo "Your token: $TOKEN"
```

Save the token value - you'll need it for the next step.

---

## Step 3: Test the API (1 min)

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

---

## Step 4: View the Dashboard (1 min)

Open your browser:
```
http://localhost:7000
```

You'll see:
- ✅ Health status
- ✅ Database connections
- ✅ Available APIs
- ✅ API documentation

---

## 🎉 You're Done!

Your AxiomNizam API is now running. What's next?

### Try Some Commands

**Create a MySQL user** (Admin only):
```bash
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "age": 28
  }'
```

**Get all users**:
```bash
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Check health** (No token needed):
```bash
curl http://localhost:8000/health
```

**Check database connections** (No token needed):
```bash
curl http://localhost:8000/status
```

---

## 📋 What You Have

| Service | URL | Status |
|---------|-----|--------|
| Frontend | http://localhost:7000 | ✅ Running |
| Backend API | http://localhost:8000 | ✅ Running |
| Keycloak | http://localhost:8080 | ✅ Running |
| Databases | localhost:3306+ | ✅ Connected |

---

## 🔑 Default Credentials

| Item | Value |
|------|-------|
| Keycloak Realm | `axiomnizam` |
| Client ID | `axiomnizam-backend` |
| Client Secret | `6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72` |
| Keycloak Admin | `admin` / `admin` |

---

## 🐛 Troubleshooting

### Services not starting?
```bash
# Check what's running
docker-compose ps

# View logs
docker-compose logs
```

### Token error?
- Make sure Keycloak is ready (wait 30 seconds)
- Check the realm URL is correct
- Verify client credentials

### API returns 401?
- Your token might be expired
- Get a new token using Step 2
- Make sure to include `Bearer ` before the token

### Can't connect to database?
- Check if services are fully running
- Wait another 30 seconds and try again
- Check docker logs: `docker-compose logs`

---

## 📚 Learn More

- **Full setup guide**: See [SETUP_GUIDE.md](SETUP_GUIDE.md)
- **All API endpoints**: See [API_REFERENCE.md](API_REFERENCE.md)
- **Authentication details**: See [AUTHENTICATION.md](AUTHENTICATION.md)
- **Postman testing**: See [POSTMAN_API_GUIDE.md](POSTMAN_API_GUIDE.md)

---

**That's it! You're ready to use AxiomNizam.** 🚀
