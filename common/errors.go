package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError is a typed representation of a non-2xx HTTP response from the
// Proof API. It captures everything callers commonly need to log, branch on,
// or surface to end users without re-parsing raw responses at every call site.
type APIError struct {
	StatusCode int
	Status     string
	Method     string
	URL        string
	Headers    http.Header
	// Body is the full response body (already drained). Safe to read multiple
	// times. Empty if the response had no body.
	Body []byte
	// Message is a best-effort extraction of a human-readable error from the
	// response body. The Proof APIs use a few different shapes
	// ({"error": "..."}, {"message": "..."}, {"errors": [...]}); this picks
	// the first one that parses. Empty if nothing could be extracted.
	Message string
	// BodyReadErr captures any error encountered while draining the response
	// body. Non-nil means Body may be partial or empty; callers that care
	// about the raw payload should check this before trusting Body.
	BodyReadErr error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("proof api: %s %s: %d %s: %s", e.Method, e.URL, e.StatusCode, e.Status, e.Message)
	}
	return fmt.Sprintf("proof api: %s %s: %d %s", e.Method, e.URL, e.StatusCode, e.Status)
}

// CheckResponse is the errors.As-friendly form of AsAPIError. It returns
// nil for 2xx/3xx responses and a non-nil error (concretely *APIError) for
// status >= 400, so callers can write:
//
//	if err := common.CheckResponse(resp); err != nil {
//	    var apiErr *common.APIError
//	    if errors.As(err, &apiErr) { ... }
//	}
//
// Body-handling semantics match AsAPIError: on a non-nil return, the body
// has been drained and replaced with a fresh reader over the captured bytes.
func CheckResponse(resp *http.Response) error {
	if apiErr, ok := AsAPIError(resp); ok {
		return apiErr
	}
	return nil
}

// AsAPIError inspects an *http.Response and returns an *APIError if the
// status is >= 400, along with true. For 2xx/3xx responses it returns
// (nil, false) and leaves the response body untouched. When it returns an
// error, the caller should NOT read resp.Body — AsAPIError has drained and
// closed it (and replaced resp.Body with an io.NopCloser over the captured
// bytes so downstream code that still tries can read the same payload).
func AsAPIError(resp *http.Response) (*APIError, bool) {
	if resp == nil || resp.StatusCode < 400 {
		return nil, false
	}

	var (
		body    []byte
		readErr error
	)
	if resp.Body != nil {
		body, readErr = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		// Replace body so callers that still try to read get the same bytes
		// rather than an already-closed reader.
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	apiErr := &APIError{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Headers:     resp.Header,
		Body:        body,
		Message:     extractMessage(body),
		BodyReadErr: readErr,
	}
	if resp.Request != nil {
		apiErr.Method = resp.Request.Method
		if resp.Request.URL != nil {
			apiErr.URL = resp.Request.URL.String()
		}
	}
	return apiErr, true
}

// extractMessage tries a handful of common error-body shapes used by the
// Proof APIs and returns the first non-empty string it finds.
func extractMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var probe struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Detail  string `json:"detail"`
		Errors  []struct {
			Message string `json:"message"`
			Detail  string `json:"detail"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return ""
	}
	switch {
	case probe.Error != "":
		return probe.Error
	case probe.Message != "":
		return probe.Message
	case probe.Detail != "":
		return probe.Detail
	case len(probe.Errors) > 0:
		if probe.Errors[0].Message != "" {
			return probe.Errors[0].Message
		}
		return probe.Errors[0].Detail
	}
	return ""
}
