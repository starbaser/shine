// Package main is the entry point for the shine application.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("✨ Shine - A Go project")

	if len(os.Args) > 1 {
		fmt.Printf("Args: %v\n", os.Args[1:])
	}
}
