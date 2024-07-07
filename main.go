package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	fileuploader "github.com/ahmadfarhanstwn/go-file-uploader-chunk/file_uploader"
)

const (
	DEFAULT_CHUNK_SIZE = 1024 * 1024 //1MB
	MAX_RETRIES        = 3
	SERVER_URL         = "http://localhost:5000"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <filepath>")
	}
	filePath := os.Args[1]

	config := fileuploader.Config{ChunkSize: DEFAULT_CHUNK_SIZE, ServerURL: SERVER_URL}
	chunker := &fileuploader.DefaultFileChunker{ChunkSize: config.ChunkSize}
	uploader := &fileuploader.DefaultUploader{ServerURL: config.ServerURL}
	metadataManager := &fileuploader.DefaultMetadataManager{}

	chunks, err := chunker.ChunkFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	metadata, err := metadataManager.LoadMetadata(fmt.Sprintf("%s.metadata.json", filePath))
	if err != nil {
		log.Println("Could not load metadata, starting fresh.")
		metadata = make(map[string]fileuploader.ChunkMetadata)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	err = fileuploader.Synchronize(chunks, metadata, uploader, &wg, &mu)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	err = metadataManager.SaveMetadata(fmt.Sprintf("%s.metadata.json", filePath), metadata)
	if err != nil {
		log.Fatal(err)
	}

	changeChan := make(chan bool)
	go fileuploader.WatchFile(filePath, changeChan)

	for {
		select {
		case <-changeChan:
			log.Println("File changed, re-chunking and synchronizing...")
			chunks, err = chunker.ChunkFile(filePath)
			if err != nil {
				log.Fatal(err)
			}

			err = fileuploader.Synchronize(chunks, metadata, uploader, &wg, &mu)
			if err != nil {
				log.Fatal(err)
			}

			wg.Wait()

			err = metadataManager.SaveMetadata(fmt.Sprintf("%s.metadata.json", filePath), metadata)
			if err != nil {
				log.Fatal(err)
			}
		case <-time.After(10 * time.Second):
			log.Println("No changes detected, checking again...")
		}
	}
}
