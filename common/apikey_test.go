package common

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKey_SendsApiKeyHeader(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
	}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := NewAuthenticatedDoer(APIKey("secret-key")).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "secret-key", got.Get("ApiKey"))
	assert.Empty(t, got.Get("Authorization"))
	assert.Equal(t, UserAgent, got.Get("User-Agent"))
}

func TestAPIKey_EmptyKeyFailsBeforeSending(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:0", nil)
	require.NoError(t, err)

	_, err = NewAuthenticatedDoer(APIKey("")).Do(req)
	assert.ErrorContains(t, err, "empty API key")
}

func TestAPIKey_HTTPClientHasTimeout(t *testing.T) {
	assert.NotZero(t, APIKey("k").HTTPClient().Timeout)
}

func TestCheckResponseBody(t *testing.T) {
	// Mimic a generated *WithResponse result: HTTPResponse.Body is already
	// consumed, the bytes are passed separately.
	resp := &http.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    httptest.NewRequest(http.MethodGet, "https://api.example.com/v1/transactions/ot_x", nil),
	}
	body := []byte(`{"error":"transaction not found"}`)

	err := CheckResponseBody(resp, body)
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Equal(t, "transaction not found", apiErr.Message)
	assert.Equal(t, body, apiErr.Body)
	assert.Equal(t, "GET", apiErr.Method)

	assert.NoError(t, CheckResponseBody(&http.Response{StatusCode: 200}, nil))
	assert.NoError(t, CheckResponseBody(nil, nil))
}
