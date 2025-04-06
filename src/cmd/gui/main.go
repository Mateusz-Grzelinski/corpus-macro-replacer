package main

import (
	"fmt"
)

//go:generate fyne bundle -o src/bundled.go assets/

const Version = "0.7"

func main() {
	fmt.Printf("Starting GUI. All command line flags are ignored")
	RunGui()
}
