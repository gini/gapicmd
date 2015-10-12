package giniapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Timing struct
type Timing struct {
	Upload     time.Duration
	Processing time.Duration
}

// Total returns the summarized timings of upload and processing
func (t *Timing) Total() time.Duration {
	return t.Upload + t.Processing
}

// Page describes a documents pages
type Page struct {
	Images     map[string]string `json:"images"`
	PageNumber int               `json:"pageNumber"`
}

// Links contains the links to a documents resources
type Links struct {
	Document    string `json:"document"`
	Extractions string `json:"extractions"`
	Layout      string `json:"layout"`
	Processed   string `json:"processed"`
}

// Document contains all informations about a single document
type Document struct {
	Timing
	client               *APIClient // client is not exported
	Owner                string
	Links                Links  `json:"_links"`
	CreationDate         int    `json:"creationDate"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Origin               string `json:"origin"`
	PageCount            int    `json:"pageCount"`
	Pages                []Page `json:"pages"`
	Progress             string `json:"progress"`
	SourceClassification string `json:"sourceClassification"`
}

// DocumentSet is a list of documents with the total count
type DocumentSet struct {
	TotalCount int         `json:"totalCount"`
	Documents  []*Document `json:"documents"`
}

// String representaion of a document
func (d *Document) String() string {
	return fmt.Sprintf(d.ID)
}

// Poll the progress state of a document and return nil when the processing
// has completed (successful or failed). On timeout return error
func (d *Document) Poll(timeout time.Duration) error {
	start := time.Now()
	defer func() { d.Timing.Processing = time.Since(start) }()

	docProgress := make(chan *Document, 1)
	quit := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				doc, err := d.client.Get(d.Links.Document, d.Owner)
				if err != nil {
					return
				}
				if doc.Progress == "COMPLETED" || doc.Progress == "ERROR" {
					docProgress <- doc
					return
				}
			}
		}
	}()

	select {
	case doc := <-docProgress:
		if doc == nil {
			return newHTTPError(ErrDocumentProcessing, "", nil, nil)
		}
		*d = *doc
		return nil
	case <-time.After(timeout):
		quit <- true
		return newHTTPError(fmt.Sprintf("%s after %f", ErrDocumentTimeout, timeout.Seconds()), "", nil, nil)
	}
}

// Update document struct from self-contained document link
func (d *Document) Update() error {
	newDoc, err := d.client.Get(d.Links.Document, d.Owner)
	if err != nil {
		return err
	}
	*d = *newDoc
	return nil
}

// Delete a document
func (d *Document) Delete() error {
	resp, err := d.client.makeAPIRequest("DELETE", d.Links.Document, nil, nil, d.Owner)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return newHTTPError(ErrDocumentDelete, d.ID, err, resp)
	}

	return nil
}

// ErrorReport creates a bug report in Gini's bugtracking system. It's a convinience way
// to help Gini learn from difficult documents
func (d *Document) ErrorReport(summary string, description string) error {
	params := map[string]interface{}{
		"summary":     summary,
		"description": description,
	}

	u := encodeURLParams(fmt.Sprintf("%s/errorreport", d.Links.Document), params)

	resp, err := d.client.makeAPIRequest("POST", u, nil, nil, d.Owner)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return newHTTPError(ErrDocumentReport, d.ID, err, resp)
	}

	return nil
}

// GetLayout returns the JSON representation of a documents layout parsed as
// Layout struct
func (d *Document) GetLayout() (*Layout, error) {
	var layout Layout

	resp, err := d.client.makeAPIRequest("GET", d.Links.Layout, nil, nil, "")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newHTTPError(ErrDocumentLayout, d.ID, err, resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&layout); err != nil {
		return nil, err
	}

	return &layout, nil
}

// GetExtractions returns a documents extractions in a Extractions struct
func (d *Document) GetExtractions() (*Extractions, error) {
	var extractions Extractions

	resp, err := d.client.makeAPIRequest("GET", d.Links.Extractions, nil, nil, d.Owner)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newHTTPError(ErrDocumentExtractions, d.ID, err, resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(&extractions); err != nil {
		return nil, err
	}

	return &extractions, nil
}

// GetProcessed returns a byte array of the processed (rectified, optimized) document
func (d *Document) GetProcessed() ([]byte, error) {
	headers := map[string]string{
		"Accept": "application/octet-stream",
	}

	resp, err := d.client.makeAPIRequest("GET", d.Links.Processed, nil, headers, d.Owner)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newHTTPError(ErrDocumentProcessed, d.ID, err, resp)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)

	if err != nil {
		return nil, newHTTPError(ErrDocumentProcessed, d.ID, err, resp)
	}

	return buf.Bytes(), nil
}

// SubmitFeedback submits feedback from map
func (d *Document) SubmitFeedback(feedback map[string]Extraction) error {
	feedbackMap := map[string]map[string]Extraction{
		"feedback": map[string]Extraction{},
	}

	for key, extraction := range feedback {
		feedbackMap["feedback"][key] = extraction
	}

	feedbackBody, err := json.Marshal(feedbackMap)
	if err != nil {
		return err
	}

	resp, err := d.client.makeAPIRequest("PUT", d.Links.Extractions, bytes.NewReader(feedbackBody), nil, d.Owner)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return newHTTPError(ErrDocumentFeedback, d.ID, err, resp)
	}

	return nil
}
