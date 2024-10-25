package main

import (
	"fmt"
	"os"

	"github.com/moov-io/watchman/internal/ui"
)

func main() {
	app, err := ui.Setup()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app.Run()
}
