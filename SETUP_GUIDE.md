# Setup Guide - AxiomNizam

Complete setup instructions for AxiomNizam multi-database API platform.

---

## Prerequisites

Before starting, ensure you have:
- ✅ Docker & Docker Compose installed
- ✅ Go 1.18 or later (for local development)
- ✅ 8GB RAM minimum
- ✅ Ports available: 7000, 8000, 8080, 3306, 5432, 27017

---

## Installation Methods

### Method 1: Docker Compose (Recommended)

#### Step 1: Clone or Download Project
```bash
cd AxiomNizam
```

#### Step 2: Start All Services
```bash
docker-compose up -d
```

This starts:
- **Keycloak** (localhost:8080) - Authentication server
- **Backend API** (localhost:8000) - Go API server  
- **Frontend** (localhost:7000) - Dashboard
- **PostgreSQL** (localhost:5432) - Database
- **MySQL** (localhost:3306) - Database
- **MongoDB** (localhost:27017) - Database

#### Step 3: Verify Services
```bash
docker-compose ps

# Should show all containers as "Up"
```

#### Step 4: Wait for Startup
- Keycloak: ~30-40 seconds
- Databases: ~10-15 seconds
- Backend/Frontend: ~5 seconds

Check logs if needed:
```bash
docker-compose logs -f keycloak    # Watch Keycloak startup
docker-compose logs -f backend     # Watch Backend
```

---

### Method 2: Local Development

#### Backend Setup

1. **Install dependencies:**
```bash
cd AxiomNizam
go mod download
```

2. **Configure environment** (edit `.env`):
```bash
BACKEND_PORT=8000
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
DATABASE_TYPE=postgres
DATABASE_HOST=localhost
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DISCORD_WEBHOOK_URL=https://discordapp.com/api/webhooks/...
```

3. **Run backend:**
```bash
go run main.go
# Server runs on http://localhost:8000
```

#### Frontend Setup

1. **Navigate to frontend:**
```bash
cd AxiomNizam/frontend
go mod download
```

2. **Configure environment** (edit `.env`):
```bash
FRONTEND_PORT=7000
BACKEND_URL=http://localhost:8000
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-frontend
```

3. **Run frontend:**
```bash
go run main.go
# Server runs on http://localhost:7000
```

#### Keycloak Setup (Manual)

If using Docker Compose, Keycloak starts automatically. Otherwise:

1. **Download Keycloak** from keycloak.org
2. **Start Keycloak:**
```bash
./bin/kc.sh start-dev
```
3. **Configure realm** (see Authentication section)

---

## Configuration Files

### Backend `.env`

```dotenv
# Server Configuration
BACKEND_PORT=8000
FRONTEND_PORT=7000

# Keycloak Configuration
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_PUBLIC_KEY=-----BEGIN PUBLIC KEY-----...

# Database Connections
# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=axiomnizam

# MySQL
MYSQL_HOST=localhost
MYSQL_USER=root
MYSQL_PASSWORD=root
MYSQL_DATABASE=axiomnizam

# MongoDB
MONGODB_URI=mongodb://localhost:27017

# Firebase (optional)
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_PRIVATE_KEY=...

# Oracle (optional)
ORACLE_HOST=localhost
ORACLE_PORT=1521
ORACLE_USER=admin
ORACLE_PASSWORD=admin
ORACLE_SID=xe

# Discord Integration (optional)
DISCORD_WEBHOOK_URL=https://discordapp.com/api/webhooks/...

# JWT Configuration
JWT_EXPIRATION=3600
JWT_REFRESH_EXPIRATION=86400
```

### Frontend `.env`

```dotenv
FRONTEND_PORT=7000
BACKEND_URL=http://localhost:8000
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-frontend
KEYCLOAK_REDIRECT_URI=http://localhost:7000/callback
```

---

## Keycloak Configuration

### Realm Setup

1. **Access Keycloak:**
   - URL: http://localhost:8080
   - Admin: admin / admin

2. **Create Realm:**
   - Click "Master" dropdown → "Create Realm"
   - Name: `axiomnizam`
   - Click "Create"

3. **Create Clients:**

   **For Backend:**
   - Go to "Clients" → "Create Client"
   - Client ID: `axiomnizam-backend`
   - Client Protocol: `openid-connect`
   - Access Type: `Confidential`
   - Valid Redirect URIs: `http://localhost:8000/*`
   - Save, then go to "Credentials" and copy the Secret

   **For Frontend:**
   - Go to "Clients" → "Create Client"
   - Client ID: `axiomnizam-frontend`
   - Client Protocol: `openid-connect`
   - Access Type: `Public`
   - Valid Redirect URIs: `http://localhost:7000/*`

4. **Create Roles:**
   - Go to "Roles" → "Create Role"
   - Create: `admin`
   - Create: `user`

5. **Create Users:**
   - Go to "Users" → "Create User"
   - Username: `admin_user`
   - Email: `admin@example.com`
   - Assign role: `admin`
   
   - Username: `regular_user`
   - Email: `user@example.com`
   - Assign role: `user`

---

## Database Initialization

### PostgreSQL
```bash
psql -U postgres -h localhost -d axiomnizam < init-postgres.sql
```

### MySQL
```bash
mysql -u root -p axiomnizam < init-postgres.sql
```

### MongoDB
```bash
mongosh localhost:27017/axiomnizam
# Collections auto-create on first use
```

---

## Verification Checklist

### Services Running
```bash
# Check all services
docker-compose ps

# Expected output:
# ✅ keycloak     Up
# ✅ backend      Up
# ✅ frontend     Up
# ✅ postgres     Up
# ✅ mysql        Up
# ✅ mongodb      Up
```

### Health Checks
```bash
# Backend health
curl http://localhost:8000/health
# Expected: {"status":"ok","message":"..."}

# Backend status
curl http://localhost:8000/status
# Expected: Database connectivity status

# Frontend
curl http://localhost:7000
# Expected: HTML dashboard

# Keycloak
curl http://localhost:8080/health
# Expected: Keycloak health info
```

### Token Generation
```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=YOUR_SECRET&grant_type=client_credentials"

# Expected: {"access_token":"...","token_type":"Bearer","expires_in":3600}
```

### API Access
```bash
TOKEN="your-token-from-above"

curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN"

# Expected: {"status":"success","data":[],...}
```

---

## Troubleshooting

### Services won't start
```bash
# Check port conflicts
netstat -tuln | grep LISTEN

# Clear Docker volumes (careful - deletes data)
docker-compose down -v
docker-compose up -d
```

### Keycloak taking too long
- Keycloak can take 30-60 seconds
- Check logs: `docker-compose logs keycloak`
- Wait longer and try again

### Database connection errors
```bash
# Verify database is running
docker-compose ps | grep postgres

# Check logs
docker-compose logs postgres

# Rebuild database
docker-compose down -v
docker-compose up -d postgres
```

### Token validation fails
- Verify PUBLIC_KEY in backend `.env`
- Get it from: `http://localhost:8080/realms/axiomnizam/protocol/openid-connect/certs`
- Ensure token matches realm

### Frontend can't connect to backend
- Check BACKEND_URL in frontend `.env`
- Verify backend is accessible: `curl http://localhost:8000/health`
- Check browser console for CORS errors

### API returns 403 Forbidden
- Your user role doesn't have permission
- Use admin token for write operations
- Check role assignment in Keycloak

---

## Next Steps

1. **Test APIs**: Import [POSTMAN_COLLECTION.json](POSTMAN_COLLECTION.json)
2. **Learn Endpoints**: Read [API_REFERENCE.md](API_REFERENCE.md)
3. **Understand Auth**: Read [AUTHENTICATION.md](AUTHENTICATION.md)
4. **Deploy**: Use docker-compose or Kubernetes

---

## Production Deployment

### Security Checklist
- [ ] Change default Keycloak admin password
- [ ] Use HTTPS (TLS certificates)
- [ ] Set strong JWT secrets
- [ ] Configure CORS properly
- [ ] Use environment variables for secrets
- [ ] Enable rate limiting
- [ ] Set up monitoring/alerting
- [ ] Regular backups

### Scaling
- Use load balancer for API
- Database replication for high availability
- Caching layer (Redis) for performance
- CDN for frontend assets

---

## Support & Resources

- **Keycloak Docs**: https://www.keycloak.org/documentation.html
- **Go Docs**: https://golang.org/doc
- **Docker Docs**: https://docs.docker.com
- **Project Issues**: Check GitHub repository

---

**Setup Complete! Now read [QUICK_START.md](QUICK_START.md) or [API_REFERENCE.md](API_REFERENCE.md)**
