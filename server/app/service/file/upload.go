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
	"time"
)

func (s *service) Upload(ctx context.Context, fileHeader *multipart.FileHeader, path string) (*UploadedFile, error) {
	s.cleanupExpiredUploadSessions()
	fileName := fileHeader.Filename
	if err := s.validateUploadFileName(fileName); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(s.uploadRootPath(), 0755); err != nil {
		return nil, fmt.Errorf("创建上传缓存失败: %w", err)
	}
	sessionPath, err := os.MkdirTemp(s.uploadRootPath(), "direct-*")
	if err != nil {
		return nil, fmt.Errorf("创建上传缓存失败: %w", err)
	}
	defer os.RemoveAll(sessionPath)

	tempPath := filepath.Join(sessionPath, "payload")
	if err = s.saveMultipartFile(fileHeader, tempPath, fileHeader.Size); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	uploadPath := filepath.Clean(path)
	if uploadPath == "." {
		return nil, fmt.Errorf("path参数不合法")
	}
	destPath := filepath.Join(uploadPath, fileName)
	targetUnlock := s.lockUploadTarget(destPath)
	defer targetUnlock()
	if err = s.commitUploadedFile(ctx, tempPath, destPath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	return &UploadedFile{Name: fileName, Path: destPath, Size: fileHeader.Size}, nil
}

func (s *service) UploadChunk(fileHeader *multipart.FileHeader, meta UploadMeta, chunkIndex int) (*UploadedFile, bool, error) {
	s.cleanupExpiredUploadSessions()
	if err := s.validateUploadMeta(meta); err != nil {
		return nil, false, err
	}
	if chunkIndex < 0 || chunkIndex >= meta.TotalChunks {
		return nil, false, fmt.Errorf("chunk_index参数不合法")
	}
	unlock := s.lockUploadSession(meta.UploadID)
	defer unlock()
	if err := s.verifyUploadSession(meta); err != nil {
		return nil, false, err
	}
	s.touchUploadSession(meta.UploadID)
	targetUnlock := s.lockUploadTarget(filepath.Join(meta.Path, meta.FileName))
	completed, found, completedErr := s.readCompletedUpload(meta)
	targetUnlock()
	if completedErr != nil {
		return nil, false, completedErr
	}
	if found {
		return completed, false, nil
	}
	exists, err := s.saveUploadChunk(fileHeader, meta, chunkIndex)
	if err != nil {
		return nil, false, fmt.Errorf("保存分片失败: %w", err)
	}
	return nil, exists, nil
}

func (s *service) CheckUpload(meta UploadMeta) ([]int, *UploadedFile, error) {
	s.cleanupExpiredUploadSessions()
	if err := s.validateUploadMeta(meta); err != nil {
		return nil, nil, err
	}
	unlock := s.lockUploadSession(meta.UploadID)
	defer unlock()
	if err := s.ensureUploadSession(meta); err != nil {
		return nil, nil, err
	}
	s.touchUploadSession(meta.UploadID)
	targetUnlock := s.lockUploadTarget(filepath.Join(meta.Path, meta.FileName))
	completed, found, completedErr := s.readCompletedUpload(meta)
	targetUnlock()
	if completedErr != nil {
		return nil, nil, completedErr
	}
	if found {
		return s.allUploadChunkIndexes(meta.TotalChunks), completed, nil
	}
	uploadedChunks, err := s.uploadedChunkIndexes(meta)
	if err != nil {
		return nil, nil, fmt.Errorf("获取分片状态失败: %w", err)
	}
	return uploadedChunks, nil, nil
}

func (s *service) MergeUpload(ctx context.Context, meta UploadMeta) (*UploadedFile, bool, error) {
	s.cleanupExpiredUploadSessions()
	if err := s.validateUploadMeta(meta); err != nil {
		return nil, false, err
	}
	unlock := s.lockUploadSession(meta.UploadID)
	defer unlock()
	if err := s.verifyUploadSession(meta); err != nil {
		return nil, false, err
	}
	s.touchUploadSession(meta.UploadID)
	targetUnlock := s.lockUploadTarget(filepath.Join(meta.Path, meta.FileName))
	defer targetUnlock()
	completed, found, completedErr := s.readCompletedUpload(meta)
	if completedErr != nil {
		return nil, false, completedErr
	}
	if found {
		return completed, true, nil
	}
	mergedPath, size, err := s.mergeUploadChunks(ctx, meta)
	if err != nil {
		return nil, false, err
	}
	destPath := filepath.Join(meta.Path, meta.FileName)
	if err = s.commitUploadedFile(ctx, mergedPath, destPath); err != nil {
		return nil, false, fmt.Errorf("保存合并文件失败: %w", err)
	}
	completed = &UploadedFile{Name: meta.FileName, Path: destPath, Size: size, FileHash: meta.FileHash}
	if err = s.writeCompletedUpload(meta.UploadID, completed); err != nil {
		return nil, false, fmt.Errorf("记录上传结果失败: %w", err)
	}
	_ = os.RemoveAll(s.uploadChunksPath(meta.UploadID))
	_ = os.Remove(mergedPath)
	return completed, false, nil
}

func (s *service) ClearUpload(uploadID string) error {
	s.cleanupExpiredUploadSessions()
	if uploadID == "" || len(uploadID) > 128 || uploadID == "." || uploadID == ".." || strings.ContainsAny(uploadID, `/\`) {
		return fmt.Errorf("upload_id参数不合法")
	}
	unlock := s.lockUploadSession(uploadID)
	defer unlock()
	if err := os.RemoveAll(s.uploadSessionPath(uploadID)); err != nil {
		return fmt.Errorf("清理上传缓存失败: %w", err)
	}
	return nil
}

func (s *service) cleanupExpiredUploadSessions() {
	s.upload.Lock()
	if time.Since(s.upload.lastCleanup) < time.Hour {
		s.upload.Unlock()
		return
	}
	s.upload.lastCleanup = time.Now()
	s.upload.Unlock()

	entries, err := os.ReadDir(s.uploadRootPath())
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		unlock := s.lockUploadSession(entry.Name())
		info, infoErr = os.Stat(s.uploadSessionPath(entry.Name()))
		if infoErr == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(s.uploadSessionPath(entry.Name()))
		}
		unlock()
	}
}

func (s *service) lockUploadSession(uploadID string) func() {
	s.upload.Lock()
	entry := s.upload.locks[uploadID]
	if entry == nil {
		entry = &uploadSessionLock{}
		s.upload.locks[uploadID] = entry
	}
	entry.refs++
	s.upload.Unlock()

	entry.mutex.Lock()
	return func() {
		entry.mutex.Unlock()
		s.upload.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(s.upload.locks, uploadID)
		}
		s.upload.Unlock()
	}
}

func (s *service) lockUploadTarget(path string) func() {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		absolutePath = filepath.Clean(path)
	}
	absolutePath = filepath.Join(s.resolveUploadLockDirectory(filepath.Dir(absolutePath)), filepath.Base(absolutePath))
	if runtime.GOOS == "windows" {
		absolutePath = strings.ToLower(absolutePath)
	}
	return s.lockUploadSession("target\x00" + absolutePath)
}

func (s *service) resolveUploadLockDirectory(path string) string {
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

func (s *service) touchUploadSession(uploadID string) {
	now := time.Now()
	_ = os.Chtimes(s.uploadSessionPath(uploadID), now, now)
}

func (s *service) validateUploadMeta(meta UploadMeta) error {
	if meta.UploadID == "" || len(meta.UploadID) > 128 || meta.UploadID == "." || meta.UploadID == ".." || strings.ContainsAny(meta.UploadID, `/\`) {
		return fmt.Errorf("upload_id参数不合法")
	}
	if err := s.validateUploadFileName(meta.FileName); err != nil {
		return err
	}
	if meta.TotalSize < 0 {
		return fmt.Errorf("total_size参数不合法")
	}
	if meta.ChunkSize <= 0 || meta.ChunkSize > 64<<20 {
		return fmt.Errorf("chunk_size参数不合法")
	}
	if meta.TotalChunks <= 0 || meta.TotalChunks > 1_000_000 {
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

func (s *service) matchesUploadMeta(meta, other UploadMeta) bool {
	return meta.UploadID == other.UploadID &&
		meta.Path == other.Path &&
		meta.FileName == other.FileName &&
		meta.TotalSize == other.TotalSize &&
		meta.ChunkSize == other.ChunkSize &&
		meta.TotalChunks == other.TotalChunks &&
		meta.FileHash == other.FileHash
}

func (s *service) validateUploadFileName(fileName string) error {
	if fileName == "" || fileName == "." || fileName == ".." || strings.ContainsRune(fileName, 0) || strings.ContainsAny(fileName, `/\`) || filepath.Base(fileName) != fileName {
		return fmt.Errorf("file_name参数不合法")
	}
	return nil
}

func (s *service) uploadSessionPath(uploadID string) string {
	return filepath.Join(s.uploadRootPath(), uploadID)
}

func (s *service) uploadMetadataPath(uploadID string) string {
	return filepath.Join(s.uploadSessionPath(uploadID), "meta.json")
}

func (s *service) completedUploadPath(uploadID string) string {
	return filepath.Join(s.uploadSessionPath(uploadID), "completed.json")
}

func (s *service) uploadChunksPath(uploadID string) string {
	return filepath.Join(s.uploadSessionPath(uploadID), "chunks")
}

func (s *service) uploadChunkPath(uploadID string, chunkIndex int) string {
	return filepath.Join(s.uploadChunksPath(uploadID), fmt.Sprintf("%d.part", chunkIndex))
}

func (s *service) ensureUploadSession(meta UploadMeta) error {
	stored, err := s.readUploadMeta(meta.UploadID)
	if err == nil {
		if !s.matchesUploadMeta(stored, meta) {
			return fmt.Errorf("上传任务信息不一致")
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("读取上传任务失败: %w", err)
	}

	sessionPath := s.uploadSessionPath(meta.UploadID)
	if err = os.RemoveAll(sessionPath); err != nil {
		return err
	}
	if err = os.MkdirAll(s.uploadChunksPath(meta.UploadID), 0755); err != nil {
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
	return os.Rename(tempPath, s.uploadMetadataPath(meta.UploadID))
}

func (s *service) verifyUploadSession(meta UploadMeta) error {
	stored, err := s.readUploadMeta(meta.UploadID)
	if os.IsNotExist(err) {
		return fmt.Errorf("上传任务不存在，请重新检查")
	}
	if err != nil {
		return fmt.Errorf("读取上传任务失败: %w", err)
	}
	if !s.matchesUploadMeta(stored, meta) {
		return fmt.Errorf("上传任务信息不一致")
	}
	return nil
}

func (s *service) readUploadMeta(uploadID string) (UploadMeta, error) {
	data, err := os.ReadFile(s.uploadMetadataPath(uploadID))
	if err != nil {
		return UploadMeta{}, err
	}
	var meta UploadMeta
	if err = json.Unmarshal(data, &meta); err != nil {
		return UploadMeta{}, err
	}
	if err = s.validateUploadMeta(meta); err != nil {
		return UploadMeta{}, err
	}
	return meta, nil
}

func (s *service) expectedUploadChunkSize(meta UploadMeta, chunkIndex int) int64 {
	if chunkIndex < meta.TotalChunks-1 {
		return meta.ChunkSize
	}
	return meta.TotalSize - int64(meta.TotalChunks-1)*meta.ChunkSize
}

func (s *service) uploadedChunkIndexes(meta UploadMeta) ([]int, error) {
	uploaded := make([]int, 0, meta.TotalChunks)
	for index := 0; index < meta.TotalChunks; index++ {
		chunkPath := s.uploadChunkPath(meta.UploadID, index)
		info, err := os.Stat(chunkPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() != s.expectedUploadChunkSize(meta, index) {
			_ = os.Remove(chunkPath)
			continue
		}
		uploaded = append(uploaded, index)
	}
	return uploaded, nil
}

func (s *service) allUploadChunkIndexes(totalChunks int) []int {
	chunks := make([]int, totalChunks)
	for index := range chunks {
		chunks[index] = index
	}
	return chunks
}

func (s *service) readCompletedUpload(meta UploadMeta) (*UploadedFile, bool, error) {
	data, err := os.ReadFile(s.completedUploadPath(meta.UploadID))
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
	fileHash, err := s.hashUploadFile(completed.Path)
	if err != nil {
		return nil, false, fmt.Errorf("校验已上传文件失败: %w", err)
	}
	if fileHash != completed.FileHash {
		return nil, false, fmt.Errorf("已上传文件校验失败")
	}
	return &completed, true, nil
}

func (s *service) writeCompletedUpload(uploadID string, completed *UploadedFile) error {
	data, err := json.Marshal(completed)
	if err != nil {
		return err
	}
	sessionPath := s.uploadSessionPath(uploadID)
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
	return os.Rename(tempPath, s.completedUploadPath(uploadID))
}

func (s *service) saveUploadChunk(fileHeader *multipart.FileHeader, meta UploadMeta, chunkIndex int) (bool, error) {
	chunkPath := s.uploadChunkPath(meta.UploadID, chunkIndex)
	expectedSize := s.expectedUploadChunkSize(meta, chunkIndex)
	if info, err := os.Stat(chunkPath); err == nil && info.Mode().IsRegular() && info.Size() == expectedSize {
		return true, nil
	}
	_ = os.Remove(chunkPath)
	if fileHeader.Size != expectedSize {
		return false, fmt.Errorf("分片大小不一致: %d/%d", fileHeader.Size, expectedSize)
	}
	if err := s.saveMultipartFile(fileHeader, chunkPath, expectedSize); err != nil {
		return false, err
	}
	return false, nil
}

func (s *service) saveMultipartFile(fileHeader *multipart.FileHeader, destPath string, expectedSize int64) error {
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

func (s *service) mergeUploadChunks(ctx context.Context, meta UploadMeta) (string, int64, error) {
	sessionPath := s.uploadSessionPath(meta.UploadID)
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
		chunkPath := s.uploadChunkPath(meta.UploadID, index)
		info, statErr := os.Stat(chunkPath)
		if statErr != nil {
			_ = merged.Close()
			if os.IsNotExist(statErr) {
				return "", 0, fmt.Errorf("分片 %d 不存在或不完整", index)
			}
			return "", 0, fmt.Errorf("读取分片 %d 失败: %w", index, statErr)
		}
		if !info.Mode().IsRegular() || info.Size() != s.expectedUploadChunkSize(meta, index) {
			_ = merged.Close()
			return "", 0, fmt.Errorf("分片 %d 不存在或不完整", index)
		}
		chunk, openErr := os.Open(chunkPath)
		if openErr != nil {
			_ = merged.Close()
			return "", 0, fmt.Errorf("打开分片 %d 失败: %w", index, openErr)
		}
		written, copyErr := s.copyUploadData(ctx, io.MultiWriter(merged, fileHash), chunk)
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

func (s *service) commitUploadedFile(ctx context.Context, sourcePath, destPath string) error {
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
	if err := s.replaceUploadedFile(sourcePath, destPath); err == nil {
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
	if _, err = s.copyUploadData(ctx, temp, source); err == nil {
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
	if err = s.replaceUploadedFile(tempPath, destPath); err != nil {
		return err
	}
	_ = os.Remove(sourcePath)
	return nil
}

func (s *service) hashUploadFile(path string) (string, error) {
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

func (s *service) replaceUploadedFile(sourcePath, destPath string) error {
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

func (s *service) copyUploadData(ctx context.Context, dest io.Writer, source io.Reader) (int64, error) {
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

func (s *service) uploadRootPath() string {
	return filepath.Clean(constant.UploadPath)
}
