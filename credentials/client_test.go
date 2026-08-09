package credentials

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestNewClient(t *testing.T) {
	t.Run("adds trailing slash to server URL", func(t *testing.T) {
		client, err := NewClient("https://api.example.com")
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/", client.Server)
		assert.NotNil(t, client.Client)
	})

	t.Run("accepts custom HTTP client", func(t *testing.T) {
		custom := &http.Client{}
		client, err := NewClient("https://api.example.com", WithHTTPClient(custom))
		require.NoError(t, err)
		assert.Equal(t, custom, client.Client)
	})
}

func TestNewAuthorizeVerifiableCredentialPresentationRequest(t *testing.T) {
	params := &AuthorizeVerifiableCredentialPresentationParams{
		ClientId:     "client-123",
		ResponseType: "vp_token",
		ResponseMode: "fragment",
		RedirectUri:  ptr("https://app.example.com/callback"),
		Scope:        "openid",
		LoginHint:    "user@example.com",
		Nonce:        "nonce-abc",
		State:        ptr("state-xyz"),
	}

	req, err := NewAuthorizeVerifiableCredentialPresentationRequest("https://api.example.com/", params)
	require.NoError(t, err)

	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "/verifiable-credentials/v1/presentation/authorize", req.URL.Path)

	q := req.URL.Query()
	assert.Equal(t, "client-123", q.Get("client_id"))
	assert.Equal(t, "vp_token", q.Get("response_type"))
	assert.Equal(t, "fragment", q.Get("response_mode"))
	assert.Equal(t, "https://app.example.com/callback", q.Get("redirect_uri"))
	assert.Equal(t, "user@example.com", q.Get("login_hint"))
	assert.Equal(t, "nonce-abc", q.Get("nonce"))
	assert.Equal(t, "state-xyz", q.Get("state"))
	assert.Empty(t, q.Get("response_uri"))
}
