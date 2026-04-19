package business

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

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

// TestEnumValidation verifies enum validation methods
func TestEnumValidation(t *testing.T) {
	t.Run("CosignerSigningRequirement valid values", func(t *testing.T) {
		assert.True(t, CosignerSigningRequirementEsign.Valid())
		assert.True(t, CosignerSigningRequirementIdentify.Valid())
		assert.True(t, CosignerSigningRequirementVerify.Valid())
	})

	t.Run("CosignerSigningRequirement invalid value", func(t *testing.T) {
		invalid := CosignerSigningRequirement("invalid")
		assert.False(t, invalid.Valid())
	})

	t.Run("DocumentSigningType valid values", func(t *testing.T) {
		assert.True(t, ESIGN.Valid())
		assert.True(t, NOTARIZATION.Valid())
	})

	t.Run("DocumentSigningType invalid value", func(t *testing.T) {
		invalid := DocumentSigningType("invalid")
		assert.False(t, invalid.Valid())
	})

	t.Run("NotaryStatuses valid values", func(t *testing.T) {
		assert.True(t, Compliant.Valid())
		assert.True(t, NeedsReview.Valid())
		assert.True(t, NotCompliant.Valid())
	})

	t.Run("NotaryStatuses invalid value", func(t *testing.T) {
		invalid := NotaryStatuses("invalid")
		assert.False(t, invalid.Valid())
	})

	t.Run("CredentialImageDescriptor valid values", func(t *testing.T) {
		assert.True(t, Front.Valid())
		assert.True(t, Back.Valid())
	})

	t.Run("CredentialImageDescriptor invalid value", func(t *testing.T) {
		invalid := CredentialImageDescriptor("invalid")
		assert.False(t, invalid.Valid())
	})
}

// TestNewGetAllTransactionsRequest verifies list transactions request generation
func TestNewGetAllTransactionsRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewGetAllTransactionsRequest("https://api.example.com/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/v1/transactions")
	})

	t.Run("includes limit parameter", func(t *testing.T) {
		params := &GetAllTransactionsParams{
			Limit: ptr(50),
		}
		req, err := NewGetAllTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=50")
	})

	t.Run("includes offset parameter", func(t *testing.T) {
		params := &GetAllTransactionsParams{
			Offset: ptr(10),
		}
		req, err := NewGetAllTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "offset=10")
	})

	t.Run("includes created_date_start parameter", func(t *testing.T) {
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		params := &GetAllTransactionsParams{
			CreatedDateStart: &startDate,
		}
		req, err := NewGetAllTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "created_date_start=")
	})

	t.Run("includes transaction_status parameter", func(t *testing.T) {
		status := GetAllTransactionsParamsTransactionStatus("active")
		params := &GetAllTransactionsParams{
			TransactionStatus: &status,
		}
		req, err := NewGetAllTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "transaction_status=active")
	})

	t.Run("includes multiple parameters", func(t *testing.T) {
		params := &GetAllTransactionsParams{
			Limit:  ptr(100),
			Offset: ptr(0),
		}
		req, err := NewGetAllTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=100")
		assert.Contains(t, req.URL.RawQuery, "offset=0")
	})
}

// TestNewGetTransactionRequest verifies get transaction request generation
func TestNewGetTransactionRequest(t *testing.T) {
	t.Run("generates request with transaction ID", func(t *testing.T) {
		req, err := NewGetTransactionRequest("https://api.example.com/", "txn-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/transactions/txn-123")
	})

	t.Run("handles UUID format ID", func(t *testing.T) {
		req, err := NewGetTransactionRequest("https://api.example.com/", "550e8400-e29b-41d4-a716-446655440000", nil)
		require.NoError(t, err)
		assert.Contains(t, req.URL.Path, "550e8400-e29b-41d4-a716-446655440000")
	})
}

// TestNewDeleteTransactionRequest verifies delete transaction request generation
func TestNewDeleteTransactionRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteTransactionRequest("https://api.example.com/", "txn-123")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/transactions/txn-123")
	})
}

// TestNewGetAllNotariesRequest verifies list notaries request generation
func TestNewGetAllNotariesRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewGetAllNotariesRequest("https://api.example.com/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/v1/notaries")
	})

	t.Run("includes organization_id parameter", func(t *testing.T) {
		params := &GetAllNotariesParams{
			OrganizationId: ptr("org-123"),
		}
		req, err := NewGetAllNotariesRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "organization_id=org-123")
	})

	t.Run("includes us_state_abbr parameter", func(t *testing.T) {
		params := &GetAllNotariesParams{
			UsStateAbbr: ptr("CA"),
		}
		req, err := NewGetAllNotariesRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "us_state_abbr=CA")
	})
}

// TestNewGetNotaryRequest verifies get notary request generation
func TestNewGetNotaryRequest(t *testing.T) {
	t.Run("generates request with notary ID", func(t *testing.T) {
		req, err := NewGetNotaryRequest("https://api.example.com/", "notary-456")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/notaries/notary-456")
	})
}

// TestNewDeleteNotaryRequest verifies delete notary request generation
func TestNewDeleteNotaryRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteNotaryRequest("https://api.example.com/", "notary-456")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/notaries/notary-456")
	})
}

// TestNewGetWebhookURLRequest verifies get webhook URL request generation
func TestNewGetWebhookURLRequest(t *testing.T) {
	t.Run("generates request", func(t *testing.T) {
		req, err := NewGetWebhookURLRequest("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/v1/webhooks")
	})
}

// TestNewDeleteWebhookURLRequest verifies delete webhook URL request generation
func TestNewDeleteWebhookURLRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteWebhookURLRequest("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.String(), "/v1/webhooks")
	})
}

// TestNewGetOrganizationInformationRequest verifies get org info request generation
func TestNewGetOrganizationInformationRequest(t *testing.T) {
	t.Run("generates request", func(t *testing.T) {
		req, err := NewGetOrganizationInformationRequest("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/v1/organization")
	})
}

// TestNewGetAllTemplatesRequest verifies list templates request generation
func TestNewGetAllTemplatesRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewGetAllTemplatesRequest("https://api.example.com/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/v1/templates")
	})

	t.Run("includes limit parameter", func(t *testing.T) {
		params := &GetAllTemplatesParams{
			Limit: ptr(10),
		}
		req, err := NewGetAllTemplatesRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=10")
	})
}

// TestNewGetNotarizationRecordsRequest verifies list notarization records request generation
func TestNewGetNotarizationRecordsRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewGetNotarizationRecordsRequest("https://api.example.com/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/v1/notarization_records")
	})

	t.Run("includes limit parameter", func(t *testing.T) {
		limit := float32(20)
		params := &GetNotarizationRecordsParams{
			Limit: &limit,
		}
		req, err := NewGetNotarizationRecordsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=20")
	})
}

// TestNewGetNotarizationRecordRequest verifies get notarization record request generation
func TestNewGetNotarizationRecordRequest(t *testing.T) {
	t.Run("generates request with record ID", func(t *testing.T) {
		req, err := NewGetNotarizationRecordRequest("https://api.example.com/", "record-789", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/notarization_records/record-789")
	})
}

// TestNewGetDocumentRequest verifies get document request generation
func TestNewGetDocumentRequest(t *testing.T) {
	t.Run("generates request with transaction and document IDs", func(t *testing.T) {
		req, err := NewGetDocumentRequest("https://api.example.com/", "txn-123", "doc-456", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/transactions/txn-123/documents/doc-456")
	})
}

// TestNewDeleteDocumentRequest verifies delete document request generation
func TestNewDeleteDocumentRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteDocumentRequest("https://api.example.com/", "doc-123")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/documents/doc-123")
	})
}

// TestParseGetAllTransactionsResponse verifies list transactions response parsing
// by asserting on actual payload fields, not just the status code.
func TestParseGetAllTransactionsResponse(t *testing.T) {
	t.Run("deserializes nested transaction fields", func(t *testing.T) {
		status := TransactionObjectDetailedStatusComplete
		body := TransactionObjects{
			Data: &[]TransactionObject{
				{
					Id:             ptr("txn-123"),
					ExternalId:     ptr("external-abc"),
					Cost:           ptr(float32(29.99)),
					DetailedStatus: &status,
				},
				{
					Id: ptr("txn-456"),
				},
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetAllTransactionsResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON200)
		require.NotNil(t, parsed.JSON200.Data)
		require.Len(t, *parsed.JSON200.Data, 2)

		first := (*parsed.JSON200.Data)[0]
		assert.Equal(t, "txn-123", *first.Id)
		assert.Equal(t, "external-abc", *first.ExternalId)
		assert.InDelta(t, 29.99, *first.Cost, 0.001)
		assert.Equal(t, TransactionObjectDetailedStatusComplete, *first.DetailedStatus)

		assert.Equal(t, "txn-456", *(*parsed.JSON200.Data)[1].Id)
	})

	t.Run("empty body yields nil JSON200 without error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 204,
			Status:     http.StatusText(204),
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			Header:     http.Header{},
		}

		parsed, err := ParseGetAllTransactionsResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 204, parsed.StatusCode())
		assert.Nil(t, parsed.JSON200)
	})
}

// TestParseGetTransactionResponse verifies get transaction response parsing
func TestParseGetTransactionResponse(t *testing.T) {
	t.Run("parses 200 response", func(t *testing.T) {
		body := TransactionObject{
			Id: ptr("txn-123"),
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetTransactionResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
		assert.Equal(t, "txn-123", *parsed.JSON200.Id)
	})
}

// TestParseGetAllNotariesResponse verifies list notaries response parsing
func TestParseGetAllNotariesResponse(t *testing.T) {
	t.Run("parses 200 response", func(t *testing.T) {
		body := NotaryObjects{
			{
				Id: ptr("notary-123"),
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetAllNotariesResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
	})

	t.Run("parses empty array", func(t *testing.T) {
		body := NotaryObjects{}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetAllNotariesResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
	})
}

// TestParseGetNotaryResponse verifies get notary response parsing
func TestParseGetNotaryResponse(t *testing.T) {
	t.Run("parses 200 response", func(t *testing.T) {
		body := NotaryObject{
			Id: ptr("notary-123"),
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetNotaryResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
	})

	t.Run("parses 404 error response", func(t *testing.T) {
		body := ErrorsObject{
			Message: ptr("Notary not found"),
		}
		resp := mockJSONResponse(404, body)

		parsed, err := ParseGetNotaryResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 404, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON404)
	})
}

// TestParseDeleteDocumentResponse verifies delete document response parsing
func TestParseDeleteDocumentResponse(t *testing.T) {
	t.Run("parses 200 response", func(t *testing.T) {
		body := struct {
			Message *string `json:"message,omitempty"`
		}{
			Message: ptr("Document deleted"),
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseDeleteDocumentResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
	})

	t.Run("parses 404 error response", func(t *testing.T) {
		body := ErrorsObject{
			Message: ptr("Document not found"),
		}
		resp := mockJSONResponse(404, body)

		parsed, err := ParseDeleteDocumentResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 404, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON404)
	})

	t.Run("parses 422 error response", func(t *testing.T) {
		body := ErrorsObject{
			Message: ptr("Cannot delete document"),
		}
		resp := mockJSONResponse(422, body)

		parsed, err := ParseDeleteDocumentResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 422, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON422)
	})
}

// TestClientWithResponsesMethods verifies client with responses wrapper methods
func TestClientWithResponsesMethods(t *testing.T) {
	mockRT := &MockRoundTripper{
		Response: mockJSONResponse(200, TransactionObjects{
			Data: &[]TransactionObject{},
		}),
	}

	client, err := NewClientWithResponses("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}))
	require.NoError(t, err)

	t.Run("GetAllTransactionsWithResponse", func(t *testing.T) {
		resp, err := client.GetAllTransactionsWithResponse(context.Background(), nil)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode())
	})
}

// TestClientMakesRequest verifies client makes correct requests
func TestClientMakesRequest(t *testing.T) {
	t.Run("GetAllTransactions with authentication", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, TransactionObjects{
				Data: &[]TransactionObject{},
			}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.GetAllTransactions(context.Background(), nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "/v1/transactions")
	})

	t.Run("GetTransaction with ID", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, TransactionObject{}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.GetTransaction(context.Background(), "txn-test-123", nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Contains(t, mockRT.LastReq.URL.Path, "txn-test-123")
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
		Response: mockJSONResponse(200, TransactionObjects{
			Data: &[]TransactionObject{},
		}),
	}

	client, err := NewClient("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}),
		WithRequestEditorFn(editor))
	require.NoError(t, err)

	_, err = client.GetAllTransactions(context.Background(), nil)
	require.NoError(t, err)

	assert.True(t, editorCalled)
	assert.Equal(t, "test-value", mockRT.LastReq.Header.Get("X-Custom-Header"))
}
