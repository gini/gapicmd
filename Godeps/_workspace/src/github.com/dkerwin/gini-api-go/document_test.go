package giniapi

import (
	"testing"
	"time"
)

func Test_TimingTotal(t *testing.T) {
	timing := Timing{
		Upload:     2,
		Processing: 5,
	}

	assertEqual(t, timing.Total(), time.Duration(7), "")
}

func Test_DocumentString(t *testing.T) {
	doc := Document{
		ID: "fb9877fc-f23c-40df-9e81-26e51f26682d",
	}

	assertEqual(t, doc.String(), "fb9877fc-f23c-40df-9e81-26e51f26682d", "Document.String() should return document ID")
}

func Test_DocumentUpdate(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Name:   "original",
		Links: Links{
			Document: testHTTPServer.URL + "/test/document/update",
		},
	}

	err := doc.Update()

	assertEqual(t, err, nil, "")
	assertEqual(t, doc.Name, "Updated!", "")
}

func Test_DocumentDelete(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Document: testHTTPServer.URL + "/test/document/delete",
		},
	}

	assertEqual(t, doc.Delete(), nil, "")
}

func Test_DocumentErrorReport(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		ID:     "12345",
		Links: Links{
			Document: testHTTPServer.URL + "/test/document",
		},
	}

	assertEqual(t, doc.ErrorReport("", ""), nil, "")
}

func Test_DocumentGetLayout(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Layout: testHTTPServer.URL + "/test/layout",
		},
	}

	_, err := doc.GetLayout()
	assertEqual(t, err, nil, "")
}

func Test_DocumentGetExtractions(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Extractions: testHTTPServer.URL + "/test/extractions",
		},
	}

	_, err := doc.GetExtractions()
	assertEqual(t, err, nil, "")
}

func Test_DocumentGetProcessed(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Processed: testHTTPServer.URL + "/test/processed",
		},
	}

	docBytes, err := doc.GetProcessed()
	assertEqual(t, err, nil, "")
	assertEqual(t, string(docBytes), "get processed", "")
}

func Test_DocumentSubmitFeedback(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Extractions: testHTTPServer.URL + "/test/feedback",
		},
	}

	feedback := map[string]Extraction{
		"iban": Extraction{
			Entity: "iban",
			Value:  "DE22222111117777766666",
		},
	}

	// single label
	assertEqual(t, doc.SubmitFeedback(feedback), nil, "")

	feedback["bic"] = Extraction{
		Entity: "bic",
		Value:  "HYVEDEMMXXX",
	}

	// multiple labels
	assertEqual(t, doc.SubmitFeedback(feedback), nil, "")
}
