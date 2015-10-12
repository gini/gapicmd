package giniapi

import (
	"testing"
)

func Test_ExtractionsGetValue(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Extractions: testHTTPServer.URL + "/test/extractions",
		},
	}

	extractions, _ := doc.GetExtractions()
	assertEqual(t, extractions.GetValue("amountToPay"), "24.99:EUR", "")
	assertEqual(t, extractions.GetValue("unknown"), "", "")
}
