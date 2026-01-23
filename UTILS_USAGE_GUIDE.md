# AxiomNizam Utils - Complete Usage Guide

This document provides comprehensive examples for using all utility functions available in `internal/utils/`.

---

## Table of Contents

1. [String Utilities](#string-utilities)
2. [Validators](#validators)
3. [Error Handling](#error-handling)
4. [Formatters](#formatters)
5. [Encryption & Hashing](#encryption--hashing)
6. [HTTP Utilities](#http-utilities)
7. [Database Utilities](#database-utilities)
8. [Bot Protection](#bot-protection)

---

## String Utilities

**File**: `internal/utils/string_utils.go`

String manipulation and analysis functions.

### Basic Operations

```go
import "example.com/axiomnizam/internal/utils"

// Trimming whitespace
utils.TrimSpaces("  hello world  ")          // "hello world"
utils.TrimAllSpaces("h e l l o")              // "hello"

// Case conversion
utils.ToLowerCase("HELLO")                    // "hello"
utils.ToUpperCase("hello")                    // "HELLO"
utils.CapitalizeString("hello world")         // "Hello world"

// Reversing
utils.ReverseString("hello")                  // "olleh"
```

### Substring Operations

```go
// Checking
utils.IsEmpty("   ")                          // true
utils.IsNotEmpty("hello")                     // true
utils.ContainsSubstring("hello", "ell")       // true

// Replacing
utils.ReplaceString("hello world", "world", "there")  // "hello there"

// Splitting & Joining
words := utils.SplitString("a,b,c", ",")     // ["a", "b", "c"]
utils.JoinStrings(words, "-")                 // "a-b-c"

// Prefix/Suffix
utils.HasPrefix("hello", "he")                // true
utils.HasSuffix("hello", "lo")                // true
utils.RemovePrefix("hello", "he")             // "llo"
utils.RemoveSuffix("hello", "lo")             // "hel"
```

### String Analysis

```go
// Length
utils.StringLength("hello")                   // 5
utils.StringLength("你好")                     // 2 (handles multi-byte)

// Counting
utils.CountOccurrences("banana", "a")         // 3

// Truncation
utils.TruncateString("Long text here", 10)   // "Long te..."

// Removing special characters
utils.RemoveSpecialChars("h@llo-w0rld!")      // "helloworld"
```

### String Formatting

```go
// Padding
utils.PadLeft("5", 3, "0")                    // "005"
utils.PadRight("5", 3, "0")                   // "500"

// Multiple replacements
replacements := map[string]string{
    "old": "new",
    "foo": "bar",
}
utils.ReplaceMultiple("old foo text", replacements)  // "new bar text"
```

---

## Validators

**File**: `internal/utils/validators.go`

Input validation for various formats and types.

### Email & Contact Validation

```go
// Email validation
utils.IsValidEmail("user@example.com")        // true
utils.IsValidEmail("invalid.email")           // false

// Phone validation
utils.IsValidPhone("+12125551234")            // true
utils.IsValidPhone("invalid")                 // false
```

### URL & Network Validation

```go
// URL validation
utils.IsValidURL("https://example.com")       // true

// IP address validation
utils.IsValidIPAddress("192.168.1.1")         // true
utils.IsValidIPAddress("999.999.999.999")     // false

// Domain validation
utils.IsValidDomain("example.com")            // true

// Port validation
utils.IsValidPort("8080")                     // true
utils.IsValidPort("99999")                    // false
```

### Format Validation

```go
// UUID validation
utils.IsValidUUID("550e8400-e29b-41d4-a716-446655440000")  // true

// Hex color validation
utils.IsValidHexColor("#FF5733")              // true
utils.IsValidHexColor("FF5733")               // true

// JSON validation
utils.IsValidJSON(`{"key": "value"}`)         // true
utils.IsValidJSON("not json")                 // false
```

### Password & Security Validation

```go
// Strong password (8+ chars, uppercase, lowercase, digit, special)
utils.IsValidPassword("MyP@ssw0rd")           // true
utils.IsValidPassword("weak")                 // false

// Credit card validation (Luhn algorithm)
utils.IsValidCreditCard("4532015112830366")   // true
utils.IsValidCreditCard("0000000000000000")   // false
```

### Username & Data Validation

```go
// Username (3-20 chars, alphanumeric, underscore, hyphen)
utils.IsValidUsername("user_name")            // true
utils.IsValidUsername("ab")                   // false (too short)

// Type validation
utils.IsValidInt("123")                       // true
utils.IsValidFloat("123.45")                  // true
utils.IsValidBoolean("true")                  // true

// Pattern validation
utils.IsValidAlpha("abc")                     // true
utils.IsValidAlphaNumeric("abc123")           // true
utils.IsValidNumeric("12345")                 // true
```

### Length & Slug Validation

```go
// Length validation
utils.IsValidLength("hello", 3, 10)           // true

// URL slug validation
utils.IsValidSlug("my-cool-slug")             // true
utils.IsValidSlug("invalid_slug")             // false
```

### Custom Validators

```go
// Using validator functions
validator := utils.ValidateMinLength(5)
validator("hello")    // nil (valid)
validator("hi")       // error (too short)

// Email validator
utils.ValidateEmail("user@example.com")       // nil
utils.ValidateEmail("invalid")                // error

// Required field validator
utils.ValidateRequired("value")               // nil
utils.ValidateRequired("")                    // error
```

---

## Error Handling

**File**: `internal/utils/error.go`

Custom error types and error handling utilities.

### Creating Custom Errors

```go
import "example.com/axiomnizam/internal/utils"

// Validation error
validationErr := utils.NewValidationError(
    "Email is invalid",
    map[string]string{"field": "email"},
)

// Not found error
notFoundErr := utils.NewNotFoundError("User not found")

// Unauthorized error
unauthorizedErr := utils.NewUnauthorizedError("Invalid credentials")

// Forbidden error
forbiddenErr := utils.NewForbiddenError("Access denied")

// Conflict error
conflictErr := utils.NewConflictError("User already exists", nil)

// Internal server error
internalErr := utils.NewInternalError("Database error", originalErr)

// Database error
dbErr := utils.NewDatabaseError("Failed to query", dbError)

// Timeout error
timeoutErr := utils.NewTimeoutError("Request timeout")
```

### Chaining Error Details

```go
err := utils.NewInternalError("Query failed", originalErr).
    WithDetails(map[string]interface{}{
        "query": "SELECT * FROM users",
        "table": "users",
    })
```

### Error Checking & Conversion

```go
// Check if error is custom error
if utils.IsCustomError(err) {
    customErr := err.(*utils.CustomError)
    fmt.Println(customErr.Type)      // ErrorTypeInternal
    fmt.Println(customErr.Message)   // "Query failed"
    fmt.Println(customErr.StatusCode) // 500
}

// Convert any error to custom error
customErr := utils.AsCustomError(err)
```

### API Error Response

```go
// Convert to API response format
customErr := utils.NewValidationError("Invalid input", nil)
apiResponse := customErr.ToErrorResponse()
// Returns: {"error": "VALIDATION_ERROR", "message": "Invalid input", ...}
```

### Multiple Validation Errors

```go
// Collect multiple validation errors
errors := utils.NewValidationErrors()
errors.AddError("email", "Invalid email format", "user@invalid")
errors.AddError("password", "Password too short", "123")

if errors.HasErrors() {
    customErr := errors.ToCustomError()
    // Now can be sent as API response
}
```

---

## Formatters

**File**: `internal/utils/formatters.go`

Format various data types for display and logging.

### Byte & Size Formatting

```go
import "example.com/axiomnizam/internal/utils"

utils.FormatBytes(1024)                       // "1.00 KB"
utils.FormatBytes(1048576)                    // "1.00 MB"
utils.FormatBytes(1073741824)                 // "1.00 GB"
```

### Time & Duration Formatting

```go
// Duration formatting
utils.FormatDuration(500 * time.Millisecond)  // "500ms"
utils.FormatDuration(45 * time.Second)        // "45.00s"
utils.FormatDuration(2 * time.Minute)         // "2.00m"

// Time formatting
now := time.Now()
utils.FormatTime(now)                         // ISO 8601 format
utils.FormatDate(now)                         // "2026-01-23"
utils.FormatDateTime(now)                     // "2026-01-23 14:30:45"
utils.FormatTimeCustom(now, "02/01/2006")    // "23/01/2026"
```

### Number Formatting

```go
utils.FormatNumber(1000)                      // "1000"
utils.FormatPercentage(95.5)                  // "95.50%"
utils.FormatCurrency(99.99, "$")              // "$99.99"
utils.FormatCount(1500)                       // "1.5K"
utils.FormatCount(5000000)                    // "5.0M"
```

### Personal Data Masking

```go
// Phone number formatting
utils.FormatPhoneNumber("1234567890")         // "(123) 456-7890"

// Social Security Number masking
utils.FormatSSN("123456789")                  // "123-45-6789"

// Credit card masking
utils.FormatCreditCard("4532015112830366")    // "****830366"

// Email masking
utils.FormatEmail("user@example.com")         // "u***r@example.com"
```

### Text Formatting

```go
// Title case
utils.FormatTitle("hello world")              // "Hello World"

// Case conversions
utils.FormatCamelCase("hello world")          // "helloWorld"
utils.FormatPascalCase("hello world")         // "HelloWorld"
utils.FormatSnakeCase("hello world")          // "hello_world"
utils.FormatKebabCase("hello world")          // "hello-world"

// Key formatting
utils.FormatKey("User Name")                  // "user_name"
```

### URL & Path Formatting

```go
// URL formatting
utils.FormatURL("https://example.com?key=value")  // "https://example.com"

// Path formatting
utils.FormatPath("C:\\Users\\Name\\file.txt")     // "C:/Users/Name/file.txt"
```

### JSON Formatting

```go
data := map[string]interface{}{
    "name": "John",
    "age": 30,
}
utils.FormatJSON(data)
// Returns:
// {
//   "age": 30,
//   "name": "John"
// }
```

### Network Metrics Formatting

```go
utils.FormatLatency(45.5)                     // "45.50ms"
utils.FormatLatency(1500)                     // "1.50s"

utils.FormatBitrate(1000000)                  // "1000.00 Kbps"
utils.FormatBitrate(50000000)                 // "50.00 Mbps"

utils.FormatUptime(86400)                     // "24h"
```

---

## Encryption & Hashing

**File**: `internal/utils/encryption.go`

Secure password hashing and token encryption.

### Password Hashing

```go
import "example.com/axiomnizam/internal/utils"

// Using password hasher
hasher := utils.NewPasswordHasher()

// Hash password
hash, err := hasher.HashPassword("myPassword123!")
if err != nil {
    log.Fatal(err)
}
// hash: "$2a$10$..."

// Verify password
err = hasher.VerifyPassword("myPassword123!", hash)
if err != nil {
    log.Println("Password mismatch")
}

// Quick check
isValid := hasher.IsPasswordValid("myPassword123!", hash)  // true
```

### Custom Cost Password Hasher

```go
// Higher cost = slower but more secure (takes longer to hash)
// Useful for important operations
hasher := utils.NewPasswordHasherWithCost(14)

hash, err := hasher.HashPassword("password")
// Will take longer but more resistant to brute force
```

### Quick Password Functions

```go
// Convenience functions (default cost)
hash, err := utils.HashPassword("myPassword")
isValid := utils.VerifyPassword("myPassword", hash)
```

### Token Encryption

```go
// Generate a 32-byte key (save this securely!)
key, err := utils.GenerateSecureKey(32)
if err != nil {
    log.Fatal(err)
}

// Create token encryptor
te, err := utils.NewTokenEncryption(key)
if err != nil {
    log.Fatal(err)
}

// Encrypt token
token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
encrypted, err := te.EncryptToken(token)
if err != nil {
    log.Fatal(err)
}
// encrypted: "base64_encoded_encrypted_data"

// Decrypt token
decrypted, err := te.DecryptToken(encrypted)
if err != nil {
    log.Fatal(err)
}
// decrypted: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### String Encryption

```go
key, _ := utils.GenerateSecureKey(32)

// Encrypt
encrypted, err := utils.EncryptString("secret data", key)
if err != nil {
    log.Fatal(err)
}

// Decrypt
plaintext, err := utils.DecryptString(encrypted, key)
if err != nil {
    log.Fatal(err)
}
```

### Token Generation

```go
// Generate random token (base64 URL-safe)
token, err := utils.GenerateRandomToken(32)
if err != nil {
    log.Fatal(err)
}
// token: "SGVsbG8gV29ybGQgSGVsbG8gV29ybGQ="

// Generate secure key
key, err := utils.GenerateSecureKey(32)
if err != nil {
    log.Fatal(err)
}
```

### Simple Hash (Not for passwords!)

```go
// For checksums, not password hashing!
hash := utils.SimpleHash("my string")
// hash: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"
```

---

## HTTP Utilities

**File**: `internal/utils/http.go`

Request parsing and response building for HTTP APIs.

### Response Building

```go
import "example.com/axiomnizam/internal/utils"

// Build success response
response := utils.NewResponseBuilder().
    WithStatus(200).
    WithMessage("User created successfully").
    WithData(map[string]interface{}{
        "id": 123,
        "name": "John",
        "email": "john@example.com",
    }).
    WithHeader("X-Custom-Header", "value").
    Success()

// In handler:
json.NewEncoder(w).Encode(response)
```

### Response Writing

```go
// Using ResponseWriter in HTTP handler
func MyHandler(w http.ResponseWriter, r *http.Request) {
    rw := utils.NewResponseWriter(w)
    
    // Write JSON success
    rw.WriteSuccess(200, map[string]interface{}{
        "id": 123,
        "name": "John",
    })
    
    // Or write JSON error
    rw.WriteError(400, "Invalid input")
    
    // Or write text
    rw.WriteText(200, "Hello World")
    
    // Or write HTML
    rw.WriteHTML(200, "<h1>Hello</h1>")
    
    // Redirect
    rw.Redirect(301, "https://example.com")
}
```

### Request Parsing

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    parser := utils.NewRequestParser(r)
    
    // Get query parameters
    userID := parser.GetQueryParam("user_id")      // "123"
    limit, _ := parser.GetQueryParamInt("limit")   // 50
    isActive, _ := parser.GetQueryParamBool("active")  // true
    
    // Get headers
    authHeader := parser.GetAuthHeader()            // "Bearer token123"
    bearerToken := parser.GetBearerToken()          // "token123"
    userAgent := parser.GetUserAgent()              // "Mozilla/5.0..."
    contentType := parser.GetContentType()          // "application/json"
    
    // Get request info
    method := parser.GetMethod()                    // "GET"
    path := parser.GetPath()                        // "/api/users"
    ip := parser.GetIP()                            // "192.168.1.100"
    
    // Parse JSON body
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    parser.GetJSON(&user)
    
    // Get all query params
    allParams := parser.GetAllQueryParams()
    
    // Check for param
    if parser.HasQueryParam("filter") {
        // ...
    }
}
```

### Content Type Checking

```go
parser := utils.NewRequestParser(r)

if parser.IsJSON() {
    // Parse as JSON
    var data map[string]interface{}
    parser.GetJSON(&data)
}

if parser.IsForm() {
    // Parse as form
    name := parser.GetFormParam("name")
}
```

---

## Database Utilities

**File**: `internal/utils/database.go`

Connection pooling and query building helpers.

### Connection Pool

```go
import "example.com/axiomnizam/internal/utils"

// Create pool
pool := utils.NewConnectionPool()

// Add connections
pool.Add("mysql", mysqlDB)
pool.Add("postgres", postgresDB)

// Get connection
mysqlConn, err := pool.Get("mysql")

// Get all
allConns := pool.GetAll()

// Check size
size := pool.Size()  // 2

// Close all
pool.Close()
```

### Query Builder (SELECT)

```go
// Simple SELECT
qb := utils.NewQueryBuilder().
    Select("id", "name", "email").
    From("users").
    OrderBy("name ASC").
    Limit(10)

query, params := qb.Build()
// SELECT id, name, email FROM users ORDER BY name ASC LIMIT 10

// With WHERE
qb := utils.NewQueryBuilder().
    Select("*").
    From("users").
    Where("age > ?", 18).
    AndWhere("city = ?", "NYC").
    OrderBy("id DESC").
    Limit(20)

// With JOIN
qb := utils.NewQueryBuilder().
    Select("u.id", "u.name", "p.title").
    From("users u").
    LeftJoin("posts p", "u.id = p.user_id").
    Where("u.active = ?", true).
    OrderBy("u.created_at DESC")

// With GROUP BY
qb := utils.NewQueryBuilder().
    Select("category", "COUNT(*) as count").
    From("products").
    GroupBy("category").
    Having("COUNT(*) > ?", 5).
    OrderBy("count DESC")

// Complex query
qb := utils.NewQueryBuilder().
    Distinct().
    Select("u.id", "u.name", "COUNT(p.id) as posts").
    From("users u").
    LeftJoin("posts p", "u.id = p.user_id").
    Where("u.created_at > ?", "2025-01-01").
    GroupBy("u.id", "u.name").
    Having("COUNT(p.id) > ?", 3).
    OrderBy("posts DESC").
    Limit(100).
    Offset(0)

query, params := qb.Build()
// Pass to database:
// db.Raw(query, params...).Scan(&result)
```

### INSERT Builder

```go
// Single insert
ib := utils.NewInsertBuilder("users").
    Columns("name", "email", "age").
    Values("John", "john@example.com", 28)

query, params := ib.BuildInsert()
// INSERT INTO users (name, email, age) VALUES (?, ?, ?)
// params: ["John", "john@example.com", 28]

db.Exec(query, params...)
```

### UPDATE Builder

```go
ub := utils.NewUpdateBuilder("users").
    Set("name", "Jane").
    Set("age", 29).
    Where("id = ?", 123)

query, params := ub.BuildUpdate()
// UPDATE users SET name = ?, age = ? WHERE id = ?
// params: ["Jane", 29, 123]

db.Exec(query, params...)
```

### DELETE Builder

```go
db := utils.NewDeleteBuilder("users").
    Where("age < ?", 18).
    AndWhere("status != ?", "active")

query, params := db.BuildDelete()
// DELETE FROM users WHERE age < ? AND status != ?
// params: [18, "active"]

db.Exec(query, params...)
```

---

## Bot Protection

**File**: `internal/utils/bot_protection.go`

Comprehensive bot detection and protection utilities.

### Bot Detection

```go
import "example.com/axiomnizam/internal/utils"

detector := utils.NewBotDetector()

// Check if request looks like a bot
if detector.IsLikelyBot(r) {
    return http.StatusForbidden
}

// Check if it's a known bot
if detector.IsKnownBot(r) {
    // Maybe allow or log differently
}

// Check if it's a search engine bot
if detector.IsKnownSearchBot(r) {
    // May want to allow with different rate limit
}

// Add custom suspicious patterns
detector.AddSuspiciousPattern("mybot")
detector.AddBotUserAgent("custombot/1.0")
```

### Rate Limiting with Bot Protection

```go
// Create limiter with 1-minute window
limiter := utils.NewBotRateLimiter(1 * time.Minute)

// Check and increment
clientIP := "192.168.1.100"
maxRequests := 100
blockDuration := 5 * time.Minute

if !limiter.CheckAndIncrement(clientIP, maxRequests, blockDuration) {
    return http.StatusTooManyRequests  // Blocked
}

// Check if currently blocked
if limiter.IsBlocked(clientIP) {
    return http.StatusTooManyRequests
}

// Reset limit
limiter.Reset(clientIP)
```

### IP Blacklist/Whitelist

```go
blacklist := utils.NewIPBlacklist()

// Add malicious IPs to blacklist
blacklist.AddToBlacklist("192.168.1.100")
blacklist.AddToBlacklist("10.0.0.50")

// Whitelist trusted IPs (cannot be blacklisted)
blacklist.AddToWhitelist("203.0.113.50")

// Check if IP is blocked
clientIP := "192.168.1.100"
if blacklist.IsBlacklisted(clientIP) {
    return http.StatusForbidden
}

// Remove from lists
blacklist.RemoveFromBlacklist("192.168.1.100")
blacklist.RemoveFromWhitelist("203.0.113.50")
```

### Honeypot Fields

```go
// In HTML form, add hidden field
honeypot := utils.NewHoneypotField("website_url")

// In handler, check if filled
if honeypot.IsHoneypotFilled(r.FormValue("website_url")) {
    return http.StatusForbidden  // Bot detected
}
```

### Request Fingerprinting

```go
// Generate fingerprint
fp := utils.GenerateFingerprint(r)

// Get hash for comparison
hash := fp.GetHash()

// Store or compare with known hashes
// Useful for detecting multiple requests from same source
```

### Behavior Analysis

```go
analyzer := utils.NewBehaviorAnalyzer()

// Record requests
clientIP := "192.168.1.100"
analyzer.RecordRequest(clientIP, "/api/users", false)
analyzer.RecordRequest(clientIP, "/api/posts", false)
analyzer.RecordRequest(clientIP, "/api/invalid", true)

// Get suspicion score
score := analyzer.GetSuspicionScore(clientIP)

// Check if suspicious
if analyzer.IsSuspicious(clientIP, 50) {  // threshold 50
    return http.StatusForbidden
}

// Reset behavior history
analyzer.Reset(clientIP)
```

### User Agent Validation

```go
validator := utils.NewUserAgentValidator()

ua := r.Header.Get("User-Agent")
if !validator.IsValidUserAgent(ua) {
    return http.StatusForbidden  // Likely a bot
}
```

### Header Anomaly Detection

```go
validator := utils.NewHeaderValidator()

if validator.HasAnomalies(r) {
    return http.StatusForbidden
}
```

### SQL Injection Detection

```go
sqlDetector := utils.NewSQLInjectionDetector()

userInput := r.FormValue("search")
if sqlDetector.IsSuspiciousInput(userInput) {
    return http.StatusBadRequest
}
```

### XSS Detection

```go
xssDetector := utils.NewXSSDetector()

userComment := r.FormValue("comment")
if xssDetector.IsSuspiciousInput(userComment) {
    return http.StatusBadRequest
}
```

### Complete Bot Protection Middleware Example

```go
func BotProtectionMiddleware(next http.Handler) http.Handler {
    detector := utils.NewBotDetector()
    limiter := utils.NewBotRateLimiter(1 * time.Minute)
    blacklist := utils.NewIPBlacklist()
    analyzer := utils.NewBehaviorAnalyzer()
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        clientIP := utils.GetClientIP(r)
        
        // Check blacklist
        if blacklist.IsBlacklisted(clientIP) {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        // Check rate limit
        if !limiter.CheckAndIncrement(clientIP, 100, 5*time.Minute) {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        
        // Check if bot
        if detector.IsLikelyBot(r) {
            blacklist.AddToBlacklist(clientIP)
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        // Record behavior
        analyzer.RecordRequest(clientIP, r.URL.Path, false)
        if analyzer.IsSuspicious(clientIP, 50) {
            blacklist.AddToBlacklist(clientIP)
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

## Complete Integration Example

```go
package handlers

import (
    "net/http"
    "example.com/axiomnizam/internal/utils"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    // Parse request
    parser := utils.NewRequestParser(r)
    
    var user struct {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := parser.GetJSON(&user); err != nil {
        rw := utils.NewResponseWriter(w)
        rw.WriteError(400, "Invalid JSON")
        return
    }
    
    // Validate input
    errors := utils.NewValidationErrors()
    
    if !utils.IsValidUsername(user.Name) {
        errors.AddError("name", "Invalid username format", user.Name)
    }
    
    if !utils.IsValidEmail(user.Email) {
        errors.AddError("email", "Invalid email format", user.Email)
    }
    
    if !utils.IsValidPassword(user.Password) {
        errors.AddError("password", "Password too weak", "")
    }
    
    if errors.HasErrors() {
        rw := utils.NewResponseWriter(w)
        rw.WriteError(400, "Validation failed")
        return
    }
    
    // Hash password
    hasher := utils.NewPasswordHasher()
    hash, err := hasher.HashPassword(user.Password)
    if err != nil {
        rw := utils.NewResponseWriter(w)
        rw.WriteError(500, "Error processing password")
        return
    }
    
    // Create user in database
    // ... database code ...
    
    // Format response
    rw := utils.NewResponseWriter(w)
    rw.WriteSuccess(201, map[string]interface{}{
        "id":    123,
        "name":  user.Name,
        "email": user.Email,
    })
}
```

---

## Best Practices

1. **Always validate user input** using appropriate validators
2. **Hash passwords** with strong cost settings
3. **Use custom errors** for consistent error handling
4. **Implement bot protection** early in middleware
5. **Log suspicious activities** for security monitoring
6. **Use encryption** for sensitive data
7. **Format output** consistently for APIs
8. **Pool database connections** for better performance

---

## Quick Reference

| Utility | Best For |
|---------|----------|
| String Utils | Text manipulation |
| Validators | Input validation |
| Error Handling | API errors |
| Formatters | Display & logging |
| Encryption | Security |
| HTTP Utils | Request/Response |
| Database Utils | SQL queries |
| Bot Protection | Security |

---

**Version**: 1.0  
**Last Updated**: January 23, 2026  
**Status**: Production Ready
