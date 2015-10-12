package giniapi

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	// "io/ioutil"
	// "log"
	"net/http"
	"net/http/httptest"
	// "strconv"
	// "time"
)

var (
	testHTTPServer *httptest.Server
)

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/ping", handlerGetPing).Methods("GET")
	r.HandleFunc("/oauth/token", handlerPostToken).Methods("POST")
	r.HandleFunc("/documents", handlerTestDocumentList).Methods("GET")
	r.HandleFunc("/documents", handlerTestDocumentUpload).Methods("POST")
	r.HandleFunc("/search", handlerTestDocumentSearch).Methods("GET")
	r.HandleFunc("/test/http/basicAuth", handlerTestHTTPBasicAuth).Methods("GET")
	r.HandleFunc("/test/http/oauth2", handlerTestHTTPOauth2).Methods("GET")
	r.HandleFunc("/test/document/get", handlerTestDocumentGet).Methods("GET")
	r.HandleFunc("/test/document/update", handlerTestDocumentUpdate).Methods("GET")
	r.HandleFunc("/test/document/delete", handlerTestDocumentDelete).Methods("DELETE")
	r.HandleFunc("/test/document/errorreport", handlerTestDocumentErrorReport).Methods("POST")
	r.HandleFunc("/test/layout", handlerTestDocumentLayout).Methods("GET")
	r.HandleFunc("/test/extractions", handlerTestDocumentExtractions).Methods("GET")
	r.HandleFunc("/test/processed", handlerTestDocumentProcessed).Methods("GET")
	r.HandleFunc("/test/feedback", handlerTestDocumentFeedback).Methods("PUT")

	testHTTPServer = httptest.NewServer(handlerAccessLog(r))
}

func handlerAccessLog(handler http.Handler) http.Handler {
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		// body, _ := ioutil.ReadAll(r.Body)
		// log.Printf("%s \"%s %s\" - %v => %v\n\n", r.RemoteAddr, r.Method, r.URL, r.Header, string(body))
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(logHandler)
}

func handlerGetPing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func handlerPostToken(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := `{
                "access_token":"760822cb-2dec-4275-8da8-fa8f5680e8d4",
                "token_type":"bearer",
                "expires_in":300,
                "refresh_token":"46463dd6-cdbb-440d-88fc-b10a34f68b26"
             }`

	w.Write([]byte(body))
}

func handlerTestHTTPBasicAuth(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/vnd.gini.v1+json" {
		writeHeaders(w, 500, "changes")
	} else {
		writeHeaders(w, 200, "changes")
	}
	body := "test completed"
	w.Write([]byte(body))
}

func handlerTestHTTPOauth2(w http.ResponseWriter, r *http.Request) {
	body := "test completed"
	if r.Header.Get("Authorization") != "Bearer 760822cb-2dec-4275-8da8-fa8f5680e8d4" {
		writeHeaders(w, 401, "invalid token")
		body = "Invalid Authorization header"
	} else {
		writeHeaders(w, 200, "changes")
	}
	w.Write([]byte(body))
}

func writeHeaders(w http.ResponseWriter, code int, jobName string) {
	h := w.Header()
	h.Add("Content-Type", "application/json")
	if jobName != "" {
		h.Add("Job-Name", jobName)
	}
	w.WriteHeader(code)
}

func handlerTestDocumentUpdate(w http.ResponseWriter, r *http.Request) {
	body := `{ "name": "Updated!" }`
	writeHeaders(w, 200, "changes")
	w.Write([]byte(body))
}

func handlerTestDocumentDelete(w http.ResponseWriter, r *http.Request) {
	body := "test completed"
	writeHeaders(w, 204, "changes")
	w.Write([]byte(body))
}

func handlerTestDocumentErrorReport(w http.ResponseWriter, r *http.Request) {
	body := "test completed"
	writeHeaders(w, 200, "changes")
	w.Write([]byte(body))
}

func handlerTestDocumentLayout(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := `{
	  "pages": [
	    {
	      "number": 1,
	      "sizeX": 595.3,
	      "sizeY": 841.9,
	      "textZones": [
	        {
	          "paragraphs": [
	            {
	              "l": 54.0,
	              "t": 158.76,
	              "w": 190.1,
	              "h": 36.55000000000001,
	              "lines": [
	                {
	                  "l": 54.0,
	                  "t": 158.76,
	                  "w": 190.1,
	                  "h": 10.810000000000002,
	                  "wds": [
	                    {
	                      "l": 54.0,
	                      "t": 158.76,
	                      "w": 18.129999999999995,
	                      "h": 9.900000000000006,
	                      "fontSize": 9.9,
	                      "fontFamily": "Arial-BoldMT",
	                      "bold":false,
	                      "text": "Ihre"
	                    },
	                    {
	                      "l": 74.86,
	                      "t": 158.76,
	                      "w": 83.91000000000001,
	                      "h": 9.900000000000006,
	                      "fontSize": 9.9,
	                      "fontFamily": "Arial-BoldMT",
	                      "bold":false,
	                      "text": "Vorgangsnummer"
	                    },
	                    {
	                      "l": 158.76,
	                      "t": 158.76,
	                      "w": 3.3000000000000114,
	                      "h": 9.900000000000006,
	                      "fontSize": 9.9,
	                      "fontFamily": "Arial-BoldMT",
	                      "bold":false,
	                      "text": ":"
	                    }
	                  ]
	                }
	              ]
	            }
	          ]
	        }
	      ],
	      "regions": [
	        {
	          "l": 20.0,
	          "t": 240.1,
	          "w": 190.0,
	          "h": 150.3,
	          "type": "RemittanceSlip"
	        }
	      ]
	    }
	  ]
	}`

	w.Write([]byte(body))
}

func handlerTestDocumentExtractions(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := `{
	    "extractions": {
	        "amountToPay": {
	            "box": {
	                "height": 9.0,
	                "left": 516.0,
	                "page": 1,
	                "top": 588.0,
	                "width": 42.0
	            },
	            "entity": "amount",
	            "value": "24.99:EUR",
	            "candidates": "amounts"
	        }
	      },
	      "candidates": {
	        "amounts": [
	          {
	              "box": {
	                  "height": 9.0,
	                  "left": 516.0,
	                  "page": 1,
	                  "top": 588.0,
	                  "width": 42.0
	              },
	              "entity": "amount",
	              "value": "24.99:EUR"
	          },
	          {
	              "box": {
	                  "height": 9.0,
	                  "left": 241.0,
	                  "page": 1,
	                  "top": 588.0,
	                  "width": 42.0
	              },
	              "entity": "amount",
	              "value": "21.0:EUR"
	          }
	        ]
	    }
	}`

	w.Write([]byte(body))
}

func handlerTestDocumentProcessed(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	w.Write([]byte("get processed"))
}

func handlerTestDocumentFeedback(w http.ResponseWriter, r *http.Request) {
	var feedbackMap map[string]map[string]Extraction

	if err := json.NewDecoder(r.Body).Decode(&feedbackMap); err != nil {
		writeHeaders(w, 500, "failed")
		return
	}

	writeHeaders(w, 204, "ok")
}

func handlerTestDocumentUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Location", fmt.Sprintf("%s/test/document/get", testHTTPServer.URL))
	writeHeaders(w, 201, "ok")

	return
}

func handlerTestDocumentGet(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := fmt.Sprintf(`{
		  "id": "626626a0-749f-11e2-bfd6-000000000000",
		  "creationDate": 1360623867402,
		  "name": "scanned.jpg",
		  "progress": "COMPLETED",
		  "origin": "UPLOAD",
		  "sourceClassification": "SCANNED",
		  "pageCount": 1,
		  "pages" : [
		    {
		      "images" : {
		        "750x900" : "http://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/pages/1/750x900",
		        "1280x1810" : "http://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/pages/1/1280x1810"
		      },
		      "pageNumber" : 1
		    }
		  ],
		  "_links": {
		    "extractions": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/extractions",
		    "layout": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/layout",
		    "document": "%s/test/document/get",
		    "processed": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/processed"
		  }
		}`, testHTTPServer.URL)

	w.Write([]byte(body))
}

func handlerTestDocumentList(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := `{
		"totalCount": 2,
		"documents": [
			{
			  "id": "626626a0-749f-11e2-bfd6-000000000000",
			  "creationDate": 1360623867402,
			  "name": "scanned.jpg",
			  "progress": "COMPLETED",
			  "origin": "UPLOAD",
			  "sourceClassification": "SCANNED",
			  "pageCount": 1,
			  "pages" : [
			    {
			      "images" : {
			        "750x900" : "http://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/pages/1/750x900",
			        "1280x1810" : "http://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/pages/1/1280x1810"
			      },
			      "pageNumber" : 1
			    }
			  ],
			  "_links": {
			    "extractions": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/extractions",
			    "layout": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/layout",
			    "document": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000",
			    "processed": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/processed"
			  }
			},
			{
			  "id": "626626a0-749f-11e2-abc2-000000000000",
			  "creationDate": 1360624287987,
			  "name": "native.pdf",
			  "progress": "COMPLETED",
			  "origin": "UPLOAD",
			  "sourceClassification": "NATIVE",
			  "pageCount": 1,
			  "pages" : [
			    {
			      "images" : {
			        "750x900" : "http://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/pages/1/750x900",
			        "1280x1810" : "http://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/pages/1/1280x1810"
			      },
			      "pageNumber" : 1
			    }
			  ],
			  "_links": {
			    "extractions": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/extractions",
			    "layout": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/layout",
			    "document": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000",
			    "processed": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/processed"
			  }
			}
		]
	}`
	w.Write([]byte(body))
}

func handlerTestDocumentSearch(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := `{
		"totalCount": 2,
		"documents": [
			{
			  "id": "626626a0-749f-11e2-bfd6-000000000000",
			  "creationDate": 1360623867402,
			  "name": "scanned.jpg",
			  "progress": "COMPLETED",
			  "origin": "UPLOAD",
			  "sourceClassification": "SCANNED",
			  "pageCount": 1,
			  "pages" : [
			    {
			      "images" : {
			        "750x900" : "http://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/pages/1/750x900",
			        "1280x1810" : "http://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/pages/1/1280x1810"
			      },
			      "pageNumber" : 1
			    }
			  ],
			  "_links": {
			    "extractions": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/extractions",
			    "layout": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/layout",
			    "document": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000",
			    "processed": "https://api.gini.net/documents/626626a0-749f-11e2-bfd6-000000000000/processed"
			  }
			},
			{
			  "id": "626626a0-749f-11e2-abc2-000000000000",
			  "creationDate": 1360624287987,
			  "name": "native.pdf",
			  "progress": "COMPLETED",
			  "origin": "UPLOAD",
			  "sourceClassification": "NATIVE",
			  "pageCount": 1,
			  "pages" : [
			    {
			      "images" : {
			        "750x900" : "http://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/pages/1/750x900",
			        "1280x1810" : "http://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/pages/1/1280x1810"
			      },
			      "pageNumber" : 1
			    }
			  ],
			  "_links": {
			    "extractions": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/extractions",
			    "layout": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/layout",
			    "document": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000",
			    "processed": "https://api.gini.net/documents/626626a0-749f-11e2-abc2-000000000000/processed"
			  }
			}
		]
	}`
	w.Write([]byte(body))
}
