package fileuploader

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchFile(filePath string, changeChan chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Modified file: ", event.Name)
				changeChan <- true
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error", err)
		}
	}
}
