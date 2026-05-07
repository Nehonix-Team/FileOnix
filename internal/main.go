package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Nehonix-Team/FileOnix/internal/config"
	"github.com/Nehonix-Team/FileOnix/internal/ui"
	"github.com/Nehonix-Team/FileOnix/internal/watcher"
)

const VERSION = "2.0.6"

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
	case "--help", "-h", "help":
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

	// Validate - No script is now optional (watch-only mode)
	if cfg.Script != "" {
		if _, err := os.Stat(cfg.Script); os.IsNotExist(err) {
			ui.Warn(fmt.Sprintf("Script file not found: %s", cfg.Script))
		}
	}

	// Validate watch directories
	existingWatch := make([]string, 0)
	for _, dir := range cfg.Watch {
		if _, err := os.Stat(dir); err == nil {
			existingWatch = append(existingWatch, dir)
		} else {
			ui.Warn(fmt.Sprintf("Watch directory not found: %s", dir))
		}
	}

	if len(existingWatch) == 0 {
		ui.Fatal("No valid watch directories", "None of the specified watch paths exist or are accessible")
	}
	cfg.Watch = existingWatch

	// Display loaded config
	ui.PrintConfig(cfg)

	// Start watcher engine
	w, err := watcher.New(cfg)
	if err != nil {
		ui.Fatal("Failed to initialize watcher", err.Error())
	}

	fmt.Println()
	ui.PrintSection("FileOnix", fmt.Sprintf("Watching %d path(s) | Runner: %s", len(cfg.Watch), cfg.Runner))
	fmt.Println()

	// Block on watcher
	if err := w.Start(); err != nil {
		ui.Fatal("Watcher crashed", err.Error())
	}
}
