package giniapi

import (
	"fmt"
	"net/http"
)

var (
	ErrConfigInvalid      = "failed to initialize config object"
	ErrMissingCredentials = "username or password cannot be empty in Oauth2 flow"

	ErrOauthAuthCodeExchange  = "failed to exchange oauth2 auth code"
	ErrOauthCredentials       = "failed to obtain token with username/password"
	ErrOauthParametersMissing = "oauth2 authentication requires AuthCode or Username + Password"

	ErrUploadFailed        = "failed to upoad document"
	ErrDocumentGet         = "failed to GET document object"
	ErrDocumentParse       = "failed to parse document json"
	ErrDocumentRead        = "failed to read document body"
	ErrDocumentList        = "failed to get document list"
	ErrDocumentSearch      = "failed to complete your search"
	ErrDocumentTimeout     = "failed to process document in time"
	ErrDocumentProcessing  = "failed to process document"
	ErrDocumentDelete      = "failed to delete document"
	ErrDocumentReport      = "failed to submit error report"
	ErrDocumentLayout      = "failed to retrieve layout"
	ErrDocumentExtractions = "failed to retrieve extractions"
	ErrDocumentProcessed   = "failed to retrieve processed document"
	ErrDocumentFeedback    = "failed to submit feedback"

	ErrHTTPPostFailed   = "failed to complete POST request"
	ErrHTTPGetFailed    = "failed to complete GET request"
	ErrHTTPDeleteFailed = "failed to complete GET request"
)

// APIError provides additional error informations
type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
	DocumentID string
	Parent     error
}

// Error satisifes the Error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s (HTTP status: %d, RequestID: %s, DocumentID: %s)",
		e.Message, e.StatusCode, e.RequestID, e.DocumentID)
}

// NewHttpError is a wrapper to simplify the error creation
func newHTTPError(message, docID string, err error, response *http.Response) *APIError {
	ae := APIError{
		Message:    message,
		DocumentID: docID,
	}

	// Sanity check for response pointer
	if response != nil {
		ae.StatusCode = response.StatusCode
		ae.RequestID = response.Header.Get("X-Request-Id")
	}

	if err != nil {
		ae.Parent = err
	}

	return &ae
}
