package main

import (
	"os"

	"github.com/bmf-san/gondola"
)

// TODO:
// This is the source code for development.
// Since we want to assume that it will be used by distributing it as a binary,
// we will delete this main.go later and edit the Dockerfile.
func main() {
	f, err := os.Open("config.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gondola.Run(f)
}
