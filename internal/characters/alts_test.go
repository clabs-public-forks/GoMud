package characters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func TestSaveAltsFixesExistingFilePermissions(t *testing.T) {
	mudlog.SetupLogger(nil, "", "", false)

	oldFilePaths := configs.GetFilePathsConfig()
	tempDir := t.TempDir()
	usersDir := filepath.Join(tempDir, "users")
	if err := os.MkdirAll(usersDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := configs.AddOverlayOverrides(map[string]any{
		"FilePaths.DataFiles":        tempDir,
		"FilePaths.CarefulSaveFiles": false,
	}); err != nil {
		t.Fatalf("AddOverlayOverrides() error = %v", err)
	}
	t.Cleanup(func() {
		_ = configs.AddOverlayOverrides(map[string]any{
			"FilePaths.DataFiles":        oldFilePaths.DataFiles,
			"FilePaths.CarefulSaveFiles": oldFilePaths.CarefulSaveFiles,
		})
	})

	altsPath := filepath.Join(usersDir, "123.alts.yaml")
	if err := os.WriteFile(altsPath, []byte("old"), 0o666); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chmod(altsPath, 0o666); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}

	if !SaveAlts(123, []Character{}) {
		t.Fatal("SaveAlts() = false, want true")
	}

	info, err := os.Stat(altsPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("SaveAlts() kept insecure permissions: %o", got)
	}
}
