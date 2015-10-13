package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"text/template"
)

func getUserIdentifier(c *cli.Context) string {
	userid := c.GlobalString("user-id")
	if userid == "" {
		userid = createUserIdentifier()
	}
	return userid
}

func createUserIdentifier() string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz=-_."
	bytes := make([]byte, 64)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes)
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

func RenderCurlCommand(c *cli.Context, method, url string, headers map[string]string, body string) error {
	tpl := fmt.Sprintf("❯❯❯ curl -v -X{{$.Method}} -u \"%s:%s\" {{range $key, $value := $.Headers}}-H \"{{$key}}: {{$value}}\" {{end}}{{$.Body}} {{$.URL}}", c.GlobalString("client-id"), c.GlobalString("client-secret"))
	var curl bytes.Buffer

	data := curlData{
		Headers: headers,
		Body:    body,
		URL:     url,
		Method:  method,
	}

	t := template.New("bozo")
	t.Parse(tpl)
	err := t.Execute(&curl, data)

	if err != nil {
		color.Red("Error: %s", err)
		return err
	}
	boldYellow := color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold).Add(color.Underline)
	boldYellow.Println("\n★★★ cURL command to replay request ★★★\n")
	color.Yellow("%s", curl.String())

	return nil
}
