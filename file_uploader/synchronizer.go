package fileuploader

import "sync"

func Synchronize(chunks []ChunkMetadata, metadata map[string]ChunkMetadata, uploader Uploader, wg *sync.WaitGroup, mu *sync.Mutex) error {
	chunkChan := make(chan ChunkMetadata, len(chunks))
	errChan := make(chan error, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		chunkChan <- chunk
	}

	close(chunkChan)

	for i := 0; i < NUM_OF_WORKERS; i++ {
		go func() {
			for chunk := range chunkChan {
				defer wg.Done()

				newHash := chunk.MD5Hash

				mu.Lock()
				oldChunk, exist := metadata[chunk.FileName]
				mu.Unlock()

				if !exist || oldChunk.MD5Hash != newHash {
					err := uploader.UploadChunk(chunk)
					if err != nil {
						errChan <- err
						return
					}

					mu.Lock()
					metadata[chunk.FileName] = chunk
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
