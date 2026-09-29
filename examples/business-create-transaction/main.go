// Command business-create-transaction creates a draft Business API transaction
// with one signer and one document, and optionally activates it (which emails
// the signer).
//
//	PROOF_API_KEY=... go run ./examples/business-create-transaction \
//	    -email signer@example.com -document https://example.com/contract.pdf [-activate]
//
// -document is a PDF URL or a local file path (sent base64-encoded).
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/tsarlewey/proof-sdk-go/business"
	"github.com/tsarlewey/proof-sdk-go/common"
)

func ptr[T any](v T) *T { return &v }

func main() {
	email := flag.String("email", "", "signer email (required)")
	document := flag.String("document", "", "PDF URL or local file path (required)")
	activate := flag.Bool("activate", false, "activate the draft, which notifies the signer")
	flag.Parse()
	if *email == "" || *document == "" {
		flag.Usage()
		os.Exit(2)
	}

	// A resource is a URL or the base64-encoded file contents.
	resource := *document
	if data, err := os.ReadFile(*document); err == nil {
		resource = base64.StdEncoding.EncodeToString(data)
	}

	baseURL := common.SandboxURL
	if u := os.Getenv("PROOF_BASE_URL"); u != "" {
		baseURL = u
	}
	auth := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
	client, err := business.NewClientWithResponses(baseURL, business.WithHTTPClient(auth))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	created, err := client.CreateTransactionWithResponse(ctx,
		&business.CreateTransactionParams{DocumentUrlVersion: ptr(business.CreateTransactionParamsDocumentUrlVersionV2)},
		business.TransactionCreateParams{
			Draft:           ptr(true),
			TransactionName: ptr("proof-sdk-go example"),
			Signer:          business.Signer{Email: *email, FirstName: ptr("Test"), LastName: ptr("Signer")},
			Documents:       &[]string{resource},
		})
	if err != nil {
		log.Fatal(err)
	}
	if err := common.CheckResponseBody(created.HTTPResponse, created.Body); err != nil {
		log.Fatal(err)
	}
	tx := created.JSON200

	if *activate {
		activated, err := client.ActivateDraftTransactionWithResponse(ctx, *tx.Id,
			&business.ActivateDraftTransactionParams{DocumentUrlVersion: ptr(business.ActivateDraftTransactionParamsDocumentUrlVersionV2)},
			business.ActivateDraftParams{})
		if err != nil {
			log.Fatal(err)
		}
		if err := common.CheckResponseBody(activated.HTTPResponse, activated.Body); err != nil {
			log.Fatal(err)
		}
		tx = activated.JSON200
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(tx); err != nil {
		log.Fatal(err)
	}
}
