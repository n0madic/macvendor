package macvendor

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/gob"
	"fmt"
)

//go:embed embedb.bin.gz
var embedb []byte

// LoadEmbeddedDB loads and decompresses the Trie from a gzip-compressed binary file
func LoadEmbeddedDB() (*Trie, error) {
	reader, err := gzip.NewReader(bytes.NewReader(embedb))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	var trie Trie
	decoder := gob.NewDecoder(reader)

	if err := decoder.Decode(&trie); err != nil {
		return nil, fmt.Errorf("failed to decode trie: %w", err)
	}

	return &trie, nil
}
