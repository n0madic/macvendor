package macvendor

import (
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		name string
		mac  string
		want *Vendor
	}{
		{
			name: "large-block",
			mac:  "54:bf:64:51:c5:44",
			want: &Vendor{
				OUI:                 "54:BF:64",
				IsPrivate:           false,
				CompanyName:         "Dell Inc.",
				AssignmentBlockSize: "MA-L",
				LastUpdate:          "2018/02/15",
			},
		},
		{
			name: "medium-block",
			mac:  "94:05:BB:94:B1:C4",
			want: &Vendor{
				OUI:                 "94:05:BB:9",
				IsPrivate:           false,
				CompanyName:         "Zimmer GmbH",
				AssignmentBlockSize: "MA-M",
				LastUpdate:          "2020/01/16",
			},
		},
		{
			name: "small-block",
			mac:  "70-b3-d5-e6-f1-e2",
			want: &Vendor{
				OUI:                 "70:B3:D5:E6:F",
				IsPrivate:           false,
				CompanyName:         "Amazon Technologies Inc.",
				AssignmentBlockSize: "MA-S",
				LastUpdate:          "2019/09/27",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Lookup(tt.mac)
			if err != nil {
				t.Errorf("Lookup() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lookup() = %v, want %v", got, tt.want)
			}
		})
	}

	_, err := Lookup("FF-FF-FF-FF-FF-FF")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Lookup() ErrNotFound fail")
	}
}

func BenchmarkLoop(b *testing.B) {
	addr, err := net.ParseMAC("A4:81:EE:FF:FF:FF")
	if err != nil {
		b.Fatalf("Unexpected error: %s", err)
	}
	embeddedDB := LoadEmbeddedDB()
	for i := 0; i < b.N; i++ {
		for _, v := range embeddedDB {
			if strings.Contains(strings.ToUpper(addr.String()), v.OUI) {
				break
			}
		}
	}
}

func BenchmarkLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Lookup("A4:81:EE:FF:FF:FF")
		if err != nil {
			b.Fatalf("Unexpected error: %s", err)
		}
	}
}
