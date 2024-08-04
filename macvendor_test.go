package macvendor

import (
	"errors"
	"reflect"
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
				OUI:                 "54:bf:64",
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
				OUI:                 "94:05:bb:9",
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
				OUI:                 "70:b3:d5:e6:f",
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

func BenchmarkLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Lookup("A4:81:EE:FF:FF:FF")
		if err != nil {
			b.Fatalf("Unexpected error: %s", err)
		}
	}
}
