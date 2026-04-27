package common

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsAPIError_Success(t *testing.T) {
	resp := MockJSONResponse(200, map[string]string{"ok": "true"})
	apiErr, ok := AsAPIError(resp)
	assert.False(t, ok)
	assert.Nil(t, apiErr)
}

func TestAsAPIError_Nil(t *testing.T) {
	apiErr, ok := AsAPIError(nil)
	assert.False(t, ok)
	assert.Nil(t, apiErr)
}

func TestAsAPIError_4xx_ExtractsErrorField(t *testing.T) {
	resp := MockJSONResponse(400, map[string]string{"error": "bad request"})
	resp.Request = &http.Request{Method: "POST", URL: &url.URL{Scheme: "https", Host: "api.example.com", Path: "/x"}}

	apiErr, ok := AsAPIError(resp)
	require.True(t, ok)
	require.NotNil(t, apiErr)

	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "bad request", apiErr.Message)
	assert.Equal(t, "POST", apiErr.Method)
	assert.Equal(t, "https://api.example.com/x", apiErr.URL)
	assert.NoError(t, apiErr.BodyReadErr)
	assert.Contains(t, apiErr.Error(), "400")
	assert.Contains(t, apiErr.Error(), "bad request")
}

func TestAsAPIError_5xx_NestedErrors(t *testing.T) {
	resp := MockJSONResponse(503, map[string]any{
		"errors": []map[string]string{{"message": "service unavailable"}},
	})
	apiErr, ok := AsAPIError(resp)
	require.True(t, ok)
	assert.Equal(t, "service unavailable", apiErr.Message)
}

func TestAsAPIError_BodyReplayable(t *testing.T) {
	resp := MockJSONResponse(404, map[string]string{"message": "missing"})
	apiErr, ok := AsAPIError(resp)
	require.True(t, ok)

	// Body should still be readable from resp after AsAPIError captured it.
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(apiErr.Body), string(b))
}

func TestAsAPIError_NonJSONBody(t *testing.T) {
	resp := MockTextResponse(500, "boom")
	apiErr, ok := AsAPIError(resp)
	require.True(t, ok)
	assert.Equal(t, "", apiErr.Message)
	assert.Equal(t, "boom", string(apiErr.Body))
}

func TestCheckResponse_AsCompatible(t *testing.T) {
	resp := MockJSONResponse(422, map[string]string{"detail": "validation failed"})
	err := CheckResponse(resp)
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 422, apiErr.StatusCode)
	assert.Equal(t, "validation failed", apiErr.Message)
}

func TestCheckResponse_NilOnSuccess(t *testing.T) {
	resp := MockJSONResponse(204, nil)
	assert.NoError(t, CheckResponse(resp))
}
