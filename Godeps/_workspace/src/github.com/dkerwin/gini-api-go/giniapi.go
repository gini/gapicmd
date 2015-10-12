// Copyright 2015 The giniapi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package giniapit interacts with Gini's API service to make sense of unstructured
documents. Please visit http://developer.gini.net/gini-api/html/index.html
for more details about the Gini API.

API features

Suppoted API calls include:

	- Upload documents (native, scanned, text)
	- List a users documents
	- Search documents
	- Get extractions (incubator is supported)
	- Download rendered pages, processed document and layout XML
	- Submit feedback on extractions
	- Submit error reports

Contributing

It's awesome that you consider contributing to gini-api-go. Here's how it's done:

	- Fork repository on Github
	- Create a topic/feature branch
	- Write code AND tests
	- Update documentation if necessary
	- Open a pull request

*/
package giniapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

const (
	// VERSION is the API client version
	VERSION string = "0.1.0"
)

// Config to setup Gini API connection
type Config struct {
	// ClientID is the application's ID.
	ClientID string
	// ClientSecret is the application's secret.
	ClientSecret string
	// Username for oauth2 password grant
	Username string
	// Password for oauth2 pssword grant
	Password string
	// Auth_code to exchange for oauth2 token
	AuthCode string
	// Scopes to use (leave empty for all assigned scopes)
	Scopes []string
	// API & Usercenter endpoints
	Endpoints
	// APIVersion to use (v1)
	APIVersion string `default:"v1"`
	// Authentication to use
	// oauth2: auth_code || password credentials
	// basicAuth: basic auth + user identifier
	Authentication APIAuthScheme
	// Debug
	HTTPDebug bool

	RequestDebug  chan []byte
	ResponseDebug chan []byte
}

func (c *Config) Verify() error {
	if c.ClientID == "" || c.ClientSecret == "" {
		return newHTTPError(ErrConfigInvalid, "", nil, nil)
	}

	if reflect.TypeOf(c.Authentication).Name() == "Oauth2" {
		if c.AuthCode == "" && (c.Username == "" || c.Password == "") {
			return newHTTPError(ErrMissingCredentials, "", nil, nil)
		}
	}

	cType := reflect.TypeOf(*c)

	// Fix potential missing APIVersion with default
	if c.APIVersion == "" {
		f, _ := cType.FieldByName("APIVersion")
		c.APIVersion = f.Tag.Get("default")
	}

	// Fix potential missing Endpoints with defaults
	cType = reflect.TypeOf(c.Endpoints)

	if c.Endpoints.API == "" {
		f, _ := cType.FieldByName("API")
		c.Endpoints.API = f.Tag.Get("default")
	}
	if c.Endpoints.UserCenter == "" {
		f, _ := cType.FieldByName("UserCenter")
		c.Endpoints.UserCenter = f.Tag.Get("default")
	}

	return nil
}

// Endpoints to access API and Usercenter
type Endpoints struct {
	API        string `default:"https://api.gini.net"`
	UserCenter string `default:"https://user.gini.net"`
}

// UploadOptions specify parameters to the Upload function
type UploadOptions struct {
	PollTimeout    time.Duration
	FileName       string
	DocType        string
	UserIdentifier string
}

// Timeout returns a default timeout of 30 when PollTimeout is uninitialized
func (o *UploadOptions) Timeout() time.Duration {
	if o.PollTimeout == 0 {
		return 30 * time.Second
	}
	return o.PollTimeout
}

// ListOptions specify parameters to the List function
type ListOptions struct {
	Limit          int
	Offset         int
	UserIdentifier string
}

// SearchOptions specify parameters to the List function
type SearchOptions struct {
	Query          string
	Type           string
	UserIdentifier string
	Limit          int
	Offset         int
}

// APIClient is the main interface for the user
type APIClient struct {
	// Config
	Config

	// Http client
	HTTPClient *http.Client
}

// NewClient validates your Config parameters and returns a APIClient object
// with a matching http client included.
func NewClient(config *Config) (*APIClient, error) {
	if err := config.Verify(); err != nil {
		return nil, err
	}

	// Get http client based on the selected Authentication
	client, err := newHTTPClient(config)
	if err != nil {
		return nil, err
	}

	return &APIClient{
		Config:     *config,
		HTTPClient: client,
	}, nil

}

// Upload a document from a given io.Reader objct (document). Additional options can be
// passed with a instance of UploadOptions. FileName and DocType are optional and can be empty.
// UserIdentifier is required if Authentication method is "basic_auth".
// Upload time is measured and stored in Timing struct (part of Document).
func (api *APIClient) Upload(document io.Reader, options UploadOptions) (*Document, error) {
	start := time.Now()
	resp, err := api.makeAPIRequest("POST", fmt.Sprintf("%s/documents", api.Config.Endpoints.API), document, nil, options.UserIdentifier)
	if err != nil {
		return nil, newHTTPError(ErrHTTPPostFailed, "", err, resp)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, newHTTPError(ErrUploadFailed, "", err, resp)
	}
	uploadDuration := time.Since(start)

	doc, err := api.Get(resp.Header.Get("Location"), options.UserIdentifier)
	if err != nil {
		return nil, err
	}
	doc.Timing.Upload = uploadDuration

	// Poll for completion or failure with timeout
	err = doc.Poll(options.Timeout())

	return doc, err
}

// Get Document struct from URL
func (api *APIClient) Get(url, userIdentifier string) (*Document, error) {
	resp, err := api.makeAPIRequest("GET", url, nil, nil, userIdentifier)
	if err != nil {
		return nil, newHTTPError(ErrHTTPGetFailed, "", err, resp)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, newHTTPError(ErrDocumentGet, "", err, resp)
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newHTTPError(ErrDocumentRead, "", err, nil)
	}

	var doc Document
	if err := json.Unmarshal(contents, &doc); err != nil {
		return nil, newHTTPError(ErrDocumentParse, doc.ID, err, nil)
	}

	// Add client and owner to doc object
	doc.client = api
	doc.Owner = userIdentifier

	return &doc, nil
}

// List returns DocumentSet
func (api *APIClient) List(options ListOptions) (*DocumentSet, error) {
	params := map[string]interface{}{
		"limit":  options.Limit,
		"offset": options.Offset,
	}

	u := encodeURLParams(fmt.Sprintf("%s/documents", api.Config.Endpoints.API), params)

	resp, err := api.makeAPIRequest("GET", u, nil, nil, options.UserIdentifier)
	if err != nil {
		return nil, newHTTPError(ErrHTTPGetFailed, "", err, resp)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newHTTPError(ErrDocumentList, "", err, resp)
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newHTTPError(ErrDocumentRead, "", err, nil)
	}

	var docs DocumentSet
	if err := json.Unmarshal(contents, &docs); err != nil {
		return nil, newHTTPError(ErrDocumentParse, "", err, nil)
	}

	// Extra round: Ingesting *APIClient into each and every doc
	for _, d := range docs.Documents {
		d.client = api
		d.Owner = options.UserIdentifier
	}

	return &docs, nil
}

// Search returns DocumentSet
func (api *APIClient) Search(options SearchOptions) (*DocumentSet, error) {
	params := map[string]interface{}{
		"q":     options.Query,
		"type":  options.Type,
		"limit": options.Limit,
		"next":  options.Offset,
	}

	u := encodeURLParams(fmt.Sprintf("%s/search", api.Config.Endpoints.API), params)

	resp, err := api.makeAPIRequest("GET", u, nil, nil, options.UserIdentifier)
	if err != nil {
		return nil, newHTTPError(ErrHTTPGetFailed, "", err, resp)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newHTTPError(ErrDocumentSearch, "", err, resp)
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newHTTPError(ErrDocumentRead, "", err, nil)
	}

	var docs DocumentSet
	if err = json.Unmarshal(contents, &docs); err != nil {
		return nil, newHTTPError(ErrDocumentParse, "", err, nil)
	}

	// Extra round: Ingesting *APIClient into each and every doc
	for _, d := range docs.Documents {
		d.client = api
		d.Owner = options.UserIdentifier
	}

	return &docs, nil
}
