package fileonix

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWatcher_DetectsChange(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileonix_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.env")
	if err := os.WriteFile(testFile, []byte("KEY=VAL1\n"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	changed := make(chan string, 1)
	stopCh := make(chan struct{})
	defer close(stopCh)

	WatchFile(testFile, 50*time.Millisecond, stopCh, func(p string) {
		changed <- p
	})

	time.Sleep(100 * time.Millisecond)

	// Update file
	if err := os.WriteFile(testFile, []byte("KEY=VAL2\n"), 0600); err != nil {
		t.Fatalf("failed to update test file: %v", err)
	}

	select {
	case p := <-changed:
		if p != testFile {
			t.Errorf("expected %s, got %s", testFile, p)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for FileOnix to detect file change")
	}
}
