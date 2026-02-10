package main

import (
	"os"

	"project-initiator/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
