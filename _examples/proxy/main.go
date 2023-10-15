package main

import (
	"os"

	"github.com/bmf-san/gondola"
)

func main() {
	f, err := os.Open("config.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gondola.New(f)
}
