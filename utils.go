package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"os"
	"strings"
	"text/template"
)

type curlData struct {
	Headers map[string]string
	Body    string
	URL     string
	Method  string
}

func (cdata *curlData) render(c *cli.Context) error {
	credentials := getClientCredentials(c)

	tpl := fmt.Sprintf("❯❯❯ curl -v -X{{$.Method}} -u \"%s:%s\" {{range $key, $value := $.Headers}}-H \"{{$key}}: {{$value}}\" {{end}}{{$.Body}} {{$.URL}}", credentials[0], credentials[1])
	var curl bytes.Buffer

	t := template.New("curl")
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
	boldMagenta.Println("★★★ Results ★★★\n")

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

func xorBytes(b1, b2 []byte) []byte {
	if len(b1) != len(b2) {
		panic("length mismatch")
	}

	rv := make([]byte, len(b1))

	for i := range b1 {
		rv[i] = b1[i] ^ b2[i]
	}

	return rv
}

func getClientCredentials(c *cli.Context) []string {
	credentials := []string{c.GlobalString("client-id"), c.GlobalString("client-secret")}

	if credentials[0] == "" || credentials[1] == "" {
		color.Yellow("No client credentials given. Fallback to builtin default...")
		color.Yellow("Keep in mind that your document might be visible to other users.")
		color.Yellow("Your unique user-id is the only secret to protect your data.\n\n")

		superSecretSecret := []byte("V;4nJvuANmoywKNYk.yewNhqwmAQctc3BvByxeozQVpiK")

		// Decode HEX default credentials
		credentialsBytes, err := hex.DecodeString(defaultClientCredentials)
		if err != nil {
			color.Red("Error: client-id and client-secret missing and fallback decoding (step 1) failed: %s\n\n", err)
			cli.ShowCommandHelp(c, c.Command.FullName())
			os.Exit(1)
		}

		decodedCredentials := strings.Split(string(xorBytes(credentialsBytes, superSecretSecret)), ":")

		if len(decodedCredentials) < 2 {
			color.Red("Error: client-id and client-secret missing and fallback decoding (step 2) failed: %s\n\n", err)
			cli.ShowCommandHelp(c, c.Command.FullName())
			os.Exit(1)
		}
		credentials = decodedCredentials
	}

	return credentials
}
