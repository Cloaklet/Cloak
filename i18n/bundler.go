//+build ignore

package main

import (
	"bytes"
	"encoding/json"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

const (
	dataFileName   string = "data.go"
	dataSourceFile string = "locales.json"
)

var tplt = template.Must(template.New("").Funcs(template.FuncMap{
	"compact": compact,
}).Parse(`
package i18n

import "github.com/tidwall/gjson"

// Code generated by go generate; DO NOT EDIT!

func init() {
	data = ` + "`{{ compact .jsonBytes }}`" + `
	if !gjson.Valid(data) {
		panic("malformed i18n data")
	}
}
`))

func compact(s []byte) string {
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, s); err != nil {
		panic(err)
	}
	return buffer.String()
}

func main() {
	jsonBytes, err := ioutil.ReadFile(dataSourceFile)
	if err != nil {
		log.Fatal("Failed to read locale JSON file:", dataSourceFile, err)
	}

	var buffer bytes.Buffer
	if err := tplt.Execute(&buffer, map[string][]byte{"jsonBytes": jsonBytes}); err != nil {
		log.Fatal("Failed to generate code:", err)
	}

	code, err := format.Source(buffer.Bytes())
	if err != nil {
		log.Fatal("Error formatting generated code:", err)
	}

	if err := ioutil.WriteFile(dataFileName, code, os.ModePerm); err != nil {
		log.Fatal("Failed to write generated code to Go source file:", dataFileName, err)
	}
}
