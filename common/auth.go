package common

import (
	"net/http"
)

// AuthProvider defines the interface for adding authentication headers to requests.
// This interface allows mocking ProofClient in tests.
type AuthProvider interface {
	AddAuthHeaders(req *http.Request) error
	HTTPClient() *http.Client
}

// AuthenticatedDoer wraps an AuthProvider to implement the HttpRequestDoer interface
// required by oapi-codegen generated clients. It automatically adds authentication
// headers (OAuth Bearer token or API key) to all requests.
type AuthenticatedDoer struct {
	client AuthProvider
}

// NewAuthenticatedDoer creates a new AuthenticatedDoer that wraps the given AuthProvider.
func NewAuthenticatedDoer(client AuthProvider) *AuthenticatedDoer {
	return &AuthenticatedDoer{
		client: client,
	}
}

// Do implements the HttpRequestDoer interface. It adds authentication headers
// to the request and then executes it using the underlying HTTP client.
func (a *AuthenticatedDoer) Do(req *http.Request) (*http.Response, error) {
	// Add authentication headers
	if err := a.client.AddAuthHeaders(req); err != nil {
		return nil, err
	}

	// Execute the request using the underlying HTTP client
	return a.client.HTTPClient().Do(req)
}
