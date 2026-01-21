# Authentication & Authorization Guide

## Overview

All CRUD operations in AxiomNizam are protected by JWT token-based authentication via **Keycloak**. The frontend dashboard and any API client must provide a valid JWT token in the `Authorization` header.

## Authentication Flow

### 1. Get Access Token from Keycloak

```powershell
# PowerShell
$clientId = "axiomnizam"
$clientSecret = "your-client-secret"  # Set in Keycloak
$username = "admin"
$password = "admin"
$realm = "master"

$tokenResponse = Invoke-RestMethod `
  -Uri "http://localhost:8080/realms/$realm/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = $clientId
    grant_type    = "password"
    username      = $username
    password      = $password
    client_secret = $clientSecret
  }

$accessToken = $tokenResponse.access_token
Write-Host "Token: $accessToken"
```

### 2. Use Token in API Requests

All CRUD endpoints require the token in the Authorization header:

```powershell
# Example: Create a user in MySQL
$headers = @{
    "Authorization" = "Bearer $accessToken"
    "Content-Type"  = "application/json"
}

$body = @{
    name  = "John Doe"
    email = "john@example.com"
    age   = 30
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8000/api/mysql/users" `
  -Method POST `
  -Headers $headers `
  -Body $body
```

## Protected Endpoints

All CRUD operations require authentication (Bearer token):

### MySQL
- `POST /api/mysql/users` - Create user (requires auth)
- `GET /api/mysql/users` - Get all users (requires auth)
- `GET /api/mysql/users/:id` - Get user by ID (requires auth)
- `PUT /api/mysql/users/:id` - Update user (requires auth)
- `DELETE /api/mysql/users/:id` - Delete user (requires auth)

### MariaDB
- `POST /api/mariadb/users` - Create user (requires auth)
- `GET /api/mariadb/users` - Get all users (requires auth)
- `GET /api/mariadb/users/:id` - Get user by ID (requires auth)
- `PUT /api/mariadb/users/:id` - Update user (requires auth)
- `DELETE /api/mariadb/users/:id` - Delete user (requires auth)

### PostgreSQL
- `POST /api/postgres/users` - Create user (requires auth)
- `GET /api/postgres/users` - Get all users (requires auth)
- `GET /api/postgres/users/:id` - Get user by ID (requires auth)
- `PUT /api/postgres/users/:id` - Update user (requires auth)
- `DELETE /api/postgres/users/:id` - Delete user (requires auth)

### Percona
- `POST /api/percona/users` - Create user (requires auth)
- `GET /api/percona/users` - Get all users (requires auth)
- `GET /api/percona/users/:id` - Get user by ID (requires auth)
- `PUT /api/percona/users/:id` - Update user (requires auth)
- `DELETE /api/percona/users/:id` - Delete user (requires auth)

### MongoDB
- `POST /api/mongodb/users` - Create user (requires auth)
- `GET /api/mongodb/users` - Get all users (requires auth)
- `GET /api/mongodb/users/:id` - Get user by ID (requires auth)
- `PUT /api/mongodb/users/:id` - Update user (requires auth)
- `DELETE /api/mongodb/users/:id` - Delete user (requires auth)

### Firebase
- `POST /api/firebase/users` - Create user (requires auth)
- `GET /api/firebase/users` - Get all users (requires auth)
- `GET /api/firebase/users/:id` - Get user by ID (requires auth)
- `PUT /api/firebase/users/:id` - Update user (requires auth)
- `DELETE /api/firebase/users/:id` - Delete user (requires auth)

### Oracle
- `POST /api/oracle/users` - Create user (requires auth)
- `GET /api/oracle/users` - Get all users (requires auth)
- `GET /api/oracle/users/:id` - Get user by ID (requires auth)
- `PUT /api/oracle/users/:id` - Update user (requires auth)
- `DELETE /api/oracle/users/:id` - Delete user (requires auth)

## Public Endpoints (No Auth Required)

- `GET /health` - Health check
- `GET /status` - Database connection status

## Token Structure

The JWT token from Keycloak contains the following claims:

```json
{
  "sub": "user-id",
  "preferred_username": "username",
  "email": "user@example.com",
  "name": "User Full Name",
  "exp": 1234567890,
  "iat": 1234567800
}
```

These claims are automatically extracted and available in the API context.

## Keycloak Configuration

### Default Setup

When using Docker Compose, Keycloak is configured with:
- **URL**: http://localhost:8080
- **Admin Console**: http://localhost:8080/admin
- **Username**: admin
- **Password**: admin
- **Realm**: master

### Create a Client for API Access

1. Go to http://localhost:8080/admin
2. Login with `admin`/`admin`
3. Select realm "master" (or create a new realm)
4. Click "Clients" → "Create"
5. Set Client ID to `axiomnizam`
6. Enable "Standard Flow Enabled" and "Direct Access Grants Enabled"
7. Set Valid Redirect URIs to `http://localhost:8000/*`
8. Save and note the Client Secret (from Credentials tab)

### Create a Test User

1. In Admin Console, go to Users → Create
2. Set username to `testuser`
3. Set password (temporary), uncheck "Temporary" before saving
4. Use this username/password to get tokens

## Error Responses

### Missing Token
```
Status: 401
{
  "error": "missing authorization header"
}
```

### Invalid Token Format
```
Status: 401
{
  "error": "invalid authorization header: invalid authorization header format"
}
```

### Token Expired
```
Status: 401
{
  "error": "invalid token: token has expired"
}
```

### Invalid Token
```
Status: 401
{
  "error": "invalid token: failed to parse token"
}
```

## Environment Variables

Add these to your `.env` file:

```
# Keycloak Configuration
KEYCLOAK_HOST=keycloak
KEYCLOAK_PORT=8080
KEYCLOAK_REALM=master
KEYCLOAK_CLIENT_ID=axiomnizam
```

## Using with Frontend Dashboard

The frontend dashboard automatically handles token management if a valid token is provided. To authenticate the dashboard:

1. Get a token from Keycloak (see "Get Access Token" section above)
2. The dashboard will include the token in all requests to the backend
3. Display real-time authenticated access to all database statuses

## Testing with cURL

```bash
# Get token
TOKEN=$(curl -s -X POST \
  http://localhost:8080/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam" \
  -d "grant_type=password" \
  -d "username=admin" \
  -d "password=admin" \
  | jq -r '.access_token')

# Use token in request
curl -X GET http://localhost:8000/api/mysql/users \
  -H "Authorization: Bearer $TOKEN"
```

## Troubleshooting

### "Failed to fetch JWKS from Keycloak"
- Verify Keycloak is running: `docker ps | grep keycloak`
- Check Keycloak URL is correct in environment variables
- Verify network connectivity between API and Keycloak containers

### "Token validation failed"
- Ensure token is from the same Keycloak instance
- Check token hasn't expired
- Verify realm and client configuration match

### "No Authorization header"
- Ensure `Authorization: Bearer <token>` is in request headers
- Check token is not empty or malformed

## Security Best Practices

1. **Never commit tokens** to version control
2. **Always use HTTPS** in production (not http://)
3. **Rotate client secrets** regularly
4. **Use strong passwords** for Keycloak admin accounts
5. **Enable CORS** only for trusted domains
6. **Set appropriate token expiration** times
7. **Implement rate limiting** on token endpoint
8. **Monitor failed authentication** attempts

## Advanced Configuration

For production environments, consider:

1. **Token caching** to reduce Keycloak queries
2. **JWKS caching** with automatic refresh
3. **Rate limiting** on authentication endpoints
4. **Multi-realm support** for different environments
5. **Role-based access control (RBAC)** for different operations
6. **Audit logging** of authentication events
7. **Certificate pinning** for HTTPS connections
