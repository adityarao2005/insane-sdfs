package filesystem_service

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Size         int64
	Path         string
	LastModified int64
	IsDirectory  bool
}

type FileSystemService struct {
}

func (fss *FileSystemService) UploadFile(path string, data chan bytes.Buffer) error {

	temp, err := os.CreateTemp(filepath.Dir(path), "upload-*")

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

	if err := os.Rename(temp.Name(), path); err != nil {
		return err
	}

	return nil
}

func (fss *FileSystemService) DownloadFile(path string) (chan bytes.Buffer, error) {

	file, err := os.Open(path)
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
	return os.RemoveAll(path)
}

func (fss *FileSystemService) GetFileInfo(path string) (FileInfo, error) {

	info, err := os.Stat(path)

	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Size:         info.Size(),
		Path:         path,
		LastModified: info.ModTime().Unix(),
		IsDirectory:  info.IsDir(),
	}, nil
}

func (fss *FileSystemService) ListFiles(directoryPath string) ([]FileInfo, error) {

	entries, err := os.ReadDir(directoryPath)
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
			Path:         filepath.Join(directoryPath, info.Name()),
			LastModified: info.ModTime().Unix(),
			IsDirectory:  info.IsDir(),
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

func (fss *FileSystemService) CreateDirectory(directoryPath string) error {
	return os.MkdirAll(directoryPath, 0755)
}

type IFileSystemServiceServer interface {
	AddService(service FileSystemService) error

	Start(port int) error

	Stop() error
}
