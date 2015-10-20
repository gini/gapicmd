package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/dkerwin/gini-api-go"
	"github.com/fatih/color"
	"io/ioutil"
	"net/url"
	"os"
)

// getApiClient create a Gini API client from cli context
func getApiClient(c *cli.Context) *giniapi.APIClient {
	credentials := getClientCredentials(c)

	apiConfig := giniapi.Config{
		ClientID:       credentials[0],
		ClientSecret:   credentials[1],
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
	u := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(u, userid)

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
	u := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(u, userid)

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
	u := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(u, userid)

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
	incubator := c.Bool("incubator")
	userid := getUserIdentifier(c)

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)
	u := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(u, userid)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	ext, err := doc.GetExtractions(incubator)
	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults(ext)

	if c.GlobalBool("curl") {
		accept := "application/vnd.gini.v1+json"
		if incubator {
			accept = "application/vnd.gini.incubator+json"
		}

		curl := curlData{
			Headers: map[string]string{
				"Accept":            accept,
				"X-User-Identifier": userid,
			},
			Body:   "",
			URL:    doc.Links.Extractions,
			Method: "GET",
		}

		curl.render(c)
	}
}

func reportError(c *cli.Context) {
	summary := c.String("summary")
	description := c.String("description")
	userid := c.GlobalString("user-id")

	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, c.Command.FullName())
		return
	}

	api := getApiClient(c)
	u := fmt.Sprintf("%s/documents/%s", api.Endpoints.API, c.Args().First())

	doc, err := api.Get(u, userid)

	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	err = doc.ErrorReport(summary, description)
	if err != nil {
		color.Red("\nError: %s\n\n", err)
		return
	}

	done <- true
	wg.Wait()

	renderResults("")

	if c.GlobalBool("curl") {
		curl := curlData{
			Headers: map[string]string{
				"Accept":            "application/vnd.gini.v1+json",
				"X-User-Identifier": userid,
			},
			Body:   fmt.Sprintf("-d \"summary=%s&description=%s\"", url.QueryEscape(summary), url.QueryEscape(description)),
			URL:    fmt.Sprintf("%s/errorreport", doc.Links.Document),
			Method: "POST",
		}

		curl.render(c)
	}
}
