package celestrak

import (
	"fmt"
	"net/http"
)

// ErrorResponse represents an error response from the Celestrak API.
type ErrorResponse struct {
	Response *http.Response
	Message  string
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("celestrak: %s: %s", r.Response.Status, r.Message)
}

// IsNotFound returns true if the error is a 404 Not Found.
func (r *ErrorResponse) IsNotFound() bool {
	return r.Response != nil && r.Response.StatusCode == http.StatusNotFound
}

// IsRateLimit returns true if the error is a rate limit (429 Too Many Requests).
func (r *ErrorResponse) IsRateLimit() bool {
	return r.Response != nil && r.Response.StatusCode == http.StatusTooManyRequests
}

// IsServerError returns true if the error is a 5xx server error.
func (r *ErrorResponse) IsServerError() bool {
	return r.Response != nil && r.Response.StatusCode >= 500
}

// IsClientError returns true if the error is a 4xx client error.
func (r *ErrorResponse) IsClientError() bool {
	return r.Response != nil && r.Response.StatusCode >= 400 && r.Response.StatusCode < 500
}

// QueryError represents an error in query construction or validation.
type QueryError struct {
	Message string
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("celestrak: query error: %s", e.Message)
}

// IsQueryError checks if an error is a QueryError.
func IsQueryError(err error) bool {
	_, ok := err.(*QueryError)
	return ok
}

// IsErrorResponse checks if an error is an ErrorResponse.
func IsErrorResponse(err error) bool {
	_, ok := err.(*ErrorResponse)
	return ok
}
