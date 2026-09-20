package fileonix

import (
	"path/filepath"
	"time"

	"github.com/Nehonix-Team/FileOnix/internal/config"
	"github.com/Nehonix-Team/FileOnix/internal/watcher"
)

// Watcher alias representing the underlying FileOnix watcher engine.
type Watcher = watcher.Watcher

// WatchFile monitors a specific file using FileOnix's internal watcher engine.
// It delegates directly to FileOnix internal configuration and hash-based change detection.
func WatchFile(filePath string, debounce time.Duration, stopCh <-chan struct{}, onChange func(path string)) (*Watcher, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	cfg := config.DefaultConfig()
	cfg.TargetFile = absPath
	cfg.Watch = []string{filepath.Dir(absPath)}
	cfg.Script = "" // watch-only mode
	cfg.Silent = true
	cfg.NoSignals = true
	cfg.UseFileHash = true
	if debounce > 0 {
		cfg.Debounce = debounce
		cfg.DebounceMs = int(debounce.Milliseconds())
	}
	cfg.OnChange = func(p, eventType string) {
		onChange(p)
	}

	w, err := watcher.New(cfg)
	if err != nil {
		return nil, err
	}

	go w.Start()

	if stopCh != nil {
		go func() {
			<-stopCh
			w.Stop()
		}()
	}

	return w, nil
}
