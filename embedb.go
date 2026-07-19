package macvendor

import (
	"cmp"
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// embedded holds the vendor database as sorted TSV lines:
// <hex prefix>\t<flags>\t<days since Unix epoch>\t<company name>\n
// The prefix length encodes the block size: 6 hex chars for 24-bit (MA-L),
// 7 for 28-bit (MA-M), 9 for 36-bit (MA-S/IAB) prefixes.
//
//go:embed db.tsv
var embedded string

// record references a single vendor entry inside the TSV data.
// Company names are not copied: they are addressed by offset/length
// into the data string, so the parsed database contains no pointers.
type record struct {
	nameOff uint32
	nameLen uint16
	days    uint16 // days since Unix epoch; 0 means unknown date
	flags   uint8
}

type entry[K uint32 | uint64] struct {
	key K
	rec record
}

// db is the parsed vendor database: one sorted slice per prefix length.
type db struct {
	data string
	mal  []entry[uint32] // 24-bit prefixes
	mam  []entry[uint32] // 28-bit prefixes
	mas  []entry[uint64] // 36-bit prefixes
}

// parseDB parses TSV data into a searchable database.
// Records keep offsets into data, so it must stay alive as long as the db.
func parseDB(data string) (*db, error) {
	d := &db{data: data}
	for off, num := 0, 1; off < len(data); num++ {
		line := data[off:]
		if end := strings.IndexByte(line, '\n'); end >= 0 {
			line = line[:end]
		}
		lineOff := off
		off += len(line) + 1
		if line == "" {
			continue
		}
		if err := d.addLine(line, uint32(lineOff)); err != nil {
			return nil, fmt.Errorf("embedded DB line %d: %w", num, err)
		}
	}
	byKey := func(a, b entry[uint32]) int { return cmp.Compare(a.key, b.key) }
	slices.SortFunc(d.mal, byKey)
	slices.SortFunc(d.mam, byKey)
	slices.SortFunc(d.mas, func(a, b entry[uint64]) int { return cmp.Compare(a.key, b.key) })
	return d, nil
}

func (d *db) addLine(line string, lineOff uint32) error {
	prefix, rest, ok1 := strings.Cut(line, "\t")
	flagsStr, rest, ok2 := strings.Cut(rest, "\t")
	daysStr, name, ok3 := strings.Cut(rest, "\t")
	if !ok1 || !ok2 || !ok3 {
		return fmt.Errorf("expected 4 tab-separated fields: %q", line)
	}

	key, err := strconv.ParseUint(prefix, 16, 64)
	if err != nil {
		return fmt.Errorf("invalid prefix %q: %w", prefix, err)
	}
	flags, err := strconv.ParseUint(flagsStr, 10, 8)
	if err != nil {
		return fmt.Errorf("invalid flags %q: %w", flagsStr, err)
	}
	days, err := strconv.ParseUint(daysStr, 10, 16)
	if err != nil {
		return fmt.Errorf("invalid date %q: %w", daysStr, err)
	}

	rec := record{
		nameOff: lineOff + uint32(len(line)-len(name)),
		nameLen: uint16(len(name)),
		days:    uint16(days),
		flags:   uint8(flags),
	}
	switch len(prefix) {
	case 6:
		d.mal = append(d.mal, entry[uint32]{uint32(key), rec})
	case 7:
		d.mam = append(d.mam, entry[uint32]{uint32(key), rec})
	case 9:
		d.mas = append(d.mas, entry[uint64]{key, rec})
	default:
		return fmt.Errorf("unsupported prefix length %d: %q", len(prefix), prefix)
	}
	return nil
}

// loadEmbeddedDB parses the embedded TSV database.
func loadEmbeddedDB() (*db, error) {
	return parseDB(embedded)
}

func search[K uint32 | uint64](entries []entry[K], key K) (record, bool) {
	i, ok := slices.BinarySearchFunc(entries, key, func(e entry[K], k K) int {
		return cmp.Compare(e.key, k)
	})
	if !ok {
		return record{}, false
	}
	return entries[i].rec, true
}

// find looks up a record by OUI in VendorItem representation, where the
// last byte of 4- and 5-byte prefixes holds a single trailing nibble.
func (d *db) find(oui []byte) (record, bool) {
	switch len(oui) {
	case 3:
		return search(d.mal, uint32(oui[0])<<16|uint32(oui[1])<<8|uint32(oui[2]))
	case 4:
		return search(d.mam, uint32(oui[0])<<20|uint32(oui[1])<<12|uint32(oui[2])<<4|uint32(oui[3]))
	case 5:
		return search(d.mas, uint64(oui[0])<<28|uint64(oui[1])<<20|uint64(oui[2])<<12|uint64(oui[3])<<4|uint64(oui[4]))
	default:
		return record{}, false
	}
}

func (d *db) name(rec record) string {
	return d.data[rec.nameOff : rec.nameOff+uint32(rec.nameLen)]
}

func (d *db) vendor(rec record, oui string) *Vendor {
	var lastUpdate time.Time // zero days means unknown date, kept as zero time
	if rec.days > 0 {
		lastUpdate = time.Unix(int64(rec.days)*86400, 0).UTC()
	}
	return &Vendor{
		AssignmentBlockSize: blockSize(rec.flags),
		CompanyName:         d.name(rec),
		IsPrivate:           rec.flags&FlagPrivate != 0,
		LastUpdate:          lastUpdate.Format("2006/01/02"),
		OUI:                 oui,
	}
}

func oui24(key uint32) string {
	return fmt.Sprintf("%02x:%02x:%02x", byte(key>>16), byte(key>>8), byte(key))
}

func oui28(key uint32) string {
	return fmt.Sprintf("%02x:%02x:%02x:%x", byte(key>>20), byte(key>>12), byte(key>>4), key&0xf)
}

func oui36(key uint64) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%x", byte(key>>28), byte(key>>20), byte(key>>12), byte(key>>4), key&0xf)
}
