package main

import "fmt"

// Version is set here
const Version string = "0.0.3"

// Get the current version
func getVersion() {
	fmt.Println(Version)
}
