package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nehonix/fileonix/internal/ui"
)

// Config holds all FileOnix runtime configuration
type Config struct {
	// Core
	Script           string   `json:"script"`
	Watch            []string `json:"watch"`
	Ignore           []string `json:"ignore"`
	Extensions       []string `json:"extensions"`

	// Runtime
	TypescriptRunner string   `json:"typescriptRunner"`
	NodeArgs         []string `json:"nodeArgs"`
	EnvFile          string   `json:"envFile"`

	// Behavior
	ClearScreen  bool          `json:"clearScreen"`
	BatchMode    bool          `json:"batchMode"`
	DebounceMs   int           `json:"debounceMs"`
	MaxRestarts  int           `json:"maxRestarts"`
	UseFileHash  bool          `json:"useFileHash"`
	GracefulMs   int           `json:"gracefulMs"`

	// Internal
	Debounce time.Duration `json:"-"`
	Graceful time.Duration `json:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		Watch:            []string{"."},
		Ignore:           []string{"node_modules", "dist", ".git", ".next", "build", "coverage"},
		Extensions:       []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"},
		TypescriptRunner: "auto",
		ClearScreen:      false,
		BatchMode:        false,
		DebounceMs:       100,
		MaxRestarts:      -1, // unlimited
		UseFileHash:      true,
		GracefulMs:       3000,
	}
}

// Load reads config from CLI args and config files
func Load(args []string) (*Config, error) {
	cfg := DefaultConfig()

	// 1. Try to load from config file
	if err := loadFromFile(cfg); err != nil {
		// Not fatal — config file is optional
		ui.Info(fmt.Sprintf("No config file found, using CLI args"))
	}

	// 2. Override with CLI args (higher priority)
	if err := parseArgs(cfg, args); err != nil {
		return nil, err
	}

	// 3. Auto-detect runtime if needed
	if cfg.TypescriptRunner == "auto" {
		cfg.TypescriptRunner = detectRuntime()
	}

	// 4. Resolve durations
	cfg.Debounce = time.Duration(cfg.DebounceMs) * time.Millisecond
	cfg.Graceful = time.Duration(cfg.GracefulMs) * time.Millisecond

	// 5. If no watch dirs set, infer from script
	if len(cfg.Watch) == 1 && cfg.Watch[0] == "." && cfg.Script != "" {
		dir := filepath.Dir(cfg.Script)
		if dir != "." {
			cfg.Watch = []string{dir}
		}
	}

	// Display config
	displayConfig(cfg)

	return cfg, nil
}

// Display shows loaded config in the terminal
func (c *Config) Display() {
	displayConfig(c)
}

func displayConfig(cfg *Config) {
	fields := []ui.ConfigField{
		{Key: "script",   Value: cfg.Script,                    Highlight: true},
		{Key: "runner",   Value: cfg.TypescriptRunner,          Highlight: true},
		{Key: "watch",    Value: strings.Join(cfg.Watch, ", "), Highlight: false},
		{Key: "ignore",   Value: strings.Join(cfg.Ignore, ", "), Dim: true},
		{Key: "ext",      Value: strings.Join(cfg.Extensions, " "), Dim: false},
		{Key: "debounce", Value: fmt.Sprintf("%dms", cfg.DebounceMs), Dim: false},
		{Key: "hash",     Value: boolLabel(cfg.UseFileHash, "enabled", "disabled"), Dim: !cfg.UseFileHash},
		{Key: "batch",    Value: boolLabel(cfg.BatchMode, "enabled", "disabled"),   Dim: !cfg.BatchMode},
		{Key: "clear",    Value: boolLabel(cfg.ClearScreen, "yes", "no"),           Dim: !cfg.ClearScreen},
	}
	ui.PrintConfigFields(fields)
}

func boolLabel(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

// ─── FILE LOADING ────────────────────────────────────────────────────────────

func loadFromFile(cfg *Config) error {
	candidates := []string{
		"fileonix.config.json",
		".fileonixrc.json",
		".fileonixrc",
	}

	for _, name := range candidates {
		data, err := os.ReadFile(name)
		if err != nil {
			continue
		}

		fileCfg := &Config{}
		if err := json.Unmarshal(data, fileCfg); err != nil {
			return fmt.Errorf("invalid config file %s: %w", name, err)
		}

		// Merge: file values overwrite defaults only if set
		mergeConfig(cfg, fileCfg)
		ui.Success(fmt.Sprintf("Loaded %s", name))
		return nil
	}

	return fmt.Errorf("no config file found")
}

func mergeConfig(dst, src *Config) {
	if src.Script != "" {
		dst.Script = src.Script
	}
	if len(src.Watch) > 0 {
		dst.Watch = src.Watch
	}
	if len(src.Ignore) > 0 {
		dst.Ignore = src.Ignore
	}
	if len(src.Extensions) > 0 {
		dst.Extensions = src.Extensions
	}
	if src.TypescriptRunner != "" {
		dst.TypescriptRunner = src.TypescriptRunner
	}
	if src.DebounceMs > 0 {
		dst.DebounceMs = src.DebounceMs
	}
	if src.GracefulMs > 0 {
		dst.GracefulMs = src.GracefulMs
	}
	if src.MaxRestarts != 0 {
		dst.MaxRestarts = src.MaxRestarts
	}
	if src.ClearScreen {
		dst.ClearScreen = true
	}
	if src.BatchMode {
		dst.BatchMode = true
	}
	if src.EnvFile != "" {
		dst.EnvFile = src.EnvFile
	}
	if len(src.NodeArgs) > 0 {
		dst.NodeArgs = src.NodeArgs
	}
	// UseFileHash: explicit false in JSON should disable it
	// We default to true, so only override if src explicitly disables it
	dst.UseFileHash = src.UseFileHash || dst.UseFileHash
}

// ─── ARG PARSING ─────────────────────────────────────────────────────────────

func parseArgs(cfg *Config, args []string) error {
	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "-script" || arg == "--script":
			i++
			if i >= len(args) {
				return fmt.Errorf("-script requires a value")
			}
			cfg.Script = args[i]

		case strings.HasPrefix(arg, "-script="):
			cfg.Script = strings.TrimPrefix(arg, "-script=")

		case arg == "-watch" || arg == "--watch":
			i++
			if i >= len(args) {
				return fmt.Errorf("-watch requires a value")
			}
			cfg.Watch = splitComma(args[i])

		case strings.HasPrefix(arg, "--watch="):
			cfg.Watch = splitComma(strings.TrimPrefix(arg, "--watch="))

		case arg == "-runner" || arg == "--runner":
			i++
			if i >= len(args) {
				return fmt.Errorf("-runner requires a value")
			}
			cfg.TypescriptRunner = args[i]

		case arg == "-ignore" || arg == "--ignore":
			i++
			cfg.Ignore = append(cfg.Ignore, splitComma(args[i])...)

		case arg == "-ext" || arg == "--ext":
			i++
			cfg.Extensions = splitComma(args[i])

		case arg == "-delay" || arg == "--delay":
			i++
			var ms int
			fmt.Sscanf(args[i], "%d", &ms)
			if ms > 0 {
				cfg.DebounceMs = ms
			}

		case arg == "-batch" || arg == "--batch" || arg == "--batch=true":
			cfg.BatchMode = true

		case arg == "-clear" || arg == "--clear":
			cfg.ClearScreen = true

		case arg == "--no-hash":
			cfg.UseFileHash = false

		case arg == "-env" || arg == "--env":
			i++
			cfg.EnvFile = args[i]

		case arg == "-max-restarts":
			i++
			fmt.Sscanf(args[i], "%d", &cfg.MaxRestarts)
		}

		i++
	}
	return nil
}

// ─── RUNTIME DETECTION ───────────────────────────────────────────────────────

func detectRuntime() string {
	// Check for bun first (fastest)
	if commandExists("bun") {
		return "bun"
	}
	// Then tsx
	if commandExists("tsx") {
		return "tsx"
	}
	// Then ts-node
	if commandExists("ts-node") {
		return "ts-node"
	}
	// Fallback
	return "node"
}

func commandExists(name string) bool {
	// Simple check via PATH
	_, err := os.Stat("/usr/local/bin/" + name)
	if err == nil {
		return true
	}
	// Try common Bun path
	home, _ := os.UserHomeDir()
	_, err = os.Stat(filepath.Join(home, ".bun", "bin", name))
	return err == nil
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ─── INIT COMMAND ────────────────────────────────────────────────────────────

func InitConfig() {
	const configFile = "fileonix.config.json"

	if _, err := os.Stat(configFile); err == nil {
		ui.Warn(configFile + " already exists — skipping")
		return
	}

	example := map[string]interface{}{
		"script":           "src/index.ts",
		"watch":            []string{"src"},
		"ignore":           []string{"node_modules", "dist", ".git"},
		"typescriptRunner": "bun",
		"clearScreen":      true,
		"debounceMs":       100,
		"batchMode":        false,
		"useFileHash":      true,
		"maxRestarts":      -1,
	}

	data, _ := json.MarshalIndent(example, "", "  ")
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		ui.Fatal("Could not create config", err.Error())
	}

	ui.Success("Created " + configFile)
	ui.Info("Edit it to match your project, then run: fileonix")
}
