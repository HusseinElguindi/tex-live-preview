package main

import (
	"os"

	"github.com/husseinelguindi/tex-live-preview/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
