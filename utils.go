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

type curlData struct {
	Headers map[string]string
	Body    string
	URL     string
	Method  string
}

func (cdata *curlData) render(c *cli.Context) error {
	tpl := fmt.Sprintf("❯❯❯ curl -v -X{{$.Method}} -u \"%s:%s\" {{range $key, $value := $.Headers}}-H \"{{$key}}: {{$value}}\" {{end}}{{$.Body}} {{$.URL}}", c.GlobalString("client-id"), c.GlobalString("client-secret"))
	var curl bytes.Buffer

	t := template.New("bozo")
	t.Parse(tpl)
	err := t.Execute(&curl, cdata)

	if err != nil {
		color.Red("Error: %s", err)
		return err
	}
	boldYellow := color.New(color.FgYellow).Add(color.Bold).Add(color.Underline)
	boldYellow.Println("\n★★★ cURL command to replay request ★★★\n")
	color.Yellow("%s", curl.String())

	return nil
}

func renderResults(obj interface{}) error {
	boldMagenta := color.New(color.FgMagenta).Add(color.Bold).Add(color.Underline)
	boldMagenta.Println("★★★ The results are in ★★★\n")

	pretty, err := prettyJSON(obj)

	if err != nil {
		color.Red("%s: %s\n", pretty, err)
	} else {
		color.Magenta("%s\n", pretty)
	}

	return err
}

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
