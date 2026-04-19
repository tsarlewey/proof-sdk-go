package certificates

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

// TestCertificateProfileEnum verifies enum validation
func TestCertificateProfileEnum(t *testing.T) {
	t.Run("V1 valid profiles", func(t *testing.T) {
		assert.True(t, PostV1CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl1.Valid())
		assert.True(t, PostV1CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl2.Valid())
		assert.True(t, PostV1CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl3.Valid())
		assert.True(t, PostV1CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl4.Valid())
	})

	t.Run("V1 invalid profile", func(t *testing.T) {
		invalid := PostV1CertificatesJSONBodyCertificateProfile("invalid")
		assert.False(t, invalid.Valid())
	})

	t.Run("V2 valid profiles", func(t *testing.T) {
		assert.True(t, PostV2CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl1.Valid())
		assert.True(t, PostV2CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl2.Valid())
		assert.True(t, PostV2CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl3.Valid())
		assert.True(t, PostV2CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl4.Valid())
	})

	t.Run("V2 invalid profile", func(t *testing.T) {
		invalid := PostV2CertificatesJSONBodyCertificateProfile("invalid")
		assert.False(t, invalid.Valid())
	})

	t.Run("algorithm OID valid", func(t *testing.T) {
		assert.True(t, N1284010045432.Valid())
	})

	t.Run("algorithm OID invalid", func(t *testing.T) {
		invalid := PostV1CertificatesIdSignJSONBodyAlgorithmOid("invalid")
		assert.False(t, invalid.Valid())
	})
}

// TestNewGetV1CertificatesRequest verifies list certificates request generation
func TestNewGetV1CertificatesRequest(t *testing.T) {
	t.Run("generates request without params", func(t *testing.T) {
		req, err := NewGetV1CertificatesRequest("https://api.example.com/", nil)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "https://api.example.com/v1/certificates", req.URL.String())
	})

	t.Run("includes limit parameter", func(t *testing.T) {
		params := &GetV1CertificatesParams{
			Limit: ptr("10"),
		}
		req, err := NewGetV1CertificatesRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=10")
	})

	t.Run("includes offset parameter", func(t *testing.T) {
		params := &GetV1CertificatesParams{
			Offset: ptr("20"),
		}
		req, err := NewGetV1CertificatesRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "offset=20")
	})

	t.Run("includes multiple parameters", func(t *testing.T) {
		params := &GetV1CertificatesParams{
			Limit:  ptr("10"),
			Offset: ptr("0"),
		}
		req, err := NewGetV1CertificatesRequest("https://api.example.com/", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "limit=10")
		assert.Contains(t, req.URL.RawQuery, "offset=0")
	})
}

// TestNewGetV1CertificatesIdRequest verifies get certificate request generation
func TestNewGetV1CertificatesIdRequest(t *testing.T) {
	t.Run("generates request with certificate ID", func(t *testing.T) {
		req, err := NewGetV1CertificatesIdRequest("https://api.example.com/", "cert-123")
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/certificates/cert-123")
	})

	t.Run("handles UUID format ID", func(t *testing.T) {
		req, err := NewGetV1CertificatesIdRequest("https://api.example.com/", "550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, err)
		assert.Contains(t, req.URL.Path, "550e8400-e29b-41d4-a716-446655440000")
	})
}

// TestNewDeleteV1CertificatesIdRequest verifies revoke certificate request generation
func TestNewDeleteV1CertificatesIdRequest(t *testing.T) {
	t.Run("generates request without reason", func(t *testing.T) {
		req, err := NewDeleteV1CertificatesIdRequest("https://api.example.com/", "cert-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/certificates/cert-123")
	})

	t.Run("includes reason parameter", func(t *testing.T) {
		params := &DeleteV1CertificatesIdParams{
			Reason: ptr("keyCompromise"),
		}
		req, err := NewDeleteV1CertificatesIdRequest("https://api.example.com/", "cert-123", params)
		require.NoError(t, err)
		assert.Contains(t, req.URL.RawQuery, "reason=keyCompromise")
	})
}

// TestNewPostV1CertificatesRequest verifies create certificate request generation
func TestNewPostV1CertificatesRequest(t *testing.T) {
	t.Run("generates request with body", func(t *testing.T) {
		body := PostV1CertificatesJSONRequestBody{
			CommonNamerequired: ptr("Test Certificate"),
			CertificateProfile: (*PostV1CertificatesJSONBodyCertificateProfile)(ptr(string(PostV1CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl1))),
		}
		req, err := NewPostV1CertificatesRequest("https://api.example.com/", body)
		require.NoError(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/certificates")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// TestNewPostV1CertificatesIdSignRequest verifies sign request generation
func TestNewPostV1CertificatesIdSignRequest(t *testing.T) {
	t.Run("generates request with digests", func(t *testing.T) {
		body := PostV1CertificatesIdSignJSONRequestBody{
			AlgorithmOid: (*PostV1CertificatesIdSignJSONBodyAlgorithmOid)(ptr(string(N1284010045432))),
			Digests:      &[]string{"YWJjZGVm", "Z2hpamts"},
		}
		req, err := NewPostV1CertificatesIdSignRequest("https://api.example.com/", "cert-123", body)
		require.NoError(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.Path, "/v1/certificates/cert-123/sign")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// TestNewPostV2CertificatesRequest verifies create certificate with CSR request generation
func TestNewPostV2CertificatesRequest(t *testing.T) {
	t.Run("generates request with CSR", func(t *testing.T) {
		body := PostV2CertificatesJSONRequestBody{
			Csrrequired:        ptr("-----BEGIN CERTIFICATE REQUEST-----\nMIIC..."),
			CertificateProfile: (*PostV2CertificatesJSONBodyCertificateProfile)(ptr(string(PostV2CertificatesJSONBodyCertificateProfileOrganizationAuthenticityAl2))),
		}
		req, err := NewPostV2CertificatesRequest("https://api.example.com/", body)
		require.NoError(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Contains(t, req.URL.Path, "/v2/certificates")
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})
}

// TestParseGetV1CertificatesResponse verifies list response parsing
func TestParseGetV1CertificatesResponse(t *testing.T) {
	t.Run("deserializes certificate metadata", func(t *testing.T) {
		body := struct {
			Results *[]OrganizationCertificateShortResponse `json:"results,omitempty"`
		}{
			Results: &[]OrganizationCertificateShortResponse{
				{
					Id:           ptr("cert-123"),
					Subject:      ptr("CN=Test,O=Test Org"),
					SerialNumber: ptr("ABC123"),
					ValidFrom:    ptr("2024-01-01T00:00:00Z"),
					ValidTo:      ptr("2025-01-01T00:00:00Z"),
					External:     ptr(false),
				},
				{
					Id:           ptr("cert-456"),
					Subject:      ptr("CN=External,O=Partner"),
					SerialNumber: ptr("DEF456"),
					External:     ptr(true),
				},
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetV1CertificatesResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON200)
		require.NotNil(t, parsed.JSON200.Results)
		require.Len(t, *parsed.JSON200.Results, 2)

		first := (*parsed.JSON200.Results)[0]
		assert.Equal(t, "cert-123", *first.Id)
		assert.Equal(t, "CN=Test,O=Test Org", *first.Subject)
		assert.Equal(t, "ABC123", *first.SerialNumber)
		assert.Equal(t, "2024-01-01T00:00:00Z", *first.ValidFrom)
		assert.Equal(t, "2025-01-01T00:00:00Z", *first.ValidTo)
		assert.False(t, *first.External)

		second := (*parsed.JSON200.Results)[1]
		assert.Equal(t, "cert-456", *second.Id)
		assert.True(t, *second.External, "External flag must round-trip as true")
	})

	t.Run("parses empty results", func(t *testing.T) {
		body := struct {
			Results *[]OrganizationCertificateShortResponse `json:"results,omitempty"`
		}{
			Results: &[]OrganizationCertificateShortResponse{},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetV1CertificatesResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
		assert.Empty(t, *parsed.JSON200.Results)
	})
}

// TestParseGetV1CertificatesIdResponse verifies get certificate response parsing
func TestParseGetV1CertificatesIdResponse(t *testing.T) {
	t.Run("deserializes full certificate including chain and issuer", func(t *testing.T) {
		chain := "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----"
		body := struct {
			Result *OrganizationCertificateFullResponse `json:"result,omitempty"`
		}{
			Result: &OrganizationCertificateFullResponse{
				Id:               ptr("cert-123"),
				Subject:          ptr("CN=Test,O=Test Org"),
				Issuer:           ptr("CN=Issuer,O=Issuer Org"),
				SerialNumber:     ptr("ABC123"),
				CertificateChain: ptr(chain),
				ValidFrom:        ptr("2024-01-01T00:00:00Z"),
				ValidTo:          ptr("2025-01-01T00:00:00Z"),
				External:         ptr(false),
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseGetV1CertificatesIdResponse(resp)
		require.NoError(t, err)
		require.NotNil(t, parsed.JSON200)
		require.NotNil(t, parsed.JSON200.Result)

		r := parsed.JSON200.Result
		assert.Equal(t, "cert-123", *r.Id)
		assert.Equal(t, "CN=Test,O=Test Org", *r.Subject)
		assert.Equal(t, "CN=Issuer,O=Issuer Org", *r.Issuer)
		assert.Equal(t, "ABC123", *r.SerialNumber)
		assert.Equal(t, chain, *r.CertificateChain)
		assert.False(t, *r.External)
	})
}

// TestParseDeleteV1CertificatesIdResponse verifies revoke response parsing
func TestParseDeleteV1CertificatesIdResponse(t *testing.T) {
	t.Run("parses 200 response for successful revocation", func(t *testing.T) {
		body := struct {
			Result *OrganizationCertificateFullResponse `json:"result,omitempty"`
		}{
			Result: &OrganizationCertificateFullResponse{
				Id:               ptr("cert-123"),
				RevokedAt:        ptr("2024-06-01T00:00:00Z"),
				RevocationReason: ptr("keyCompromise"),
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParseDeleteV1CertificatesIdResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
		assert.NotNil(t, parsed.JSON200.Result.RevokedAt)
	})

	t.Run("parses 422 error response", func(t *testing.T) {
		body := struct {
			Errors *[]string `json:"errors,omitempty"`
		}{
			Errors: &[]string{"Certificate already revoked"},
		}
		resp := mockJSONResponse(422, body)

		parsed, err := ParseDeleteV1CertificatesIdResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 422, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON422)
		assert.Contains(t, *parsed.JSON422.Errors, "Certificate already revoked")
	})
}

// TestParsePostV1CertificatesResponse verifies create certificate response parsing
func TestParsePostV1CertificatesResponse(t *testing.T) {
	t.Run("parses 200 response with new certificate", func(t *testing.T) {
		body := struct {
			Result *OrganizationCertificateFullResponse `json:"result,omitempty"`
		}{
			Result: &OrganizationCertificateFullResponse{
				Id:               ptr("new-cert-456"),
				Subject:          ptr("CN=New Cert,O=Test Org"),
				CertificateChain: ptr("-----BEGIN CERTIFICATE-----\n..."),
				External:         ptr(false),
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParsePostV1CertificatesResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
		assert.Equal(t, "new-cert-456", *parsed.JSON200.Result.Id)
		assert.False(t, *parsed.JSON200.Result.External)
	})
}

// TestParsePostV2CertificatesResponse verifies create certificate with CSR response parsing
func TestParsePostV2CertificatesResponse(t *testing.T) {
	t.Run("parses 200 response with external certificate", func(t *testing.T) {
		body := struct {
			Result *OrganizationCertificateFullResponse `json:"result,omitempty"`
		}{
			Result: &OrganizationCertificateFullResponse{
				Id:               ptr("ext-cert-789"),
				Subject:          ptr("CN=External Cert,O=Test Org"),
				CertificateChain: ptr("-----BEGIN CERTIFICATE-----\n..."),
				External:         ptr(true),
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParsePostV2CertificatesResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
		assert.True(t, *parsed.JSON200.Result.External)
	})
}

// TestParsePostV1CertificatesIdSignResponse verifies sign response parsing
func TestParsePostV1CertificatesIdSignResponse(t *testing.T) {
	t.Run("parses 200 response with signatures", func(t *testing.T) {
		body := struct {
			Result *[]string `json:"result,omitempty"`
		}{
			Result: &[]string{
				"MEUCIQDbase64signature1...",
				"MEUCIQDbase64signature2...",
			},
		}
		resp := mockJSONResponse(200, body)

		parsed, err := ParsePostV1CertificatesIdSignResponse(resp)
		require.NoError(t, err)
		assert.Equal(t, 200, parsed.StatusCode())
		assert.NotNil(t, parsed.JSON200)
		assert.Len(t, *parsed.JSON200.Result, 2)
	})
}

// TestClientWithResponsesMethods verifies client with responses wrapper methods
func TestClientWithResponsesMethods(t *testing.T) {
	mockRT := &MockRoundTripper{
		Response: mockJSONResponse(200, struct {
			Results *[]OrganizationCertificateShortResponse `json:"results,omitempty"`
		}{
			Results: &[]OrganizationCertificateShortResponse{},
		}),
	}

	client, err := NewClientWithResponses("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}))
	require.NoError(t, err)

	t.Run("GetV1CertificatesWithResponse", func(t *testing.T) {
		resp, err := client.GetV1CertificatesWithResponse(context.Background(), nil)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode())
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
			Results *[]OrganizationCertificateShortResponse `json:"results,omitempty"`
		}{
			Results: &[]OrganizationCertificateShortResponse{},
		}),
	}

	client, err := NewClient("https://api.example.com",
		WithHTTPClient(&http.Client{Transport: mockRT}),
		WithRequestEditorFn(editor))
	require.NoError(t, err)

	_, err = client.GetV1Certificates(context.Background(), nil)
	require.NoError(t, err)

	assert.True(t, editorCalled)
	assert.Equal(t, "test-value", mockRT.LastReq.Header.Get("X-Custom-Header"))
}
