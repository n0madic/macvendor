package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/n0madic/macvendor"
)

func main() {
	inputPtr := flag.String("input", macvendor.SourceURL, "source database.json (file or url)")
	outputPtr := flag.String("output", "db.tsv", "destination TSV file with embedded DB")

	flag.Parse()

	db, err := macvendor.LoadSourceDB(*inputPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading source DB: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Create(*outputPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}

	if err := macvendor.WriteTSV(file, db); err != nil {
		file.Close()
		fmt.Fprintf(os.Stderr, "Error writing TSV: %v\n", err)
		os.Exit(1)
	}

	if err := file.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing output file: %v\n", err)
		os.Exit(1)
	}
}
