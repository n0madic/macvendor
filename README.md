# macvendor

Go library with embedded DB for getting MAC address vendor information.

Source database from [macaddress.io](https://macaddress.io/database-download).

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

When the vendor base is no longer required, you can free memory in the heap using

```go
macvendor.FreeEmbeddedDB()
```

Frees more than 10 megabytes.
