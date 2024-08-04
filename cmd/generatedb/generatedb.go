package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/n0madic/macvendor"
)

func main() {
	inputPtr := flag.String("input", "https://maclookup.app/downloads/json-database/get-db", "source database.json (file or url)")
	outputPtr := flag.String("output", "embedb.go", "destination file with embedded DB")

	flag.Parse()

	db, err := macvendor.LoadSourceDB(*inputPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading source DB: %v\n", err)
		os.Exit(1)
	}

	keys := make([]string, 0, len(db))
	for k := range db {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("package macvendor\n\nfunc LoadEmbeddedDB() map[string]*Vendor {\n  return map[string]*Vendor{\n")

	for _, k := range keys {
		v := db[k]
		fmt.Fprintf(&sb, "    %q: {OUI: %q, AssignmentBlockSize: %q, IsPrivate: %t, CompanyName: %q, LastUpdate: %q},\n",
			k,
			v.OUI,
			v.AssignmentBlockSize,
			v.IsPrivate,
			v.CompanyName,
			v.LastUpdate)
	}

	sb.WriteString("  }\n}")

	err = os.WriteFile(*outputPtr, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
}
