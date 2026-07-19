package macvendor

import (
	"errors"
	"net"
	"sync"
)

// Vendor holds MAC vendor information
type Vendor struct {
	AssignmentBlockSize string `json:"blockType"`  // Assignment block size
	CompanyName         string `json:"vendorName"` // Name of the company
	IsPrivate           bool   `json:"private"`    // Privacy flag
	LastUpdate          string `json:"lastUpdate"` // Last update in human-readable form
	OUI                 string `json:"macPrefix"`  // MAC Prefix in human-readable form
}

// Bit flags for assignment block size
const (
	FlagMAL     = 1 << iota // MA-L for MAC Address Block Large
	FlagMAM                 // MA-M for MAC Address Block Medium
	FlagMAS                 // MA-S for MAC Address Block Small
	FlagIAB                 // IAB for Individual Address Block
	FlagPrivate             // Privacy flag
)

var (
	dbMutex     sync.RWMutex
	vendorDB    *db
	ErrNotFound = errors.New("MAC not found in DB")
)

// getDB lazily parses the embedded database on first use
func getDB() (*db, error) {
	dbMutex.RLock()
	d := vendorDB
	dbMutex.RUnlock()
	if d != nil {
		return d, nil
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()
	if vendorDB == nil {
		d, err := loadEmbeddedDB()
		if err != nil {
			return nil, err
		}
		vendorDB = d
	}
	return vendorDB, nil
}

// Lookup finds the OUI the address belongs to and converts it to a readable format
func Lookup(mac string) (*Vendor, error) {
	addr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	d, err := getDB()
	if err != nil {
		return nil, err
	}

	key24 := uint32(addr[0])<<16 | uint32(addr[1])<<8 | uint32(addr[2])
	key28 := key24<<4 | uint32(addr[3]>>4)
	key36 := uint64(key28)<<8 | uint64(addr[3]&0xf)<<4 | uint64(addr[4]>>4)

	// Check the most specific prefixes first
	if rec, ok := search(d.mas, key36); ok {
		return d.vendor(rec, oui36(key36)), nil
	}
	if rec, ok := search(d.mam, key28); ok {
		return d.vendor(rec, oui28(key28)), nil
	}
	if rec, ok := search(d.mal, key24); ok {
		return d.vendor(rec, oui24(key24)), nil
	}

	return nil, ErrNotFound
}

// FreeEmbeddedDB frees memory used by the parsed database index
func FreeEmbeddedDB() {
	dbMutex.Lock()
	vendorDB = nil
	dbMutex.Unlock()
}

// blockSize converts block size flags to a readable form
func blockSize(flags uint8) string {
	switch {
	case flags&FlagMAL != 0:
		return "MA-L"
	case flags&FlagMAM != 0:
		return "MA-M"
	case flags&FlagMAS != 0:
		return "MA-S"
	case flags&FlagIAB != 0:
		return "IAB"
	default:
		return "Unknown"
	}
}
