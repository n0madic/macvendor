package macvendor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

// decodeJsonLines decodes JSON lines from the reader into a map of VendorItems
func decodeJsonLines(r io.Reader) (map[string]*VendorItem, error) {
	var vendors []Vendor
	dec := json.NewDecoder(r)
	if err := dec.Decode(&vendors); err != nil {
		return nil, err
	}

	result := make(map[string]*VendorItem, len(vendors))
	for i := range vendors {
		vendor := &vendors[i]
		vendor.CompanyName = strings.ReplaceAll(vendor.CompanyName, "`", "'")

		flags := ComputeFlags(vendor.AssignmentBlockSize, vendor.IsPrivate)
		ouiBytes := MacToByteSlice(vendor.OUI)

		lastUpdateUnix, err := parseDate(vendor.LastUpdate)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for vendor %s: %v", vendor.OUI, err)
		}

		result[vendor.OUI] = &VendorItem{
			Flags:       flags,
			CompanyName: vendor.CompanyName,
			LastUpdate:  lastUpdateUnix,
			OUI:         ouiBytes,
		}
	}
	return result, nil
}

// LoadSourceDB loads the vendor database from a file or URL
func LoadSourceDB(source string) (map[string]*VendorItem, error) {
	r, err := getSourceReader(source)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return decodeJsonLines(r)
}

// ComputeFlags calculates the flag value for the given block size and privacy status
func ComputeFlags(blockType string, isPrivate bool) uint8 {
	var flags uint8
	switch blockType {
	case "MA-L":
		flags |= FlagMAL
	case "MA-M":
		flags |= FlagMAM
	case "MA-S":
		flags |= FlagMAS
	case "IAB":
		flags |= FlagIAB
	}
	if isPrivate {
		flags |= FlagPrivate
	}
	return flags
}

// MacToByteSlice converts a MAC address string to a byte slice
func MacToByteSlice(macStr string) []byte {
	macParts := strings.Split(macStr, ":")
	bytes := make([]byte, len(macParts))
	for i, part := range macParts {
		b, _ := strconv.ParseUint(part, 16, 8)
		bytes[i] = byte(b)
	}
	return bytes
}

// ByteSliceToMac converts a byte slice to a MAC address string
func ByteSliceToMac(mac []byte) string {
	hexParts := make([]string, len(mac))

	// Iterate over each byte and format it as a hexadecimal string
	for i, b := range mac {
		// For the last byte, use %x to potentially remove the leading zero
		if i == len(mac)-1 {
			hexParts[i] = fmt.Sprintf("%x", b) // Single digit without leading zero
		} else {
			hexParts[i] = fmt.Sprintf("%02x", b) // Two-digit with leading zero
		}
	}

	// Join the parts with a colon separator
	return strings.Join(hexParts, ":")

}

// parseDate attempts to parse the date string into a Unix timestamp
func parseDate(dateStr string) (int64, error) {
	// Define multiple date formats to try
	formats := []string{
		"2006-01-02 15:04:05", // Full date-time
		"2006-01-02",          // Date only
		"2006/01/02",          // Date with slashes
		"2006/01/02 15:04:05", // Full date-time with slashes
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, fmt.Errorf("unsupported date format: %s", dateStr)
}
