# AxiomNizam Frontend - Multi-Interface Project Structure

## 📋 Project Overview

A comprehensive Go-based frontend application with three distinct interfaces designed for different user roles:
- **Public Dashboard** (http://localhost:7000/) - Public access, no authentication required
- **Admin Interface** (http://localhost:7000/admin) - API testing & configuration
- **System Manager** (http://localhost:7000/system-manager) - Monitoring & control center

---

## 🏗️ Project Structure

```
frontend/
├── main.go                          # Main entry point (Multi-page router)
├── templates/
│   ├── layout.html                 # Base layout (shared navigation & modals)
│   ├── public-dashboard.html       # Public dashboard content
│   ├── admin.html                  # Admin panel content
│   ├── system-manager.html         # System manager content
│   ├── auth.js                     # Authentication module (Keycloak)
│   ├── dashboard.js                # Public dashboard logic
│   ├── admin.js                    # Admin interface logic
│   ├── system-manager.js           # System manager logic
│   ├── style.css                   # Main stylesheet (dark theme)
│   ├── responsive.css              # Responsive design & media queries
│   ├── admin-dashboard.html        # Legacy admin interface (deprecated)
│   ├── admin-dashboard.js          # Legacy admin JS (deprecated)
│   └── other files...
```

---

## 🎨 Design Features

### Visual Theme
- **Color Scheme**: Dark mode with blue accent (primary: #3b82f6)
- **Components**: Cards, badges, progress bars, modals, tables
- **Typography**: System fonts with clear hierarchy
- **Responsiveness**: Fully responsive from 320px to 4K screens

### Navigation
- **Sticky navbar** with logo, page links, user info
- **Active page indicator** 
- **Role-based visibility** (Admin/System Manager only visible when authenticated)
- **Logout button** with session management

### Shared Elements
- **Login Modal** - Keycloak authentication form
- **Response Modal** - Display API responses
- **Modals close on background click**
- **Loading spinners** during data fetch

---

## 📄 Template Files

### layout.html
Base HTML layout with:
- Navigation bar with branding
- Login/Logout buttons
- Tab-based navigation
- Content area (filled by each page)
- Footer with links
- Authentication forms
- Response display modals

**Key Features:**
- Conditional rendering based on authentication
- Dynamic page titles and active page highlighting
- Responsive grid system

---

### public-dashboard.html
Public-facing dashboard showing:
- **System Status Stats**: Health, databases, response time, uptime
- **Database Connections**: Visual grid showing connection status
- **API Endpoints**: Categorized list of all available endpoints
- **System Information**: Backend URL, version, last update time

**Auto-refresh every 30 seconds** for real-time updates

---

### admin.html
Admin control panel with:
- **Tab 1 - API Testing**
  - Search functionality for APIs
  - Filter by HTTP method (GET, POST, PUT, DELETE)
  - Click-to-test buttons for all endpoints
  - Response display in modal

- **Tab 2 - Logs**
  - System operation log viewer
  - Clear logs button
  - Timestamped entries

- **Tab 3 - Settings**
  - Backend URL configuration
  - Authentication type display
  - API timeout settings
  - Auto-refresh toggle

---

### system-manager.html
System monitoring & control with:
- **Live Status Indicator** - Real-time system health with pulsing dot
- **Tab 1 - Overview**
  - CPU, Memory, Disk, Network usage meters
  - Service status (Running/Stopped)
  - Performance metrics

- **Tab 2 - Databases**
  - Database connection status
  - Create/Backup/Restore operations
  - Database details grid

- **Tab 3 - Monitoring**
  - Performance charts (placeholder)
  - Detailed metrics table
  - Response times, query times, error rates

- **Tab 4 - Operations**
  - Database operations (Optimize, Cleanup, Reindex)
  - System maintenance (Cache, Memory, Logs)
  - System control (Restart, Stop, Emergency Restart)
  - Operation log with timestamps

---

## 🎯 JavaScript Modules

### auth.js
**Authentication Management**
- Keycloak OAuth2 integration
- Token storage in localStorage
- Authorization header injection
- Login/Logout handlers
- Session persistence

**Key Functions:**
- `handleLogin(event)` - Process login form
- `logout()` - Clear session
- `getAuthHeaders()` - Return auth headers
- `isAuthenticated()` - Check token validity
- `isAdmin()` - Check admin role

---

### dashboard.js
**Public Dashboard Logic**
- Auto-load health and database status
- 30-second refresh interval
- Response time calculation
- Database connection counting
- Status color coding

**Key Functions:**
- `loadDashboardData()` - Fetch all data
- `loadHealthStatus()` - Check system health
- `loadDatabaseStatus()` - Get DB connections
- `loadResponseTime()` - Measure backend latency

---

### admin.js
**Admin Interface Logic**
- Load and display all API endpoints
- Search and filter capabilities
- Execute API requests with auth headers
- Display responses in modal
- Log all operations

**Key Functions:**
- `loadAPIs()` - Load API categories
- `filterAPIs()` - Search/filter results
- `testAPI()` - Execute endpoint
- `addLog()` - Add log entry

---

### system-manager.js
**System Manager Logic**
- Load real-time system metrics
- Database status monitoring
- Operation execution and logging
- Tab switching

**Key Functions:**
- `loadStatusData()` - Get system health
- `updateMetrics()` - Simulate metric updates
- `loadDatabases()` - Load database status
- `executeOp()` - Run maintenance operations
- `addOperationLog()` - Log operations

---

## 🎨 Stylesheet Organization

### style.css (1200+ lines)
Main stylesheet with:
- CSS Variables for consistent theming
- Component-specific styles
- Dark mode colors and gradients
- Animations and transitions
- Form styling
- Table styling
- Modal styling
- Card and grid systems
- Badge/status indicators

**CSS Variables:**
```css
--primary-color: #3b82f6
--secondary-color: #8b5cf6
--success-color: #10b981
--warning-color: #f59e0b
--danger-color: #ef4444
--dark-bg: #0f172a
--card-bg: #1e293b
```

### responsive.css (700+ lines)
Responsive design breakpoints:
- **Tablets (≤768px)**
- **Mobile (≤480px)**
- **Extra Small (≤320px)**
- **Landscape orientation**
- **Print styles**
- **Accessibility (prefers-reduced-motion)**
- **Touch device adjustments**
- **High DPI screens**

---

## 🚀 Frontend Routes

| Route | Page | Auth Required | Description |
|-------|------|---------------|-------------|
| `/` | Public Dashboard | No | System status & API list |
| `/admin` | Admin Interface | Optional | API testing & logs |
| `/system-manager` | System Manager | Optional | Monitoring & control |
| `/favicon.ico` | Icon | No | SVG favicon |
| `/api/health` | JSON | No | Backend health proxy |
| `/api/status` | JSON | No | Backend status proxy |
| `/static/*` | Static Files | No | CSS, JS, assets |

---

## 🔄 Main.go Routing

```go
// Public routes
router.GET("/", dashboardHandler)           // Public dashboard
router.GET("/favicon.ico", faviconHandler)  // Favicon

// Protected routes
router.GET("/admin", adminHandler)          // Admin panel
router.GET("/system-manager", systemManagerHandler)

// API proxy routes
router.GET("/api/health", apiHealthHandler)
router.GET("/api/status", apiStatusHandler)

// Static files
router.Static("/static", "templates/")
```

**Rendering Strategy:**
- Load layout.html as base template
- Load page-specific content template
- Render content into layout
- Return complete HTML to client

---

## 🔐 Authentication Flow

1. **Login Modal** appears on page load if not authenticated
2. **User enters** username/password
3. **Keycloak request** sent to backend
4. **Token received** and stored in localStorage
5. **Authorization header** added to all subsequent requests
6. **Page reloads** to show authenticated UI
7. **Admin/System Manager** tabs appear in navigation

---

## 📊 API Integration

### Endpoint Categories Displayed

**Health & Status**
- GET `/health` - System health (no auth)
- GET `/status` - All connections (no auth)

**Notifications**
- POST `/api/notifications/send` - Custom notifications
- POST `/api/notifications/health` - Health checks
- POST `/api/notifications/status` - Status reports

**Database CRUD (All auth required)**
- GET `/api/mysql/users` - List MySQL users
- GET `/api/postgres/users` - List PostgreSQL users
- GET `/api/mongodb/users` - List MongoDB users
- POST endpoints for creating records
- Similar operations for other databases

---

## 🎯 Features & Capabilities

### Public Dashboard
✅ Real-time system status  
✅ Database connection monitoring  
✅ API endpoint discovery  
✅ Response time metrics  
✅ No authentication required  

### Admin Interface
✅ Complete API testing suite  
✅ Search & filter endpoints  
✅ One-click API execution  
✅ Response display with formatting  
✅ Operation logging  
✅ Settings management  

### System Manager
✅ Live system metrics (CPU, Memory, Disk)  
✅ Service status monitoring  
✅ Database management operations  
✅ Performance metrics dashboard  
✅ System maintenance operations  
✅ Operation execution & logging  

---

## 📱 Responsive Breakpoints

| Breakpoint | Screen Size | Adjustments |
|-----------|------------|------------|
| Desktop | > 768px | Full layout, all features |
| Tablet | 481-768px | 2-column grids, simplified nav |
| Mobile | 321-480px | Single column, touch-friendly |
| Small | < 320px | Minimal layout, stacked elements |

---

## 🎨 Color Palette

| Name | Hex | Usage |
|------|-----|-------|
| Primary | #3b82f6 | Links, active states, primary buttons |
| Secondary | #8b5cf6 | Accents, gradient elements |
| Success | #10b981 | Status OK, connected databases |
| Warning | #f59e0b | Warning states, pending operations |
| Danger | #ef4444 | Errors, disconnected, delete actions |
| Dark BG | #0f172a | Main background |
| Card BG | #1e293b | Card backgrounds |
| Text Primary | #f1f5f9 | Main text |
| Text Secondary | #cbd5e1 | Secondary text |

---

## 🛠️ Running the Frontend

```bash
# Start the frontend server
cd frontend
go run main.go

# Frontend runs on http://localhost:7000
# Dashboard: http://localhost:7000/
# Admin: http://localhost:7000/admin
# System Manager: http://localhost:7000/system-manager
```

---

## 📦 Dependencies

- **Gin** - Web framework
- **godotenv** - Environment configuration
- **text/template** - HTML templating

---

## 🔄 Multi-Page Rendering

Each page follows the pattern:
1. Load layout.html (shared structure)
2. Load page-specific HTML (dashboard/admin/manager)
3. Render page content into layout content area
4. Pass template data (auth state, backend URL, etc.)
5. Include page-specific JavaScript

This ensures consistent navigation and styling across all pages while allowing independent functionality.

---

## 📝 Notes

- **Authentication** is optional for Dashboard, required for Admin/Manager
- **Dark theme** optimized for reduced eye strain
- **All endpoints** categorized and easily accessible
- **API responses** displayed with pretty-printing
- **Logs** timestamped for debugging
- **Mobile-first** responsive design
- **Accessibility** features included (reduced motion, keyboard nav)

