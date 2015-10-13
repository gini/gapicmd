package main

import (
	"crypto/rand"
	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
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
