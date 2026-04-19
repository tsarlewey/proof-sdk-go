package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

// MockRoundTripper is a mock implementation of http.RoundTripper for testing
type MockRoundTripper struct {
	Response *http.Response
	Err      error
	LastReq  *http.Request
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.LastReq = req
	return m.Response, m.Err
}

// mockJSONResponse creates a mock HTTP response with the given status code and body
func mockJSONResponse(statusCode int, body any) *http.Response {
	jsonBytes, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// TestNewClient verifies client creation
func TestNewClient(t *testing.T) {
	t.Run("creates client with default http client", func(t *testing.T) {
		client, err := NewClient("https://api.example.com")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://api.example.com/", client.Server)
		assert.NotNil(t, client.Client)
	})

	t.Run("adds trailing slash to server URL", func(t *testing.T) {
		client, err := NewClient("https://api.example.com")
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/", client.Server)
	})

	t.Run("preserves existing trailing slash", func(t *testing.T) {
		client, err := NewClient("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/", client.Server)
	})

	t.Run("accepts custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{}
		client, err := NewClient("https://api.example.com", WithHTTPClient(customClient))
		require.NoError(t, err)
		assert.Equal(t, customClient, client.Client)
	})
}

// TestNewClientWithResponses verifies response client creation
func TestNewClientWithResponses(t *testing.T) {
	client, err := NewClientWithResponses("https://api.example.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestNewListUsersRequest verifies list users request generation
func TestNewListUsersRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users")
	})

	t.Run("includes startIndex parameter", func(t *testing.T) {
		params := &ListUsersParams{
			StartIndex: ptr(int32(1)),
		}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "startIndex=1")
	})

	t.Run("includes count parameter", func(t *testing.T) {
		params := &ListUsersParams{
			Count: ptr(int32(50)),
		}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "count=50")
	})

	t.Run("includes filter parameter", func(t *testing.T) {
		params := &ListUsersParams{
			Filter: ptr(`userName eq "test@example.com"`),
		}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "filter=")
	})

	t.Run("includes multiple parameters", func(t *testing.T) {
		params := &ListUsersParams{
			StartIndex: ptr(int32(1)),
			Count:      ptr(int32(100)),
		}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "startIndex=1")
		assert.Contains(t, req.URL.RawQuery, "count=100")
	})
}

// TestNewGetUserRequest verifies get user request generation
func TestNewGetUserRequest(t *testing.T) {
	t.Run("generates GET request for specific user", func(t *testing.T) {
		req, err := NewGetUserRequest("https://api.example.com/", "org-123", "user-456", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
	})

	t.Run("includes accept header parameter", func(t *testing.T) {
		params := &GetUserParams{
			Accept: ptr("application/json"),
		}
		req, err := NewGetUserRequest("https://api.example.com/", "org-123", "user-456", params)
		require.NoError(t, err)
		assert.Equal(t, "application/json", req.Header.Get("accept"))
	})
}

// TestNewDeleteUserRequest verifies delete user request generation
func TestNewDeleteUserRequest(t *testing.T) {
	t.Run("generates DELETE request for user", func(t *testing.T) {
		req, err := NewDeleteUserRequest("https://api.example.com/", "org-123", "user-456", nil)
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
	})
}

// TestNewPatchUserRequest verifies patch user request generation
func TestNewPatchUserRequest(t *testing.T) {
	t.Run("generates PATCH request for user", func(t *testing.T) {
		body := PatchUserJSONRequestBody{
			Operations: struct {
				Op    string  `json:"op"`
				Path  *string `json:"path,omitempty"`
				Value *string `json:"value,omitempty"`
			}{
				Op:    "replace",
				Path:  ptr("active"),
				Value: ptr("false"),
			},
		}
		req, err := NewPatchUserRequest("https://api.example.com/", "org-123", "user-456", nil, body)
		require.NoError(t, err)
		assert.Equal(t, "PATCH", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// TestNewCreateUserRequest verifies create user request generation
func TestNewCreateUserRequest(t *testing.T) {
	t.Run("generates POST request to create user", func(t *testing.T) {
		body := CreateUserJSONRequestBody{
			UserName: "newuser@example.com",
			Name: &struct {
				FamilyName *string `json:"familyName,omitempty"`
				GivenName  *string `json:"givenName,omitempty"`
			}{
				GivenName:  ptr("John"),
				FamilyName: ptr("Doe"),
			},
			Emails: &[]string{"newuser@example.com"},
		}
		req, err := NewCreateUserRequest("https://api.example.com/", "org-123", nil, body)
		require.NoError(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// TestNewRetrieveUsersSchemaRequest verifies user schema retrieval
func TestNewRetrieveUsersSchemaRequest(t *testing.T) {
	t.Run("generates request for user schema", func(t *testing.T) {
		req, err := NewRetrieveUsersSchemaRequest("https://api.example.com/", "org-123")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Schemas/Users")
	})
}

// TestNewGetResourceTypesRequest verifies resource types retrieval
func TestNewGetResourceTypesRequest(t *testing.T) {
	t.Run("generates request for resource types", func(t *testing.T) {
		req, err := NewGetResourceTypesRequest("https://api.example.com/", "org-123")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/ResourceTypes")
	})
}

// TestNewGetServiceProviderConfigRequest verifies service provider config retrieval
func TestNewGetServiceProviderConfigRequest(t *testing.T) {
	t.Run("generates request for service provider config", func(t *testing.T) {
		req, err := NewGetServiceProviderConfigRequest("https://api.example.com/", "org-123")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/ServiceProviderConfig")
	})
}

// TestParseListUsersResponse verifies list users response parsing — payload
// fields are asserted (not just the status code) so a parser that silently
// dropped them would fail the test.
func TestParseListUsersResponse(t *testing.T) {
	t.Run("deserializes 200 payload fields", func(t *testing.T) {
		rawJSON := []byte(`{
			"schemas": ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
			"totalResults": 2,
			"itemsPerPage": 50,
			"startIndex": 1,
			"Resources": [
				{
					"id": "user-1",
					"userName": "alice@example.com",
					"active": true,
					"name": {"givenName": "Alice", "familyName": "Smith"},
					"emails": [{"value": "alice@example.com", "primary": true}],
					"roles": [{"value": "admin", "display": "Admin"}]
				},
				{
					"id": "user-2",
					"userName": "bob@example.com",
					"active": false
				}
			]
		}`)
		resp := &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Body:       io.NopCloser(bytes.NewReader(rawJSON)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}

		parsed, err := ParseListUsersResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON200)

		assert.Equal(t, 2, *parsed.JSON200.TotalResults)
		assert.Equal(t, 50, *parsed.JSON200.ItemsPerPage)
		assert.Equal(t, 1, *parsed.JSON200.StartIndex)

		require.NotNil(t, parsed.JSON200.Resources)
		require.Len(t, *parsed.JSON200.Resources, 2)

		alice := (*parsed.JSON200.Resources)[0]
		assert.Equal(t, "user-1", *alice.Id)
		assert.Equal(t, "alice@example.com", *alice.UserName)
		assert.True(t, *alice.Active)
		require.NotNil(t, alice.Name)
		assert.Equal(t, "Alice", *alice.Name.GivenName)
		assert.Equal(t, "Smith", *alice.Name.FamilyName)
		require.NotNil(t, alice.Emails)
		require.Len(t, *alice.Emails, 1)
		assert.Equal(t, "alice@example.com", *(*alice.Emails)[0].Value)
		assert.True(t, *(*alice.Emails)[0].Primary)

		bob := (*parsed.JSON200.Resources)[1]
		assert.Equal(t, "bob@example.com", *bob.UserName)
		assert.False(t, *bob.Active)

		// Body bytes are preserved so callers that PrintResponse() see the raw JSON.
		assert.Contains(t, string(parsed.Body), `"alice@example.com"`)
	})

	t.Run("deserializes 403 errors", func(t *testing.T) {
		resp := mockJSONResponse(403, struct {
			Errors []string `json:"errors"`
		}{Errors: []string{"Access denied"}})

		parsed, err := ParseListUsersResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON403)
		require.NotNil(t, parsed.JSON403.Errors)
		assert.Equal(t, []string{"Access denied"}, *parsed.JSON403.Errors)
	})

	t.Run("deserializes 404 errors", func(t *testing.T) {
		resp := mockJSONResponse(404, struct {
			Errors []string `json:"errors"`
		}{Errors: []string{"Not found"}})

		parsed, err := ParseListUsersResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON404)
		require.NotNil(t, parsed.JSON404.Errors)
		assert.Equal(t, []string{"Not found"}, *parsed.JSON404.Errors)
	})

	t.Run("empty body yields nil JSON200 without error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 204,
			Status:     http.StatusText(204),
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			Header:     http.Header{},
		}

		parsed, err := ParseListUsersResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 204, parsed.StatusCode())
		assert.Nil(t, parsed.JSON200)
	})
}

// TestClientWithResponsesMethods verifies client with responses wrapper methods
func TestClientWithResponsesMethods(t *testing.T) {
	mockRT := &MockRoundTripper{
		Response: mockJSONResponse(200, struct {
			Resources []interface{} `json:"Resources"`
		}{
			Resources: []interface{}{},
		}),
	}

	client, err := NewClientWithResponses("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}))
	require.NoError(t, err)

	t.Run("ListUsersWithResponse", func(t *testing.T) {
		resp, err := client.ListUsersWithResponse(context.Background(), "org-123", nil)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode())
	})
}

// TestClientMakesRequest verifies client makes correct requests
func TestClientMakesRequest(t *testing.T) {
	t.Run("ListUsers", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, struct {
				Resources []interface{} `json:"Resources"`
			}{
				Resources: []interface{}{},
			}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.ListUsers(context.Background(), "org-123", nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "/org-123/Users")
	})

	t.Run("GetUser", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, struct{}{}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.GetUser(context.Background(), "org-123", "user-456", nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "user-456")
	})
}

// TestWithBaseURL verifies base URL override
func TestWithBaseURL(t *testing.T) {
	client, err := NewClient("https://api.example.com",
		WithBaseURL("https://custom.api.example.com/"))
	require.NoError(t, err)
	assert.Equal(t, "https://custom.api.example.com/", client.Server)
}

// TestWithRequestEditorFn verifies request editor functionality
func TestWithRequestEditorFn(t *testing.T) {
	editorCalled := false
	editor := func(ctx context.Context, req *http.Request) error {
		editorCalled = true
		req.Header.Set("X-Custom-Header", "test-value")
		return nil
	}

	mockRT := &MockRoundTripper{
		Response: mockJSONResponse(200, struct {
			Resources []interface{} `json:"Resources"`
		}{
			Resources: []interface{}{},
		}),
	}

	client, err := NewClient("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}),
		WithRequestEditorFn(editor))
	require.NoError(t, err)

	_, err = client.ListUsers(context.Background(), "org-123", nil)
	require.NoError(t, err)

	assert.True(t, editorCalled)
	assert.Equal(t, "test-value", mockRT.LastReq.Header.Get("X-Custom-Header"))
}
