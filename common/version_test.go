package common

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatedDoer_SetsDefaultUserAgent(t *testing.T) {
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)

	req, err := http.NewRequest("GET", "https://api.example.com/x", nil)
	require.NoError(t, err)

	mockProvider.On("AddAuthHeaders", req).Return(nil)
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.Header.Get("User-Agent") == UserAgent
	})).Return(MockJSONResponse(200, nil), nil)

	doer := NewAuthenticatedDoer(mockProvider)
	_, err = doer.Do(req)
	require.NoError(t, err)
}

func TestAuthenticatedDoer_PreservesCallerUserAgent(t *testing.T) {
	mockProvider := new(MockAuthProvider)
	mockTransport := new(MockRoundTripper)

	req, err := http.NewRequest("GET", "https://api.example.com/x", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "my-app/1.2.3")

	mockProvider.On("AddAuthHeaders", req).Return(nil)
	mockProvider.On("HTTPClient").Return(&http.Client{Transport: mockTransport})
	mockTransport.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.Header.Get("User-Agent") == "my-app/1.2.3"
	})).Return(MockJSONResponse(200, nil), nil)

	doer := NewAuthenticatedDoer(mockProvider)
	_, err = doer.Do(req)
	require.NoError(t, err)
}
