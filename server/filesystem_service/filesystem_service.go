package filesystem_service

import (
	"bytes"
)

type FileInfo struct {
	Size int64
	Path string
	LastModified int64
	IsDirectory bool
}

type IFileSystemService interface {

	UploadFile(path string, data chan bytes.Buffer) error

	DownloadFile(path string) (chan bytes.Buffer, error)

	DeleteFile(path string) error

	GetFileInfo(path string) (FileInfo, error)

	ListFiles(directoryPath string) ([]FileInfo, error)

	CreateDirectory(directoryPath string) error
}

type IFileSystemServiceServer interface {

	AddService(service IFileSystemService) error

	Start(port int) error

	Stop() error
}