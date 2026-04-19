package common

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAuthenticatedDoer_Provider(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)

	// Act
	doer := NewAuthenticatedDoer(mockProvider)

	// Assert
	assert.NotNil(t, doer)
	assert.Equal(t, mockProvider, doer.client)
}

func TestAuthenticatedDoer_Do_WithOAuth(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)

	req, err := http.NewRequest("GET", "https://api.example.com/test", nil)
	require.NoError(t, err)

	expectedResp := MockJSONResponse(200, map[string]string{"status": "ok"})

	// Setup expectations
	mockProvider.On("AddAuthHeaders", req).Return(nil).Run(func(args mock.Arguments) {
		r := args.Get(0).(*http.Request)
		r.Header.Set("Authorization", "Bearer test-oauth-token")
	})
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.Anything).Return(expectedResp, nil)

	doer := NewAuthenticatedDoer(mockProvider)

	// Act
	resp, err := doer.Do(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Bearer test-oauth-token", req.Header.Get("Authorization"))
	mockProvider.AssertExpectations(t)
	mockTransport.AssertExpectations(t)
}

func TestAuthenticatedDoer_Do_WithAPIKey(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)

	req, err := http.NewRequest("GET", "https://api.example.com/test", nil)
	require.NoError(t, err)

	expectedResp := MockJSONResponse(200, map[string]string{"status": "ok"})

	// Setup expectations - API key auth sets ApiKey header
	mockProvider.On("AddAuthHeaders", req).Return(nil).Run(func(args mock.Arguments) {
		r := args.Get(0).(*http.Request)
		r.Header.Set("ApiKey", "test-api-key")
	})
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.Anything).Return(expectedResp, nil)

	doer := NewAuthenticatedDoer(mockProvider)

	// Act
	resp, err := doer.Do(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "test-api-key", req.Header.Get("ApiKey"))
	mockProvider.AssertExpectations(t)
	mockTransport.AssertExpectations(t)
}

func TestAuthenticatedDoer_Do_AuthError(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)
	authError := errors.New("failed to get OAuth token")

	req, err := http.NewRequest("GET", "https://api.example.com/test", nil)
	require.NoError(t, err)

	// Setup expectations - auth fails
	mockProvider.On("AddAuthHeaders", req).Return(authError)

	doer := NewAuthenticatedDoer(mockProvider)

	// Act
	resp, err := doer.Do(req)

	// Assert
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, authError, err)
	mockProvider.AssertExpectations(t)
}

func TestAuthenticatedDoer_Do_HTTPError(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)
	httpError := errors.New("connection refused")

	req, err := http.NewRequest("GET", "https://api.example.com/test", nil)
	require.NoError(t, err)

	// Setup expectations - auth succeeds but HTTP request fails
	mockProvider.On("AddAuthHeaders", req).Return(nil)
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.Anything).Return(nil, httpError)

	doer := NewAuthenticatedDoer(mockProvider)

	// Act
	resp, err := doer.Do(req)

	// Assert
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	mockProvider.AssertExpectations(t)
	mockTransport.AssertExpectations(t)
}

func TestAuthenticatedDoer_Do_HTTP400Error(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)

	req, err := http.NewRequest("POST", "https://api.example.com/test", nil)
	require.NoError(t, err)

	expectedResp := MockJSONResponse(400, map[string]string{"error": "bad request"})

	// Setup expectations
	mockProvider.On("AddAuthHeaders", req).Return(nil)
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.Anything).Return(expectedResp, nil)

	doer := NewAuthenticatedDoer(mockProvider)

	// Act
	resp, err := doer.Do(req)

	// Assert - HTTP 400 is not an error from the transport perspective
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	mockProvider.AssertExpectations(t)
	mockTransport.AssertExpectations(t)
}

func TestAuthenticatedDoer_Do_HTTP500Error(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)

	req, err := http.NewRequest("GET", "https://api.example.com/test", nil)
	require.NoError(t, err)

	expectedResp := MockJSONResponse(500, map[string]string{"error": "internal server error"})

	// Setup expectations
	mockProvider.On("AddAuthHeaders", req).Return(nil)
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.Anything).Return(expectedResp, nil)

	doer := NewAuthenticatedDoer(mockProvider)

	// Act
	resp, err := doer.Do(req)

	// Assert - HTTP 500 is not an error from the transport perspective
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	mockProvider.AssertExpectations(t)
	mockTransport.AssertExpectations(t)
}

func TestAuthenticatedDoer_Do_PreservesRequestHeaders(t *testing.T) {
	// Arrange
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)

	req, err := http.NewRequest("POST", "https://api.example.com/test", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom-Header", "custom-value")

	expectedResp := MockJSONResponse(200, map[string]string{"status": "ok"})

	// Setup expectations
	mockProvider.On("AddAuthHeaders", req).Return(nil).Run(func(args mock.Arguments) {
		r := args.Get(0).(*http.Request)
		r.Header.Set("Authorization", "Bearer token")
	})
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.Header.Get("Content-Type") == "application/json" &&
			r.Header.Get("X-Custom-Header") == "custom-value" &&
			r.Header.Get("Authorization") == "Bearer token"
	})).Return(expectedResp, nil)

	doer := NewAuthenticatedDoer(mockProvider)

	// Act
	resp, err := doer.Do(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockProvider.AssertExpectations(t)
	mockTransport.AssertExpectations(t)
}
