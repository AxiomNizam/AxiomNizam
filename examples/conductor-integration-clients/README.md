# Conductor + IAM Integration Clients

This folder contains ready-to-use client templates that authenticate through IAM first, then call Conductor APIs.

Supported templates:
- Node.js: `nodejs/conductor-iam-client.js`
- Python FastAPI: `python-fastapi/main.py`
- Java Spring Boot: `java-springboot/com/example/axiomnizam/conductor/AxiomConductorClient.java`
- .NET: `dotnet/AxiomConductorClient.cs`

## IAM and Conductor flow

1. Login via `POST /auth/login` with:
   - `username`
   - `password`
2. Use returned `access_token` for Conductor API calls:
   - Header: `Authorization: Bearer <token>`
3. For browser-style WebSocket clients, use token query param:
   - `/ws/conductor?token=<token>`

## Important role requirements

Conductor read endpoints (list/stats/messages/stream) require authenticated token.

Conductor write endpoints require elevated roles (`admin`, `system-manager`, `sysadmin`, `system_admin`, `system-admin`), including:
- create/update/delete producers and consumers
- publish
- pause/resume
- connect/disconnect backends
- replay DLQ

## Common environment variables

- `AXIOM_BASE_URL` (default: `http://localhost:8000`)
- `AXIOM_USERNAME`
- `AXIOM_PASSWORD`

## Quick endpoint map

- `POST /auth/login`
- `POST /auth/refresh`
- `GET /api/v1/conductor/stats`
- `GET /api/v1/conductor/producers`
- `POST /api/v1/conductor/producers`
- `GET /api/v1/conductor/consumers`
- `POST /api/v1/conductor/consumers`
- `POST /api/v1/conductor/publish`
- `GET /api/v1/conductor/messages?limit=100`
- `GET /api/v1/conductor/dlq`
- `POST /api/v1/conductor/dlq/:id/replay`
- `GET /api/v1/conductor/stream` (SSE)
- `GET /ws/conductor` (WebSocket)

## Run notes

### Node.js
- Node 18+ required (`fetch` built in).
- Run:
  - `set AXIOM_BASE_URL=http://localhost:8000`
  - `set AXIOM_USERNAME=admin@example.com`
  - `set AXIOM_PASSWORD=your-password`
  - `node examples/conductor-integration-clients/nodejs/conductor-iam-client.js`

### Python FastAPI
- Install:
  - `pip install fastapi uvicorn httpx pydantic`
- Run:
  - `set AXIOM_BASE_URL=http://localhost:8000`
  - `set AXIOM_USERNAME=admin@example.com`
  - `set AXIOM_PASSWORD=your-password`
  - `uvicorn main:app --app-dir examples/conductor-integration-clients/python-fastapi --reload --port 8101`

### Java Spring Boot
- Add dependencies:
  - `spring-boot-starter-webflux`
  - `jackson-databind`
- Register `AxiomConductorClient` as a bean/service and inject into your controllers/services.

### .NET
- Target .NET 8+.
- Register `AxiomConductorClient` as a singleton/typed client with `HttpClient`.
