package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/dkerwin/gini-api-go"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"sync"
)

var (
	wg       sync.WaitGroup
	once     sync.Once
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
					boldBlue.Println("★★★ HTTP requests ★★★\n")
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
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{Name: "Daniel Kerwin",
			Email: "d.kerwin@gini.net",
		},
	}
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
			Usage: "upload document to Gini's API",
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
			Usage: "get document details from Gini's API",
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
			Name:  "get-processed",
			Usage: "get processed document details from Gini's API",
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
			Name:  "extractions",
			Usage: "get document extractions and candidates",
			Description: `Get document extractions for given documentId.
   See http://developer.gini.net/gini-api/html/documents.html#retrieving-extractions for details.`,
			ArgsUsage: "[documentId]",
			Aliases:   []string{"e"},
			Action: func(c *cli.Context) {
				disableColors(c)
				getExtractions(c)
			},
		},
	}

	fmt.Printf("\n")

	app.Run(os.Args)
}

// getApiClient create a Gini API client from cli context
func getApiClient(c *cli.Context) *giniapi.APIClient {
	apiConfig := giniapi.Config{
		ClientID:       c.GlobalString("client-id"),
		ClientSecret:   c.GlobalString("client-secret"),
		Authentication: giniapi.UseBasicAuth,
		Endpoints: giniapi.Endpoints{
			API: "https://api.gini.net",
		},
	}

	if c.GlobalBool("debug") {
		apiConfig.HTTPDebug = true
		apiConfig.RequestDebug = request
		apiConfig.ResponseDebug = response
	}

	api, err := giniapi.NewClient(&apiConfig)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		cli.ShowCommandHelp(c, c.Command.FullName())
		os.Exit(1)
	}

	return api
}

func uploadDocument(c *cli.Context) {
	filename := c.String("filename")
	doctype := c.String("doctype")
	userid := getUserIdentifier(c)

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	if _, err := os.Stat(c.Args().First()); os.IsNotExist(err) {
		color.Red("\nError: cannot find %s\n\n", c.Args().First())
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	bodyBuf, err := os.Open(c.Args().First())
	if err != nil {
		color.Red("\nError: failed to read %s\n\n", c.Args().First())
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)

	doc, err := api.Upload(bodyBuf, giniapi.UploadOptions{
		FileName:       filename,
		DocType:        doctype,
		UserIdentifier: userid,
	})

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults(doc)

	if c.GlobalBool("curl") {
		curl := curlData{
			Headers: map[string]string{
				"Accept":            "application/vnd.gini.v1+json",
				"X-User-Identifier": userid,
			},
			Body:   fmt.Sprintf("--data-binary '@%s'", c.Args().First()),
			URL:    fmt.Sprintf("%s/documents", api.Endpoints.API),
			Method: "GET",
		}

		curl.render(c)
	}
}

func getDocument(c *cli.Context) {
	userid := getUserIdentifier(c)

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)
	url := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(url, userid)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults(doc)

	if c.GlobalBool("curl") {
		curl := curlData{
			Headers: map[string]string{
				"Accept":            "application/vnd.gini.v1+json",
				"X-User-Identifier": userid,
			},
			Body:   "",
			URL:    doc.Links.Document,
			Method: "GET",
		}

		curl.render(c)
	}
}

func getProcessed(c *cli.Context) {
	userid := getUserIdentifier(c)

	if len(c.Args()) != 2 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)
	url := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(url, userid)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	body, err := doc.GetProcessed()

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	err = ioutil.WriteFile(c.Args()[1], body, 0644)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults(doc)

	if c.GlobalBool("curl") {
		curl := curlData{
			Headers: map[string]string{
				"Accept":            "application/vnd.gini.v1+json,application/octet-stream",
				"X-User-Identifier": userid,
			},
			Body:   "",
			URL:    doc.Links.Processed,
			Method: "GET",
		}

		curl.render(c)
	}
}

func deleteDocument(c *cli.Context) {
	userid := getUserIdentifier(c)

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)
	url := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(url, userid)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	err = doc.Delete()

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults("empty response")

	if c.GlobalBool("curl") {
		curl := curlData{
			Headers: map[string]string{
				"Accept":            "application/vnd.gini.v1+json",
				"X-User-Identifier": userid,
			},
			Body:   "",
			URL:    doc.Links.Document,
			Method: "DELETE",
		}

		curl.render(c)
	}
}

func listDocuments(c *cli.Context) {
	limit := c.Int("limit")
	offset := c.Int("offset")
	userid := c.GlobalString("user-id")

	api := getApiClient(c)

	doc, err := api.List(giniapi.ListOptions{
		Limit:          limit,
		Offset:         offset,
		UserIdentifier: userid,
	})

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults(doc)

	if c.GlobalBool("curl") {
		curl := curlData{
			Headers: map[string]string{
				"Accept":            "application/vnd.gini.v1+json",
				"X-User-Identifier": userid,
			},
			Body:   "",
			URL:    fmt.Sprintf("%s/documents?limit=%d&offset=%d", api.Endpoints.API, limit, offset),
			Method: "GET",
		}

		curl.render(c)
	}
}

func getExtractions(c *cli.Context) {
	userid := getUserIdentifier(c)

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)
	url := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(url, userid)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	ext, err := doc.GetExtractions()
	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults(ext)

	if c.GlobalBool("curl") {
		curl := curlData{
			Headers: map[string]string{
				"Accept":            "application/vnd.gini.v1+json",
				"X-User-Identifier": userid,
			},
			Body:   "",
			URL:    doc.Links.Extractions,
			Method: "GET",
		}

		curl.render(c)
	}
}
