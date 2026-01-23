package utils

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// BotDetector detects and blocks bot traffic
type BotDetector struct {
	suspiciousPatterns []string
	botUserAgents      []string
	mu                 sync.RWMutex
}

// NewBotDetector creates a new bot detector
func NewBotDetector() *BotDetector {
	return &BotDetector{
		suspiciousPatterns: []string{
			"curl", "wget", "python", "perl", "php", "java", "node",
			"bot", "crawler", "spider", "scraper", "urllib", "httplib",
			"scrapy", "selenium", "puppeteer", "playwright",
		},
		botUserAgents: []string{
			"googlebot", "bingbot", "yandexbot", "baidubot",
			"facebookexternalhit", "twitterbot", "linkedinbot",
			"slurp", "duckduckbot", "baiduspider", "sogou",
			"bytespider", "semrushbot", "ahrefs", "mj12bot",
		},
	}
}

// IsLikelyBot checks if request is likely from a bot
func (bd *BotDetector) IsLikelyBot(r *http.Request) bool {
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))

	// Check if User-Agent is empty or suspicious
	if userAgent == "" {
		return true
	}

	// Check against bot patterns
	if bd.matchesPattern(userAgent, bd.suspiciousPatterns) {
		return true
	}

	// Check for missing standard headers that real browsers have
	if r.Header.Get("Accept-Language") == "" && r.Header.Get("Accept-Encoding") == "" {
		return true
	}

	// Check for suspicious header combinations
	if bd.hasNoReferer(r) && bd.hasNoAcceptLanguage(r) {
		return true
	}

	return false
}

// IsKnownBot checks if request is from a known bot (search engines, etc.)
func (bd *BotDetector) IsKnownBot(r *http.Request) bool {
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))
	return bd.matchesPattern(userAgent, bd.botUserAgents)
}

// IsKnownSearchBot checks if request is from a search engine bot
func (bd *BotDetector) IsKnownSearchBot(r *http.Request) bool {
	searchBots := []string{
		"googlebot", "bingbot", "yandexbot", "baidubot",
		"slurp", "duckduckbot", "baiduspider", "sogou",
	}
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))
	return bd.matchesPattern(userAgent, searchBots)
}

// matchesPattern checks if text matches any pattern
func (bd *BotDetector) matchesPattern(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// hasNoReferer checks if request has no referer
func (bd *BotDetector) hasNoReferer(r *http.Request) bool {
	return r.Header.Get("Referer") == "" && r.Header.Get("Referrer") == ""
}

// hasNoAcceptLanguage checks if request has no accept-language
func (bd *BotDetector) hasNoAcceptLanguage(r *http.Request) bool {
	return r.Header.Get("Accept-Language") == ""
}

// AddSuspiciousPattern adds a custom suspicious pattern
func (bd *BotDetector) AddSuspiciousPattern(pattern string) {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	bd.suspiciousPatterns = append(bd.suspiciousPatterns, pattern)
}

// AddBotUserAgent adds a custom bot user agent
func (bd *BotDetector) AddBotUserAgent(agent string) {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	bd.botUserAgents = append(bd.botUserAgents, agent)
}

// RateLimiter for bot protection
type BotRateLimiter struct {
	limits map[string]*RateLimit
	mu     sync.RWMutex
	window time.Duration
}

// RateLimit stores rate limit information
type RateLimit struct {
	Count      int
	FirstSeen  time.Time
	LastSeen   time.Time
	Blocked    bool
	BlockUntil time.Time
}

// NewBotRateLimiter creates a new bot rate limiter
func NewBotRateLimiter(window time.Duration) *BotRateLimiter {
	return &BotRateLimiter{
		limits: make(map[string]*RateLimit),
		window: window,
	}
}

// CheckAndIncrement checks rate limit and increments counter
func (brl *BotRateLimiter) CheckAndIncrement(key string, maxRequests int, blockDuration time.Duration) bool {
	brl.mu.Lock()
	defer brl.mu.Unlock()

	limit, exists := brl.limits[key]

	// Check if currently blocked
	if exists && limit.Blocked {
		if time.Now().Before(limit.BlockUntil) {
			return false // Still blocked
		}
		// Block period expired, reset
		limit.Blocked = false
		limit.Count = 0
	}

	now := time.Now()

	if !exists {
		// First request from this key
		brl.limits[key] = &RateLimit{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}
		return true
	}

	// Check if window has expired
	if now.Sub(limit.FirstSeen) > brl.window {
		// Reset counter
		limit.Count = 1
		limit.FirstSeen = now
		limit.LastSeen = now
		return true
	}

	// Increment counter
	limit.Count++
	limit.LastSeen = now

	// Check if exceeded limit
	if limit.Count > maxRequests {
		limit.Blocked = true
		limit.BlockUntil = now.Add(blockDuration)
		return false
	}

	return true
}

// IsBlocked checks if key is currently blocked
func (brl *BotRateLimiter) IsBlocked(key string) bool {
	brl.mu.RLock()
	defer brl.mu.RUnlock()

	limit, exists := brl.limits[key]
	if !exists {
		return false
	}

	if !limit.Blocked {
		return false
	}

	if time.Now().After(limit.BlockUntil) {
		return false
	}

	return true
}

// Reset resets the limit for a key
func (brl *BotRateLimiter) Reset(key string) {
	brl.mu.Lock()
	defer brl.mu.Unlock()
	delete(brl.limits, key)
}

// IPBlacklist manages IP-based blocking
type IPBlacklist struct {
	blacklist map[string]bool
	whitelist map[string]bool
	mu        sync.RWMutex
}

// NewIPBlacklist creates a new IP blacklist
func NewIPBlacklist() *IPBlacklist {
	return &IPBlacklist{
		blacklist: make(map[string]bool),
		whitelist: make(map[string]bool),
	}
}

// AddToBlacklist adds an IP to blacklist
func (ipb *IPBlacklist) AddToBlacklist(ip string) {
	ipb.mu.Lock()
	defer ipb.mu.Unlock()
	ipb.blacklist[ip] = true
}

// AddToWhitelist adds an IP to whitelist
func (ipb *IPBlacklist) AddToWhitelist(ip string) {
	ipb.mu.Lock()
	defer ipb.mu.Unlock()
	ipb.whitelist[ip] = true
}

// RemoveFromBlacklist removes IP from blacklist
func (ipb *IPBlacklist) RemoveFromBlacklist(ip string) {
	ipb.mu.Lock()
	defer ipb.mu.Unlock()
	delete(ipb.blacklist, ip)
}

// RemoveFromWhitelist removes IP from whitelist
func (ipb *IPBlacklist) RemoveFromWhitelist(ip string) {
	ipb.mu.Lock()
	defer ipb.mu.Unlock()
	delete(ipb.whitelist, ip)
}

// IsBlacklisted checks if IP is blacklisted
func (ipb *IPBlacklist) IsBlacklisted(ip string) bool {
	ipb.mu.RLock()
	defer ipb.mu.RUnlock()

	// Whitelisted IPs cannot be blacklisted
	if ipb.whitelist[ip] {
		return false
	}

	return ipb.blacklist[ip]
}

// IsWhitelisted checks if IP is whitelisted
func (ipb *IPBlacklist) IsWhitelisted(ip string) bool {
	ipb.mu.RLock()
	defer ipb.mu.RUnlock()
	return ipb.whitelist[ip]
}

// CAPTCHAValidator validates CAPTCHA responses (interface for extensibility)
type CAPTCHAValidator interface {
	Validate(token string) (bool, error)
}

// HoneypotField detects honeypot field fills (bot indicator)
type HoneypotField struct {
	fieldName string
}

// NewHoneypotField creates a new honeypot field
func NewHoneypotField(fieldName string) *HoneypotField {
	return &HoneypotField{fieldName: fieldName}
}

// IsHoneypotFilled checks if honeypot field was filled (indicates bot)
func (hf *HoneypotField) IsHoneypotFilled(value string) bool {
	return value != "" && value != "0"
}

// RequestFingerprint creates a fingerprint of a request
type RequestFingerprint struct {
	IP        string
	UserAgent string
	Headers   string
	Timestamp int64
}

// GenerateFingerprint generates a fingerprint for a request
func GenerateFingerprint(r *http.Request) *RequestFingerprint {
	headerStr := fmt.Sprintf(
		"%s-%s-%s",
		r.Header.Get("Accept"),
		r.Header.Get("Accept-Language"),
		r.Header.Get("Accept-Encoding"),
	)

	return &RequestFingerprint{
		IP:        GetClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Headers:   headerStr,
		Timestamp: time.Now().Unix(),
	}
}

// GetClientIP extracts client IP from request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// GetFingerprintHash returns MD5 hash of fingerprint
func (rf *RequestFingerprint) GetHash() string {
	data := fmt.Sprintf("%s:%s:%s", rf.IP, rf.UserAgent, rf.Headers)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// BehaviorAnalyzer detects suspicious behavior patterns
type BehaviorAnalyzer struct {
	requestPatterns map[string]*BehaviorPattern
	mu              sync.RWMutex
}

// BehaviorPattern tracks behavior for an entity
type BehaviorPattern struct {
	Requests        int
	FirstSeen       time.Time
	LastSeen        time.Time
	UniqueEndpoints map[string]int
	ErrorCount      int
	SuspicionScore  int
}

// NewBehaviorAnalyzer creates a new behavior analyzer
func NewBehaviorAnalyzer() *BehaviorAnalyzer {
	return &BehaviorAnalyzer{
		requestPatterns: make(map[string]*BehaviorPattern),
	}
}

// RecordRequest records a request for behavior analysis
func (ba *BehaviorAnalyzer) RecordRequest(key string, endpoint string, hasError bool) *BehaviorPattern {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	pattern, exists := ba.requestPatterns[key]
	if !exists {
		pattern = &BehaviorPattern{
			UniqueEndpoints: make(map[string]int),
			FirstSeen:       time.Now(),
		}
		ba.requestPatterns[key] = pattern
	}

	pattern.Requests++
	pattern.LastSeen = time.Now()
	pattern.UniqueEndpoints[endpoint]++

	if hasError {
		pattern.ErrorCount++
	}

	// Update suspicion score
	ba.updateSuspicionScore(pattern)

	return pattern
}

// updateSuspicionScore updates the suspicion score based on behavior
func (ba *BehaviorAnalyzer) updateSuspicionScore(pattern *BehaviorPattern) {
	score := 0

	// Too many requests in short time
	duration := pattern.LastSeen.Sub(pattern.FirstSeen).Seconds()
	if duration < 60 && pattern.Requests > 50 {
		score += 30
	}

	// Too many different endpoints accessed
	if len(pattern.UniqueEndpoints) > 50 {
		score += 25
	}

	// High error rate
	if pattern.Requests > 0 {
		errorRate := float64(pattern.ErrorCount) / float64(pattern.Requests)
		if errorRate > 0.5 {
			score += 20
		}
	}

	// Accessing endpoints sequentially (pattern of a bot)
	if pattern.Requests > 100 {
		score += 15
	}

	pattern.SuspicionScore = score
}

// GetSuspicionScore returns the suspicion score for a key
func (ba *BehaviorAnalyzer) GetSuspicionScore(key string) int {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	pattern, exists := ba.requestPatterns[key]
	if !exists {
		return 0
	}

	return pattern.SuspicionScore
}

// IsSuspicious checks if behavior is suspicious (score > threshold)
func (ba *BehaviorAnalyzer) IsSuspicious(key string, threshold int) bool {
	return ba.GetSuspicionScore(key) > threshold
}

// Reset clears behavior history for a key
func (ba *BehaviorAnalyzer) Reset(key string) {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	delete(ba.requestPatterns, key)
}

// UserAgentValidator validates user agent strings
type UserAgentValidator struct {
	validPatterns   []string
	invalidPatterns []string
}

// NewUserAgentValidator creates a new user agent validator
func NewUserAgentValidator() *UserAgentValidator {
	return &UserAgentValidator{
		validPatterns: []string{
			"mozilla", "chrome", "safari", "firefox", "edge", "opera",
		},
		invalidPatterns: []string{
			"curl", "wget", "python", "perl", "java",
		},
	}
}

// IsValidUserAgent checks if user agent is valid
func (uav *UserAgentValidator) IsValidUserAgent(ua string) bool {
	if ua == "" {
		return false
	}

	ua = strings.ToLower(ua)

	// Check for invalid patterns
	for _, pattern := range uav.invalidPatterns {
		if strings.Contains(ua, pattern) {
			return false
		}
	}

	// Check for valid patterns
	hasValid := false
	for _, pattern := range uav.validPatterns {
		if strings.Contains(ua, pattern) {
			hasValid = true
			break
		}
	}

	return hasValid
}

// HeaderValidator validates HTTP headers for anomalies
type HeaderValidator struct{}

// NewHeaderValidator creates a new header validator
func NewHeaderValidator() *HeaderValidator {
	return &HeaderValidator{}
}

// HasAnomalies checks if request headers have anomalies
func (hv *HeaderValidator) HasAnomalies(r *http.Request) bool {
	// Check for missing critical headers
	if r.Header.Get("Accept") == "" {
		return true
	}

	// Check for suspicious header values
	if hv.hasSuspiciousLength(r.Header.Get("User-Agent")) {
		return true
	}

	// Check for malformed headers
	if hv.hasMalformedAccept(r.Header.Get("Accept")) {
		return true
	}

	return false
}

// hasSuspiciousLength checks if header value has suspicious length
func (hv *HeaderValidator) hasSuspiciousLength(header string) bool {
	// User-Agent should be between 50-500 chars for legitimate browsers
	length := len(header)
	return length < 20 || length > 500
}

// hasMalformedAccept checks if Accept header is malformed
func (hv *HeaderValidator) hasMalformedAccept(accept string) bool {
	if accept == "" {
		return true
	}

	// Should contain common mime types
	validMimeTypes := []string{"text/html", "application/json", "application/", "text/"}
	found := false

	for _, mimeType := range validMimeTypes {
		if strings.Contains(accept, mimeType) {
			found = true
			break
		}
	}

	return !found
}

// SQLInjectionDetector detects SQL injection attempts
type SQLInjectionDetector struct {
	patterns []*regexp.Regexp
}

// NewSQLInjectionDetector creates a new SQL injection detector
func NewSQLInjectionDetector() *SQLInjectionDetector {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`),
		regexp.MustCompile(`(?i)(-{2}|;|/\*|\*/)`),
		regexp.MustCompile(`(?i)(or|and)\s*('|")?[^'"]+(=|<|>)`),
		regexp.MustCompile(`(?i)(xp_|sp_)`),
	}

	return &SQLInjectionDetector{patterns: patterns}
}

// IsSuspiciousInput checks if input looks like SQL injection
func (sid *SQLInjectionDetector) IsSuspiciousInput(input string) bool {
	for _, pattern := range sid.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// XSSDetector detects XSS injection attempts
type XSSDetector struct {
	patterns []*regexp.Regexp
}

// NewXSSDetector creates a new XSS detector
func NewXSSDetector() *XSSDetector {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)<iframe[^>]*>`),
		regexp.MustCompile(`(?i)<embed[^>]*>`),
		regexp.MustCompile(`(?i)<object[^>]*>`),
	}

	return &XSSDetector{patterns: patterns}
}

// IsSuspiciousInput checks if input looks like XSS
func (xd *XSSDetector) IsSuspiciousInput(input string) bool {
	for _, pattern := range xd.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}
