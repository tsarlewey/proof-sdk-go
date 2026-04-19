package realestate

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

// TestEnumValidation verifies enum validation methods
func TestEnumValidation(t *testing.T) {
	t.Run("ContactRole valid values", func(t *testing.T) {
		assert.True(t, ContactRoleAttorney.Valid())
		assert.True(t, ContactRoleCloser.Valid())
		assert.True(t, ContactRoleLoanOfficer.Valid())
		assert.True(t, ContactRoleTitleAgent.Valid())
	})

	t.Run("ContactRole invalid value", func(t *testing.T) {
		invalid := ContactRole("invalid")
		assert.False(t, invalid.Valid())
	})

	t.Run("CosignerSigningRequirement valid values", func(t *testing.T) {
		assert.True(t, CosignerSigningRequirementEsign.Valid())
		assert.True(t, CosignerSigningRequirementIdentify.Valid())
		assert.True(t, CosignerSigningRequirementVerify.Valid())
	})

	t.Run("CosignerSigningRequirement invalid value", func(t *testing.T) {
		invalid := CosignerSigningRequirement("invalid")
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
}

// TestNewGetAllMortgageTransactionsRequest verifies list transactions request generation
func TestNewGetAllMortgageTransactionsRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewGetAllMortgageTransactionsRequest("https://api.example.com/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/mortgage/v2/transactions")
	})

	t.Run("includes limit parameter", func(t *testing.T) {
		params := &GetAllMortgageTransactionsParams{
			Limit: ptr(50),
		}
		req, err := NewGetAllMortgageTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=50")
	})

	t.Run("includes offset parameter", func(t *testing.T) {
		params := &GetAllMortgageTransactionsParams{
			Offset: ptr(10),
		}
		req, err := NewGetAllMortgageTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "offset=10")
	})

	t.Run("includes organization_id parameter", func(t *testing.T) {
		params := &GetAllMortgageTransactionsParams{
			OrganizationId: ptr("org-123"),
		}
		req, err := NewGetAllMortgageTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "organization_id=org-123")
	})

	t.Run("includes loan_number parameter", func(t *testing.T) {
		params := &GetAllMortgageTransactionsParams{
			LoanNumber: ptr("LN-12345"),
		}
		req, err := NewGetAllMortgageTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "loan_number=LN-12345")
	})

	t.Run("includes transaction_status parameter", func(t *testing.T) {
		status := GetAllMortgageTransactionsParamsTransactionStatus("active")
		params := &GetAllMortgageTransactionsParams{
			TransactionStatus: &status,
		}
		req, err := NewGetAllMortgageTransactionsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "transaction_status=active")
	})
}

// TestNewGetMortgageTransactionRequest verifies get transaction request generation
func TestNewGetMortgageTransactionRequest(t *testing.T) {
	t.Run("generates request with transaction ID", func(t *testing.T) {
		req, err := NewGetMortgageTransactionRequest("https://api.example.com/", "txn-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/mortgage/v2/transactions/txn-123")
	})
}

// TestNewDeleteMortgageTransactionRequest verifies delete transaction request generation
func TestNewDeleteMortgageTransactionRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteMortgageTransactionRequest("https://api.example.com/", "txn-123")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/mortgage/v2/transactions/txn-123")
	})
}

// TestNewGetRecordingLocationsRequest verifies recording locations request generation
func TestNewGetRecordingLocationsRequest(t *testing.T) {
	t.Run("generates request with transaction type", func(t *testing.T) {
		txnType := GetRecordingLocationsParamsTransactionType("refinance")
		params := &GetRecordingLocationsParams{
			TransactionType: txnType,
		}
		req, err := NewGetRecordingLocationsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/mortgage/v2/recording_locations")
		assert.Contains(t, req.URL.RawQuery, "transaction_type=refinance")
	})

	t.Run("includes address parameters", func(t *testing.T) {
		txnType := GetRecordingLocationsParamsTransactionType("purchase_buyer_loan")
		params := &GetRecordingLocationsParams{
			TransactionType:      txnType,
			StreetAddressLine1:   ptr("123 Main St"),
			StreetAddressCity:    ptr("Los Angeles"),
			StreetAddressState:   ptr("CA"),
			StreetAddressPostal:  ptr("90001"),
			StreetAddressCountry: ptr("US"),
		}
		req, err := NewGetRecordingLocationsRequest("https://api.example.com/", params)
		require.NoError(t, err)
		// Encoded parameters
		assert.Contains(t, req.URL.RawQuery, "street_address")
	})
}

// TestNewGetAllMortgageNotariesRequest verifies list notaries request generation
func TestNewGetAllMortgageNotariesRequest(t *testing.T) {
	t.Run("generates request", func(t *testing.T) {
		req, err := NewGetAllMortgageNotariesRequest("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/mortgage/v1/notaries")
	})
}

// TestNewGetMortgageNotaryRequest verifies get notary request generation
func TestNewGetMortgageNotaryRequest(t *testing.T) {
	t.Run("generates request with notary ID", func(t *testing.T) {
		req, err := NewGetMortgageNotaryRequest("https://api.example.com/", "notary-456")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/mortgage/v1/notaries/notary-456")
	})
}

// TestNewDeleteMortgageNotaryRequest verifies delete notary request generation
func TestNewDeleteMortgageNotaryRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteMortgageNotaryRequest("https://api.example.com/", "notary-456")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/mortgage/v1/notaries/notary-456")
	})
}

// TestNewGetMortgageWebhookURLRequest verifies get webhook URL request generation
func TestNewGetMortgageWebhookURLRequest(t *testing.T) {
	t.Run("generates request", func(t *testing.T) {
		req, err := NewGetMortgageWebhookURLRequest("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/mortgage/v1/webhooks")
	})
}

// TestNewDeleteMortgageWebhookURLRequest verifies delete webhook URL request generation
func TestNewDeleteMortgageWebhookURLRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteMortgageWebhookURLRequest("https://api.example.com/")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.String(), "/mortgage/v1/webhooks")
	})
}

// TestNewGetMortgageNotarizationRecordRequest verifies get notarization record request generation
func TestNewGetMortgageNotarizationRecordRequest(t *testing.T) {
	t.Run("generates request with record ID", func(t *testing.T) {
		req, err := NewGetMortgageNotarizationRecordRequest("https://api.example.com/", "record-789", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/mortgage/v2/notarization_records/record-789")
	})
}

// TestNewGetMortgageDocumentRequest verifies get document request generation
func TestNewGetMortgageDocumentRequest(t *testing.T) {
	t.Run("generates request with transaction and document IDs", func(t *testing.T) {
		req, err := NewGetMortgageDocumentRequest("https://api.example.com/", "txn-123", "doc-456", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/mortgage/v2/transactions/txn-123/documents/doc-456")
	})
}

// TestNewDeleteMortgageDocumentRequest verifies delete document request generation
func TestNewDeleteMortgageDocumentRequest(t *testing.T) {
	t.Run("generates DELETE request", func(t *testing.T) {
		req, err := NewDeleteMortgageDocumentRequest("https://api.example.com/", "doc-123")
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/mortgage/v2/documents/doc-123")
	})
}

// TestNewVerifyPropertyAddressRequest verifies address verification request generation
func TestNewVerifyPropertyAddressRequest(t *testing.T) {
	t.Run("generates request with address params", func(t *testing.T) {
		txnType := VerifyPropertyAddressParamsTransactionType("refinance")
		params := &VerifyPropertyAddressParams{
			TransactionType:     txnType,
			StreetAddressLine1:  "456 Oak Ave",
			StreetAddressCity:   "San Francisco",
			StreetAddressState:  "CA",
			StreetAddressPostal: "94102",
		}
		req, err := NewVerifyPropertyAddressRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/mortgage/v2/transactions/verify_address")
	})
}

// TestParseGetAllMortgageTransactionsResponse verifies list transactions
// response parsing by asserting on deserialized payload fields.
func TestParseGetAllMortgageTransactionsResponse(t *testing.T) {
	t.Run("deserializes nested transaction fields", func(t *testing.T) {
		body := TransactionObjects{
			Data: &[]TransactionObject{
				{
					Id:         ptr("txn-123"),
					ExternalId: ptr("ext-abc"),
					FileNumber: ptr("FILE-42"),
					LoanNumber: ptr("LOAN-99"),
					Cost:       ptr(float32(149.95)),
				},
				{Id: ptr("txn-456")},
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetAllMortgageTransactionsResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON200)
		require.NotNil(t, parsed.JSON200.Data)
		require.Len(t, *parsed.JSON200.Data, 2)

		first := (*parsed.JSON200.Data)[0]
		assert.Equal(t, "txn-123", *first.Id)
		assert.Equal(t, "ext-abc", *first.ExternalId)
		assert.Equal(t, "FILE-42", *first.FileNumber)
		assert.Equal(t, "LOAN-99", *first.LoanNumber)
		assert.InDelta(t, 149.95, *first.Cost, 0.001)

		assert.Equal(t, "txn-456", *(*parsed.JSON200.Data)[1].Id)
	})

	t.Run("empty body yields nil JSON200 without error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 204,
			Status:     http.StatusText(204),
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
			Header:     http.Header{},
		}

		parsed, err := ParseGetAllMortgageTransactionsResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 204, parsed.StatusCode())
		assert.Nil(t, parsed.JSON200)
	})
}

// TestParseGetMortgageTransactionResponse verifies get transaction response parsing
func TestParseGetMortgageTransactionResponse(t *testing.T) {
	t.Run("parses 200 response", func(t *testing.T) {
		body := TransactionObject{
			Id: ptr("txn-123"),
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetMortgageTransactionResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
		assert.Equal(t, "txn-123", *parsed.JSON200.Id)
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

	t.Run("GetAllMortgageTransactionsWithResponse", func(t *testing.T) {
		resp, err := client.GetAllMortgageTransactionsWithResponse(context.Background(), nil)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode())
	})
}

// TestClientMakesRequest verifies client makes correct requests
func TestClientMakesRequest(t *testing.T) {
	t.Run("GetAllMortgageTransactions", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, TransactionObjects{
				Data: &[]TransactionObject{},
			}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.GetAllMortgageTransactions(context.Background(), nil)
		require.NoError(t, err)

		assert.NotNil(t, mockRT.LastReq)
		assert.Equal(t, "GET", mockRT.LastReq.Method)
		assert.Contains(t, mockRT.LastReq.URL.Path, "/mortgage/v2/transactions")
	})

	t.Run("GetMortgageTransaction with ID", func(t *testing.T) {
		mockRT := &MockRoundTripper{
			Response: mockJSONResponse(200, TransactionObject{}),
		}

		client, err := NewClient("https://api.example.com",
			WithHTTPClient(&http.Client{Transport: mockRT}))
		require.NoError(t, err)

		_, err = client.GetMortgageTransaction(context.Background(), "txn-test-123", nil)
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

	_, err = client.GetAllMortgageTransactions(context.Background(), nil)
	require.NoError(t, err)

	assert.True(t, editorCalled)
	assert.Equal(t, "test-value", mockRT.LastReq.Header.Get("X-Custom-Header"))
}
