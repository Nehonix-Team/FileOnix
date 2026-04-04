package watcher

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nehonix/fileonix/internal/config"
	"github.com/nehonix/fileonix/internal/ui"
)

// Watcher is the core file watching and process management engine
type Watcher struct {
	cfg          *config.Config
	process      *exec.Cmd
	processMu    sync.Mutex
	restartCount int
	startedAt    time.Time
	fileHashes   map[string]string
	hashesMu     sync.RWMutex
	events       chan FileEvent
	done         chan struct{}
	restart      chan string
}

type FileEvent struct {
	Path      string
	EventType string
	Hash      string
}

// New creates a new Watcher instance
func New(cfg *config.Config) (*Watcher, error) {
	return &Watcher{
		cfg:        cfg,
		fileHashes: make(map[string]string),
		events:     make(chan FileEvent, 64),
		done:       make(chan struct{}),
		restart:    make(chan string, 1),
	}, nil
}

// Start begins watching and running the process
func (w *Watcher) Start() error {
	// Handle OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Initial file hash scan
	if w.cfg.UseFileHash {
		w.scanHashes()
	}

	// Start the process
	w.spawnProcess()

	// Start filesystem polling
	go w.pollFilesystem()

	// Event processing loop
	go w.processEvents()

	// Wait for signal or done
	select {
	case sig := <-sigCh:
		ui.Log(ui.EventInfo, fmt.Sprintf("Received %s — shutting down gracefully", sig))
		w.shutdown()
	case <-w.done:
	}

	return nil
}

// ─── PROCESS MANAGEMENT ──────────────────────────────────────────────────────

func (w *Watcher) spawnProcess() {
	w.processMu.Lock()
	defer w.processMu.Unlock()

	if w.cfg.ClearScreen {
		clearScreen()
	}

	runner, runnerArgs := w.buildCommand()
	ui.PrintProcessStart(runner, w.cfg.Script)

	cmd := exec.Command(runner, runnerArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = w.buildEnv()

	w.startedAt = time.Now()

	if err := cmd.Start(); err != nil {
		ui.Log(ui.EventError, "Failed to start process", err.Error())
		return
	}

	w.process = cmd

	// Monitor process in background
	go func() {
		err := cmd.Wait()
		duration := time.Since(w.startedAt)

		w.processMu.Lock()
		// Only report if this is still the current process
		if w.process == cmd {
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}
			ui.PrintProcessStop(exitCode, duration)
			w.process = nil
		}
		w.processMu.Unlock()
	}()
}

func (w *Watcher) killProcess() {
	w.processMu.Lock()
	defer w.processMu.Unlock()

	if w.process == nil || w.process.Process == nil {
		return
	}

	// Try graceful SIGTERM first
	w.process.Process.Signal(syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		w.process.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Graceful exit
	case <-time.After(w.cfg.Graceful):
		// Force kill
		w.process.Process.Kill()
	}
	w.process = nil
}

func (w *Watcher) triggerRestart(reason string) {
	// Debounce: non-blocking send
	select {
	case w.restart <- reason:
	default:
	}
}

func (w *Watcher) processEvents() {
	var batchTimer *time.Timer
	var pendingReasons []string

	flush := func() {
		if len(pendingReasons) == 0 {
			return
		}
		reason := pendingReasons[len(pendingReasons)-1]
		if len(pendingReasons) > 1 {
			reason = fmt.Sprintf("%d files changed", len(pendingReasons))
		}
		pendingReasons = nil

		w.restartCount++
		ui.PrintRestartDivider(w.restartCount)
		w.killProcess()
		w.spawnProcess()
		_ = reason
	}

	for {
		select {
		case reason := <-w.restart:
			if w.cfg.BatchMode {
				pendingReasons = append(pendingReasons, reason)
				if batchTimer != nil {
					batchTimer.Reset(w.cfg.Debounce * 3)
				} else {
					batchTimer = time.AfterFunc(w.cfg.Debounce*3, flush)
				}
			} else {
				w.restartCount++
				ui.PrintRestartDivider(w.restartCount)
				w.killProcess()
				w.spawnProcess()
			}
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) shutdown() {
	ui.Log(ui.EventInfo, "Killing process...")
	w.killProcess()
	fmt.Println()
	ui.Log(ui.EventSuccess, "FileOnix shutdown complete")
	fmt.Println()
	close(w.done)
}

// ─── FILESYSTEM POLLING ───────────────────────────────────────────────────────

func (w *Watcher) pollFilesystem() {
	ticker := time.NewTicker(w.cfg.Debounce)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkForChanges()
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) checkForChanges() {
	for _, watchDir := range w.cfg.Watch {
		if err := filepath.WalkDir(watchDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // skip errors
			}

			// Skip ignored dirs
			if d.IsDir() && w.shouldIgnore(path) {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}

			// Check extension
			if !w.hasWatchedExt(path) {
				return nil
			}

			// Check modification time
			info, err := d.Info()
			if err != nil {
				return nil
			}

			// Hash-based detection (more precise)
			if w.cfg.UseFileHash {
				w.checkFileHash(path, info)
			} else {
				w.checkFileMtime(path, info)
			}

			return nil
		}); err != nil {
			// Walk error - ignore silently
		}
	}
}

func (w *Watcher) checkFileHash(path string, info os.FileInfo) {
	if info.Size() > 10*1024*1024 { // Skip files >10MB
		return
	}

	hash, err := hashFile(path)
	if err != nil {
		return
	}

	w.hashesMu.Lock()
	oldHash, exists := w.fileHashes[path]
	w.fileHashes[path] = hash
	w.hashesMu.Unlock()

	if !exists {
		// New file
		ui.PrintFileChange(path, "created")
		w.triggerRestart(path)
	} else if oldHash != hash {
		// Modified
		ui.PrintFileChange(path, "modified")
		w.triggerRestart(path)
	}
}

var mtimes = make(map[string]time.Time)
var mtimesMu sync.Mutex

func (w *Watcher) checkFileMtime(path string, info os.FileInfo) {
	mtimesMu.Lock()
	old, exists := mtimes[path]
	mtimes[path] = info.ModTime()
	mtimesMu.Unlock()

	if !exists {
		return // First seen
	}
	if info.ModTime().After(old) {
		ui.PrintFileChange(path, "modified")
		w.triggerRestart(path)
	}
}

func (w *Watcher) scanHashes() {
	spinner := ui.NewSpinner("Scanning files...")
	spinner.Start()

	count := 0
	for _, watchDir := range w.cfg.Watch {
		filepath.WalkDir(watchDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				if d != nil && d.IsDir() && w.shouldIgnore(path) {
					return filepath.SkipDir
				}
				return nil
			}
			if !w.hasWatchedExt(path) {
				return nil
			}
			hash, err := hashFile(path)
			if err == nil {
				w.hashesMu.Lock()
				w.fileHashes[path] = hash
				w.hashesMu.Unlock()
				count++
			}
			return nil
		})
	}

	spinner.Stop()
	ui.Log(ui.EventInfo, fmt.Sprintf("Indexed %d files", count))
}

// ─── HELPERS ─────────────────────────────────────────────────────────────────

func (w *Watcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)
	for _, ignore := range w.cfg.Ignore {
		if base == ignore || strings.Contains(path, "/"+ignore+"/") || strings.HasSuffix(path, "/"+ignore) {
			return true
		}
	}
	return false
}

func (w *Watcher) hasWatchedExt(path string) bool {
	ext := filepath.Ext(path)
	for _, e := range w.cfg.Extensions {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

func (w *Watcher) buildCommand() (string, []string) {
	runner := w.cfg.TypescriptRunner
	script := w.cfg.Script

	switch runner {
	case "bun":
		return "bun", append([]string{"run", script}, w.cfg.NodeArgs...)
	case "tsx":
		return "tsx", append([]string{script}, w.cfg.NodeArgs...)
	case "ts-node":
		return "ts-node", append([]string{script}, w.cfg.NodeArgs...)
	default:
		return "node", append([]string{script}, w.cfg.NodeArgs...)
	}
}

func (w *Watcher) buildEnv() []string {
	env := os.Environ()
	env = append(env, "FILEONIX=1")
	env = append(env, fmt.Sprintf("FILEONIX_RESTART=%d", w.restartCount))
	return env
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
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

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
