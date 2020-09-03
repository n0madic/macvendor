package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/n0madic/macvendor"
)

func main() {
	inputPtr := flag.String("input", "https://macaddress.io/database/macaddress.io-db.json", "source database.json (file or url)")
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
		embedb += fmt.Sprintf("    \"%s\": &Vendor{OUI: `%s`, AssignmentBlockSize: `%s`, IsPrivate: %t, CompanyName: `%s`, CompanyAddress: `%s`, CountryCode: `%s`},\n",
			k,
			db[k].OUI,
			db[k].AssignmentBlockSize,
			db[k].IsPrivate,
			db[k].CompanyName,
			db[k].CompanyAddress,
			db[k].CountryCode,
		)
	}
	embedb += fmt.Sprintf("  }\n}\n")

	err = ioutil.WriteFile(*outputPtr, []byte(embedb), 0644)
	if err != nil {
		panic(err)
	}
}
