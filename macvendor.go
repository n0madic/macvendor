package macvendor

import (
	"errors"
	"net"
	"sync"
	"time"
)

// VendorItem info with optimized structure for internal storage
type VendorItem struct {
	Flags       uint8  // Flags for block size and privacy
	CompanyName string // Name of the company
	LastUpdate  int64  // Unix timestamp for the last update
	OUI         []byte // OUI stored as a byte slice
}

// Original Vendor struct for output purposes
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
	vendorTrie  *Trie
	ErrNotFound = errors.New("MAC not found in DB")
)

// TrieNode represents a node in the trie
type TrieNode struct {
	Children map[byte]*TrieNode
	Vendor   *VendorItem
}

// Trie structure to store MAC prefixes
type Trie struct {
	Root *TrieNode
}

// Insert adds a MAC prefix and its Vendor information into the trie
func (t *Trie) Insert(prefix []byte, vendor *VendorItem) {
	node := t.Root
	for _, b := range prefix {
		if node.Children == nil {
			node.Children = make(map[byte]*TrieNode)
		}
		if node.Children[b] == nil {
			node.Children[b] = &TrieNode{}
		}
		node = node.Children[b]
	}
	node.Vendor = vendor
}

// Search finds the Vendor for a given MAC address prefix
func (t *Trie) Search(mac []byte) (*VendorItem, bool) {
	node := t.Root
	for _, b := range mac {
		if node.Children == nil || node.Children[b] == nil {
			return nil, false
		}
		node = node.Children[b]
	}
	return node.Vendor, node.Vendor != nil
}

// Lookup finds the OUI the address belongs to and converts it to a readable format
func Lookup(mac string) (*Vendor, error) {
	// Try to parse the MAC address
	addr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	dbMutex.RLock()
	if vendorTrie == nil {
		dbMutex.RUnlock()
		dbMutex.Lock()
		if vendorTrie == nil {
			vendorTrie, err = LoadEmbeddedDB()
			if err != nil {
				dbMutex.Unlock()
				return nil, err
			}
		}
		dbMutex.Unlock()
		dbMutex.RLock()
	}
	defer dbMutex.RUnlock()

	// Convert the MAC address to a byte slice
	addrBytes := []byte(addr.String())

	// Attempt to find the vendor by checking prefixes in the trie
	prefixLengths := []int{13, 10, 8}
	for _, length := range prefixLengths {
		if len(addrBytes) >= length {
			if vendorItem, found := vendorTrie.Search(addrBytes[:length]); found {
				return convertVendorItemToVendor(vendorItem), nil
			}
		}
	}

	return nil, ErrNotFound
}

// ConvertVendorItemToVendor converts a VendorItem to a Vendor
func convertVendorItemToVendor(vendorItem *VendorItem) *Vendor {
	return &Vendor{
		AssignmentBlockSize: vendorItem.BlockSize(),
		CompanyName:         vendorItem.CompanyName,
		IsPrivate:           vendorItem.IsPrivate(),
		LastUpdate:          time.Unix(vendorItem.LastUpdate, 0).UTC().Format("2006/01/02"),
		OUI:                 ByteSliceToMac(vendorItem.OUI),
	}
}

// FreeEmbeddedDB frees memory used by the database (if not already needed)
func FreeEmbeddedDB() {
	dbMutex.Lock()
	vendorTrie = nil
	dbMutex.Unlock()
}

// Additional helper function to check block size flags
func (v *VendorItem) BlockSize() string {
	switch {
	case v.Flags&FlagMAL != 0:
		return "MA-L"
	case v.Flags&FlagMAM != 0:
		return "MA-M"
	case v.Flags&FlagMAS != 0:
		return "MA-S"
	case v.Flags&FlagIAB != 0:
		return "IAB"
	default:
		return "Unknown"
	}
}

// Check if the vendor is private
func (v *VendorItem) IsPrivate() bool {
	return v.Flags&FlagPrivate != 0
}
