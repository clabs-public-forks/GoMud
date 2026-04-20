package fileloader

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type testFlatFile struct {
	File string `yaml:"-"`
	Name string `yaml:"name"`
}

func (t testFlatFile) Filepath() string {
	return t.File
}

func (t testFlatFile) Validate() error {
	return nil
}

type testLoadableFile struct {
	testFlatFile `yaml:",inline"`
	ID           string `yaml:"id"`
}

func (t testLoadableFile) Id() string {
	return t.ID
}

func TestSaveFlatFileReplacesExistingPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exact Unix permission bits are not portable on Windows")
	}

	basePath := t.TempDir()
	path := filepath.Join(basePath, "thing.yaml")

	if err := os.WriteFile(path, []byte("old"), 0o666); err != nil {
		t.Fatalf("unable to create existing file: %v", err)
	}
	if err := os.Chmod(path, 0o777); err != nil {
		t.Fatalf("unable to widen file permissions: %v", err)
	}

	if err := SaveFlatFile(basePath, testFlatFile{File: "thing.yaml", Name: "new"}); err != nil {
		t.Fatalf("SaveFlatFile failed: %v", err)
	}

	assertFileMode(t, path, saveFilePerm)
}

func TestSaveAllFlatFilesReplacesExistingPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exact Unix permission bits are not portable on Windows")
	}

	basePath := t.TempDir()
	path := filepath.Join(basePath, "thing.yaml")

	if err := os.WriteFile(path, []byte("old"), 0o666); err != nil {
		t.Fatalf("unable to create existing file: %v", err)
	}
	if err := os.Chmod(path, 0o777); err != nil {
		t.Fatalf("unable to widen file permissions: %v", err)
	}

	data := map[string]testLoadableFile{
		"thing": {
			testFlatFile: testFlatFile{File: "thing.yaml", Name: "new"},
			ID:           "thing",
		},
	}
	if _, err := SaveAllFlatFiles(basePath, data); err != nil {
		t.Fatalf("SaveAllFlatFiles failed: %v", err)
	}

	assertFileMode(t, path, saveFilePerm)
}

func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("unable to stat saved file: %v", err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("file permissions = %o, want %o", got, want)
	}
}
