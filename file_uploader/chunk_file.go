package fileuploader

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	NUM_OF_WORKERS = 4
)

func (c *DefaultFileChunker) ChunkFile(filePath string) ([]ChunkMetadata, error) {
	var chunks []ChunkMetadata

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, c.ChunkSize)
	index := 0

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}

		//if all inside file is chunked, will break
		if bytesRead == 0 {
			break
		}

		hash := md5.Sum(buffer[:bytesRead])
		hashString := hex.EncodeToString(hash[:])

		fileName := fmt.Sprintf("%s.chunk.%d", filePath, index)

		fileChunk, err := os.Create(fileName)
		if err != nil {
			return nil, err
		}

		_, err = fileChunk.Write(buffer[:bytesRead])
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, ChunkMetadata{FileName: fileName, MD5Hash: hashString, Index: index})

		fileChunk.Close()

		index++
	}

	return chunks, nil
}

func (c *DefaultFileChunker) ChunkLargeFile(filePath string) ([]ChunkMetadata, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var chunks []ChunkMetadata

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	numOfChunks := int64(fileInfo.Size()) / int64(c.ChunkSize)
	if int64(fileInfo.Size())%int64(c.ChunkSize) != 0 {
		numOfChunks++
	}

	chunkChan := make(chan ChunkMetadata, numOfChunks)
	errChan := make(chan error, numOfChunks)
	indexChan := make(chan int, numOfChunks)

	for i := 0; i < int(numOfChunks); i++ {
		indexChan <- i
	}
	close(indexChan)

	for i := 0; i < NUM_OF_WORKERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range indexChan {
				offset := index * c.ChunkSize
				buffer := make([]byte, c.ChunkSize)

				file.Seek(int64(offset), 0)

				bytesRead, err := file.Read(buffer)
				if err != nil {
					errChan <- err
					return
				}

				if bytesRead > 0 {
					hash := md5.Sum(buffer[:bytesRead])
					hashString := hex.EncodeToString(hash[:])

					fileName := fmt.Sprintf("%s.chunk.%d", filePath, index)

					fileChunk, err := os.Create(fileName)
					if err != nil {
						errChan <- err
						return
					}

					_, err = fileChunk.Write(buffer[:bytesRead])
					if err != nil {
						errChan <- err
						return
					}

					mu.Lock()
					chunks = append(chunks, ChunkMetadata{FileName: fileName, MD5Hash: hashString, Index: index})
					mu.Unlock()

					fileChunk.Close()
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(chunkChan)
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return chunks, nil
}
