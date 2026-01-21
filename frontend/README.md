# AxiomNizam Frontend Dashboard

A lightweight Go-based web dashboard to monitor the health and status of the AxiomNizam backend API.

## Features

- 🎨 Beautiful, responsive web interface
- 💚 Real-time health status monitoring
- 🗄️ Database connection status display
- 🔄 Auto-refresh capability (default: every 5 seconds)
- 📱 Mobile-friendly design
- 🚀 Lightweight and fast

## Running the Frontend

### Prerequisites

- Go 1.21 or higher
- The AxiomNizam backend running on `http://localhost:8000`

### Development Mode

```bash
cd frontend

# Download dependencies
go mod download

# Run the frontend
go run main.go
```

The dashboard will be available at `http://localhost:8080`

### With Custom Backend URL

```bash
BACKEND_URL=http://your-backend-host:8000 go run main.go
```

### Using Docker

Build the Docker image:

```bash
docker build -t axiomnizam-frontend:latest .
```

Run the container:

```bash
docker run -d \
  -p 8080:8080 \
  -e BACKEND_URL=http://axiomnizam:8000 \
  --name axiomnizam-frontend \
  axiomnizam-frontend:latest
```

### With Docker Compose

The frontend service is included in the main `docker-compose.yml`:

```bash
docker-compose up -d axiomnizam-frontend
```

## Environment Variables

- `BACKEND_URL`: The URL of the AxiomNizam backend (default: `http://localhost:8000`)
- `FRONTEND_PORT`: The port to run the dashboard on (default: `8080`)

## API Endpoints

The frontend exposes the following endpoints:

- `GET /` - Main dashboard page
- `GET /api/health` - Fetch backend health status (JSON)
- `GET /api/status` - Fetch database connection status (JSON)

## Dashboard Features

### Health Status
Displays the overall health of the backend API:
- Status indicator (OK/Error)
- Status message

### Database Connections
Shows connection status for all configured databases:
- MySQL
- MariaDB
- PostgreSQL
- Percona
- MongoDB
- Firebase
- Oracle
- Valkey (Redis)
- Elasticsearch
- etcd
- Keycloak

### Auto-Refresh
- Automatically refreshes data every 5 seconds (configurable via checkbox)
- Manual refresh button for immediate updates
- Last update timestamp

## Customization

Edit `templates/dashboard.html` to customize:
- Colors and styling
- Refresh interval
- Additional information to display
- Layout and organization

## Troubleshooting

### Cannot connect to backend
- Verify the backend is running: `curl http://localhost:8000/health`
- Check `BACKEND_URL` environment variable is set correctly
- Ensure the backend and frontend are on the same network (if using Docker)

### Databases show as disconnected
- Check the backend logs: `docker logs axiomnizam`
- Verify database services are running
- Check database connection credentials in `.env`

### Port already in use
- Change the port with `FRONTEND_PORT=8888 go run main.go`
- Or kill the existing process: `lsof -i :8080`

## Example Usage

1. Start the main backend:
   ```bash
   docker-compose up -d
   ```

2. Start the frontend:
   ```bash
   cd frontend
   go run main.go
   ```

3. Open your browser to `http://localhost:8080`

4. Monitor the backend health and database connections in real-time!

## License

MIT License
