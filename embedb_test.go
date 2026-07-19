package macvendor

import (
	"strings"
	"testing"
)

func TestParseDB(t *testing.T) {
	data := "000000\t1\t16756\tXEROX CORPORATION\n" +
		"8c5db29\t2\t18277\tZimmer GmbH\n" +
		"70b3d5e6f\t4\t18166\tAmazon Technologies Inc.\n" +
		"4c424c\t16\t0\t\n"

	d, err := parseDB(data)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		oui   []byte
		name  string
		flags uint8
		days  uint16
	}{
		{[]byte{0x00, 0x00, 0x00}, "XEROX CORPORATION", FlagMAL, 16756},
		{[]byte{0x8c, 0x5d, 0xb2, 0x09}, "Zimmer GmbH", FlagMAM, 18277},
		{[]byte{0x70, 0xb3, 0xd5, 0xe6, 0x0f}, "Amazon Technologies Inc.", FlagMAS, 18166},
		{[]byte{0x4c, 0x42, 0x4c}, "", FlagPrivate, 0},
	}
	for _, tt := range tests {
		rec, ok := d.find(tt.oui)
		if !ok {
			t.Fatalf("OUI %x not found", tt.oui)
		}
		if d.name(rec) != tt.name || rec.flags != tt.flags || rec.days != tt.days {
			t.Errorf("OUI %x: got (%q, %d, %d), want (%q, %d, %d)",
				tt.oui, d.name(rec), rec.flags, rec.days, tt.name, tt.flags, tt.days)
		}
	}

	if _, ok := d.find([]byte{0xff, 0xff, 0xff}); ok {
		t.Error("unexpected match for missing OUI")
	}
}

func TestParseDBErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"missing-fields", "000000\t1\t16756\n"},
		{"bad-prefix", "00zz00\t1\t16756\tACME\n"},
		{"bad-prefix-length", "00000\t1\t16756\tACME\n"},
		{"bad-flags", "000000\tx\t16756\tACME\n"},
		{"bad-date", "000000\t1\t99999999\tACME\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseDB(tt.data); err == nil {
				t.Error("expected parse error")
			}
		})
	}
}

func TestDatabaseRelevance(t *testing.T) {
	onlineDB, err := LoadSourceDB(SourceURL)
	if err != nil {
		t.Fatal(err)
	}

	d, err := loadEmbeddedDB()
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range onlineDB {
		rec, ok := d.find(want.OUI)
		if !ok {
			t.Fatalf("Missing OUI: %v", want)
		}
		days := want.LastUpdate / 86400
		if want.LastUpdate <= 0 {
			days = 0
		}
		name := strings.Map(func(r rune) rune {
			if r == '\t' || r == '\n' || r == '\r' {
				return ' '
			}
			return r
		}, want.CompanyName)
		if d.name(rec) != name || rec.flags != want.Flags || int64(rec.days) != days {
			t.Errorf("Changed OUI %x: get (%q, %d, %d), present (%q, %d, %d)",
				want.OUI, name, want.Flags, days, d.name(rec), rec.flags, rec.days)
		}
	}
}
