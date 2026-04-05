package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"server/filesystem_service"
)

func TestUploadAndDownloadFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service, err := filesystem_service.NewFileSystemService(rootDir)
	if err != nil {
		t.Fatalf("create file system service: %v", err)
	}

	targetPath := filepath.Join(rootDir, "nested", "payload.txt")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("create target directory: %v", err)
	}

	input := make(chan bytes.Buffer)
	go func() {
		input <- *bytes.NewBufferString("hello ")
		input <- *bytes.NewBufferString("world")
		close(input)
	}()

	if err := service.UploadFile(targetPath, input); err != nil {
		t.Fatalf("upload file: %v", err)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(content) != "hello world" {
		t.Fatalf("unexpected uploaded content: %q", string(content))
	}

	output, err := service.DownloadFile(targetPath)
	if err != nil {
		t.Fatalf("download file: %v", err)
	}

	var downloaded bytes.Buffer
	for chunk := range output {
		downloaded.Write(chunk.Bytes())
	}

	if downloaded.String() != "hello world" {
		t.Fatalf("unexpected downloaded content: %q", downloaded.String())
	}
}

func TestGetFileInfoAndListFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service, err := filesystem_service.NewFileSystemService(rootDir)
	if err != nil {
		t.Fatalf("create file system service: %v", err)
	}

	filePath := filepath.Join(rootDir, "alpha.txt")
	dirPath := filepath.Join(rootDir, "nested")

	if err := os.WriteFile(filePath, []byte("abc"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}

	info, err := service.GetFileInfo(filePath)
	if err != nil {
		t.Fatalf("get file info: %v", err)
	}
	if info.Path != filePath || info.Size != 3 || info.IsDirectory {
		t.Fatalf("unexpected file info: %+v", info)
	}

	dirInfo, err := service.GetFileInfo(dirPath)
	if err != nil {
		t.Fatalf("get directory info: %v", err)
	}
	if dirInfo.Path != dirPath || !dirInfo.IsDirectory {
		t.Fatalf("unexpected directory info: %+v", dirInfo)
	}

	listed, err := service.ListFiles(rootDir)
	if err != nil {
		t.Fatalf("list files: %v", err)
	}

	if len(listed) != 2 {
		t.Fatalf("unexpected file count: got %d want %d", len(listed), 2)
	}
	if listed[0].Path != filePath || listed[1].Path != dirPath {
		t.Fatalf("unexpected listing order or paths: %+v", listed)
	}
	if listed[0].Size != 3 || listed[0].IsDirectory {
		t.Fatalf("unexpected first entry: %+v", listed[0])
	}
	if !listed[1].IsDirectory {
		t.Fatalf("unexpected second entry: %+v", listed[1])
	}
}

func TestCreateAndDeleteDirectory(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	service, err := filesystem_service.NewFileSystemService(rootDir)
	if err != nil {
		t.Fatalf("create file system service: %v", err)
	}

	targetDir := filepath.Join(rootDir, "a", "b", "c")

	if err := service.CreateDirectory(targetDir); err != nil {
		t.Fatalf("create directory: %v", err)
	}
	if _, err := os.Stat(targetDir); err != nil {
		t.Fatalf("stat created directory: %v", err)
	}

	filePath := filepath.Join(targetDir, "delete-me.txt")
	if err := os.WriteFile(filePath, []byte("remove me"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := service.DeleteFile(targetDir); err != nil {
		t.Fatalf("delete file: %v", err)
	}
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Fatalf("expected directory to be deleted, got err=%v", err)
	}
}
