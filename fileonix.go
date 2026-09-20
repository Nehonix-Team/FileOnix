package fileonix

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// FileWatcher provides high-precision file monitoring using FileOnix detection algorithms.
type FileWatcher struct {
	path      string
	debounce  time.Duration
	lastHash  string
	lastMtime time.Time
	mu        sync.Mutex
}

// NewFileWatcher initializes a watcher for a specific target file.
func NewFileWatcher(filePath string, debounce time.Duration) *FileWatcher {
	if debounce <= 0 {
		debounce = 150 * time.Millisecond
	}
	fw := &FileWatcher{
		path:     filePath,
		debounce: debounce,
	}
	fw.recordInitialState()
	return fw
}

// CheckModified checks whether the target file has changed since the last check.
func (fw *FileWatcher) CheckModified() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	info, err := os.Stat(fw.path)
	if err != nil {
		return false
	}

	// Check modification time
	if !info.ModTime().After(fw.lastMtime) {
		return false
	}

	// Content hash check for files under 10MB
	if info.Size() <= 10*1024*1024 {
		newHash, err := hashFile(fw.path)
		if err == nil {
			if newHash != fw.lastHash {
				fw.lastHash = newHash
				fw.lastMtime = info.ModTime()
				return true
			}
			fw.lastMtime = info.ModTime()
			return false
		}
	}

	fw.lastMtime = info.ModTime()
	return true
}

func (fw *FileWatcher) recordInitialState() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	info, err := os.Stat(fw.path)
	if err == nil {
		fw.lastMtime = info.ModTime()
		if info.Size() <= 10*1024*1024 {
			if h, err := hashFile(fw.path); err == nil {
				fw.lastHash = h
			}
		}
	}
}

// Watch runs the monitoring loop until stopCh is closed, calling onChange on modification.
func (fw *FileWatcher) Watch(stopCh <-chan struct{}, onChange func(path string)) {
	ticker := time.NewTicker(fw.debounce)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if fw.CheckModified() {
				onChange(fw.path)
			}
		}
	}
}

// WatchFile starts background monitoring of a file with FileOnix.
func WatchFile(filePath string, debounce time.Duration, stopCh <-chan struct{}, onChange func(path string)) *FileWatcher {
	fw := NewFileWatcher(filePath, debounce)
	go fw.Watch(stopCh, onChange)
	return fw
}

func hashFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
