package macvendor

import (
	"errors"
	"net"
	"strings"
	"sync"
)

// Vendor info
type Vendor struct {
	// Assignment block size, one of the following:
	// 'MA-L' for MAC Address Block Large
	// 'MA-M' for MAC Address Block Medium
	// 'MA-S' for MAC Address Block Small
	// 'IAB' for Individual Address Block
	AssignmentBlockSize string `json:"blockType"`
	// Name of the company which registered the MAC addresses block.
	CompanyName string `json:"vendorName"`
	// For an extra fee to IEEE, vendors can hide their details.
	// In this case, this flag is set to 'true' and companyName,
	// companyAddress and countryCode are 'private'.
	IsPrivate bool `json:"private"`
	// Last update record in the database.
	LastUpdate string `json:"lastUpdate"`
	// Organization Unique Identifier
	OUI string `json:"macPrefix"`
}

var (
	dbMutex     sync.RWMutex
	embeddedDB  map[string]*Vendor
	ErrNotFound = errors.New("MAC not found in DB")
)

// Lookup finds the OUI the address belongs to
func Lookup(mac string) (*Vendor, error) {
	// Try to parse the MAC address
	addr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	dbMutex.RLock()
	defer dbMutex.RUnlock()

	// Initialize the embedded database if it's nil
	if embeddedDB == nil {
		dbMutex.RUnlock()
		dbMutex.Lock()
		if embeddedDB == nil {
			embeddedDB = LoadEmbeddedDB()
		}
		dbMutex.Unlock()
		dbMutex.RLock()
	}

	// Precompute the address string once
	addrStr := strings.ToUpper(addr.String())

	// Attempt to find the vendor by checking prefixes in the map
	for _, i := range []int{13, 10, 8} {
		if len(addrStr) >= i {
			prefix := addrStr[:i]
			if vendor, ok := embeddedDB[prefix]; ok {
				return vendor, nil
			}
		}
	}

	return nil, ErrNotFound
}

// FreeEmbeddedDB frees memory used by the database (if not already needed)
func FreeEmbeddedDB() {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	embeddedDB = nil
}
