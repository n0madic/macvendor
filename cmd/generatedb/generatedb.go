package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/n0madic/macvendor"
)

func main() {
	inputPtr := flag.String("input", "https://maclookup.app/downloads/json-database/get-db", "source database.json (file or url)")
	outputPtr := flag.String("output", "embedb.go", "destination file with embedded DB")

	flag.Parse()

	db, err := macvendor.LoadSourceDB(*inputPtr)
	if err != nil {
		panic(err)
	}

	keys := []string{}
	for k := range db {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	embedb := fmt.Sprintf("package macvendor\n\nfunc LoadEmbeddedDB() map[string]*Vendor {\n  return map[string]*Vendor{\n")
	for _, k := range keys {
		embedb += fmt.Sprintf("    \"%s\": {OUI: `%s`, AssignmentBlockSize: `%s`, IsPrivate: %t, CompanyName: `%s`, LastUpdate: `%s`},\n",
			k,
			db[k].OUI,
			db[k].AssignmentBlockSize,
			db[k].IsPrivate,
			db[k].CompanyName,
			db[k].LastUpdate,
		)
	}
	embedb += fmt.Sprintf("  }\n}\n")

	err = ioutil.WriteFile(*outputPtr, []byte(embedb), 0644)
	if err != nil {
		panic(err)
	}
}
