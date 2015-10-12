package giniapi

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func Test_UploadOptionsTimeout(t *testing.T) {
	u := UploadOptions{}
	assertEqual(t, u.Timeout(), 30*time.Second, "")

	u.PollTimeout = 1 * time.Second
	assertEqual(t, u.Timeout(), 1*time.Second, "")
}

func Test_ConfigVerify(t *testing.T) {
	c := Config{}

	// Empty config fails
	assertNotEqual(t, c.Verify(), nil, "")

	// Minimal Oauth2 config
	c.ClientID = "client"
	c.ClientSecret = "secret"
	c.Authentication = UseOauth2

	// Fail without auth_code || username & password
	assertNotEqual(t, c.Verify(), nil, "")

	c.Username = "user1"
	assertNotEqual(t, c.Verify(), nil, "")

	c.AuthCode = "12345"
	assertEqual(t, c.Verify(), nil, "")

	c.Password = "secret"
	assertEqual(t, c.Verify(), nil, "")

	// Verify defaults
	c.Verify()

	assertEqual(t, c.APIVersion, "v1", "")
	assertEqual(t, c.Endpoints.API, "https://api.gini.net", "")
	assertEqual(t, c.Endpoints.UserCenter, "https://user.gini.net", "")
}

func Test_NewClient(t *testing.T) {
	// BasicAuth case
	config := Config{
		ClientID:       "c",
		ClientSecret:   "s",
		Authentication: UseBasicAuth,
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	client, err := NewClient(&config)

	assertEqual(t, err, nil, "")
	assertEqual(t, reflect.TypeOf(*client).Name(), "APIClient", "")
	assertEqual(t, reflect.TypeOf(client.HTTPClient.Transport).Name(), "BasicAuthTransport", "")

	// OAuth2
	config.Authentication = UseOauth2
	config.Username = "user1"
	config.Password = "secret"

	client, err = NewClient(&config)

	assertEqual(t, err, nil, "")
	assertEqual(t, reflect.TypeOf(*client).Name(), "APIClient", "")
}

func Test_DocumentUpload(t *testing.T) {
	config := Config{
		ClientID:       "c",
		ClientSecret:   "s",
		Authentication: UseBasicAuth,
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	client, err := NewClient(&config)
	document, err := client.Upload(bytes.NewReader([]byte("test")), UploadOptions{UserIdentifier: "user1"})

	assertEqual(t, err, nil, "")
	assertEqual(t, document.ID, "626626a0-749f-11e2-bfd6-000000000000", "")
}

func Test_DocumentGet(t *testing.T) {
	config := Config{
		ClientID:       "c",
		ClientSecret:   "s",
		Authentication: UseBasicAuth,
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	client, err := NewClient(&config)
	document, err := client.Get(fmt.Sprintf("%s/test/document/get", testHTTPServer.URL), "user1")

	assertEqual(t, err, nil, "")
	assertEqual(t, document.Owner, "user1", "")
	assertEqual(t, document.Progress, "COMPLETED", "")
}

func Test_DocumentList(t *testing.T) {
	config := Config{
		ClientID:       "c",
		ClientSecret:   "s",
		Authentication: UseBasicAuth,
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	client, err := NewClient(&config)
	documents, err := client.List(ListOptions{UserIdentifier: "user1"})

	assertEqual(t, err, nil, "")
	assertEqual(t, documents.TotalCount, 2, "")
	assertEqual(t, documents.Documents[1].String(), "626626a0-749f-11e2-abc2-000000000000", "")
}

func Test_DocumentSearch(t *testing.T) {
	config := Config{
		ClientID:       "c",
		ClientSecret:   "s",
		Authentication: UseBasicAuth,
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	client, err := NewClient(&config)
	documents, err := client.Search(SearchOptions{Query: "invoice", UserIdentifier: "user1"})

	assertEqual(t, err, nil, "")
	assertEqual(t, documents.TotalCount, 2, "")
	assertEqual(t, documents.Documents[1].String(), "626626a0-749f-11e2-abc2-000000000000", "")
}
