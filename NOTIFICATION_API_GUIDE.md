# 🔔 Notification API - Discord Integration Guide

**Date**: January 22, 2026  
**Status**: ✅ Ready for Testing  
**Authentication**: Keycloak JWT (Required)

---

## 📋 Overview

The Notification API enables authenticated users to send real-time notifications to Discord. Notifications include system health checks, status reports, and custom messages. All notifications are sent to a configured Discord webhook with beautiful formatted embeds.

**Features**:
- ✅ Keycloak authentication required
- ✅ Discord webhook integration
- ✅ Health & status monitoring
- ✅ Custom notifications
- ✅ Color-coded message types
- ✅ Automatic database status inclusion
- ✅ Timestamped messages

---

## 🔐 Security

✅ **Authentication Requirements**:
- Valid JWT token from Keycloak
- Token validation on all endpoints (except GET /api/notifications/status)
- Webhook URL stored securely in environment variables

---

## 🚀 Endpoints

### 1️⃣ Send Custom Notification

**Endpoint**: `POST /api/notifications/send`  
**Authentication**: ✅ Required (any authenticated user)

**Request**:
```json
{
  "title": "System Alert",
  "message": "Custom notification message",
  "type": "info",
  "include_data": true
}
```

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| title | string | Yes | Notification title |
| message | string | Yes | Notification body |
| type | string | No | Type: info, success, warning, error (default: info) |
| include_data | boolean | No | Include health/status data (default: false) |

**Notification Types & Colors**:
| Type | Color | Hex | Use Case |
|------|-------|-----|----------|
| info | Blue | #3447003 | General information |
| success | Green | #65280 | Successful operations |
| warning | Yellow | #16776960 | Warnings |
| error | Red | #16711680 | Errors/failures |

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "Notification sent to Discord",
  "title": "System Alert"
}
```

**Discord Embed Output**:
```
━━━━━━━━━━━━━━━━━━━━━━━
  🔔 System Alert - info
━━━━━━━━━━━━━━━━━━━━━━━
Custom notification message

System Status: healthy
Timestamp: 2026-01-22T10:30:45Z
━━━━━━━━━━━━━━━━━━━━━━━
```

---

### 2️⃣ Send Health Check Notification

**Endpoint**: `POST /api/notifications/health`  
**Authentication**: ✅ Required (any authenticated user)

**Request**: No body required

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "Health notification sent",
  "health_data": {
    "timestamp": "2026-01-22T10:30:45Z",
    "status": "healthy",
    "databases": {
      "mysql": "✅ connected",
      "mariadb": "✅ connected",
      "postgres": "✅ connected",
      "percona": "✅ connected",
      "oracle": "✅ connected"
    }
  }
}
```

**Discord Embed Output**:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━
  🏥 System Health Check
━━━━━━━━━━━━━━━━━━━━━━━━━━
Automated health status notification

mysql Status: ✅ connected
mariadb Status: ✅ connected
postgres Status: ✅ connected
percona Status: ✅ connected
oracle Status: ✅ connected

Timestamp: 2026-01-22T10:30:45Z
━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### 3️⃣ Send Status Report Notification

**Endpoint**: `POST /api/notifications/status`  
**Authentication**: ✅ Required (any authenticated user)

**Request**: No body required

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "Status notification sent",
  "status_data": {
    "timestamp": "2026-01-22T10:30:45Z",
    "status": "healthy",
    "databases": {
      "mysql": "✅ connected",
      "mariadb": "✅ connected",
      "postgres": "✅ connected",
      "percona": "✅ connected",
      "oracle": "✅ connected"
    }
  }
}
```

---

### 4️⃣ Get Notification Service Status

**Endpoint**: `GET /api/notifications/status`  
**Authentication**: ❌ Not Required

**Response** (200 OK):
```json
{
  "status": "active",
  "webhook_url": "https://discord.com/api/webhooks/...",
  "notification_types": [
    "custom",
    "health",
    "status"
  ],
  "supported_types": [
    "info",
    "success",
    "warning",
    "error"
  ]
}
```

---

## 💻 PowerShell Examples

### Get Admin Token

```powershell
$adminToken = (Invoke-RestMethod -Uri "http://localhost:8080/realms/master/protocol/openid-connect/token" `
  -Method POST `
  -ContentType "application/x-www-form-urlencoded" `
  -Body @{
    client_id     = "axiomnizam"
    client_secret = "uzqxRJUEI44gpURiytWtCujKwQ1ESZrv"
    grant_type    = "password"
    username      = "admin"
    password      = "admin"
  }).access_token

Write-Host "✅ Token obtained"
```

### Send Custom Notification

```powershell
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type"  = "application/json"
}

$body = @{
    title       = "Database Backup"
    message     = "Nightly backup completed successfully"
    type        = "success"
    include_data = $true
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/send" `
    -Method POST `
    -Headers $headers `
    -Body $body

Write-Host "✅ Notification sent: $($response.message)"
```

### Send Health Check Notification

```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/health" `
    -Method POST `
    -Headers $headers

Write-Host "✅ Health notification sent"
Write-Host "Status: $($response.health_data.status)"
foreach ($db in $response.health_data.databases.PSObject.Properties) {
    Write-Host "  $($db.Name): $($db.Value)"
}
```

### Send Status Report Notification

```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/status" `
    -Method POST `
    -Headers $headers

Write-Host "✅ Status report sent"
Write-Host "Report Time: $($response.status_data.timestamp)"
```

### Check Service Status

```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/status" `
    -Method GET

Write-Host "Service Status: $($response.status)"
Write-Host "Webhook URL: $($response.webhook_url)"
Write-Host "Supported Types: $($response.supported_types -join ', ')"
```

---

## 🧪 Test Scenarios

### Test 1: Send Success Notification ✅

```powershell
$body = @{
    title   = "Deployment Successful"
    message = "API v2.0 deployed to production"
    type    = "success"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/send" `
    -Method POST `
    -Headers $headers `
    -Body $body

# Expected: 200 OK with success message in Discord
```

### Test 2: Send Warning Notification ⚠️

```powershell
$body = @{
    title   = "High Memory Usage"
    message = "Database server memory at 85%"
    type    = "warning"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/send" `
    -Method POST `
    -Headers $headers `
    -Body $body

# Expected: 200 OK with yellow embed in Discord
```

### Test 3: Send Error Notification ❌

```powershell
$body = @{
    title   = "Connection Failed"
    message = "Failed to connect to Oracle database"
    type    = "error"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/send" `
    -Method POST `
    -Headers $headers `
    -Body $body

# Expected: 200 OK with red embed in Discord
```

### Test 4: Send Health Check ✅

```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/health" `
    -Method POST `
    -Headers $headers

# Expected: 200 OK with all database statuses
# Discord: 🏥 System Health Check with all DB statuses
```

### Test 5: Send Status Report ✅

```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/status" `
    -Method POST `
    -Headers $headers

# Expected: 200 OK with current system status
# Discord: 📊 System Status Report
```

### Test 6: Unauthenticated Request (Should Fail) ❌

```powershell
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/send" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body '{"title":"Test","message":"Test"}'
} catch {
    # Expected: 401 Unauthorized
    Write-Host "✅ Correctly rejected: $($_.Exception.Response.StatusCode)"
}
```

### Test 7: Check Service Status (No Auth) ✅

```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/status" `
    -Method GET

# Expected: 200 OK with service info
Write-Host "Status: $($response.status)"
```

---

## 📊 Complete Test Workflow

```powershell
function Test-NotificationAPI {
    param([string]$Token)
    
    $baseUrl = "http://localhost:8000"
    $headers = @{
        "Authorization" = "Bearer $Token"
        "Content-Type"  = "application/json"
    }
    
    Write-Host "`n🔔 Notification API Testing" -ForegroundColor Yellow
    Write-Host "=" * 60
    
    # Test 1: Custom Success
    Write-Host "`n[1/7] Custom Success Notification" -ForegroundColor Green
    try {
        $body = @{
            title   = "Test Success"
            message = "This is a success test"
            type    = "success"
        } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/notifications/send" `
            -Method POST -Headers $headers -Body $body
        Write-Host "✅ Sent: $($response.message)"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 2: Custom Warning
    Write-Host "`n[2/7] Custom Warning Notification" -ForegroundColor Yellow
    try {
        $body = @{
            title   = "Test Warning"
            message = "This is a warning test"
            type    = "warning"
        } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/notifications/send" `
            -Method POST -Headers $headers -Body $body
        Write-Host "✅ Sent: $($response.message)"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 3: Custom Error
    Write-Host "`n[3/7] Custom Error Notification" -ForegroundColor Red
    try {
        $body = @{
            title   = "Test Error"
            message = "This is an error test"
            type    = "error"
        } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/notifications/send" `
            -Method POST -Headers $headers -Body $body
        Write-Host "✅ Sent: $($response.message)"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 4: Health Check
    Write-Host "`n[4/7] Health Check Notification" -ForegroundColor Green
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/notifications/health" `
            -Method POST -Headers $headers
        Write-Host "✅ Sent: Health check with $($response.health_data.databases.Count) databases"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 5: Status Report
    Write-Host "`n[5/7] Status Report Notification" -ForegroundColor Green
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/notifications/status" `
            -Method POST -Headers $headers
        Write-Host "✅ Sent: Status report at $($response.status_data.timestamp)"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 6: With Health Data
    Write-Host "`n[6/7] Notification with Health Data" -ForegroundColor Green
    try {
        $body = @{
            title       = "API Status"
            message     = "Current API health snapshot"
            type        = "info"
            include_data = $true
        } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/notifications/send" `
            -Method POST -Headers $headers -Body $body
        Write-Host "✅ Sent: $($response.message)"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    # Test 7: Check Service Status (No Auth)
    Write-Host "`n[7/7] Service Status Check (No Auth)" -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/notifications/status" -Method GET
        Write-Host "✅ Service is $($response.status)"
        Write-Host "   Types available: $($response.notification_types -join ', ')"
    } catch {
        Write-Host "❌ Failed: $($_.Exception.Message)"
    }
    
    Write-Host "`n" + "=" * 60
    Write-Host "🎉 Notification API Testing Complete!" -ForegroundColor Yellow
}

# Run tests
Test-NotificationAPI -Token $token
```

---

## 🔗 Discord Setup

### Step 1: Create Webhook
1. Go to Discord Server → Settings → Webhooks
2. Click "New Webhook"
3. Copy the webhook URL
4. Paste in `.env` as `DISCORD_WEBHOOK_URL`

### Step 2: Test Webhook
```powershell
$webhookURL = "https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN"

$body = @{
    content = "🔔 AxiomNizam Notification Test"
    embeds = @(@{
        title = "Test Message"
        description = "Webhook is working!"
        color = 3447003
    })
} | ConvertTo-Json

Invoke-RestMethod -Uri $webhookURL -Method POST -ContentType "application/json" -Body $body
```

---

## 🗄️ Database Status Monitoring

When `include_data=true`, notifications include real-time database status:

```json
{
  "timestamp": "2026-01-22T10:30:45Z",
  "status": "healthy",
  "databases": {
    "mysql": "✅ connected",
    "mariadb": "✅ connected",
    "postgres": "✅ connected",
    "percona": "✅ connected",
    "oracle": "✅ connected"
  }
}
```

**Status Values**:
- `✅ connected` - Database is reachable
- `❌ error` - Connection failed
- `⚠️ not configured` - Database not configured

---

## 🛡️ Best Practices

✅ **Do**:
- Include meaningful titles and messages
- Use appropriate notification types (info, success, warning, error)
- Enable health data for critical notifications
- Log notification sending in your application
- Test webhooks before deployment

❌ **Don't**:
- Send too many notifications (use rate limiting)
- Expose webhook URL in logs or client code
- Send sensitive data in notifications
- Forget to validate token before sending
- Use hardcoded webhook URLs

---

## 🐛 Troubleshooting

### Error: "Invalid request: title required"
**Cause**: Missing required field  
**Solution**: Include both `title` and `message` in request

### Error: "Invalid notification type: xxx"
**Cause**: Invalid type specified  
**Solution**: Use one of: info, success, warning, error

### Error: "failed to send to Discord: 401"
**Cause**: Invalid webhook URL  
**Solution**: Check DISCORD_WEBHOOK_URL in .env

### Error: "failed to send to Discord: 429"
**Cause**: Rate limited by Discord  
**Solution**: Wait before sending more notifications (Discord limit: 10 per 10 seconds)

### No notification received
**Cause**: Multiple possible issues  
**Solution**: 
1. Check webhook URL is correct
2. Verify Discord channel exists and is accessible
3. Check application logs for errors
4. Test with GET /api/notifications/status

---

## 📚 Integration Examples

### Scheduled Health Checks

```powershell
# Schedule PowerShell task to send health every hour
$taskAction = New-ScheduledTaskAction -Execute powershell.exe `
    -Argument "-File C:\scripts\send-health-notification.ps1"

$taskTrigger = New-ScheduledTaskTrigger -At 0am -RepetitionInterval (New-TimeSpan -Hours 1) -RepetitionDuration (New-TimeSpan -Days 365)

Register-ScheduledTask -TaskName "AxiomNizam-Health-Check" -Action $taskAction -Trigger $taskTrigger
```

### Error Handler Integration

```powershell
function Send-ErrorNotification {
    param([string]$ErrorMessage)
    
    $body = @{
        title   = "API Error"
        message = $ErrorMessage
        type    = "error"
    } | ConvertTo-Json
    
    Invoke-RestMethod -Uri "http://localhost:8000/api/notifications/send" `
        -Method POST `
        -Headers @{"Authorization"="Bearer $token";"Content-Type"="application/json"} `
        -Body $body
}

# Usage
try {
    # Your code here
} catch {
    Send-ErrorNotification -ErrorMessage $_.Exception.Message
}
```

---

## 📊 Summary

| Feature | Status | Details |
|---------|--------|---------|
| Custom Notifications | ✅ | info, success, warning, error types |
| Health Check | ✅ | Real-time database status |
| Status Reports | ✅ | System snapshot |
| Discord Integration | ✅ | Webhook-based delivery |
| Authentication | ✅ | Keycloak JWT required |
| Rate Limiting | ⚠️ | Discord API limits apply |
| Logging | ✅ | All notifications logged |

---

**Configuration**: Add `DISCORD_WEBHOOK_URL` to `.env`  
**Implementation Date**: January 22, 2026  
**Status**: ✅ Complete & Ready  
**Testing**: Use Test Workflow above  
**Production**: Ready to deploy! 🚀
