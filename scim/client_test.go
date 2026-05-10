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

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// MockRoundTripper is a mock implementation of http.RoundTripper for testing.
type MockRoundTripper struct {
	Response *http.Response
	Err      error
	LastReq  *http.Request
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.LastReq = req
	return m.Response, m.Err
}

func mockJSONResponse(statusCode int, body any) *http.Response {
	jsonBytes, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
		Header:     http.Header{"Content-Type": []string{"application/scim+json"}},
	}
}

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

func TestNewClientWithResponses(t *testing.T) {
	client, err := NewClientWithResponses("https://api.example.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewListUsersRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/org-123/Users")
	})

	t.Run("includes startIndex parameter", func(t *testing.T) {
		params := &ListUsersParams{StartIndex: ptr(1)}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "startIndex=1")
	})

	t.Run("includes count parameter", func(t *testing.T) {
		params := &ListUsersParams{Count: ptr(50)}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "count=50")
	})

	t.Run("includes filter parameter", func(t *testing.T) {
		params := &ListUsersParams{Filter: ptr(`userName eq "test@example.com"`)}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "filter=")
	})

	t.Run("includes multiple parameters", func(t *testing.T) {
		params := &ListUsersParams{StartIndex: ptr(1), Count: ptr(100)}
		req, err := NewListUsersRequest("https://api.example.com/", "org-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "startIndex=1")
		assert.Contains(t, req.URL.RawQuery, "count=100")
	})
}

func TestNewGetUserRequest(t *testing.T) {
	req, err := NewGetUserRequest("https://api.example.com/", "org-123", "user-456")
	require.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
}

func TestNewDeleteUserRequest(t *testing.T) {
	req, err := NewDeleteUserRequest("https://api.example.com/", "org-123", "user-456")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", req.Method)
	assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
}

func TestNewPatchUserRequest(t *testing.T) {
	body := PatchUserApplicationScimPlusJSONRequestBody{
		Operations: &[]struct {
			Op    *string                 `json:"op,omitempty"`
			Path  *string                 `json:"path,omitempty"`
			Value *map[string]interface{} `json:"value,omitempty"`
		}{
			{Op: ptr("replace"), Path: ptr("active"), Value: &map[string]interface{}{"value": false}},
		},
	}
	req, err := NewPatchUserRequestWithApplicationScimPlusJSONBody("https://api.example.com/", "org-123", "user-456", body)
	require.NoError(t, err)
	assert.Equal(t, "PATCH", req.Method)
	assert.Contains(t, req.URL.Path, "/org-123/Users/user-456")
	assert.Equal(t, "application/scim+json", req.Header.Get("Content-Type"))
}

func TestNewCreateUserRequest(t *testing.T) {
	body := CreateUserApplicationScimPlusJSONRequestBody{
		UserName: "newuser@example.com",
		Name: &struct {
			FamilyName *string `json:"familyName,omitempty"`
			GivenName  *string `json:"givenName,omitempty"`
		}{
			GivenName:  ptr("John"),
			FamilyName: ptr("Doe"),
		},
		Emails: &[]struct {
			Value *string `json:"value,omitempty"`
		}{
			{Value: ptr("newuser@example.com")},
		},
	}
	req, err := NewCreateUserRequestWithApplicationScimPlusJSONBody("https://api.example.com/", "org-123", body)
	require.NoError(t, err)
	assert.Equal(t, "POST", req.Method)
	assert.Contains(t, req.URL.Path, "/org-123/Users")
	assert.Equal(t, "application/scim+json", req.Header.Get("Content-Type"))
}

func TestNewRetrieveUsersSchemaRequest(t *testing.T) {
	req, err := NewRetrieveUsersSchemaRequest("https://api.example.com/", "org-123")
	require.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Contains(t, req.URL.Path, "/org-123/Schemas/Users")
}

func TestNewGetResourceTypesRequest(t *testing.T) {
	req, err := NewGetResourceTypesRequest("https://api.example.com/", "org-123")
	require.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Contains(t, req.URL.Path, "/org-123/ResourceTypes")
}

func TestNewGetServiceProviderConfigRequest(t *testing.T) {
	req, err := NewGetServiceProviderConfigRequest("https://api.example.com/", "org-123")
	require.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Contains(t, req.URL.Path, "/org-123/ServiceProviderConfig")
}

// TestParseListUsersResponse verifies the raw body is preserved so callers can
// unmarshal SCIM-shaped payloads themselves. The generated client does not
// produce typed JSON200/JSON403 fields because the spec uses
// application/scim+json rather than application/json for responses.
func TestParseListUsersResponse(t *testing.T) {
	t.Run("preserves body bytes for 200 response", func(t *testing.T) {
		rawJSON := []byte(`{"totalResults":2,"itemsPerPage":50,"startIndex":1}`)
		resp := &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Body:       io.NopCloser(bytes.NewReader(rawJSON)),
			Header:     http.Header{"Content-Type": []string{"application/scim+json"}},
		}
		parsed, err := ParseListUsersResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.JSONEq(t, string(rawJSON), string(parsed.Body))
	})

	t.Run("preserves body bytes for 403 response", func(t *testing.T) {
		resp := mockJSONResponse(403, map[string][]string{"errors": {"Access denied"}})
		parsed, err := ParseListUsersResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 403, parsed.StatusCode())
		assert.Contains(t, string(parsed.Body), "Access denied")
	})

	t.Run("empty body yields no error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 204,
			Status:     http.StatusText(204),
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			Header:     http.Header{},
		}
		parsed, err := ParseListUsersResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 204, parsed.StatusCode())
		assert.Empty(t, parsed.Body)
	})
}

func TestClientWithResponsesMethods(t *testing.T) {
	mockRT := &MockRoundTripper{
		Response: mockJSONResponse(200, struct {
			Resources []interface{} `json:"Resources"`
		}{Resources: []interface{}{}}),
	}

	client, err := NewClientWithResponses("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}))
	require.NoError(t, err)

	resp, err := client.ListUsersWithResponse(context.Background(), "org-123", nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode())
}

func TestClientMakesRequest(t *testing.T) {
	t.Run("ListUsers", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, struct {
				Resources []interface{} `json:"Resources"`
			}{Resources: []interface{}{}}),
		}
		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.ListUsers(context.Background(), "org-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "/org-123/Users")
	})

	t.Run("GetUser", func(t *testing.T) {
		mockRT := &MockRoundTripper{Response: mockJSONResponse(200, struct{}{})}
		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.GetUser(context.Background(), "org-123", "user-456")
		require.NoError(t, err)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "user-456")
	})
}

func TestWithBaseURL(t *testing.T) {
	client, err := NewClient("https://api.example.com",
		WithBaseURL("https://custom.api.example.com/"))
	require.NoError(t, err)
	assert.Equal(t, "https://custom.api.example.com/", client.Server)
}

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
		}{Resources: []interface{}{}}),
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
