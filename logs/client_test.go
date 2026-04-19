package logs

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

// TestNewListSecurityEventsRequest verifies request generation
func TestNewListSecurityEventsRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewListSecurityEventsRequest("https://api.example.com/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "https://api.example.com/logs/v1/security-events", req.URL.String())
	})

	t.Run("includes limit parameter", func(t *testing.T) {
		params := &ListSecurityEventsParams{
			Limit: ptr(50),
		}
		req, err := NewListSecurityEventsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=50")
	})

	t.Run("includes cursor parameter", func(t *testing.T) {
		params := &ListSecurityEventsParams{
			Cursor: ptr("abc123"),
		}
		req, err := NewListSecurityEventsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "cursor=abc123")
	})

	t.Run("includes since parameter", func(t *testing.T) {
		params := &ListSecurityEventsParams{
			Since: ptr("2024-01-01T00:00:00Z"),
		}
		req, err := NewListSecurityEventsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.String(), "since=")
	})

	t.Run("includes class_uid parameter", func(t *testing.T) {
		params := &ListSecurityEventsParams{
			ClassUid: ptr(1001),
		}
		req, err := NewListSecurityEventsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "class_uid=1001")
	})

	t.Run("includes severity_id parameter", func(t *testing.T) {
		params := &ListSecurityEventsParams{
			SeverityId: ptr(3),
		}
		req, err := NewListSecurityEventsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "severity_id=3")
	})

	t.Run("includes multiple parameters", func(t *testing.T) {
		params := &ListSecurityEventsParams{
			Limit:      ptr(100),
			ClassUid:   ptr(1001),
			SeverityId: ptr(2),
		}
		req, err := NewListSecurityEventsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=100")
		assert.Contains(t, req.URL.RawQuery, "class_uid=1001")
		assert.Contains(t, req.URL.RawQuery, "severity_id=2")
	})
}

// TestParseListSecurityEventsResponse verifies response parsing
func TestParseListSecurityEventsResponse(t *testing.T) {
	t.Run("deserializes events and pagination meta", func(t *testing.T) {
		body := SecurityEventsResponse{
			Data: &[]SecurityEventObject{
				{
					ActivityId:   ptr(1),
					ActivityName: ptr("Login"),
					ClassName:    ptr("Authentication"),
					Severity:     ptr("Informational"),
				},
				{
					ActivityId:   ptr(2),
					ActivityName: ptr("Logout"),
					ClassName:    ptr("Authentication"),
					Severity:     ptr("Informational"),
				},
			},
			Meta: &struct {
				Count      *int    `json:"count,omitempty"`
				HasMore    *bool   `json:"has_more,omitempty"`
				NextCursor *string `json:"next_cursor,omitempty"`
			}{
				Count:      ptr(2),
				HasMore:    ptr(true),
				NextCursor: ptr("cursor-xyz"),
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseListSecurityEventsResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON200)
		require.NotNil(t, parsed.JSON200.Data)
		require.Len(t, *parsed.JSON200.Data, 2)

		first := (*parsed.JSON200.Data)[0]
		assert.Equal(t, 1, *first.ActivityId)
		assert.Equal(t, "Login", *first.ActivityName)
		assert.Equal(t, "Authentication", *first.ClassName)
		assert.Equal(t, "Informational", *first.Severity)

		assert.Equal(t, "Logout", *(*parsed.JSON200.Data)[1].ActivityName)

		require.NotNil(t, parsed.JSON200.Meta)
		assert.Equal(t, 2, *parsed.JSON200.Meta.Count)
		assert.True(t, *parsed.JSON200.Meta.HasMore)
		assert.Equal(t, "cursor-xyz", *parsed.JSON200.Meta.NextCursor)
	})

	t.Run("parses 400 error response", func(t *testing.T) {
		body := ErrorsObject{
			Errors: &[]string{"Invalid parameter: limit must be positive"},
		}
		resp := mockJSONResponse(400, body)

		parsed, err := ParseListSecurityEventsResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 400, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON400)
		assert.Contains(t, *parsed.JSON400.Errors, "Invalid parameter: limit must be positive")
	})

	t.Run("parses 403 error response", func(t *testing.T) {
		body := ErrorsObject{
			Errors: &[]string{"Access denied"},
		}
		resp := mockJSONResponse(403, body)

		parsed, err := ParseListSecurityEventsResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 403, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON403)
		assert.Contains(t, *parsed.JSON403.Errors, "Access denied")
	})

	t.Run("handles empty response body", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 204,
			Status:     http.StatusText(204),
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			Header:     http.Header{},
		}

		parsed, err := ParseListSecurityEventsResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 204, parsed.StatusCode())
		assert.Nil(t, parsed.JSON200)
	})
}

// TestClientListSecurityEvents verifies the full client workflow
func TestClientListSecurityEvents(t *testing.T) {
	t.Run("makes request with authentication", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, SecurityEventsResponse{
				Data: &[]SecurityEventObject{},
			}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.ListSecurityEvents(context.Background(), nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "/logs/v1/security-events")
	})

	t.Run("passes parameters correctly", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, SecurityEventsResponse{}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		params := &ListSecurityEventsParams{
			Limit: ptr(25),
		}
		_, err = client.ListSecurityEvents(context.Background(), params)
		require.NoError(t, err)

		assert.Contains(t, mockRT.LastReq.URL.RawQuery, "limit=25")
	})
}

// TestClientWithResponsesListSecurityEvents verifies the response wrapper
func TestClientWithResponsesListSecurityEvents(t *testing.T) {
	t.Run("returns parsed response", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, SecurityEventsResponse{
				Data: &[]SecurityEventObject{
					{ActivityName: ptr("Test Event")},
				},
				Meta: &struct {
					Count      *int    `json:"count,omitempty"`
					HasMore    *bool   `json:"has_more,omitempty"`
					NextCursor *string `json:"next_cursor,omitempty"`
				}{
					Count:   ptr(1),
					HasMore: ptr(false),
				},
			}),
		}

		client, err := NewClientWithResponses("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		resp, err := client.ListSecurityEventsWithResponse(context.Background(), nil)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode())
		assert.NotNil(t, resp.JSON200)
		assert.Len(t, *resp.JSON200.Data, 1)
		assert.Equal(t, "Test Event", *(*resp.JSON200.Data)[0].ActivityName)
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
		Response: mockJSONResponse(200, SecurityEventsResponse{}),
	}

	client, err := NewClient("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}),
		WithRequestEditorFn(editor))
	require.NoError(t, err)

	_, err = client.ListSecurityEvents(context.Background(), nil)
	require.NoError(t, err)

	assert.True(t, editorCalled)
	assert.Equal(t, "test-value", mockRT.LastReq.Header.Get("X-Custom-Header"))
}
