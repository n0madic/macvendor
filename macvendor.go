package macvendor

import (
	"fmt"
	"net"
	"strings"
)

// Vendor info
type Vendor struct {
	// Assignment block size, one of the following:
	// 'MA-L' for MAC Address Block Large
	// 'MA-M' for MAC Address Block Medium
	// 'MA-S' for MAC Address Block Small
	// 'IAB' for Individual Address Block
	AssignmentBlockSize string `json:"assignmentBlockSize"`
	// Name of the company which registered the MAC addresses block.
	CompanyName string `json:"companyName"`
	// Company's full address.
	CompanyAddress string `json:"companyAddress"`
	// Company's country code in ISO 3166 format.
	CountryCode string `json:"countryCode"`
	// For an extra fee to IEEE, vendors can hide their details.
	// In this case, this flag is set to 'true' and companyName,
	// companyAddress and countryCode are 'private'.
	IsPrivate bool `json:"isPrivate"`
	// Organization Unique Identifier
	OUI string `json:"oui"`
}

var (
	embeddedDB  map[string]*Vendor
	ErrNotFound = fmt.Errorf("MAC not found in DB")
)

// Lookup finds the OUI the address belongs to
func Lookup(mac string) (*Vendor, error) {
	addr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}
	if embeddedDB == nil || len(embeddedDB) == 0 {
		embeddedDB = LoadEmbeddedDB()
	}
	for _, i := range []int{13, 10, 8} {
		prefix := strings.ToUpper(addr.String()[:i])
		vendor, ok := embeddedDB[prefix]
		if ok {
			return vendor, nil
		}
	}
	return nil, ErrNotFound
}

// FreeEmbeddedDB frees memory used by the database (if not already needed)
func FreeEmbeddedDB() {
	embeddedDB = make(map[string]*Vendor)
}
