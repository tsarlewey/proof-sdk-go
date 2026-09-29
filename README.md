# proof-sdk-go

Auto-generated Go SDK clients for the Proof API, plus a shared authentication adapter.

This module contains the SDK pieces extracted from [proof-cli](https://github.com/tsarlewey/proof-cli). It is intended to be imported by any Go application that needs to talk to the Proof API — the CLI itself, internal services, scripts, or third-party tooling.

Integrating with a coding agent? Start with [AGENTS.md](AGENTS.md): quickstart, rules and quirks on one page.

## Packages

| Package | Purpose |
|---------|---------|
| `business` | Business API client (transactions, documents, webhooks, notaries, templates, referrals, integrations) |
| `realestate` | Real Estate / Mortgage API client |
| `scim` | SCIM identity management client |
| `logs` | Security Events API client |
| `certificates` | Organization Certificates API client |
| `credentials` | Verifiable Credentials API: presentation authorize (browser redirect — use `NewAuthorizeVerifiableCredentialPresentationRequest` to build the URL, not the client) and x401 handoff |
| `common` | `APIKey` auth, `SandboxURL`/`ProductionURL`, `AuthenticatedDoer` + `AuthProvider`, error and retry helpers shared across SDK packages |

All of the per-API clients are generated from OpenAPI specs via [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen). Do not edit `*/client.gen.go` by hand — regenerate instead (see below).

## Installation

```bash
go get github.com/tsarlewey/proof-sdk-go@latest
```

## Quick start

```go
import (
    "github.com/tsarlewey/proof-sdk-go/business"
    "github.com/tsarlewey/proof-sdk-go/common"
)

authDoer := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
client, err := business.NewClientWithResponses(
    common.SandboxURL, // https://api.fairfax.proof.com; use common.ProductionURL in production
    business.WithHTTPClient(authDoer),
)
```

`common.APIKey` sends the `ApiKey` header through an `*http.Client` with a 60s timeout.

### Custom authentication

For OAuth 2.0 client credentials, or to control the HTTP client, implement `common.AuthProvider` (two methods: `AddAuthHeaders(*http.Request) error` and `HTTPClient() *http.Client`) and pass it to `common.NewAuthenticatedDoer` in place of `common.APIKey(...)`. See `proof-cli`'s `pkg/utils/client.go` for a complete implementation covering both OAuth 2.0 client-credentials and API-key authentication.

## Examples

Runnable programs live in [`examples/`](examples). Each one reads `PROOF_API_KEY`, targets the sandbox unless `PROOF_BASE_URL` is set, and prints JSON:

```bash
PROOF_API_KEY=... go run ./examples/business-list-transactions
```

`business-create-transaction`, `business-list-transactions`, `realestate-list-transactions`, `webhook-events`, `scim-list-users`.

## User-Agent

`AuthenticatedDoer` sets `User-Agent: proof-sdk-go/<version>` on every request, but only when the caller hasn't already set one. To brand requests for your own application, set `User-Agent` on the request (or via `WithRequestEditorFn`) before it reaches the doer; your value is preserved.

The version is exported as `common.Version`.

## Error helpers

The generated clients return a nil error for every HTTP status. The `*WithResponse` methods return typed response structs whose `.JSONxxx` fields are nil on an error status. To turn a status of 400 or higher into a typed error:

```go
resp, err := client.GetTransactionWithResponse(ctx, id, nil)
if err != nil { return err }

if apiErr := common.CheckResponseBody(resp.HTTPResponse, resp.Body); apiErr != nil {
    var e *common.APIError
    if errors.As(apiErr, &e) {
        log.Printf("proof api %d: %s", e.StatusCode, e.Message)
    }
    return apiErr
}
```

Use `CheckResponseBody` with `*WithResponse` results, because the generated parser has already read and closed `resp.HTTPResponse.Body` and the bytes are on `resp.Body`. For a raw `*http.Response` from the plain `Client`, use `CheckResponse(resp)` or its comma-ok form `AsAPIError(resp) (*APIError, bool)`. Both drain the response body and replace it with a re-readable buffer, so the body is captured on `APIError.Body` and still readable from `resp.Body` afterwards. `Message` is a best-effort extraction across the common Proof error shapes (`error`, `message`, `detail`, `errors[]`).

## Retries

`common.RetryDoer` wraps any `HTTPDoer` (including `AuthenticatedDoer`) with exponential-backoff retries on retryable status codes and network errors. It honors `Retry-After` (seconds form) and respects request context cancellation while sleeping.

```go
authDoer := common.NewAuthenticatedDoer(common.APIKey(apiKey))
retrying := common.NewRetryDoer(authDoer, common.DefaultRetryConfig())

client, _ := business.NewClientWithResponses(
    common.ProductionURL,
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
- `scripts/fix-scim-operation-ids.py` — rewrites SCIM `operationId` values so generated methods are named `ListUsersWithResponse`, `GetUserWithResponse`, `CreateUser…`, `ReplaceUser…`, `PatchUser…`, `DeleteUserWithResponse`, `RetrieveUsersSchemaWithResponse`, `GetResourceTypesWithResponse`, `GetServiceProviderConfigWithResponse`.

Individual regeneration is possible via `make generate` without re-downloading. `generate` also detaches oapi-codegen's generic package comment so each package's hand-written `doc.go` is the one pkg.go.dev shows.

## Development

```bash
make check       # fmt + vet + build + test-race
make test        # just the tests
```

## Known quirks

### SCIM PATCH operation `Value` is typed as an object
`Operations` is an array, but each operation's `Value` is `*map[string]interface{}`. To send a string, bool or array value with a `path`, call `PatchUserWithBodyWithResponse` with a manually marshaled `application/scim+json` body (see `proof-cli/cmd/scim.go` for an example).

### SCIM responses are untyped
SCIM results have no `JSON200` field. Decode `resp.Body` yourself (see `examples/scim-list-users`).

### Verifiable Credentials authorize is a browser redirect
Build the URL with `credentials.NewAuthorizeVerifiableCredentialPresentationRequest` and redirect the End-User to it rather than calling it with the client. The x401 handoff endpoint (`CreatePresentationX401HandoffWithResponse`) is a normal JSON POST.

## License

Same as the consuming CLI (`proof-cli`).
