package fileuploader

type ChunkMetadata struct {
	FileName string `json:"file_name"`
	MD5Hash  string `json:"md5_hash"`
	Index    int    `json:"index"`
}

type Config struct {
	ChunkSize int
	ServerURL string
}

type DefaultFileChunker struct {
	ChunkSize int
}

type DefaultUploader struct {
	ServerURL string
}

type DefaultMetadataManager struct {
}

type FileChunker interface {
	ChunkFile(filePath string) ([]ChunkMetadata, error)
	ChunkLargeFile(filePath string) ([]ChunkMetadata, error)
}

type Uploader interface {
	UploadChunk(chunk ChunkMetadata) error
}

type MetadataManager interface {
	SaveMetadata(filePath string) (map[string]ChunkMetadata, error)
	LoadMetadata(filePath string, chunkMetadata map[string]ChunkMetadata) error
}
