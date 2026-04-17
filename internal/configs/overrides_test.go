package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"gopkg.in/yaml.v2"
)

// TestOverlay tests overlaying a nested map into the Config struct.
func TestOverlay(t *testing.T) {
	// Start with a default config.
	cfg := Config{
		Validation: Validation{
			NameRejectRegex: "test",
		},
	}

	newValues := map[string]any{
		"Validation": map[string]any{
			"NameRejectRegex": "test-changed",
		},
	}

	if err := cfg.OverlayOverrides(newValues); err != nil {
		t.Fatalf("Overlay failed: %v", err)
	}

	if cfg.Validation.NameRejectRegex != "test-changed" {
		t.Errorf("Expected NameRejectRegex to be \"test-changed\", got \"%s\"", cfg.Validation.NameRejectRegex)
	}
}

// TestOverlayDotMap tests overlaying a configuration using dot-syntax keys.
func TestOverlayDotMap(t *testing.T) {
	// Start with a default config.
	cfg := Config{
		Validation: Validation{
			NameRejectRegex: "test",
		},
	}

	dotValues := map[string]any{
		"Validation.NameRejectRegex": "test-changed",
	}

	if err := cfg.OverlayOverrides(dotValues); err != nil {
		t.Fatalf("OverlayDotMap failed: %v", err)
	}

	if cfg.Validation.NameRejectRegex != "test-changed" {
		t.Errorf("Expected LeaderboardSize to be \"test-changed\", got \"%s\"", cfg.Validation.NameRejectRegex)
	}
}

// TestOverlayDotMapMultipleFields demonstrates overlaying multiple fields using dot-syntax.
// Here, we extend the configuration to have an additional field.
func TestOverlayDotMapMultipleFields(t *testing.T) {
	// Define an extended configuration.
	type ExtendedStatistics struct {
		LeaderboardSize int    `yaml:"LeaderboardSize"`
		SomeField       string `yaml:"SomeField"`
	}

	type ExtendedConfig struct {
		Statistics ExtendedStatistics `yaml:"Statistics"`
	}

	cfg := ExtendedConfig{
		Statistics: ExtendedStatistics{
			LeaderboardSize: 5,
			SomeField:       "default",
		},
	}

	dotValues := map[string]any{
		"Statistics.LeaderboardSize": 25,
		"Statistics.SomeField":       "updated",
	}

	// Unflatten the dot-syntax map.
	nestedMap := unflattenMap(dotValues)
	// Marshal to YAML and then unmarshal into the extended config.
	b, err := yaml.Marshal(nestedMap)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if cfg.Statistics.LeaderboardSize != 25 {
		t.Errorf("Expected LeaderboardSize to be 25, got %d", cfg.Statistics.LeaderboardSize)
	}
	if cfg.Statistics.SomeField != "updated" {
		t.Errorf("Expected SomeField to be 'updated', got '%s'", cfg.Statistics.SomeField)
	}
}

func TestLoadOverridesPrecedence(t *testing.T) {
	mudlog.SetupLogger(nil, "LOW", "", false)

	tmpDir := t.TempDir()

	cfg := Config{
		Server: Server{
			MudName: "Base",
		},
		Network: Network{
			HttpPort: 80,
		},
	}

	globalPath := filepath.Join(tmpDir, "global.yaml")
	worldPath := filepath.Join(tmpDir, "world.yaml")
	envPath := filepath.Join(tmpDir, "env.yaml")

	if err := os.WriteFile(globalPath, []byte("Server:\n  MudName: Global\nNetwork:\n  HttpPort: 8080\n"), 0o644); err != nil {
		t.Fatalf("writing global override: %v", err)
	}
	if err := os.WriteFile(worldPath, []byte("Server:\n  MudName: World\n"), 0o644); err != nil {
		t.Fatalf("writing world override: %v", err)
	}
	if err := os.WriteFile(envPath, []byte("Network:\n  HttpPort: 9090\n"), 0o644); err != nil {
		t.Fatalf("writing env override: %v", err)
	}

	loaded, loadedOverrides, err := loadOverrides(&cfg, []string{globalPath, worldPath, envPath}, nil)
	if err != nil {
		t.Fatalf("loadOverrides failed: %v", err)
	}
	if !loaded {
		t.Fatal("expected overrides to load")
	}

	if got := cfg.Server.MudName.String(); got != "World" {
		t.Fatalf("expected world override to win for Server.MudName, got %q", got)
	}
	if got := int(cfg.Network.HttpPort); got != 9090 {
		t.Fatalf("expected env override to win for Network.HttpPort, got %d", got)
	}

	if got := loadedOverrides["Server"].(map[string]any)["MudName"]; got != "World" {
		t.Fatalf("expected loaded overrides to preserve winning Server.MudName, got %q", got)
	}
	if got := loadedOverrides["Network"].(map[string]any)["HttpPort"]; got != 9090 {
		t.Fatalf("expected loaded overrides to preserve winning Network.HttpPort, got %v", got)
	}
}

func TestOverridePathsIncludesGlobalWorldAndEnv(t *testing.T) {
	t.Setenv("CONFIG_PATH", "/tmp/config.custom.yaml")

	cfg := Config{
		FilePaths: FilePaths{
			DataFiles: "_datafiles/world/custom",
		},
		validated: true,
	}

	paths := overridePathsForConfig(cfg)
	expected := []string{
		"_datafiles/config-overrides.yaml",
		"_datafiles/world/custom/config-overrides.yaml",
		"/tmp/config.custom.yaml",
	}

	if len(paths) != len(expected) {
		t.Fatalf("expected %d override paths, got %d: %v", len(expected), len(paths), paths)
	}

	for i := range expected {
		if paths[i] != expected[i] {
			t.Fatalf("expected override path %d to be %q, got %q", i, expected[i], paths[i])
		}
	}
}

func TestOverridePathsUsesGlobalOverrideDataFiles(t *testing.T) {
	mudlog.SetupLogger(nil, "LOW", "", false)
	t.Setenv("CONFIG_PATH", "")

	tmpDir := t.TempDir()

	cfg := Config{
		FilePaths: FilePaths{
			DataFiles: "_datafiles/world/default",
		},
	}

	globalPath := filepath.Join(tmpDir, "global.yaml")
	if err := os.WriteFile(globalPath, []byte("FilePaths:\n  DataFiles: _datafiles/world/custom\n"), 0o644); err != nil {
		t.Fatalf("writing global override: %v", err)
	}

	loaded, loadedOverrides, err := loadOverrides(&cfg, []string{globalPath}, nil)
	if err != nil {
		t.Fatalf("loadOverrides failed: %v", err)
	}
	if !loaded {
		t.Fatal("expected global override to load")
	}

	paths := overridePathsForConfig(cfg)
	if got, expected := paths[1], "_datafiles/world/custom/config-overrides.yaml"; got != expected {
		t.Fatalf("expected world override path after global override to be %q, got %q", expected, got)
	}
	if got := loadedOverrides["FilePaths"].(map[string]any)["DataFiles"]; got != "_datafiles/world/custom" {
		t.Fatalf("expected loaded overrides to preserve FilePaths.DataFiles, got %q", got)
	}
}
