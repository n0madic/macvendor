package main

import (
	"fmt"
	"os"

	"github.com/n0madic/macvendor"
)

func main() {
	vendor, err := macvendor.Lookup(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println("OUI:", vendor.OUI)
	fmt.Println("AssignmentBlockSize:", vendor.AssignmentBlockSize)
	fmt.Println("IsPrivate:", vendor.IsPrivate)
	fmt.Println("CompanyName:", vendor.CompanyName)
	fmt.Println("LastUpdate:", vendor.LastUpdate)
}
