package filesystem_service

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Size         int64
	Path         string
	LastModified int64
	IsDirectory  bool
}

type FileSystemService struct {
	basePath string
}

func NewFileSystemService(basePath string) (*FileSystemService, error) {
	if basePath == "" {
		basePath = "."
	}

	absBasePath, err := filepath.Abs(filepath.Clean(basePath))
	if err != nil {
		return nil, fmt.Errorf("resolve base path: %w", err)
	}

	return &FileSystemService{basePath: absBasePath}, nil
}

func (fss *FileSystemService) sanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	basePath := fss.basePath
	if basePath == "" {
		basePath = "."
	}

	absBasePath, err := filepath.Abs(filepath.Clean(basePath))
	if err != nil {
		return "", fmt.Errorf("resolve base path: %w", err)
	}

	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		cleanPath = filepath.Join(absBasePath, cleanPath)
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	relPath, err := filepath.Rel(absBasePath, absPath)
	if err != nil {
		return "", fmt.Errorf("compute relative path: %w", err)
	}

	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes base path %q", path, absBasePath)
	}

	return absPath, nil
}

func (fss *FileSystemService) UploadFile(path string, data chan bytes.Buffer) error {
	sanitizedPath, err := fss.sanitizePath(path)
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp(filepath.Dir(sanitizedPath), "upload-*")

	if err != nil {
		return err
	}

	defer os.Remove(temp.Name())

	for chunk := range data {
		if _, err := temp.Write(chunk.Bytes()); err != nil {
			return err
		}
	}

	if err := temp.Close(); err != nil {
		return err
	}

	if err := os.Rename(temp.Name(), sanitizedPath); err != nil {
		return err
	}

	return nil
}

func (fss *FileSystemService) DownloadFile(path string) (chan bytes.Buffer, error) {
	sanitizedPath, err := fss.sanitizePath(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(sanitizedPath)
	if err != nil {
		return nil, err
	}

	dataChannel := make(chan bytes.Buffer)

	go func() {
		defer file.Close()
		defer close(dataChannel)

		for {
			buffer := make([]byte, 1024*1024) // 1MB buffer
			n, err := file.Read(buffer)
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			chunk := bytes.NewBuffer(buffer[:n])
			dataChannel <- *chunk
		}
	}()

	return dataChannel, nil
}

func (fss *FileSystemService) DeleteFile(path string) error {
	sanitizedPath, err := fss.sanitizePath(path)
	if err != nil {
		return err
	}

	return os.RemoveAll(sanitizedPath)
}

func (fss *FileSystemService) GetFileInfo(path string) (FileInfo, error) {
	sanitizedPath, err := fss.sanitizePath(path)
	if err != nil {
		return FileInfo{}, err
	}

	info, err := os.Stat(sanitizedPath)

	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Size:         info.Size(),
		Path:         sanitizedPath,
		LastModified: info.ModTime().Unix(),
		IsDirectory:  info.IsDir(),
	}, nil
}

func (fss *FileSystemService) ListFiles(directoryPath string) ([]FileInfo, error) {
	sanitizedDirectoryPath, err := fss.sanitizePath(directoryPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(sanitizedDirectoryPath)
	if err != nil {
		return nil, err
	}

	var fileInfos []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		fileInfo := FileInfo{
			Size:         info.Size(),
			Path:         filepath.Join(sanitizedDirectoryPath, info.Name()),
			LastModified: info.ModTime().Unix(),
			IsDirectory:  info.IsDir(),
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

func (fss *FileSystemService) CreateDirectory(directoryPath string) error {
	sanitizedDirectoryPath, err := fss.sanitizePath(directoryPath)
	if err != nil {
		return err
	}

	return os.MkdirAll(sanitizedDirectoryPath, 0755)
}

type IFileSystemServiceServer interface {
	AddService(service *FileSystemService) error

	Start(port string) error

	Stop() error
}
