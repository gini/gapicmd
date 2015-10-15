package giniapi

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strconv"
)

// MakeAPIRequest is a wrapper around http.NewRequest to create http
// request and inject required headers.
func (api *APIClient) makeAPIRequest(verb, url string, body io.Reader, headers map[string]string, userIdentifier string) (*http.Response, error) {
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %s", err)
	}

	if _, ok := headers["Accept"]; !ok {
		req.Header.Add("Accept", fmt.Sprintf("application/vnd.gini.%s+json", api.Config.APIVersion))
	}

	req.Header.Add("User-Agent", fmt.Sprintf("gini-api-go/%s", VERSION))

	if reflect.TypeOf(api.Config.Authentication).Name() == "BasicAuth" {
		if userIdentifier == "" {
			return nil, fmt.Errorf("userIdentifier required (Authentication=BasicAuth)")
		}
		req.Header.Add("X-User-Identifier", userIdentifier)
	}

	// Append additional headers
	for h, v := range headers {
		req.Header.Add(h, v)
	}

	resp, err := api.HTTPClient.Do(req)

	// Debug HTTP calls?
	if api.Config.HTTPDebug {
		debug, err := httputil.DumpRequest(resp.Request, false)
		if err != nil {
			api.Config.RequestDebug <- []byte(fmt.Sprintf("Failed to dump request: %s", err))
		} else {
			api.Config.RequestDebug <- debug
		}

		debug, err = httputil.DumpResponse(resp, true)
		if err != nil {
			api.Config.ResponseDebug <- []byte(fmt.Sprintf("Failed to dump response: %s", err))
		} else {
			api.Config.ResponseDebug <- debug
		}
	}

	return resp, err
}

func encodeURLParams(baseURL string, queryParams map[string]interface{}) string {
	u, _ := url.Parse(baseURL)

	params := url.Values{}

	for key, value := range queryParams {
		switch value := value.(type) {
		case string:
			params.Add(key, value)
		case int:
			params.Add(key, strconv.Itoa(value))
		}
	}

	u.RawQuery = params.Encode()
	return u.String()
}
