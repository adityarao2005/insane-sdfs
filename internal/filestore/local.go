package filestore

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ErrInvalidPath = errors.New("invalid file path")
var ErrFileNotFound = errors.New("file not found")

type Entry struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

type LocalStore struct {
	root string
}

func NewLocalStore(root string) *LocalStore {
	if root == "" {
		root = "./data"
	}
	return &LocalStore{root: root}
}

func (s *LocalStore) Save(deviceID, relativePath string, r io.Reader) error {
	safePath, err := sanitizePath(relativePath)
	if err != nil {
		return err
	}

	full := filepath.Join(s.root, deviceID, filepath.FromSlash(safePath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}

	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func (s *LocalStore) Load(deviceID, relativePath string) ([]byte, error) {
	safePath, err := sanitizePath(relativePath)
	if err != nil {
		return nil, err
	}

	full := filepath.Join(s.root, deviceID, filepath.FromSlash(safePath))
	b, err := os.ReadFile(full)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *LocalStore) List(deviceID, prefix string) ([]Entry, error) {
	deviceRoot := filepath.Join(s.root, deviceID)
	safePrefix := ""
	if strings.TrimSpace(prefix) != "" {
		p, err := sanitizePath(prefix)
		if err != nil {
			return nil, err
		}
		safePrefix = p
	}

	entries := make([]Entry, 0)
	err := filepath.WalkDir(deviceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(deviceRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if safePrefix != "" && !strings.HasPrefix(rel, safePrefix) {
			return nil
		}
		st, err := d.Info()
		if err != nil {
			return err
		}
		entries = append(entries, Entry{Path: rel, Size: st.Size(), ModTime: st.ModTime().UTC()})
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Entry{}, nil
		}
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

func sanitizePath(p string) (string, error) {
	trimmed := strings.TrimSpace(p)
	if trimmed == "" {
		return "", ErrInvalidPath
	}
	for _, part := range strings.FieldsFunc(trimmed, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return "", ErrInvalidPath
		}
	}
	clean := filepath.ToSlash(filepath.Clean("/" + trimmed))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || clean == "" || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", ErrInvalidPath
	}
	return clean, nil
}
