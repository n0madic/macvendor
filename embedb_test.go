package macvendor

import (
	"reflect"
	"testing"
)

func TestDatabaseRelevance(t *testing.T) {
	onlineDB, err := LoadSourceDB("https://maclookup.app/downloads/json-database/get-db")
	if err != nil {
		t.Fatal(err)
	}

	embeddedDB := LoadEmbeddedDB()

	for key, vendor := range onlineDB {
		if emvendor, ok := embeddedDB[key]; ok {
			if !reflect.DeepEqual(*vendor, *emvendor) {
				t.Errorf("Changed OUI: get %v, present %v", vendor, *emvendor)
			}
			delete(embeddedDB, key)
		} else {
			t.Errorf("Found new OUI in online DB: %s", key)
		}
	}
	if len(embeddedDB) != 0 {
		t.Errorf("Found %d deleted keys in embedded DB", len(embeddedDB))
	}
}
