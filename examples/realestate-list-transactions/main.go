// Command realestate-list-transactions prints the most recent Real Estate
// (mortgage) API transactions as JSON, optionally filtered by loan number.
//
//	PROOF_API_KEY=... go run ./examples/realestate-list-transactions [-limit 10] [-loan 12345]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/tsarlewey/proof-sdk-go/common"
	"github.com/tsarlewey/proof-sdk-go/realestate"
)

func ptr[T any](v T) *T { return &v }

func main() {
	limit := flag.Int("limit", 10, "number of transactions to return (max 1000)")
	loan := flag.String("loan", "", "only transactions with this loan number")
	flag.Parse()

	baseURL := common.SandboxURL
	if u := os.Getenv("PROOF_BASE_URL"); u != "" {
		baseURL = u
	}
	auth := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
	client, err := realestate.NewClientWithResponses(baseURL, realestate.WithHTTPClient(auth))
	if err != nil {
		log.Fatal(err)
	}

	params := &realestate.GetAllMortgageTransactionsParams{
		Limit:              limit,
		DocumentUrlVersion: ptr(realestate.GetAllMortgageTransactionsParamsDocumentUrlVersionV2),
	}
	if *loan != "" {
		params.LoanNumber = loan
	}
	resp, err := client.GetAllMortgageTransactionsWithResponse(context.Background(), params)
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
