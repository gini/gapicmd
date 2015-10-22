package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"os"
	"sync"
)

var (
	wg      sync.WaitGroup
	once    sync.Once
	Version = "0.0.0-dev"

	defaultClientCredentials string

	request  = make(chan []byte)
	response = make(chan []byte)
	done     = make(chan bool)
)

func main() {
	go func() {
		wg.Add(1)
		for {
			select {
			case r := <-request:
				once.Do(func() {
					boldBlue := color.New(color.FgBlue).Add(color.Bold).Add(color.Underline)
					boldBlue.Printf("★★★ HTTP requests ★★★\n\n")
				})

				color.Green("client ❯❯❯ gini API\n\n")
				color.Green("%s\n\n", r)
			case r := <-response:
				color.Cyan("client ❮❮❮ gini API\n\n")
				color.Cyan("%s\n\n", r)
			case <-done:
				wg.Done()
				return
			}
		}
	}()

	app := cli.NewApp()
	app.Name = "gapicmd"
	app.Usage = "interact with Gini's API service from the command line"
	app.Version = Version
	app.Authors = []cli.Author{
		cli.Author{Name: "Daniel Kerwin",
			Email: "d.kerwin@gini.net",
		},
	}
	app.Copyright = "2015 - Gini GmbH (https://www.gini.net/developers/)"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "curl, c",
			Usage: "Show curl command to replay",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Show HTTP requests and responses",
		},
		cli.BoolFlag{
			Name:  "no-color, n",
			Usage: "Disable colorized output",
		},
		cli.StringFlag{
			Name:   "client-id",
			EnvVar: "CLIENT_ID",
			Usage:  "Gini API client ID",
		},
		cli.StringFlag{
			Name:   "client-secret",
			EnvVar: "CLIENT_SECRET",
			Usage:  "Gini API client secret",
		},
		cli.StringFlag{
			Name:   "user-id",
			EnvVar: "USER_ID",
			Usage:  "Random user identfier string #freestyle",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "upload",
			Usage: "upload a new document",
			Description: `Upload the given PDF/image argument and keep polling until the processing is complete. Result is displayed in pretty-printed JSON.
   See http://developer.gini.net/gini-api/html/documents.html#submitting-files for details.`,
			ArgsUsage: "[path to PDF/Image]",
			Aliases:   []string{"u"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "filename",
					EnvVar: "FILENAME",
					Usage:  "file name of the submitted document",
				},
				cli.StringFlag{
					Name:   "doctype",
					EnvVar: "DOCTYPE",
					Usage:  "doctype hint",
				},
			},
			Action: func(c *cli.Context) {
				disableColors(c)
				uploadDocument(c)
			},
		},
		{
			Name:  "get",
			Usage: "get document details",
			Description: `Get document details for given documentId.
   See http://developer.gini.net/gini-api/html/documents.html#checking-processing-status-and-getting-document-information for details.`,
			ArgsUsage: "[doumentId]",
			Aliases:   []string{"g"},
			Action: func(c *cli.Context) {
				disableColors(c)
				getDocument(c)
			},
		},
		{
			Name:  "get-extractions",
			Usage: "get document extractions and candidates",
			Description: `Get document extractions for given documentId.
   See http://developer.gini.net/gini-api/html/documents.html#retrieving-extractions for details.`,
			ArgsUsage: "[documentId]",
			Aliases:   []string{"e"},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "incubator",
					EnvVar: "INCUBATOR",
					Usage:  "access immature extractions which are still in research or under development",
				},
			},
			Action: func(c *cli.Context) {
				disableColors(c)
				getExtractions(c)
			},
		},
		{
			Name:  "get-processed",
			Usage: "get processed document",
			Description: `Get processed document (e.g. deskewed) for given documentId.
   See http://developer.gini.net/gini-api/html/documents.html#retrieving-the-processed-document for details.`,
			ArgsUsage: "[doumentId] [target filename]",
			Aliases:   []string{"p"},
			Action: func(c *cli.Context) {
				disableColors(c)
				getProcessed(c)
			},
		},
		{
			Name:  "delete",
			Usage: "delete a document",
			Description: `Delete document with given documentId.
   See http://developer.gini.net/gini-api/html/documents.html#deleting-documents for details.`,
			ArgsUsage: "[doumentId]",
			Aliases:   []string{"d"},
			Action: func(c *cli.Context) {
				disableColors(c)
				deleteDocument(c)
			},
		},
		{
			Name:        "list",
			Usage:       "list a user's documents",
			Description: "List a user's documents with pagination and offset.",
			Aliases:     []string{"l"},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:   "limit",
					EnvVar: "LIMIT",
					Value:  25,
					Usage:  "limit number of documents to return",
				},
				cli.IntFlag{
					Name:   "offset",
					EnvVar: "OFFSET",
					Value:  0,
					Usage:  "start offset",
				},
			},
			Action: func(c *cli.Context) {
				disableColors(c)
				listDocuments(c)
			},
		},
		{
			Name:  "report",
			Usage: "submit an error report",
			Description: `Provide error details for a given document.
   This helps us creating a even better experience for you. See
   http://developer.gini.net/gini-api/html/documents.html#create-an-error-report-for-a-document for details`,
			ArgsUsage: "[doumentId]",
			Aliases:   []string{"r"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "summary",
					EnvVar: "SUMMARY",
					Usage:  "short summary of the error found",
				},
				cli.StringFlag{
					Name:   "description",
					EnvVar: "DESCRIPTION",
					Usage:  "more detailed description of the error found",
				},
			},
			Action: func(c *cli.Context) {
				disableColors(c)
				reportError(c)
			},
		},
	}

	fmt.Printf("\n")

	app.Run(os.Args)
}
