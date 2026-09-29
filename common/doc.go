// Package common holds the hand-written pieces shared by every Proof API
// client in this module: authentication, base URLs, error handling, retries
// and the SDK version.
//
// Wire an API key into any generated client (sandbox shown):
//
//	auth := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
//	client, err := business.NewClientWithResponses(common.SandboxURL, business.WithHTTPClient(auth))
//	if err != nil { ... }
//	resp, err := client.GetAllTransactionsWithResponse(ctx, nil)
//	if err == nil { err = common.CheckResponseBody(resp.HTTPResponse, resp.Body) }
//
// Use the *WithResponse methods of the generated clients. They return a nil
// error for any HTTP status and leave JSON200 nil on other statuses, so check
// every result with CheckResponseBody(resp.HTTPResponse, resp.Body), which
// returns an *APIError for status >= 400. Use CheckResponse only with a raw
// *http.Response whose body has not been read yet.
//
// APIKey covers API-key auth. For OAuth 2.0 client credentials, implement
// AuthProvider yourself; pkg/utils/client.go in
// https://github.com/tsarlewey/proof-cli has a complete implementation.
//
// Wrap the doer in a RetryDoer to retry 408/429/5xx and transport errors:
//
//	doer := common.NewRetryDoer(auth, common.DefaultRetryConfig())
package common
