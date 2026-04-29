// Package validate provides shared input validation helpers for REST
// handlers across the platform.
//
// All path parameters, query parameters, and user-provided identifiers
// should be validated through these helpers before use.
package validate

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// resourceNameRe matches DNS-1123 subdomain names (K8s convention).
// Allows lowercase alphanumeric, hyphens, dots. Max 253 chars.
var resourceNameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9.\-]*[a-z0-9])?$`)

// maxResourceNameLen is the K8s-standard max for resource names.
const maxResourceNameLen = 253

// PathParam extracts and validates a path parameter.
// Returns the trimmed value, or writes a 400 response and returns "".
func PathParam(c *gin.Context, key string) string {
	val := strings.TrimSpace(c.Param(key))
	if val == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("path parameter '%s' is required", key)})
		return ""
	}
	if len(val) > maxResourceNameLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("path parameter '%s' exceeds max length %d", key, maxResourceNameLen)})
		return ""
	}
	return val
}

// ResourceName extracts a path parameter and validates it as a valid
// resource name (DNS-1123 subdomain: lowercase alphanumeric + hyphens/dots).
func ResourceName(c *gin.Context, key string) string {
	val := PathParam(c, key)
	if val == "" {
		return ""
	}
	if !resourceNameRe.MatchString(val) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid resource name '%s': must match [a-z0-9][a-z0-9.-]*[a-z0-9]", val),
		})
		return ""
	}
	return val
}

// PathParamInt extracts a path parameter and parses it as an integer.
func PathParamInt(c *gin.Context, key string) (int64, bool) {
	val := PathParam(c, key)
	if val == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("path parameter '%s' must be an integer", key)})
		return 0, false
	}
	return n, true
}

// QueryString extracts and trims a query parameter. Returns "" if not present (no error).
func QueryString(c *gin.Context, key string) string {
	return strings.TrimSpace(c.Query(key))
}

// QueryInt extracts a query parameter as an integer. Returns the default if not present.
func QueryInt(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

// RequiredBody binds JSON and validates required fields via binding tags.
// Returns false and writes a 400 response if validation fails.
func RequiredBody[T any](c *gin.Context, dest *T) bool {
	if err := c.ShouldBindJSON(dest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
		return false
	}
	return true
}

// StringNotEmpty checks that a string field is not empty after trimming.
func StringNotEmpty(val, fieldName string) error {
	if strings.TrimSpace(val) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// StringMaxLen checks that a string doesn't exceed a maximum length.
func StringMaxLen(val, fieldName string, max int) error {
	if len(val) > max {
		return fmt.Errorf("%s exceeds maximum length %d", fieldName, max)
	}
	return nil
}
