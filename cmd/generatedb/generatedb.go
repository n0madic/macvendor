package main

import (
	"compress/gzip"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/n0madic/macvendor"
)

func main() {
	inputPtr := flag.String("input", "https://maclookup.app/downloads/json-database/get-db", "source database.json (file or url)")
	outputPtr := flag.String("output", "embedb.bin.gz", "destination compressed binary file with embedded DB")

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

	trie := buildTrie(keys, db)

	if err := serializeTrieToGzip(*outputPtr, trie); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding trie to file: %v\n", err)
		os.Exit(1)
	}
}

// buildTrie constructs a Trie from the given keys and vendor data
func buildTrie(keys []string, db map[string]*macvendor.VendorItem) *macvendor.Trie {
	trie := &macvendor.Trie{Root: &macvendor.TrieNode{Children: make(map[byte]*macvendor.TrieNode)}}
	for _, key := range keys {
		v := db[key]
		key = strings.ToLower(key)
		trie.Insert([]byte(key), v)
	}
	return trie
}

// serializeTrieToGzip encodes the Trie into a gzip-compressed file
func serializeTrieToGzip(filename string, trie *macvendor.Trie) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	encoder := gob.NewEncoder(gzWriter)
	if err := encoder.Encode(trie); err != nil {
		return fmt.Errorf("failed to encode trie: %w", err)
	}

	return nil
}
