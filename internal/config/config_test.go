package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHierarchicalLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileonix-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create structure:
	// root/package.json (BatchMode: true)
	// root/src/ (no config)
	// root/src/nested/fileonix.config.json (MaxRestarts: 10)
	
	root := tempDir
	src := filepath.Join(root, "src")
	nested := filepath.Join(src, "nested")
	os.MkdirAll(nested, 0755)

	// 1. Root package.json
	pkgData := map[string]interface{}{
		"fileonix": map[string]interface{}{
			"batchMode": true,
		},
	}
	pkgJSON, _ := json.Marshal(pkgData)
	os.WriteFile(filepath.Join(root, "package.json"), pkgJSON, 0644)

	// 2. Nested config
	nestedCfg := map[string]interface{}{
		"maxRestarts": 10,
	}
	nestedJSON, _ := json.Marshal(nestedCfg)
	os.WriteFile(filepath.Join(nested, "fileonix.config.json"), nestedJSON, 0644)

	// Change to a different CWD to test that it respects anchor
	oldCWD, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCWD)

	t.Run("Resolve from nested script", func(t *testing.T) {
		scriptPath := filepath.Join(nested, "app.ts")
		args := []string{"--script", scriptPath}
		
		cfg, err := Load(args)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Should pick nested config because it's closest
		if cfg.MaxRestarts != 10 {
			t.Errorf("Expected MaxRestarts=10, got %d", cfg.MaxRestarts)
		}
		// Should NOT pick root batchMode because it stopped at the first config found
		if cfg.BatchMode != false {
			t.Errorf("Expected BatchMode=false, got %v", cfg.BatchMode)
		}
	})

	t.Run("Resolve from src (no config, should go up to root)", func(t *testing.T) {
		scriptPath := filepath.Join(src, "main.ts")
		args := []string{"--script", scriptPath}

		cfg, err := Load(args)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Should pick root package.json
		if cfg.BatchMode != true {
			t.Errorf("Expected BatchMode=true, got %v", cfg.BatchMode)
		}
		if cfg.MaxRestarts != -1 { // Default
			t.Errorf("Expected MaxRestarts=-1, got %d", cfg.MaxRestarts)
		}
	})

	t.Run("CLI overrides file", func(t *testing.T) {
		scriptPath := filepath.Join(src, "main.ts")
		args := []string{"--script", scriptPath, "--batch=false"}

		cfg, err := Load(args)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// CLI should win over root package.json
		if cfg.BatchMode != false {
			t.Errorf("Expected BatchMode=false, got %v", cfg.BatchMode)
		}
	})
}
