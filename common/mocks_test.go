package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockAuthProvider is a mock implementation of AuthProvider for testing
type MockAuthProvider struct {
	mock.Mock
	httpClient *http.Client
}

// AddAuthHeaders mocks adding authentication headers to a request
func (m *MockAuthProvider) AddAuthHeaders(req *http.Request) error {
	args := m.Called(req)
	return args.Error(0)
}

// HTTPClient returns the mock HTTP client
func (m *MockAuthProvider) HTTPClient() *http.Client {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*http.Client)
}

// MockRoundTripper is a mock implementation of http.RoundTripper for testing HTTP requests
type MockRoundTripper struct {
	mock.Mock
}

// RoundTrip mocks HTTP request execution
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockJSONResponse creates a mock HTTP response with the given status code and body
func MockJSONResponse(statusCode int, body interface{}) *http.Response {
	jsonBytes, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// MockTextResponse creates a mock HTTP response with plain text body
func MockTextResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
	}
}

// MockEmptyResponse creates a mock HTTP response with an empty body
func MockEmptyResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
		Header:     http.Header{},
	}
}
