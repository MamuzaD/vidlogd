package main

import (
	"fmt"
	"os"

	"github.com/mamuzad/vidlogd/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
