package common

import (
	"errors"
	"net/http"
	"time"
)

// Base URLs for the Proof API. Pass one as the server argument to any
// generated NewClientWithResponses.
const (
	// ProductionURL is the live Proof API.
	ProductionURL = "https://api.proof.com"
	// SandboxURL is Proof's Fairfax sandbox. Sandbox API keys only work here.
	SandboxURL = "https://api.fairfax.proof.com"
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

	// Set the default SDK User-Agent only when the caller hasn't supplied one,
	// so server-side telemetry can attribute traffic without overriding
	// downstream tools that need their own UA.
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", UserAgent)
	}

	// Execute the request using the underlying HTTP client
	return a.client.HTTPClient().Do(req)
}

// APIKey returns an AuthProvider that authenticates every request with a
// Proof API key sent in the "ApiKey" header. Requests go through an
// *http.Client with a 60s timeout; implement AuthProvider yourself if you need
// a different client or OAuth client-credentials auth.
//
//	doer := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
func APIKey(key string) AuthProvider {
	return apiKeyAuth{key: key, client: &http.Client{Timeout: 60 * time.Second}}
}

type apiKeyAuth struct {
	key    string
	client *http.Client
}

func (a apiKeyAuth) AddAuthHeaders(req *http.Request) error {
	// Fail before sending rather than letting the API answer 401 for an
	// unset environment variable.
	if a.key == "" {
		return errors.New("proof-sdk-go: empty API key")
	}
	req.Header.Set("ApiKey", a.key)
	return nil
}

func (a apiKeyAuth) HTTPClient() *http.Client { return a.client }
