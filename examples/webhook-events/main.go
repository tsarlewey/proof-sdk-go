// Command webhook-events lists the organization's Business API v2 webhooks,
// or, given a webhook ID, that webhook's recent delivery events, as JSON.
//
//	PROOF_API_KEY=... go run ./examples/webhook-events            # list webhooks
//	PROOF_API_KEY=... go run ./examples/webhook-events <webhook-id> # list its events
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

func main() {
	limit := flag.Int("limit", 20, "number of events to return (max 100)")
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
	ctx := context.Background()

	var out any
	if id := flag.Arg(0); id != "" {
		resp, err := client.GetWebhookEventsV2WithResponse(ctx, id, &business.GetWebhookEventsV2Params{Limit: limit})
		if err != nil {
			log.Fatal(err)
		}
		if err := common.CheckResponseBody(resp.HTTPResponse, resp.Body); err != nil {
			log.Fatal(err)
		}
		out = resp.JSON200
	} else {
		resp, err := client.GetAllWebhooksV2WithResponse(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if err := common.CheckResponseBody(resp.HTTPResponse, resp.Body); err != nil {
			log.Fatal(err)
		}
		out = resp.JSON200
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		log.Fatal(err)
	}
}
