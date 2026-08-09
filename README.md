# proof-sdk-go

Auto-generated Go SDK clients for the Proof API, plus a shared authentication adapter.

This module contains the SDK pieces extracted from [proof-cli](https://github.com/tsarlewey/proof-cli). It is intended to be imported by any Go application that needs to talk to the Proof API — the CLI itself, internal services, scripts, or third-party tooling.

## Packages

| Package | Purpose |
|---------|---------|
| `business` | Business API client (transactions, documents, webhooks, notaries, templates, referrals, integrations) |
| `realestate` | Real Estate / Mortgage API client |
| `scim` | SCIM identity management client |
| `logs` | Security Events API client |
| `certificates` | Organization Certificates API client |
| `credentials` | Verifiable Credentials presentation authorize endpoint (browser redirect — use `NewAuthorizeVerifiableCredentialPresentationRequest` to build the URL, not the client) |
| `common` | `AuthenticatedDoer` + `AuthProvider` interface shared across SDK packages |

All of the per-API clients are generated from OpenAPI specs via [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen). Do not edit `*/client.gen.go` by hand — regenerate instead (see below).

## Installation

```bash
go get github.com/tsarlewey/proof-sdk-go@latest
```

## Quick start

Implement `common.AuthProvider` (two methods: `AddAuthHeaders(*http.Request) error` and `HTTPClient() *http.Client`), then wrap it with `common.NewAuthenticatedDoer` and pass to any of the generated clients.

```go
import (
    "github.com/tsarlewey/proof-sdk-go/business"
    "github.com/tsarlewey/proof-sdk-go/common"
)

authDoer := common.NewAuthenticatedDoer(myAuthProvider)
client, err := business.NewClientWithResponses(
    "https://api.proof.com",
    business.WithHTTPClient(authDoer),
)
```

See `proof-cli`'s `pkg/utils/client.go` for a complete `AuthProvider` implementation covering both OAuth 2.0 client-credentials and API-key authentication.

## User-Agent

`AuthenticatedDoer` sets `User-Agent: proof-sdk-go/<version>` on every request, but only when the caller hasn't already set one. To brand requests for your own application, set `User-Agent` on the request (or via `WithRequestEditorFn`) before it reaches the doer; your value is preserved.

The version is exported as `common.Version`.

## Error helpers

The generated clients return raw `*http.Response` (or typed response structs whose `.JSONxxx` fields are nil on error status). To uniformly extract a non-2xx response into a typed error:

```go
resp, err := client.GetTransactionWithResponse(ctx, id)
if err != nil { return err }

if apiErr := common.CheckResponse(resp.HTTPResponse); apiErr != nil {
    var e *common.APIError
    if errors.As(apiErr, &e) {
        log.Printf("proof api %d: %s", e.StatusCode, e.Message)
    }
    return apiErr
}
```

`AsAPIError(resp) (*APIError, bool)` is the comma-ok form. Both functions drain the response body and replace it with a re-readable buffer, so the body is captured on `APIError.Body` and still readable from `resp.Body` afterwards. `Message` is a best-effort extraction across the common Proof error shapes (`error`, `message`, `detail`, `errors[]`).

## Retries

`common.RetryDoer` wraps any `HTTPDoer` (including `AuthenticatedDoer`) with exponential-backoff retries on retryable status codes and network errors. It honors `Retry-After` (seconds form) and respects request context cancellation while sleeping.

```go
authDoer := common.NewAuthenticatedDoer(myAuthProvider)
retrying := common.NewRetryDoer(authDoer, common.DefaultRetryConfig())

client, _ := business.NewClientWithResponses(
    "https://api.proof.com",
    business.WithHTTPClient(retrying),
)
```

`DefaultRetryConfig()` retries up to 3 times on 408/429/5xx and on transport errors, with full-jitter exponential backoff capped at 5s. Override `RetryConfig` fields to tune.

Requests with bodies must be replayable (`req.GetBody` non-nil — the standard library populates this for the body types `oapi-codegen` uses). Otherwise the doer returns `common.ErrBodyNotReplayable` rather than a stale response, so callers can distinguish a real terminal status from an abandoned retry.

## Regenerating SDKs

```bash
make tools           # installs oapi-codegen (once)
make regenerate      # download-specs + generate + build + test
```

`make download-specs` fetches specs from `dev.proof.com` and applies two fixup scripts:

- `scripts/fix-openapi-refs.py` — inlines deeply nested `$ref` chains that `oapi-codegen` cannot resolve.
- `scripts/fix-scim-operation-ids.py` — rewrites SCIM `operationId` values so generated methods are named `ListUsersWithResponse`, `GetUserWithResponse`, `UpdateUserWithResponse`, `PatchUserWithResponse`, `DeleteUserWithResponse`, `GetResourceTypesWithResponse`, `GetServiceProviderConfigWithResponse`.

Individual regeneration is possible via `make generate` without re-downloading.

## Development

```bash
make check       # fmt + vet + build + test-race
make test        # just the tests
```

## Known quirks

### SCIM PATCH `Operations` is an object, not an array
The upstream SCIM OpenAPI spec types the PATCH `Operations` field as a single object. SCIM requires an array. Callers that need to send PATCH should use `PatchUserWithBodyWithResponse` with a manually marshaled JSON body (see `proof-cli/cmd/scim.go` for an example).

### SCIM user `active` field is `*string`, not `*bool`
The upstream spec types the `active` field on users as `*string` (`"true"` / `"false"`). Callers should convert `bool` to `*string` at the call site.

## License

Same as the consuming CLI (`proof-cli`).
