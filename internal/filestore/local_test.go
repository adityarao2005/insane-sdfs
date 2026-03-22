package filestore

import (
	"errors"
	"strings"
	"testing"
)

func TestLocalStoreSaveLoadAndList(t *testing.T) {
	tmp := t.TempDir()
	store := NewLocalStore(tmp)

	if err := store.Save("dev-1", "photos/a.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store.Save("dev-1", "photos/b.txt", strings.NewReader("world")); err != nil {
		t.Fatalf("save2: %v", err)
	}

	b, err := store.Load("dev-1", "photos/a.txt")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if string(b) != "hello" {
		t.Fatalf("unexpected content: %q", string(b))
	}

	entries, err := store.List("dev-1", "photos")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestLocalStoreRejectsTraversal(t *testing.T) {
	tmp := t.TempDir()
	store := NewLocalStore(tmp)

	err := store.Save("dev-1", "../../etc/passwd", strings.NewReader("x"))
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got %v", err)
	}
}
