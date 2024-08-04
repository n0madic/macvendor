package macvendor

import (
	"reflect"
	"strings"
	"testing"
)

func TestDatabaseRelevance(t *testing.T) {
	onlineDB, err := LoadSourceDB("https://maclookup.app/downloads/json-database/get-db")
	if err != nil {
		t.Fatal(err)
	}

	embeddedDB, err := LoadEmbeddedDB()
	if err != nil {
		t.Fatal(err)
	}

	for key, vendor := range onlineDB {
		key = strings.ToLower(key)
		if emvendor, ok := embeddedDB.Search([]byte(key)); ok {
			if !reflect.DeepEqual(*vendor, *emvendor) {
				t.Errorf("Changed OUI: get %v, present %v", vendor, *emvendor)
			}
		} else {
			t.Fatalf("Missing OUI: %v", vendor)
		}
	}
}
