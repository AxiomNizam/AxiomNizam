# AxiomNizam - Multi-Database API Platform

> A comprehensive, production-ready API platform with Keycloak authentication, Role-Based Access Control (RBAC), and support for 7 major databases.

**Status**: ✅ **FULLY OPERATIONAL** | Latest Update: January 22, 2026

---

## 🎯 What is AxiomNizam?

AxiomNizam is a unified API platform that allows you to manage data across multiple database systems through a single, authenticated REST API interface.

### Supported Databases
- ✅ MySQL
- ✅ MariaDB
- ✅ PostgreSQL
- ✅ Percona (MySQL fork)
- ✅ MongoDB
- ✅ Firebase
- ✅ Oracle

---

## 🚀 Quick Start (5 minutes)

### Prerequisites
- Docker & Docker Compose
- Go 1.18+
- Postman (optional, for testing)

### 1. Start All Services
```bash
docker-compose up -d
# Keycloak runs on: http://localhost:8080
# Backend runs on: http://localhost:8000
# Frontend runs on: http://localhost:7000
```

### 2. Get an Auth Token
```bash
curl -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72&grant_type=client_credentials"
```

### 3. Make an API Call
```bash
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 4. View Dashboard
Open: **http://localhost:7000**

---

## 📋 Main Features

### ✅ Authentication & Authorization
- **Keycloak Integration** - OpenID Connect authentication
- **JWT Tokens** - Secure token-based access
- **Role-Based Access Control** - Admin/User role separation
- **Multi-Realm Support** - Flexible authentication configuration

### ✅ API Capabilities
- **35+ API Endpoints** - Full CRUD operations across all databases
- **Request/Response Validation** - Comprehensive error handling
- **CORS Enabled** - Frontend integration ready
- **Rate Limiting Ready** - Scalable architecture

### ✅ Frontend Dashboard
- **Real-time Health Monitoring** - Live database connection status
- **API Documentation** - Built-in API reference
- **Dark/Light/Default Themes** - Customizable UI
- **Admin Panel** - Database and table management
- **System Manager** - Advanced operations

### ✅ Notification System
- **Discord Integration** - Real-time alerts
- **Health Notifications** - Automatic status updates
- **Custom Notifications** - Send custom messages

---

## 📚 Documentation Guide

Choose your learning path:

### 🏃 **5-Minute Quick Start**
→ **[QUICK_START.md](QUICK_START.md)**
- Basic setup
- Public endpoints
- Token acquisition
- First API call

### 🔧 **30-Minute Setup Guide**
→ **[SETUP_GUIDE.md](SETUP_GUIDE.md)**
- Complete installation
- Environment configuration
- Keycloak setup
- Database connectivity

### 📡 **API Reference (Bookmarks)**
→ **[API_REFERENCE.md](API_REFERENCE.md)**
- All 35+ endpoints
- Request/response examples
- PowerShell commands
- cURL examples

### 🔐 **Authentication & RBAC**
→ **[AUTHENTICATION.md](AUTHENTICATION.md)**
- Token flows
- Protected endpoints
- Role-based access
- JWT validation

### 📮 **Postman Testing**
→ **[POSTMAN_API_GUIDE.md](POSTMAN_API_GUIDE.md)** + **[POSTMAN_COLLECTION.json](POSTMAN_COLLECTION.json)**
- Pre-configured requests
- Auto-token scripts
- All endpoints ready

---

## 🔑 Key Credentials

### Keycloak
- **URL**: http://localhost:8080
- **Realm**: `axiomnizam`
- **Admin User**: `admin` / `admin`

### Backend Service
- **Client ID**: `axiomnizam-backend`
- **Client Secret**: `6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72`
- **Grant Type**: Client Credentials

### Frontend Service
- **Client ID**: `axiomnizam-frontend`
- **Realm**: `axiomnizam`

---

## 🏗️ System Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Browser / Client                      │
└─────────────────────────┬────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ↓                 ↓                 ↓
    Frontend          Health Check      Status Check
    :7000             :8000/health       :8000/status
    
        ↓─────────────────┴─────────────────↓
        
┌──────────────────────────────────────────────────────────┐
│                    API Gateway (Go)                      │
│                    :8000                                 │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Authentication & RBAC Middleware                │   │
│  │  - Validate JWT tokens                           │   │
│  │  - Check role permissions                        │   │
│  │  - Rate limiting                                 │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────┘
                      │
    ┌─────────────────┼────────────────┬──────────────┬──────────────┐
    ↓                 ↓                ↓              ↓              ↓
  MySQL           PostgreSQL       MongoDB       Firebase        Oracle
  :3306           :5432            :27017        (Cloud)         :1521
    │                 │                │              │              │
    ├─ MariaDB        └─ Percona      └─────────────┘              │
    └─ (compatible)       (fork)
```

---

## 🔐 Role-Based Access Control (RBAC)

### Permission Matrix

| Operation | Admin | Non-Admin | Public |
|-----------|-------|-----------|--------|
| **GET** (Read) | ✅ | ✅ | N/A |
| **POST** (Create) | ✅ | ❌ | N/A |
| **PUT** (Update) | ✅ | ❌ | N/A |
| **DELETE** (Delete) | ✅ | ❌ | N/A |
| **/health** | ✅ | ✅ | ✅ |
| **/status** | ✅ | ✅ | ✅ |

---

## 📡 API Examples

### Health Check (No Auth Required)
```bash
curl http://localhost:8000/health
```

### Get All MySQL Users (Auth Required)
```bash
curl http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN"
```

### Create MySQL User (Admin Only)
```bash
curl -X POST http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }'
```

---

## 📊 Services & Ports

| Service | Port | URL |
|---------|------|-----|
| Frontend | 7000 | http://localhost:7000 |
| Backend API | 8000 | http://localhost:8000 |
| Keycloak | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | localhost |
| MySQL | 3306 | localhost |
| MongoDB | 27017 | localhost |
| Firebase | - | Cloud-based |

---

## ✅ Verification Checklist

- [x] Keycloak running and configured
- [x] Backend API operational
- [x] Frontend dashboard accessible
- [x] All 7 databases connected
- [x] RBAC implemented and tested
- [x] JWT authentication working
- [x] Postman collection ready
- [x] Documentation complete

---

## 🐛 Troubleshooting

### Services won't start?
```bash
# Check Docker status
docker-compose ps

# View logs
docker-compose logs keycloak
docker-compose logs backend
docker-compose logs frontend
```

### Can't get token?
```bash
# Verify Keycloak is running
curl http://localhost:8080/health

# Check realm configuration
curl http://localhost:8080/realms/axiomnizam/.well-known/openid-configuration
```

### API returns 401 Unauthorized?
- Token expired? Get a new one
- Wrong token? Verify client credentials
- Role insufficient? Use admin token for write operations

---

## 📞 Getting Help

### Questions?
1. Check **[QUICK_START.md](QUICK_START.md)** for fast answers
2. Read **[SETUP_GUIDE.md](SETUP_GUIDE.md)** for configuration
3. Review **[API_REFERENCE.md](API_REFERENCE.md)** for endpoints
4. See **[AUTHENTICATION.md](AUTHENTICATION.md)** for auth details

### Still stuck?
- Check `.env` files for correct configuration
- Review container logs: `docker-compose logs`
- Verify all services are running: `docker-compose ps`

---

## 🎓 Learning Paths

### Path 1: Just Make It Work (5 min)
1. Start services: `docker-compose up -d`
2. Get token (see above)
3. Make API call (see above)
4. Done! ✅

### Path 2: Understand the System (30 min)
1. Read QUICK_START.md
2. Read SETUP_GUIDE.md
3. Read AUTHENTICATION.md
4. Import POSTMAN_COLLECTION.json
5. Test all endpoints
6. Done! ✅

### Path 3: Master Everything (60+ min)
1. Read all documentation files
2. Review system architecture
3. Study RBAC implementation
4. Explore database schemas
5. Deploy to production
6. Implement monitoring
7. Done! ✅

---

## 📦 Project Structure

```
AxiomNizam/
├── README.md                          # You are here
├── QUICK_START.md                     # 5-minute guide
├── SETUP_GUIDE.md                     # Complete setup
├── API_REFERENCE.md                   # All endpoints
├── AUTHENTICATION.md                  # Auth & RBAC
├── POSTMAN_API_GUIDE.md              # Postman setup
├── POSTMAN_COLLECTION.json           # Ready-to-import
│
├── main.go                            # Backend entry point
├── go.mod                             # Dependencies
├── docker-compose.yml                 # Service orchestration
├── init-postgres.sql                  # Database initialization
│
├── frontend/
│   ├── main.go                        # Frontend server
│   ├── go.mod
│   └── templates/
│       ├── layout.html                # Base layout
│       ├── public-dashboard.html      # Public view
│       ├── admin.html                 # Admin panel
│       ├── system-manager.html        # System manager
│       ├── auth.js                    # Auth module
│       ├── dashboard.js               # Dashboard logic
│       ├── admin.js                   # Admin logic
│       ├── style.css                  # Styles
│       └── responsive.css             # Mobile styles
│
└── internal/
    ├── auth/
    │   ├── auth.go                    # JWT validation
    │   └── middleware.go              # Auth middleware
    ├── config/
    │   └── config.go                  # Configuration
    ├── database/
    │   └── connections.go             # DB connections
    ├── handlers/
    │   ├── handlers.go                # Route handlers
    │   ├── mysql.go                   # MySQL handlers
    │   ├── postgres.go                # PostgreSQL handlers
    │   ├── mongodb.go                 # MongoDB handlers
    │   ├── firebase.go                # Firebase handlers
    │   ├── oracle.go                  # Oracle handlers
    │   ├── admin_handler.go           # Admin handlers
    │   └── notification_handler.go    # Notification handlers
    ├── models/
    │   └── models.go                  # Data models
    └── utils/
        └── ...                        # Utility functions
```

---

## 🚀 Next Steps

1. **Get Started**: Follow [QUICK_START.md](QUICK_START.md)
2. **Configure**: Use [SETUP_GUIDE.md](SETUP_GUIDE.md)
3. **Test APIs**: Import [POSTMAN_COLLECTION.json](POSTMAN_COLLECTION.json)
4. **Learn More**: Read [AUTHENTICATION.md](AUTHENTICATION.md)
5. **Explore**: Check [API_REFERENCE.md](API_REFERENCE.md)

---

## 📄 License

MIT License - See LICENSE file for details

---

**Made with ❤️ for seamless multi-database management**
