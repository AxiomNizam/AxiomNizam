package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Response represents a standard API response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// ResponseBuilder builds HTTP responses
type ResponseBuilder struct {
	status  int
	headers map[string]string
	data    interface{}
	error   string
	message string
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		status:  http.StatusOK,
		headers: make(map[string]string),
	}
}

// WithStatus sets the HTTP status code
func (rb *ResponseBuilder) WithStatus(status int) *ResponseBuilder {
	rb.status = status
	return rb
}

// WithHeader adds a custom header
func (rb *ResponseBuilder) WithHeader(key, value string) *ResponseBuilder {
	rb.headers[key] = value
	return rb
}

// WithHeaders adds multiple headers
func (rb *ResponseBuilder) WithHeaders(headers map[string]string) *ResponseBuilder {
	for key, value := range headers {
		rb.headers[key] = value
	}
	return rb
}

// WithData sets the response data
func (rb *ResponseBuilder) WithData(data interface{}) *ResponseBuilder {
	rb.data = data
	return rb
}

// WithMessage sets the response message
func (rb *ResponseBuilder) WithMessage(message string) *ResponseBuilder {
	rb.message = message
	return rb
}

// WithError sets the error message
func (rb *ResponseBuilder) WithError(error string) *ResponseBuilder {
	rb.error = error
	return rb
}

// Success creates a success response
func (rb *ResponseBuilder) Success() *Response {
	return &Response{
		Status:  "success",
		Message: rb.message,
		Data:    rb.data,
		Code:    rb.status,
	}
}

// Error creates an error response
func (rb *ResponseBuilder) ErrorResponse() *Response {
	return &Response{
		Status:  "error",
		Message: rb.message,
		Error:   rb.error,
		Code:    rb.status,
	}
}

// Build writes the response to http.ResponseWriter
func (rb *ResponseBuilder) Build(w http.ResponseWriter, isError bool) error {
	// Set default content type
	if _, ok := rb.headers["Content-Type"]; !ok {
		rb.headers["Content-Type"] = "application/json"
	}

	// Apply headers
	for key, value := range rb.headers {
		w.Header().Set(key, value)
	}

	// Set status
	w.WriteHeader(rb.status)

	// Build response
	var response *Response
	if isError {
		response = rb.ErrorResponse()
	} else {
		response = rb.Success()
	}

	// Encode response
	return json.NewEncoder(w).Encode(response)
}

// RequestParser parses HTTP requests
type RequestParser struct {
	r *http.Request
}

// NewRequestParser creates a new request parser
func NewRequestParser(r *http.Request) *RequestParser {
	return &RequestParser{r: r}
}

// GetJSON parses JSON body into interface
func (rp *RequestParser) GetJSON(v interface{}) error {
	defer rp.r.Body.Close()
	return json.NewDecoder(rp.r.Body).Decode(v)
}

// GetQueryParam gets a query parameter
func (rp *RequestParser) GetQueryParam(key string) string {
	return rp.r.URL.Query().Get(key)
}

// GetQueryParamInt gets a query parameter as integer
func (rp *RequestParser) GetQueryParamInt(key string) (int, error) {
	value := rp.r.URL.Query().Get(key)
	if value == "" {
		return 0, fmt.Errorf("query parameter '%s' not found", key)
	}
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("query parameter '%s' is not a valid integer", key)
	}
	return result, nil
}

// GetQueryParamBool gets a query parameter as boolean
func (rp *RequestParser) GetQueryParamBool(key string) (bool, error) {
	value := rp.r.URL.Query().Get(key)
	if value == "" {
		return false, fmt.Errorf("query parameter '%s' not found", key)
	}
	return strings.ToLower(value) == "true", nil
}

// GetFormParam gets a form parameter
func (rp *RequestParser) GetFormParam(key string) string {
	return rp.r.FormValue(key)
}

// GetHeader gets a request header
func (rp *RequestParser) GetHeader(key string) string {
	return rp.r.Header.Get(key)
}

// GetAuthHeader gets the Authorization header
func (rp *RequestParser) GetAuthHeader() string {
	return rp.r.Header.Get("Authorization")
}

// GetBearerToken extracts bearer token from Authorization header
func (rp *RequestParser) GetBearerToken() string {
	authHeader := rp.r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetContentType gets the Content-Type header
func (rp *RequestParser) GetContentType() string {
	return rp.r.Header.Get("Content-Type")
}

// IsJSON checks if request content type is JSON
func (rp *RequestParser) IsJSON() bool {
	contentType := rp.GetContentType()
	return strings.Contains(contentType, "application/json")
}

// IsForm checks if request content type is form-encoded
func (rp *RequestParser) IsForm() bool {
	contentType := rp.GetContentType()
	return strings.Contains(contentType, "application/x-www-form-urlencoded")
}

// GetBody reads the entire request body
func (rp *RequestParser) GetBody() ([]byte, error) {
	defer rp.r.Body.Close()
	return io.ReadAll(rp.r.Body)
}

// GetMethod returns the HTTP method
func (rp *RequestParser) GetMethod() string {
	return rp.r.Method
}

// GetPath returns the request path
func (rp *RequestParser) GetPath() string {
	return rp.r.URL.Path
}

// GetURL returns the full request URL
func (rp *RequestParser) GetURL() *url.URL {
	return rp.r.URL
}

// GetIP gets the client IP address
func (rp *RequestParser) GetIP() string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := rp.r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := rp.r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return strings.Split(rp.r.RemoteAddr, ":")[0]
}

// GetUserAgent gets the User-Agent header
func (rp *RequestParser) GetUserAgent() string {
	return rp.r.Header.Get("User-Agent")
}

// GetReferer gets the Referer header
func (rp *RequestParser) GetReferer() string {
	return rp.r.Header.Get("Referer")
}

// GetAllQueryParams returns all query parameters
func (rp *RequestParser) GetAllQueryParams() url.Values {
	return rp.r.URL.Query()
}

// HasQueryParam checks if query parameter exists
func (rp *RequestParser) HasQueryParam(key string) bool {
	return rp.r.URL.Query().Has(key)
}

// GetQueryParams gets multiple query parameters
func (rp *RequestParser) GetQueryParams(keys ...string) map[string]string {
	params := make(map[string]string)
	for _, key := range keys {
		params[key] = rp.r.URL.Query().Get(key)
	}
	return params
}

// ResponseWriter wraps http.ResponseWriter for easier response writing
type ResponseWriter struct {
	w http.ResponseWriter
}

// NewResponseWriter creates a new response writer wrapper
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w: w}
}

// WriteJSON writes a JSON response
func (rw *ResponseWriter) WriteJSON(status int, data interface{}) error {
	rw.w.Header().Set("Content-Type", "application/json")
	rw.w.WriteHeader(status)
	return json.NewEncoder(rw.w).Encode(data)
}

// WriteJSONWithHeader writes JSON with custom headers
func (rw *ResponseWriter) WriteJSONWithHeader(status int, headers map[string]string, data interface{}) error {
	for key, value := range headers {
		rw.w.Header().Set(key, value)
	}
	return rw.WriteJSON(status, data)
}

// WriteText writes a text response
func (rw *ResponseWriter) WriteText(status int, text string) error {
	rw.w.Header().Set("Content-Type", "text/plain")
	rw.w.WriteHeader(status)
	_, err := rw.w.Write([]byte(text))
	return err
}

// WriteHTML writes an HTML response
func (rw *ResponseWriter) WriteHTML(status int, html string) error {
	rw.w.Header().Set("Content-Type", "text/html")
	rw.w.WriteHeader(status)
	_, err := rw.w.Write([]byte(html))
	return err
}

// WriteError writes an error response
func (rw *ResponseWriter) WriteError(status int, errorMsg string) error {
	errResponse := map[string]interface{}{
		"status": "error",
		"error":  errorMsg,
		"code":   status,
	}
	return rw.WriteJSON(status, errResponse)
}

// WriteSuccess writes a success response
func (rw *ResponseWriter) WriteSuccess(status int, data interface{}) error {
	successResponse := map[string]interface{}{
		"status": "success",
		"data":   data,
		"code":   status,
	}
	return rw.WriteJSON(status, successResponse)
}

// SetHeader sets a response header
func (rw *ResponseWriter) SetHeader(key, value string) {
	rw.w.Header().Set(key, value)
}

// SetHeaders sets multiple response headers
func (rw *ResponseWriter) SetHeaders(headers map[string]string) {
	for key, value := range headers {
		rw.w.Header().Set(key, value)
	}
}

// Redirect writes a redirect response
func (rw *ResponseWriter) Redirect(status int, location string) {
	rw.w.Header().Set("Location", location)
	rw.w.WriteHeader(status)
}

// WriteBytes writes raw bytes with status
func (rw *ResponseWriter) WriteBytes(status int, contentType string, data []byte) error {
	rw.w.Header().Set("Content-Type", contentType)
	rw.w.WriteHeader(status)
	_, err := rw.w.Write(data)
	return err
}
