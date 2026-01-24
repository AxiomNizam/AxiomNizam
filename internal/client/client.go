package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents the AxiomNizam API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	userAgent  string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "axiomnizamctl/1.0.0",
	}
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetBaseURL changes the base URL
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// Request represents an API request
type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers map[string]string
	Body    interface{}
}

// Response represents an API response
type Response struct {
	StatusCode int
	Status     string
	Header     http.Header
	Body       []byte
}

// Do executes an API request
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	// Build URL
	u, err := url.Parse(c.baseURL + req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	if req.Query != nil {
		u.RawQuery = req.Query.Encode()
	}

	// Prepare body
	var body io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("User-Agent", c.userAgent)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Add authorization token
	if c.token != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	// Add custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	response := &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Header:     httpResp.Header,
		Body:       respBody,
	}

	// Check for errors
	if httpResp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return response, fmt.Errorf("API error (%d): %s", httpResp.StatusCode, errResp.Message)
		}
		return response, fmt.Errorf("API error (%d): %s", httpResp.StatusCode, string(respBody))
	}

	return response, nil
}

// ErrorResponse represents an error from the API
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Get makes a GET request
func (c *Client) Get(ctx context.Context, path string, query url.Values) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
	})
}

// Post makes a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
}

// Put makes a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPut,
		Path:   path,
		Body:   body,
	})
}

// Delete makes a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodDelete,
		Path:   path,
	})
}

// ParseResponse parses the response body into a struct
func (resp *Response) ParseJSON(v interface{}) error {
	return json.Unmarshal(resp.Body, v)
}

// String returns the response body as string
func (resp *Response) String() string {
	return string(resp.Body)
}

// HealthCheck checks if the server is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	resp, err := c.Get(ctx, "/health", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server not healthy: %s", resp.Status)
	}

	return nil
}
