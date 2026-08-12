package file

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"runtime"
	"server/app/constant"
	"strings"
	"sync"
	"time"
)

const (
	uploadMetaFileName = "meta.json"
	uploadDoneFileName = "completed.json"
	uploadChunksDir    = "chunks"
	maxUploadChunks    = 1000000
	maxUploadChunkSize = 64 << 20
	uploadSessionTTL   = 30 * 24 * time.Hour
	uploadCleanupDelay = time.Hour
)

type UploadMeta struct {
	UploadID    string `json:"upload_id"`
	Path        string `json:"path"`
	FileName    string `json:"file_name"`
	TotalSize   int64  `json:"total_size"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
	FileHash    string `json:"file_hash"`
}

type UploadedFile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	FileHash string `json:"file_hash"`
}

type uploadSessionLock struct {
	mutex sync.Mutex
	refs  int
}

var uploadSessionLockMap = struct {
	sync.Mutex
	items map[string]*uploadSessionLock
}{items: make(map[string]*uploadSessionLock)}

var uploadCleanupState struct {
	sync.Mutex
	last time.Time
}

func (s *service) Upload(ctx context.Context, fileHeader *multipart.FileHeader, path string) (*UploadedFile, error) {
	cleanupExpiredUploadSessions()
	fileName := fileHeader.Filename
	if err := validateUploadFileName(fileName); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(uploadRootPath(), 0755); err != nil {
		return nil, fmt.Errorf("创建上传缓存失败: %w", err)
	}
	sessionPath, err := os.MkdirTemp(uploadRootPath(), "direct-*")
	if err != nil {
		return nil, fmt.Errorf("创建上传缓存失败: %w", err)
	}
	defer os.RemoveAll(sessionPath)

	tempPath := filepath.Join(sessionPath, "payload")
	if err = saveMultipartFile(fileHeader, tempPath, fileHeader.Size); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	uploadPath := filepath.Clean(path)
	if uploadPath == "." {
		return nil, fmt.Errorf("path参数不合法")
	}
	destPath := filepath.Join(uploadPath, fileName)
	targetUnlock := lockUploadTarget(destPath)
	defer targetUnlock()
	if err = commitUploadedFile(ctx, tempPath, destPath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	return &UploadedFile{Name: fileName, Path: destPath, Size: fileHeader.Size}, nil
}

func (s *service) UploadChunk(fileHeader *multipart.FileHeader, meta UploadMeta, chunkIndex int) (*UploadedFile, bool, error) {
	cleanupExpiredUploadSessions()
	if err := meta.validate(); err != nil {
		return nil, false, err
	}
	if chunkIndex < 0 || chunkIndex >= meta.TotalChunks {
		return nil, false, fmt.Errorf("chunk_index参数不合法")
	}
	unlock := lockUploadSession(meta.UploadID)
	defer unlock()
	if err := verifyUploadSession(meta); err != nil {
		return nil, false, err
	}
	touchUploadSession(meta.UploadID)
	targetUnlock := lockUploadTarget(filepath.Join(meta.Path, meta.FileName))
	completed, found, completedErr := readCompletedUpload(meta)
	targetUnlock()
	if completedErr != nil {
		return nil, false, completedErr
	}
	if found {
		return completed, false, nil
	}
	exists, err := saveUploadChunk(fileHeader, meta, chunkIndex)
	if err != nil {
		return nil, false, fmt.Errorf("保存分片失败: %w", err)
	}
	return nil, exists, nil
}

func (s *service) CheckUpload(meta UploadMeta) ([]int, *UploadedFile, error) {
	cleanupExpiredUploadSessions()
	if err := meta.validate(); err != nil {
		return nil, nil, err
	}
	unlock := lockUploadSession(meta.UploadID)
	defer unlock()
	if err := ensureUploadSession(meta); err != nil {
		return nil, nil, err
	}
	touchUploadSession(meta.UploadID)
	targetUnlock := lockUploadTarget(filepath.Join(meta.Path, meta.FileName))
	completed, found, completedErr := readCompletedUpload(meta)
	targetUnlock()
	if completedErr != nil {
		return nil, nil, completedErr
	}
	if found {
		return allUploadChunkIndexes(meta.TotalChunks), completed, nil
	}
	uploadedChunks, err := uploadedChunkIndexes(meta)
	if err != nil {
		return nil, nil, fmt.Errorf("获取分片状态失败: %w", err)
	}
	return uploadedChunks, nil, nil
}

func (s *service) MergeUpload(ctx context.Context, meta UploadMeta) (*UploadedFile, bool, error) {
	cleanupExpiredUploadSessions()
	if err := meta.validate(); err != nil {
		return nil, false, err
	}
	unlock := lockUploadSession(meta.UploadID)
	defer unlock()
	if err := verifyUploadSession(meta); err != nil {
		return nil, false, err
	}
	touchUploadSession(meta.UploadID)
	targetUnlock := lockUploadTarget(filepath.Join(meta.Path, meta.FileName))
	defer targetUnlock()
	completed, found, completedErr := readCompletedUpload(meta)
	if completedErr != nil {
		return nil, false, completedErr
	}
	if found {
		return completed, true, nil
	}
	mergedPath, size, err := mergeUploadChunks(ctx, meta)
	if err != nil {
		return nil, false, err
	}
	destPath := filepath.Join(meta.Path, meta.FileName)
	if err = commitUploadedFile(ctx, mergedPath, destPath); err != nil {
		return nil, false, fmt.Errorf("保存合并文件失败: %w", err)
	}
	completed = &UploadedFile{Name: meta.FileName, Path: destPath, Size: size, FileHash: meta.FileHash}
	if err = writeCompletedUpload(meta.UploadID, completed); err != nil {
		return nil, false, fmt.Errorf("记录上传结果失败: %w", err)
	}
	_ = os.RemoveAll(filepath.Join(uploadSessionPath(meta.UploadID), uploadChunksDir))
	_ = os.Remove(mergedPath)
	return completed, false, nil
}

func (s *service) ClearUpload(uploadID string) error {
	cleanupExpiredUploadSessions()
	if uploadID == "" || len(uploadID) > 128 || uploadID == "." || uploadID == ".." || strings.ContainsAny(uploadID, `/\`) {
		return fmt.Errorf("upload_id参数不合法")
	}
	unlock := lockUploadSession(uploadID)
	defer unlock()
	if err := os.RemoveAll(uploadSessionPath(uploadID)); err != nil {
		return fmt.Errorf("清理上传缓存失败: %w", err)
	}
	return nil
}

func cleanupExpiredUploadSessions() {
	uploadCleanupState.Lock()
	if time.Since(uploadCleanupState.last) < uploadCleanupDelay {
		uploadCleanupState.Unlock()
		return
	}
	uploadCleanupState.last = time.Now()
	uploadCleanupState.Unlock()

	entries, err := os.ReadDir(uploadRootPath())
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-uploadSessionTTL)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		unlock := lockUploadSession(entry.Name())
		info, infoErr = os.Stat(uploadSessionPath(entry.Name()))
		if infoErr == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(uploadSessionPath(entry.Name()))
		}
		unlock()
	}
}

func lockUploadSession(uploadID string) func() {
	uploadSessionLockMap.Lock()
	entry := uploadSessionLockMap.items[uploadID]
	if entry == nil {
		entry = &uploadSessionLock{}
		uploadSessionLockMap.items[uploadID] = entry
	}
	entry.refs++
	uploadSessionLockMap.Unlock()

	entry.mutex.Lock()
	return func() {
		entry.mutex.Unlock()
		uploadSessionLockMap.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(uploadSessionLockMap.items, uploadID)
		}
		uploadSessionLockMap.Unlock()
	}
}

func lockUploadTarget(path string) func() {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		absolutePath = filepath.Clean(path)
	}
	absolutePath = filepath.Join(resolveUploadLockDirectory(filepath.Dir(absolutePath)), filepath.Base(absolutePath))
	if runtime.GOOS == "windows" {
		absolutePath = strings.ToLower(absolutePath)
	}
	return lockUploadSession("target\x00" + absolutePath)
}

func resolveUploadLockDirectory(path string) string {
	cleanPath := filepath.Clean(path)
	currentPath := cleanPath
	missingParts := make([]string, 0)
	for {
		resolvedPath, err := filepath.EvalSymlinks(currentPath)
		if err == nil {
			for index := len(missingParts) - 1; index >= 0; index-- {
				resolvedPath = filepath.Join(resolvedPath, missingParts[index])
			}
			return resolvedPath
		}
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return cleanPath
		}
		missingParts = append(missingParts, filepath.Base(currentPath))
		currentPath = parentPath
	}
}

func touchUploadSession(uploadID string) {
	now := time.Now()
	_ = os.Chtimes(uploadSessionPath(uploadID), now, now)
}

func (meta UploadMeta) validate() error {
	if meta.UploadID == "" || len(meta.UploadID) > 128 || meta.UploadID == "." || meta.UploadID == ".." || strings.ContainsAny(meta.UploadID, `/\`) {
		return fmt.Errorf("upload_id参数不合法")
	}
	if err := validateUploadFileName(meta.FileName); err != nil {
		return err
	}
	if meta.TotalSize < 0 {
		return fmt.Errorf("total_size参数不合法")
	}
	if meta.ChunkSize <= 0 || meta.ChunkSize > maxUploadChunkSize {
		return fmt.Errorf("chunk_size参数不合法")
	}
	if meta.TotalChunks <= 0 || meta.TotalChunks > maxUploadChunks {
		return fmt.Errorf("total_chunks参数不合法")
	}
	decodedHash, err := hex.DecodeString(meta.FileHash)
	if err != nil || len(decodedHash) != md5.Size {
		return fmt.Errorf("file_hash参数不合法")
	}
	if meta.Path == "." {
		return fmt.Errorf("path参数不合法")
	}
	expectedChunks := int64(1)
	if meta.TotalSize > 0 {
		expectedChunks = meta.TotalSize / meta.ChunkSize
		if meta.TotalSize%meta.ChunkSize != 0 {
			expectedChunks++
		}
	}
	if expectedChunks != int64(meta.TotalChunks) {
		return fmt.Errorf("分片数量与文件大小不一致")
	}
	return nil
}

func (meta UploadMeta) matches(other UploadMeta) bool {
	return meta.UploadID == other.UploadID &&
		meta.Path == other.Path &&
		meta.FileName == other.FileName &&
		meta.TotalSize == other.TotalSize &&
		meta.ChunkSize == other.ChunkSize &&
		meta.TotalChunks == other.TotalChunks &&
		meta.FileHash == other.FileHash
}

func validateUploadFileName(fileName string) error {
	if fileName == "" || fileName == "." || fileName == ".." || strings.ContainsRune(fileName, 0) || strings.ContainsAny(fileName, `/\`) || filepath.Base(fileName) != fileName {
		return fmt.Errorf("file_name参数不合法")
	}
	return nil
}

func uploadSessionPath(uploadID string) string {
	return filepath.Join(uploadRootPath(), uploadID)
}

func uploadMetadataPath(uploadID string) string {
	return filepath.Join(uploadSessionPath(uploadID), uploadMetaFileName)
}

func completedUploadPath(uploadID string) string {
	return filepath.Join(uploadSessionPath(uploadID), uploadDoneFileName)
}

func uploadChunkPath(uploadID string, chunkIndex int) string {
	return filepath.Join(uploadSessionPath(uploadID), uploadChunksDir, fmt.Sprintf("%d.part", chunkIndex))
}

func ensureUploadSession(meta UploadMeta) error {
	stored, err := readUploadMeta(meta.UploadID)
	if err == nil {
		if !stored.matches(meta) {
			return fmt.Errorf("上传任务信息不一致")
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("读取上传任务失败: %w", err)
	}

	sessionPath := uploadSessionPath(meta.UploadID)
	if err = os.RemoveAll(sessionPath); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(sessionPath, uploadChunksDir), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(sessionPath, ".meta-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err = temp.Write(data); err == nil {
		err = temp.Sync()
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tempPath, uploadMetadataPath(meta.UploadID))
}

func verifyUploadSession(meta UploadMeta) error {
	stored, err := readUploadMeta(meta.UploadID)
	if os.IsNotExist(err) {
		return fmt.Errorf("上传任务不存在，请重新检查")
	}
	if err != nil {
		return fmt.Errorf("读取上传任务失败: %w", err)
	}
	if !stored.matches(meta) {
		return fmt.Errorf("上传任务信息不一致")
	}
	return nil
}

func readUploadMeta(uploadID string) (UploadMeta, error) {
	data, err := os.ReadFile(uploadMetadataPath(uploadID))
	if err != nil {
		return UploadMeta{}, err
	}
	var meta UploadMeta
	if err = json.Unmarshal(data, &meta); err != nil {
		return UploadMeta{}, err
	}
	if err = meta.validate(); err != nil {
		return UploadMeta{}, err
	}
	return meta, nil
}

func expectedUploadChunkSize(meta UploadMeta, chunkIndex int) int64 {
	if chunkIndex < meta.TotalChunks-1 {
		return meta.ChunkSize
	}
	return meta.TotalSize - int64(meta.TotalChunks-1)*meta.ChunkSize
}

func uploadedChunkIndexes(meta UploadMeta) ([]int, error) {
	uploaded := make([]int, 0, meta.TotalChunks)
	for index := 0; index < meta.TotalChunks; index++ {
		chunkPath := uploadChunkPath(meta.UploadID, index)
		info, err := os.Stat(chunkPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() != expectedUploadChunkSize(meta, index) {
			_ = os.Remove(chunkPath)
			continue
		}
		uploaded = append(uploaded, index)
	}
	return uploaded, nil
}

func allUploadChunkIndexes(totalChunks int) []int {
	chunks := make([]int, totalChunks)
	for index := range chunks {
		chunks[index] = index
	}
	return chunks
}

func readCompletedUpload(meta UploadMeta) (*UploadedFile, bool, error) {
	data, err := os.ReadFile(completedUploadPath(meta.UploadID))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var completed UploadedFile
	if err = json.Unmarshal(data, &completed); err != nil {
		return nil, false, err
	}
	expectedPath := filepath.Join(meta.Path, meta.FileName)
	if completed.Name != meta.FileName || completed.Path != expectedPath || completed.Size != meta.TotalSize ||
		completed.FileHash != meta.FileHash {
		return nil, false, fmt.Errorf("上传完成记录不合法")
	}
	info, err := os.Stat(completed.Path)
	if err != nil {
		return nil, false, fmt.Errorf("校验已上传文件失败: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() != completed.Size {
		return nil, false, fmt.Errorf("已上传文件不存在或不完整")
	}
	fileHash, err := hashUploadFile(completed.Path)
	if err != nil {
		return nil, false, fmt.Errorf("校验已上传文件失败: %w", err)
	}
	if fileHash != completed.FileHash {
		return nil, false, fmt.Errorf("已上传文件校验失败")
	}
	return &completed, true, nil
}

func writeCompletedUpload(uploadID string, completed *UploadedFile) error {
	data, err := json.Marshal(completed)
	if err != nil {
		return err
	}
	sessionPath := uploadSessionPath(uploadID)
	temp, err := os.CreateTemp(sessionPath, ".completed-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err = temp.Write(data); err == nil {
		err = temp.Sync()
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tempPath, completedUploadPath(uploadID))
}

func saveUploadChunk(fileHeader *multipart.FileHeader, meta UploadMeta, chunkIndex int) (bool, error) {
	chunkPath := uploadChunkPath(meta.UploadID, chunkIndex)
	expectedSize := expectedUploadChunkSize(meta, chunkIndex)
	if info, err := os.Stat(chunkPath); err == nil && info.Mode().IsRegular() && info.Size() == expectedSize {
		return true, nil
	}
	_ = os.Remove(chunkPath)
	if fileHeader.Size != expectedSize {
		return false, fmt.Errorf("分片大小不一致: %d/%d", fileHeader.Size, expectedSize)
	}
	if err := saveMultipartFile(fileHeader, chunkPath, expectedSize); err != nil {
		return false, err
	}
	return false, nil
}

func saveMultipartFile(fileHeader *multipart.FileHeader, destPath string, expectedSize int64) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	source, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	temp, err := os.CreateTemp(filepath.Dir(destPath), ".upload-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	written, err := io.Copy(temp, source)
	if err == nil && expectedSize >= 0 && written != expectedSize {
		err = fmt.Errorf("上传文件大小不一致: %d/%d", written, expectedSize)
	}
	if err == nil {
		err = temp.Sync()
	}
	if err == nil {
		err = temp.Chmod(0644)
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(destPath)
	}
	return os.Rename(tempPath, destPath)
}

func mergeUploadChunks(ctx context.Context, meta UploadMeta) (string, int64, error) {
	sessionPath := uploadSessionPath(meta.UploadID)
	merged, err := os.CreateTemp(sessionPath, ".merged-*")
	if err != nil {
		return "", 0, err
	}
	mergedTempPath := merged.Name()
	defer os.Remove(mergedTempPath)

	var size int64
	fileHash := md5.New()
	for index := 0; index < meta.TotalChunks; index++ {
		if err = ctx.Err(); err != nil {
			_ = merged.Close()
			return "", 0, err
		}
		chunkPath := uploadChunkPath(meta.UploadID, index)
		info, statErr := os.Stat(chunkPath)
		if statErr != nil {
			_ = merged.Close()
			if os.IsNotExist(statErr) {
				return "", 0, fmt.Errorf("分片 %d 不存在或不完整", index)
			}
			return "", 0, fmt.Errorf("读取分片 %d 失败: %w", index, statErr)
		}
		if !info.Mode().IsRegular() || info.Size() != expectedUploadChunkSize(meta, index) {
			_ = merged.Close()
			return "", 0, fmt.Errorf("分片 %d 不存在或不完整", index)
		}
		chunk, openErr := os.Open(chunkPath)
		if openErr != nil {
			_ = merged.Close()
			return "", 0, fmt.Errorf("打开分片 %d 失败: %w", index, openErr)
		}
		written, copyErr := copyUploadData(ctx, io.MultiWriter(merged, fileHash), chunk)
		closeErr := chunk.Close()
		if copyErr != nil {
			_ = merged.Close()
			return "", 0, fmt.Errorf("合并分片 %d 失败: %w", index, copyErr)
		}
		if closeErr != nil {
			_ = merged.Close()
			return "", 0, closeErr
		}
		size += written
	}
	if size != meta.TotalSize {
		_ = merged.Close()
		return "", 0, fmt.Errorf("合并文件大小不一致: %d/%d", size, meta.TotalSize)
	}
	if hex.EncodeToString(fileHash.Sum(nil)) != meta.FileHash {
		_ = merged.Close()
		_ = os.RemoveAll(sessionPath)
		return "", 0, fmt.Errorf("合并文件校验失败，请重新上传")
	}
	if err = merged.Sync(); err == nil {
		err = merged.Chmod(0644)
	}
	if closeErr := merged.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", 0, err
	}

	mergedPath := filepath.Join(sessionPath, "merged")
	_ = os.Remove(mergedPath)
	if err = os.Rename(mergedTempPath, mergedPath); err != nil {
		return "", 0, err
	}
	return mergedPath, size, nil
}

func commitUploadedFile(ctx context.Context, sourcePath, destPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	if info, statErr := os.Lstat(destPath); statErr == nil && info.IsDir() {
		return fmt.Errorf("目标路径是目录")
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return statErr
	}
	if err := replaceUploadedFile(sourcePath, destPath); err == nil {
		return nil
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	temp, err := os.CreateTemp(filepath.Dir(destPath), "."+filepath.Base(destPath)+".upload-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err = copyUploadData(ctx, temp, source); err == nil {
		err = temp.Sync()
	}
	if err == nil {
		err = temp.Chmod(0644)
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = replaceUploadedFile(tempPath, destPath); err != nil {
		return err
	}
	_ = os.Remove(sourcePath)
	return nil
}

func hashUploadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func replaceUploadedFile(sourcePath, destPath string) error {
	if runtime.GOOS != "windows" {
		return os.Rename(sourcePath, destPath)
	}
	if _, err := os.Lstat(destPath); os.IsNotExist(err) {
		return os.Rename(sourcePath, destPath)
	}
	backup, err := os.CreateTemp(filepath.Dir(destPath), "."+filepath.Base(destPath)+".backup-*")
	if err != nil {
		return err
	}
	backupPath := backup.Name()
	_ = backup.Close()
	_ = os.Remove(backupPath)
	if err = os.Rename(destPath, backupPath); err != nil {
		return err
	}
	if err = os.Rename(sourcePath, destPath); err != nil {
		if restoreErr := os.Rename(backupPath, destPath); restoreErr != nil {
			return fmt.Errorf("替换目标文件失败: %v，恢复原文件失败: %v", err, restoreErr)
		}
		return err
	}
	_ = os.Remove(backupPath)
	return nil
}

func copyUploadData(ctx context.Context, dest io.Writer, source io.Reader) (int64, error) {
	buffer := make([]byte, 128*1024)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		read, readErr := source.Read(buffer)
		if read > 0 {
			count, writeErr := dest.Write(buffer[:read])
			written += int64(count)
			if writeErr != nil {
				return written, writeErr
			}
			if count != read {
				return written, io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return written, nil
		}
		if readErr != nil {
			return written, readErr
		}
	}
}

func uploadRootPath() string {
	return filepath.Join(constant.TempPath, "upload")
}
