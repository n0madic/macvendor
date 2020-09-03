package macvendor

import (
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
				OUI:                 "54:BF:64",
				IsPrivate:           false,
				CompanyName:         "Dell Inc",
				CompanyAddress:      "One Dell Way Round Rock TX 78682 US",
				CountryCode:         "US",
				AssignmentBlockSize: "MA-L",
			},
		},
		{
			name: "medium-block",
			mac:  "94:05:BB:94:B1:C4",
			want: &Vendor{
				OUI:                 "94:05:BB:9",
				IsPrivate:           false,
				CompanyName:         "Zimmer GmbH",
				CompanyAddress:      "Im Salmenkopf 5 Rheinau Baden-Württemberg 77866 DE",
				CountryCode:         "DE",
				AssignmentBlockSize: "MA-M",
			},
		},
		{
			name: "small-block",
			mac:  "70-b3-d5-e6-f1-e2",
			want: &Vendor{
				OUI:                 "70:B3:D5:E6:F",
				IsPrivate:           false,
				CompanyName:         "Amazon Tech Inc",
				CompanyAddress:      "P.O Box 8102 Reno NV 89507 US",
				CountryCode:         "US",
				AssignmentBlockSize: "MA-S",
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
}
