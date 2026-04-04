package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nehonix/fileonix/internal/config"
	"github.com/nehonix/fileonix/internal/ui"
	"github.com/nehonix/fileonix/internal/watcher"
)

const VERSION = "2.0.0"

func main() {
	// Parse CLI args
	args := os.Args[1:]

	// Handle special commands first
	if len(args) == 0 {
		ui.ShowHelp(VERSION)
		os.Exit(0)
	}

	switch args[0] {
	case "--version", "-v":
		ui.PrintBanner(VERSION)
		os.Exit(0)
	case "--help", "-h":
		ui.ShowHelp(VERSION)
		os.Exit(0)
	case "init":
		ui.PrintBanner(VERSION)
		config.InitConfig()
		os.Exit(0)
	}

	// Boot sequence
	ui.PrintBanner(VERSION)
	time.Sleep(80 * time.Millisecond)

	// Load configuration
	cfg, err := config.Load(args)
	if err != nil {
		ui.Fatal("Configuration error", err.Error())
	}

	// Validate
	if cfg.Script == "" {
		ui.Fatal("No script specified", "Use -script <file> or define 'script' in fileonix.config.json")
	}

	// Display loaded config
	ui.PrintConfig(cfg)

	// Start watcher engine
	w, err := watcher.New(cfg)
	if err != nil {
		ui.Fatal("Failed to initialize watcher", err.Error())
	}

	fmt.Println()
	ui.PrintSection("ENGINE READY", fmt.Sprintf("Watching %d path(s) | Runtime: %s", len(cfg.Watch), cfg.TypescriptRunner))
	fmt.Println()

	// Block on watcher
	if err := w.Start(); err != nil {
		ui.Fatal("Watcher crashed", err.Error())
	}
}
