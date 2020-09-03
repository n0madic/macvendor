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
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	} else {
		return os.Open(source)
	}
}

func decodeJsonLines(r io.Reader) (map[string]*Vendor, error) {
	vendors := map[string]*Vendor{}
	dec := json.NewDecoder(r)
	for {
		var vendor Vendor
		err := dec.Decode(&vendor)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		vendor.CompanyAddress = strings.ReplaceAll(vendor.CompanyAddress, "`", "'")
		vendor.CompanyName = strings.ReplaceAll(vendor.CompanyName, "`", "'")
		vendors[vendor.OUI] = &Vendor{
			AssignmentBlockSize: vendor.AssignmentBlockSize,
			CompanyName:         vendor.CompanyName,
			CompanyAddress:      vendor.CompanyAddress,
			CountryCode:         vendor.CountryCode,
			IsPrivate:           vendor.IsPrivate,
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
