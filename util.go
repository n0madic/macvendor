package macvendor

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SourceURL is the default location of the source vendor database
const SourceURL = "https://maclookup.app/downloads/json-database/get-db"

// VendorItem info with optimized structure for internal storage
type VendorItem struct {
	Flags       uint8  // Flags for block size and privacy
	CompanyName string // Name of the company
	LastUpdate  int64  // Unix timestamp for the last update
	OUI         []byte // OUI stored as a byte slice
}

// BlockSize converts block size flags to a readable form
func (v *VendorItem) BlockSize() string {
	return blockSize(v.Flags)
}

// IsPrivate checks if the vendor is private
func (v *VendorItem) IsPrivate() bool {
	return v.Flags&FlagPrivate != 0
}

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

// WriteTSV writes the vendor database in the embedded TSV format:
// sorted lines of <hex prefix>\t<flags>\t<days since Unix epoch>\t<company name>
func WriteTSV(w io.Writer, db map[string]*VendorItem) error {
	lines := make([]string, 0, len(db))
	for _, v := range db {
		line, err := tsvLine(v)
		if err != nil {
			return err
		}
		lines = append(lines, line)
	}
	sort.Strings(lines)

	bw := bufio.NewWriter(w)
	for _, line := range lines {
		bw.WriteString(line)
	}
	return bw.Flush()
}

// tsvLine formats a single vendor entry as a TSV line
func tsvLine(v *VendorItem) (string, error) {
	prefix, err := prefixHex(v.OUI)
	if err != nil {
		return "", err
	}
	days := v.LastUpdate / 86400
	if v.LastUpdate <= 0 {
		days = 0 // unknown date sentinel
	}
	if days > math.MaxUint16 {
		return "", fmt.Errorf("date out of range for OUI %s: %d", prefix, v.LastUpdate)
	}
	name := strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, v.CompanyName)
	return fmt.Sprintf("%s\t%d\t%d\t%s\n", prefix, v.Flags, days, name), nil
}

// prefixHex converts an OUI byte slice to its hex form without separators,
// where the last byte of 4- and 5-byte prefixes holds a single trailing nibble
func prefixHex(oui []byte) (string, error) {
	switch len(oui) {
	case 3:
		return hex.EncodeToString(oui), nil
	case 4, 5:
		last := oui[len(oui)-1]
		if last > 0xf {
			return "", fmt.Errorf("last byte of OUI %x is not a single nibble", oui)
		}
		return hex.EncodeToString(oui[:len(oui)-1]) + string("0123456789abcdef"[last]), nil
	default:
		return "", fmt.Errorf("unsupported OUI length %d: %x", len(oui), oui)
	}
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
