# Using proof-sdk-go

Guidance for developers, and their coding agents, who call the [Proof API](https://dev.proof.com) (notarization, e-signing, identity) from Go with this module. It is an unofficial SDK: `oapi-codegen` output from Proof's OpenAPI specs, plus a small hand-written `common` package.

## Install

```bash
go get github.com/tsarlewey/proof-sdk-go@latest
```

## Quickstart (sandbox)

```go
auth := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
client, err := business.NewClientWithResponses(common.SandboxURL, business.WithHTTPClient(auth))
if err != nil {
	return err
}
resp, err := client.GetAllTransactionsWithResponse(ctx, nil)
if err != nil {
	return err
}
if err := common.CheckResponseBody(resp.HTTPResponse, resp.Body); err != nil {
	return err // *common.APIError: StatusCode, Message, Body
}
fmt.Println(*resp.JSON200.TotalCount)
```

`common.SandboxURL` is `https://api.fairfax.proof.com` (Proof's Fairfax sandbox; sandbox keys only work there). Switch to `common.ProductionURL` (`https://api.proof.com`) with a production key.

## Packages

| Import `github.com/tsarlewey/proof-sdk-go/...` | API | Typical entry points |
|---|---|---|
| `business` | Business API | `CreateTransactionWithResponse`, `GetAllTransactionsWithResponse`, `ActivateDraftTransactionWithResponse`, `GetAllWebhooksV2WithResponse` |
| `realestate` | Real Estate / mortgage API | `CreateMortgageTransactionWithResponse`, `GetAllMortgageTransactionsWithResponse`, `PlaceOrderWithResponse` |
| `scim` | SCIM 2.0 user provisioning | `ListUsersWithResponse`, `GetUserWithResponse`, `PatchUserWithBodyWithResponse` (all take the organization ID) |
| `logs` | Security events | `ListSecurityEventsWithResponse` |
| `certificates` | Organization certificates | `GetV1CertificatesWithResponse`, `PostV2CertificatesWithResponse`, `PostV1CertificatesIdSignWithResponse` |
| `credentials` | Verifiable Credentials (OID4VP) | `NewAuthorizeVerifiableCredentialPresentationRequest` (build a redirect URL), `CreatePresentationX401HandoffWithResponse` |
| `common` | Auth, base URLs, errors, retries | `APIKey`, `NewAuthenticatedDoer`, `CheckResponseBody`, `NewRetryDoer` |

## Rules that prevent most mistakes

1. **Call the `*WithResponse` methods** on the client from `NewClientWithResponses`. They return a typed result with `Body []byte`, `HTTPResponse`, and `JSON200` / `JSON201` / `JSON4xx` fields.
2. **A nil `err` does not mean success.** Any HTTP status comes back with `err == nil`, and `JSON200` is nil unless the status was 200. Check every call with `common.CheckResponseBody(resp.HTTPResponse, resp.Body)` before touching `JSON200`. Don't use `common.CheckResponse(resp.HTTPResponse)` on a `*WithResponse` result: the generated parser has already read and closed that body, so the error message gets lost. `CheckResponse` is for raw `*http.Response` values from the plain `Client`.
3. **Pass `DocumentUrlVersion: "v2"`** (the `...DocumentUrlVersionV2` constants) on calls that return documents. You get Proof secure URLs; `v1` returns AWS S3 pre-signed URLs.
4. **Optional fields are pointers.** The SDK has no pointer helper. Add this one-liner:
   ```go
   func ptr[T any](v T) *T { return &v }
   // business.TransactionCreateParams{Draft: ptr(true), Signer: business.Signer{Email: "a@b.com"}}
   ```
5. **Authentication:** `common.APIKey(key)` sends the `ApiKey` header and returns an error before sending if `key` is empty. For OAuth 2.0 client credentials, implement `common.AuthProvider` (two methods). `pkg/utils/client.go` in [proof-cli](https://github.com/tsarlewey/proof-cli) has a working implementation.
6. **Retries:** wrap the doer with `common.NewRetryDoer(auth, common.DefaultRetryConfig())` to retry 408/429/5xx and transport errors.
7. **Never edit `*/client.gen.go`.** It is regenerated from the specs.

## Where the field docs are

- The generated structs carry the API's field descriptions as comments. Run `go doc github.com/tsarlewey/proof-sdk-go/business.TransactionCreateParams` or `go doc ...business.Signer`, or read `business/client.gen.go`.
- Package overviews: https://pkg.go.dev/github.com/tsarlewey/proof-sdk-go
- API guides and reference: https://dev.proof.com
- A struct's JSON tags are the wire names used in the dev.proof.com docs (for example `TransactionCreateParams.Signer` is `signer`).

## Known quirks

- **Business create transaction:** `TransactionCreateParams.Signer` is required. `Documents` is `*[]string`, where each entry is a PDF URL, base64 file contents, or a template permalink. Create with `Draft: ptr(true)` to review first, then send it with `ActivateDraftTransactionWithResponse` (the `notarization_ready` endpoint).
- **SCIM responses are untyped:** the results have no `JSON200`, so decode `resp.Body` (a SCIM `ListResponse` or `User`) yourself.
- **SCIM PATCH:** `Operations` is an array, but each operation's `Value` is typed `*map[string]interface{}`. To send a string, bool or array value with a `path`, marshal the body yourself and call `PatchUserWithBodyWithResponse(ctx, orgID, userID, "application/scim+json", body)`.
- **Credentials authorize** answers with a 302 to a Proof-hosted page, so don't call it through the client. Build the URL with `credentials.NewAuthorizeVerifiableCredentialPresentationRequest(common.ProductionURL, &params)` and redirect the browser to `req.URL.String()`. The x401 handoff (`CreatePresentationX401HandoffWithResponse`) is a normal JSON POST that returns `JSON201.RequestUri`. The spec declares no API-key security for either endpoint.
- **Certificates** accepts either an API key or an OAuth client-credentials token.

## Examples

Each example is runnable (`go run ./examples/<name>`). It reads `PROOF_API_KEY`, targets the sandbox unless `PROOF_BASE_URL` is set, and prints JSON.

| Directory | Shows |
|---|---|
| [`examples/business-create-transaction`](examples/business-create-transaction/main.go) | Draft transaction with a signer and a document (URL or local file), optional activation |
| [`examples/business-list-transactions`](examples/business-list-transactions/main.go) | Listing transactions with params |
| [`examples/realestate-list-transactions`](examples/realestate-list-transactions/main.go) | Mortgage transactions, filtered by loan number |
| [`examples/webhook-events`](examples/webhook-events/main.go) | Webhooks v2 and their delivery events |
| [`examples/scim-list-users`](examples/scim-list-users/main.go) | SCIM users with a filter, decoding the untyped body (needs `PROOF_ORGANIZATION_ID`) |

## Shell and MCP access

[proof-cli](https://github.com/tsarlewey/proof-cli) is built on this SDK. Use it to call the Proof API from a shell, or expose it to an agent as an MCP server, without writing Go.

## For SDK contributors

- `make regenerate` downloads the specs from dev.proof.com, applies the fixup scripts in `scripts/`, regenerates, builds and tests. `make generate` regenerates from the checked-in `openapi/` specs.
- `make check` runs fmt, vet, build and `test -race`. CI runs it, plus a `gofmt -l` check.
- Never hand-edit `*/client.gen.go`. Fix the spec with a script in `scripts/`, or post-process in the Makefile's `generate` target.
- Hand-written code lives in `common/` and in each package's `doc.go` and `*_test.go`. Bump `common.Version` when the exported API changes.
