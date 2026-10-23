package file

import (
	"fmt"
	"path/filepath"
	serviceFile "server/app/service/file"
	"strconv"
	"strings"
)

func parseUploadMeta(uploadID, path, fileName, totalSizeValue, chunkSizeValue, totalChunksValue, fileHash string) (serviceFile.UploadMeta, error) {
	totalSize, err := strconv.ParseInt(totalSizeValue, 10, 64)
	if err != nil {
		return serviceFile.UploadMeta{}, fmt.Errorf("total_size参数不合法")
	}
	chunkSize, err := strconv.ParseInt(chunkSizeValue, 10, 64)
	if err != nil {
		return serviceFile.UploadMeta{}, fmt.Errorf("chunk_size参数不合法")
	}
	totalChunks, err := strconv.Atoi(totalChunksValue)
	if err != nil {
		return serviceFile.UploadMeta{}, fmt.Errorf("total_chunks参数不合法")
	}
	return serviceFile.UploadMeta{
		UploadID:    uploadID,
		Path:        filepath.Clean(path),
		FileName:    fileName,
		TotalSize:   totalSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		FileHash:    strings.ToLower(fileHash),
	}, nil
}
