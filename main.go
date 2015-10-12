package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/dkerwin/gini-api-go"
	"github.com/fatih/color"
	"os"
	"sync"
	"text/template"
)

type curlData struct {
	Headers map[string]string
	Body    string
}

var (
	wg       sync.WaitGroup
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
			Name:        "upload",
			Usage:       "upload document to Gini's API",
			Description: "Upload the given PDF/image argument and keep polling until the processing is complete. Result is displayed in pretty-printed JSON",
			Aliases:     []string{"u"},
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
			Name:        "get",
			Usage:       "get document details from Gini's API", // of an account
			Description: "Get deocument deatils for documentId",
			Aliases:     []string{"g"},
			Action: func(c *cli.Context) {
				disableColors(c)
				getDocument(c)
			},
		},
		{
			Name:        "list",
			Usage:       "list a user's documents", // of an account
			Description: "List a user's documents with pagination and offset",
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
	}

	fmt.Printf("\n")

	app.Run(os.Args)
}

func RenderCurlCommand(headers map[string]string, body string) (string, error) {
	// if c.GlobalBool("curl") {
	//  headers := map[string]string{
	//      "xxxx": "aaa",
	//      "yyyy": "bbb",
	//  }

	//  curl, _ := RenderCurlCommand(headers, "")
	//  fmt.Println(curl)
	// }

	tpl := ">> curl -v {{range $key, $value := $.Headers}}-H \"{{$key}} :{{$value}}\" {{end}}https://api.gini.net/documents"
	var curl bytes.Buffer

	data := curlData{
		Headers: headers,
	}

	t := template.New("bozo")
	t.Parse(tpl)
	err := t.Execute(&curl, data)

	return curl.String(), err
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
	userid := c.GlobalString("user-id")

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

	boldMagenta := color.New(color.BgMagenta).Add(color.Bold).Add(color.Underline)
	boldMagenta.Println("★★★ The results are in ★★★\n")

	pretty, err := prettyJSON(doc)

	if err != nil {
		color.Red("%s: %s", pretty, err)
	} else {
		color.Magenta("%s", pretty)
	}
}

func getDocument(c *cli.Context) {
	userid := c.GlobalString("user-id")

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)

	doc, err := api.Get(fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First()), userid)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	boldMagenta := color.New(color.BgMagenta).Add(color.Bold).Add(color.Underline)
	boldMagenta.Println("★★★ The results are in ★★★\n")

	pretty, err := prettyJSON(doc)

	if err != nil {
		color.Red("%s: %s", pretty, err)
	} else {
		color.Magenta("%s", pretty)
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

	boldMagenta := color.New(color.BgMagenta).Add(color.Bold).Add(color.Underline)
	boldMagenta.Println("★★★ The results are in ★★★\n")

	pretty, err := prettyJSON(doc)

	if err != nil {
		color.Red("%s: %s", pretty, err)
	} else {
		color.Magenta("%s", pretty)
	}
}

func disableColors(c *cli.Context) {
	if c.GlobalBool("no-color") {
		color.NoColor = true
	}
}

func prettyJSON(obj interface{}) ([]byte, error) {
	result, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return []byte("Failed to prettify JSON object"), err
	}
	return result, nil
}
