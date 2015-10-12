package giniapi

import (
	"fmt"
	"io/ioutil"
	"testing"
)

// Test helpers
func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func assertNotEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a != b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func testOauthClient(t *testing.T) *APIClient {
	config := Config{
		ClientID:       "testclient",
		ClientSecret:   "secret",
		Authentication: UseOauth2,
		AuthCode:       "123456",
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	client, err := NewClient(&config)
	if err != nil {
		t.Fatal("Cannot init the api client")
	}
	return client
}

func testBasicAuthClient(t *testing.T) *APIClient {
	config := Config{
		ClientID:       "testclient",
		ClientSecret:   "secret",
		Authentication: UseBasicAuth,
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	client, err := NewClient(&config)
	if err != nil {
		t.Fatal("Cannot init the api client")
	}
	return client
}

// Real tests from here
func Test_makeAPIRequest(t *testing.T) {
	// Basic config
	config := Config{
		ClientID:     "testclient",
		ClientSecret: "secret",
		Endpoints: Endpoints{
			API:        testHTTPServer.URL,
			UserCenter: testHTTPServer.URL,
		},
	}

	// basicAuth
	config.Authentication = UseBasicAuth
	api, err := NewClient(&config)
	if err != nil {
		t.Errorf("Failed to setup NewClient: %s", err)
	}

	// Fail without userIdentifier
	if response, err := api.makeAPIRequest("GET", testHTTPServer.URL+"/test/http/basicAuth", nil, nil, ""); response != nil || err == nil {
		t.Errorf("Missing userIdentifier should raise err")
	}

	// Succeed with userIdentifier
	response, err := api.makeAPIRequest("GET", testHTTPServer.URL+"/test/http/basicAuth", nil, nil, "user123")
	if response == nil || err != nil {
		t.Errorf("HTTP call with supplied userIdentifier failed: %s", err)
	}

	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || string(body) != "test completed" {
		t.Errorf("Body (%s) or statusCode(%d) mismatch", string(body), response.StatusCode)
	}

	// oauth2
	config.Authentication = UseOauth2
	config.AuthCode = "123456"

	api, err = NewClient(&config)
	if err != nil {
		t.Errorf("Failed to setup NewClient: %s", err)
	}

	// Make oauth2 call
	if response, err := api.makeAPIRequest("GET", testHTTPServer.URL+"/test/http/oauth2", nil, nil, ""); response == nil || err != nil {
		t.Errorf("Call failed: %#v", err)
	}

	// Pass additional headers
	headers := map[string]string{
		"X-Dummy-Header": "Ignored",
	}
	if response, err := api.makeAPIRequest("GET", testHTTPServer.URL+"/test/http/oauth2", nil, headers, ""); response == nil || err != nil {
		t.Errorf("Call failed: %#v", err)
	}
}

func Test_encodeURLParams(t *testing.T) {
	params := map[string]interface{}{
		"aInt":             9,
		"aStrWithSpaces":   "Just a string",
		"aStrWithEncoding": "test20%25gn%3B-%2F",
	}

	u := encodeURLParams("https://www.example.com", params)

	assertEqual(t, u, "https://www.example.com?aInt=9&aStrWithEncoding=test20%2525gn%253B-%252F&aStrWithSpaces=Just+a+string", "")
}
