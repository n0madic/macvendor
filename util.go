package macvendor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// getSourceReader returns an io.Reader for the given source, which can be a URL or a file path
func getSourceReader(source string) (io.ReadCloser, error) {
	if strings.HasPrefix(source, "http") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
	return os.Open(source)
}

// decodeJsonLines decodes JSON lines from the reader into a map of Vendors
func decodeJsonLines(r io.Reader) (map[string]*Vendor, error) {
	var vendors []Vendor
	dec := json.NewDecoder(r)
	if err := dec.Decode(&vendors); err != nil {
		return nil, err
	}

	result := make(map[string]*Vendor, len(vendors))
	for i := range vendors {
		vendor := &vendors[i]
		vendor.CompanyName = strings.ReplaceAll(vendor.CompanyName, "`", "'")
		result[vendor.OUI] = vendor
	}
	return result, nil
}

// LoadSourceDB loads the vendor database from a file or URL
func LoadSourceDB(source string) (map[string]*Vendor, error) {
	r, err := getSourceReader(source)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return decodeJsonLines(r)
}
