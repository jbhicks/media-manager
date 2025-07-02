package main

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
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

	// Setup file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("Watcher events channel closed.")
					return
				}
				log.Printf("FSNotify Event: Name=%s, Op=%s", event.Name, event.Op)
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) {
					log.Printf("Detected relevant file system change for %s, triggering rescan.", event.Name)
					application.RescanMediaDirectory()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Keep main goroutine alive
	<-make(chan struct{})
}
