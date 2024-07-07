package fileuploader

import (
	"encoding/json"
	"os"
)

func (m *DefaultMetadataManager) LoadMetadata(filePath string) (map[string]ChunkMetadata, error) {
	metadata := make(map[string]ChunkMetadata)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func (m *DefaultMetadataManager) SaveMetadata(filePath string, chunkMetadata map[string]ChunkMetadata) error {
	data, err := json.MarshalIndent(chunkMetadata, "", "")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
