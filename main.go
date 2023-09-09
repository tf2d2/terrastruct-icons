package main

import (
	"os"

	"github.com/tf2d2/terrastruct-icons/icons"
)

func main() {
	if err := icons.Generate(); err != nil {
		os.Exit(1)
	}
}
