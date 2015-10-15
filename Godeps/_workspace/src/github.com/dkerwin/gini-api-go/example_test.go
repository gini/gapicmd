package giniapi_test

import (
	"fmt"
	"github.com/dkerwin/gini-api-go"
	"log"
	"os"
	"time"
)

// Very simplistic example. You shoud have a lot more error handling in place
func ExampleNewClient() {

	//////////////////////////////////
	// Oauth2
	//////////////////////////////////

	// Setup api connection
	api, err := giniapi.NewClient(&giniapi.Config{
		ClientID:       "MY_CLIENT_ID",
		ClientSecret:   "********",
		Username:       "user1",
		Password:       "secret",
		Authentication: giniapi.UseOauth2,
	})

	if err != nil {
		log.Panicf("Gini API login failed: %s", err)
	}

	// Read a PDF document
	document, _ := os.Open("/tmp/invoice.pdf")

	// Upload document to gini without doctype hint and user identifier
	doc, _ := api.Upload(document, giniapi.UploadOptions{FileName: "invoice.pdf", PollTimeout: 10 * time.Second})

	// Get extractions from our uploaded document
	extractions, _ := doc.GetExtractions(false)

	// Print IBAN
	fmt.Printf("IBAN has been found: %s Woohoo!\n", extractions.GetValue("iban"))

	//////////////////////////////////
	// basic Auth
	//////////////////////////////////

	// Setup api connection
	api, err = giniapi.NewClient(&giniapi.Config{
		ClientID:       "MY_CLIENT_ID",
		ClientSecret:   "********",
		Authentication: giniapi.UseBasicAuth,
	})

	if err != nil {
		log.Panicf("Gini API login failed: %s", err)
	}

	// Read a PDF document
	document, _ = os.Open("/tmp/invoice.pdf")

	// Upload document to gini without doctype hint and user identifier
	doc, _ = api.Upload(document, giniapi.UploadOptions{FileName: "invoice.pdf", UserIdentifier: "user123", PollTimeout: 10 * time.Second})

	// Get extractions from our uploaded document
	extractions, _ = doc.GetExtractions(false)

	// Print IBAN
	fmt.Printf("IBAN has been found: %s Woohoo!\n", extractions.GetValue("iban"))
}
