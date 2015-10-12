package giniapi

import (
	"golang.org/x/oauth2"
	"net/http"
)

// APIAuthScheme interface simplifies the addition of new auth mechanisms
type APIAuthScheme interface {
	Authenticate(config *Config) (*http.Client, error)
}

type Oauth2 struct{}
type BasicAuth struct{}

// Handy vars to simplify the initialization in a new API clients
var (
	UseOauth2    Oauth2
	UseBasicAuth BasicAuth
)

// Authenticate satisfies the APIAuthScheme interface for Oauth2
func (_ Oauth2) Authenticate(config *Config) (*http.Client, error) {
	conf := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.Endpoints.UserCenter + "/oauth/authorize",
			TokenURL: config.Endpoints.UserCenter + "/oauth/token",
		},
	}

	if config.AuthCode != "" {
		token, err := conf.Exchange(oauth2.NoContext, config.AuthCode)
		if err != nil {
			return nil, newHTTPError(ErrOauthAuthCodeExchange, "", err, nil)
		}
		client := conf.Client(oauth2.NoContext, token)
		return client, nil

	} else if config.Username != "" && config.Password != "" {
		token, err := conf.PasswordCredentialsToken(oauth2.NoContext, config.Username, config.Password)
		if err != nil {
			return nil, newHTTPError(ErrOauthCredentials, "", err, nil)
		}
		client := conf.Client(oauth2.NoContext, token)
		return client, nil
	}

	return nil, newHTTPError(ErrOauthParametersMissing, "", nil, nil)
}

// BasicAuthTransport is a net/http transport that automatically adds a matching authorization
// header for Gini's basic auth system.
type BasicAuthTransport struct {
	Transport http.RoundTripper
	Config    *Config
}

// RoundTrip to add basic auth header to all requests
func (bat BasicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(bat.Config.ClientID, bat.Config.ClientSecret)

	t := bat.Transport
	if t == nil {
		t = http.DefaultTransport
	}

	res, err := t.RoundTrip(r)
	return res, err
}

// Authenticate satisfies the APIAuthScheme interface for BasicAuth
func (_ BasicAuth) Authenticate(config *Config) (*http.Client, error) {
	client := &http.Client{Transport: BasicAuthTransport{Config: config}}
	return client, nil
}

// NewHTTPClient returns a custom http.Client for gini's oauth2 or basicAuth
// based authentication. Supports auth_code and password credentials oauth flows.
func newHTTPClient(config *Config) (*http.Client, error) {
	return config.Authentication.Authenticate(config)
}
