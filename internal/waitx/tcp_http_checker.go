package waitx

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// TCPChecker waits for TCP connectivity.
type TCPChecker struct {
	Address     string
	DialTimeout time.Duration
}

func (c TCPChecker) Name() string {
	return "tcp:" + strings.TrimSpace(c.Address)
}

func (c TCPChecker) Check(_ context.Context) error {
	address := strings.TrimSpace(c.Address)
	if address == "" {
		return fmt.Errorf("tcp address is required")
	}
	timeout := c.DialTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("tcp dial failed: %w", err)
	}
	_ = conn.Close()
	return nil
}

// HTTPChecker waits for HTTP endpoint conditions.
type HTTPChecker struct {
	URL              string
	Method           string
	Headers          map[string]string
	Body             string
	ExpectStatusCode int
	ExpectBodyRegex  *regexp.Regexp
	InsecureSkipTLS  bool
	RequestTimeout   time.Duration
}

func (c HTTPChecker) Name() string {
	return "http:" + strings.TrimSpace(c.URL)
}

func (c HTTPChecker) Check(ctx context.Context) error {
	url := strings.TrimSpace(c.URL)
	if url == "" {
		return fmt.Errorf("http url is required")
	}

	method := strings.ToUpper(strings.TrimSpace(c.Method))
	if method == "" {
		method = http.MethodGet
	}

	bodyReader := io.Reader(nil)
	if c.Body != "" {
		bodyReader = strings.NewReader(c.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("http request build failed: %w", err)
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	transport := &http.Transport{}
	if c.InsecureSkipTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{
		Transport: transport,
	}
	if c.RequestTimeout > 0 {
		client.Timeout = c.RequestTimeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if c.ExpectStatusCode > 0 && resp.StatusCode != c.ExpectStatusCode {
		return fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, c.ExpectStatusCode)
	}

	if c.ExpectBodyRegex != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("http response read failed: %w", readErr)
		}
		if !c.ExpectBodyRegex.Match(body) {
			return fmt.Errorf("response body did not match regex %q", c.ExpectBodyRegex.String())
		}
	}

	return nil
}
