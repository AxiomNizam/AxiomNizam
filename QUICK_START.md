# Quick Start Guide - AxiomNizam Frontend

## The Problem & Solution

### What Was Fixed:
1. **Data Not Loading** → Backend CORS + Frontend fetch improvements
2. **No Theme Support** → Added 3 themes (Dark/Light/Default)

---

## How to Run

### Backend
```bash
cd AxiomNizam
go run main.go
# Runs on http://localhost:8000
```

### Frontend
```bash
cd AxiomNizam/frontend
go run main.go
# Runs on http://localhost:7000
```

### Open in Browser
```
http://localhost:7000
```

---

## What You'll See

### Dashboard Loading Data
✅ System Status: **OK** or **OFFLINE**
✅ Databases Connected: **0-6** (counts connected databases)
✅ Response Time: **--ms** (when backend responds)
✅ Uptime: **--** (system uptime)
✅ Database Connections: Shows each DB status (✅ or ❌)

### Theme Button
🌙 **Dark** - Blue + Dark Gray (default)
☀️ **Light** - Professional light colors
💻 **Default** - Cyan + Indigo blend

Click the icon in top-right to switch themes!

---

## API Endpoints Being Queried

### Health Check (No Auth Required)
```
GET http://localhost:8000/health
Response: {"status":"ok","message":"AxiomNizam API is running"}
```

### System Status (No Auth Required)
```
GET http://localhost:8000/status
Response: {
    "status":"ok",
    "message":"System status",
    "data":{
        "mysql":"connected",
        "postgres":"connected",
        "mongodb":"disconnected",
        ...
    }
}
```

---

## Debugging

### Check Backend Connectivity
```bash
# In PowerShell
curl.exe http://localhost:8000/health
curl.exe http://localhost:8000/status
```

### Check Frontend Console
1. Open browser (F12)
2. Go to **Console** tab
3. Look for logs like:
   ```
   Backend URL: http://localhost:8000
   Fetching health from: http://localhost:8000/health
   Health data: {status: "ok", ...}
   ```

### Check Network Calls
1. Open browser (F12)
2. Go to **Network** tab
3. Should see requests to:
   - `http://localhost:8000/health`
   - `http://localhost:8000/status`
4. Both should return **Status 200**

---

## Theme Persistence

Themes are saved to browser's localStorage:
- Close and reopen browser → Theme stays same
- Clear cache → Resets to Dark theme
- Works offline ✅

---

## If Data Still Shows "Loading..."

1. **Check backend is running:**
   ```bash
   curl.exe http://localhost:8000/health
   ```

2. **Check browser console (F12):**
   - Look for red errors
   - Common: "Failed to fetch" or "Cannot read property"

3. **Check network tab (F12):**
   - Verify requests are going to correct URL
   - Verify responses return HTTP 200
   - Check response body is valid JSON

4. **Restart both:**
   ```bash
   # Stop backend & frontend (Ctrl+C)
   # Run backend first, then frontend
   ```

---

## File Structure

```
AxiomNizam/
├── main.go                          # Backend API
├── go.mod, go.sum                  # Backend dependencies
├── frontend/
│   ├── main.go                     # Frontend server
│   ├── go.mod, go.sum             # Frontend dependencies
│   └── templates/
│       ├── layout.html             # Navigation + Theme toggle
│       ├── public-dashboard.html   # Dashboard page
│       ├── admin.html              # Admin page
│       ├── system-manager.html     # System manager page
│       ├── style.css               # Styles (with theme variables)
│       ├── dashboard.js            # Dashboard data loading ✨
│       ├── admin.js                # Admin logic
│       ├── system-manager.js       # System manager logic
│       └── auth.js                 # Authentication
└── FRONTEND_FIXES.md               # Detailed fix documentation
```

---

## What Changed

### Backend (main.go)
- ✅ Added CORS middleware for cross-origin requests
- ✅ Enabled all HTTP methods (GET, POST, PUT, DELETE, OPTIONS)

### Frontend (dashboard.js)
- ✅ Fixed fetch requests with `mode: 'cors'`
- ✅ Better error handling for API responses
- ✅ Proper data extraction from backend responses
- ✅ Console logging for debugging

### Frontend (style.css)
- ✅ Added theme CSS variables (3 complete palettes)
- ✅ Smooth transitions between themes
- ✅ All components respect theme variables

### Frontend (layout.html)
- ✅ Added theme toggle button
- ✅ Theme persistence logic
- ✅ Theme icons that change with selection

---

## Test It Out!

1. Start backend: `go run main.go` (in backend folder)
2. Start frontend: `go run main.go` (in frontend/folder)
3. Open http://localhost:7000
4. Watch data load automatically! 📊
5. Click theme button to change colors! 🎨

Enjoy! 🚀
