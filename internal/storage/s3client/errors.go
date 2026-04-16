package s3client

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// S3Error represents an error response from an S3-compatible service.
type S3Error struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`
	Message    string   `xml:"Message"`
	Resource   string   `xml:"Resource"`
	RequestID  string   `xml:"RequestId"`
	StatusCode int      `xml:"-"`
}

func (e *S3Error) Error() string {
	return fmt.Sprintf("s3 error: %s (%s) status=%d resource=%s", e.Message, e.Code, e.StatusCode, e.Resource)
}

// parseS3Error reads an S3 XML error response from the given reader.
func parseS3Error(resp *http.Response) error {
	if resp.Body == nil {
		return &S3Error{
			Code:       http.StatusText(resp.StatusCode),
			Message:    "empty response body",
			StatusCode: resp.StatusCode,
		}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return &S3Error{
			Code:       "ReadError",
			Message:    fmt.Sprintf("failed to read error body: %v", err),
			StatusCode: resp.StatusCode,
		}
	}

	var s3Err S3Error
	if xml.Unmarshal(body, &s3Err) != nil {
		return &S3Error{
			Code:       http.StatusText(resp.StatusCode),
			Message:    string(body),
			StatusCode: resp.StatusCode,
		}
	}
	s3Err.StatusCode = resp.StatusCode
	return &s3Err
}

// isNotFoundError returns true if the error represents a 404 Not Found.
func isNotFoundError(err error) bool {
	if s3Err, ok := err.(*S3Error); ok {
		return s3Err.StatusCode == http.StatusNotFound ||
			s3Err.Code == "NoSuchBucket" ||
			s3Err.Code == "NoSuchKey"
	}
	return false
}

// isBucketExistsError returns true if the error represents a bucket already existing.
func isBucketExistsError(err error) bool {
	if s3Err, ok := err.(*S3Error); ok {
		return s3Err.Code == "BucketAlreadyOwnedByYou" || s3Err.Code == "BucketAlreadyExists"
	}
	return false
}
