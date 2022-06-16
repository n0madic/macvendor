package macvendor

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

func getSourceReader(source string) (io.Reader, error) {
	if strings.HasPrefix(source, "http") {
		resp, err := http.Get(source)
		if err != nil && resp.StatusCode != 200 {
			return nil, err
		}
		return resp.Body, nil
	} else {
		return os.Open(source)
	}
}

func decodeJsonLines(r io.Reader) (map[string]*Vendor, error) {
	var v []Vendor
	dec := json.NewDecoder(r)
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	vendors := map[string]*Vendor{}
	for _, vendor := range v {
		vendor.CompanyName = strings.ReplaceAll(vendor.CompanyName, "`", "'")
		vendors[vendor.OUI] = &Vendor{
			AssignmentBlockSize: vendor.AssignmentBlockSize,
			CompanyName:         vendor.CompanyName,
			IsPrivate:           vendor.IsPrivate,
			LastUpdate:          vendor.LastUpdate,
			OUI:                 vendor.OUI,
		}
	}
	return vendors, nil
}

// LoadSourceDB from file or URL
func LoadSourceDB(source string) (map[string]*Vendor, error) {
	r, err := getSourceReader(source)
	if err != nil {
		return nil, err
	}
	return decodeJsonLines(r)
}
