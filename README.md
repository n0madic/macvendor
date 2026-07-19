# macvendor

Go library with embedded DB for getting MAC address vendor information.

Source database from [maclookup.app](https://maclookup.app/).

## Installation

`go get -u github.com/n0madic/macvendor`

## Usage

To get detailed information about a MAC address, use `Lookup` function:

```go
import "github.com/n0madic/macvendor"

vendor, err := macvendor.Lookup("00:00:5e:00:53:01")
if err != nil {
    panic(err)
}
fmt.Println(vendor.CompanyName)
```

The database is embedded as a plain-text TSV file and parsed lazily on the
first `Lookup` call into a compact binary-search index (about 1 MB of heap;
company names are served directly from the embedded data without copying).

When the vendor base is no longer required, you can free the index memory using

```go
macvendor.FreeEmbeddedDB()
```
