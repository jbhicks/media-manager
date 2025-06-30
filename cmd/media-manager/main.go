package main

import (
	"log"
	"os"

	"github.com/user/media-manager/internal/app"
)

func main() {
	args := os.Args
	var dir string
	if len(args) > 1 {
		dir = args[1]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
		dir = cwd
	}
	log.Printf("Opening directory: %s", dir)
	log.Printf("[DEBUG] main.go: Passing dir to app: %s", dir)
	application, err := app.NewMediaManagerApp(dir)
	if err != nil {
		log.Fatalf("Failed to create application!: %v", err)
	}
	application.Run()
}
