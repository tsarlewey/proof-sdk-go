// Command scim-list-users prints an organization's SCIM users as JSON.
//
//	PROOF_API_KEY=... PROOF_ORGANIZATION_ID=orXXXXXXX go run ./examples/scim-list-users \
//	    [-filter 'userName eq "someone@example.com"']
//
// SCIM results are untyped in the generated client (no JSON200), so this
// decodes resp.Body itself.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/tsarlewey/proof-sdk-go/common"
	"github.com/tsarlewey/proof-sdk-go/scim"
)

func main() {
	filter := flag.String("filter", "", `SCIM filter, e.g. userName eq "someone@example.com"`)
	flag.Parse()

	orgID := os.Getenv("PROOF_ORGANIZATION_ID")
	if orgID == "" {
		log.Fatal("PROOF_ORGANIZATION_ID is required")
	}
	baseURL := common.SandboxURL
	if u := os.Getenv("PROOF_BASE_URL"); u != "" {
		baseURL = u
	}
	auth := common.NewAuthenticatedDoer(common.APIKey(os.Getenv("PROOF_API_KEY")))
	client, err := scim.NewClientWithResponses(baseURL, scim.WithHTTPClient(auth))
	if err != nil {
		log.Fatal(err)
	}

	params := &scim.ListUsersParams{}
	if *filter != "" {
		params.Filter = filter
	}
	resp, err := client.ListUsersWithResponse(context.Background(), orgID, params)
	if err != nil {
		log.Fatal(err)
	}
	if err := common.CheckResponseBody(resp.HTTPResponse, resp.Body); err != nil {
		log.Fatal(err)
	}

	// A SCIM ListResponse: {"schemas": [...], "totalResults": n, "Resources": [...]}.
	var users struct {
		TotalResults int               `json:"totalResults"`
		Resources    []json.RawMessage `json:"Resources"`
	}
	if err := json.Unmarshal(resp.Body, &users); err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(users); err != nil {
		log.Fatal(err)
	}
}
