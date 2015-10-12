package giniapi

import (
	"fmt"
	"testing"
)

func Test_Error(t *testing.T) {
	err := APIError{
		StatusCode: 500,
		Message:    "Error test",
		RequestID:  "12345",
		DocumentID: "67890",
	}

	assertEqual(t, err.Error(), "Error test (HTTP status: 500, RequestID: 12345, DocumentID: 67890)", "")
}

func Test_newHTTPError(t *testing.T) {
	err := fmt.Errorf("Something went wrong")
	e := newHTTPError("Error test", "12345", err, nil)

	assertEqual(t, e.Message, "Error test", "")
	assertEqual(t, e.DocumentID, "12345", "")
	assertEqual(t, e.Parent.Error(), "Something went wrong", "")
}
