// Command business-list-transactions prints the most recent Business API
// transactions as JSON.
//
//	PROOF_API_KEY=... go run ./examples/business-list-transactions [-limit 10]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/tsarlewey/proof-sdk-go/business"
	"github.com/tsarlewey/proof-sdk-go/common"
)

func ptr[T any](v T) *T { return &v }

func main() {
	limit := flag.Int("limit", 10, "number of transactions to return (max 1000)")
	flag.Parse()

	baseURL := common.SandboxURL
	if u := os.Getenv("PROOF_BASE_URL"); u != "" {
		baseURL = u
	}
	auth := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
	client, err := business.NewClientWithResponses(baseURL, business.WithHTTPClient(auth))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.GetAllTransactionsWithResponse(context.Background(), &business.GetAllTransactionsParams{
		Limit:              limit,
		DocumentUrlVersion: ptr(business.GetAllTransactionsParamsDocumentUrlVersionV2),
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := common.CheckResponseBody(resp.HTTPResponse, resp.Body); err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp.JSON200); err != nil {
		log.Fatal(err)
	}
}
