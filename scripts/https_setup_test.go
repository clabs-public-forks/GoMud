package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const sampleHTTPSConfig = `FilePaths:
  HttpsCertFile: "server.crt"
  HttpsKeyFile: "server.key"
Network:
  HttpPort: 8080
  HttpsPort: 8443
  HttpsRedirect: false
`

func TestHTTPSSetupManualModePreservesExistingDefaults(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"1",
		"",
		"",
		"",
		"",
		"",
		"Y",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input)
	if !strings.Contains(output, "HttpsPort: 8443") {
		t.Fatalf("https-setup output did not preserve HTTPS port:\n%s", output)
	}

	updated := readHTTPSSetupConfig(t, configPath)
	if !strings.Contains(updated, "  HttpsCertFile: \"server.crt\"\n") {
		t.Fatalf("https-setup config did not preserve cert file:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsKeyFile: \"server.key\"\n") {
		t.Fatalf("https-setup config did not preserve key file:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpPort: 8080\n") {
		t.Fatalf("https-setup config did not preserve HTTP port:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsPort: 8443\n") {
		t.Fatalf("https-setup config did not preserve HTTPS port:\n%s", updated)
	}
}

func TestHTTPSSetupManualModeAllowsCustomValues(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"1",
		"/etc/ssl/fullchain.pem",
		"/etc/ssl/privkey.pem",
		"80",
		"443",
		"true",
		"Y",
		"",
	}, "\n")

	runHTTPSSetup(t, configPath, input)

	updated := readHTTPSSetupConfig(t, configPath)
	if !strings.Contains(updated, "  HttpsCertFile: \"/etc/ssl/fullchain.pem\"\n") {
		t.Fatalf("https-setup config did not update cert file:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsKeyFile: \"/etc/ssl/privkey.pem\"\n") {
		t.Fatalf("https-setup config did not update key file:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsPort: 443\n") {
		t.Fatalf("https-setup config did not update HTTPS port:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsRedirect: true\n") {
		t.Fatalf("https-setup config did not update redirect setting:\n%s", updated)
	}
}

func TestHTTPSSetupHTTPOnlyClearsHTTPS(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"2",
		"8080",
		"Y",
		"",
	}, "\n")

	runHTTPSSetup(t, configPath, input)

	updated := readHTTPSSetupConfig(t, configPath)
	if !strings.Contains(updated, "  HttpsCertFile: \"\"\n") {
		t.Fatalf("https-setup config did not clear cert file:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsKeyFile: \"\"\n") {
		t.Fatalf("https-setup config did not clear key file:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsPort: 0\n") {
		t.Fatalf("https-setup config did not disable HTTPS port:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsRedirect: false\n") {
		t.Fatalf("https-setup config did not disable redirect:\n%s", updated)
	}
}

func writeHTTPSSetupTempConfig(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return configPath
}

func readHTTPSSetupConfig(t *testing.T, configPath string) string {
	t.Helper()

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	return string(data)
}

func runHTTPSSetup(t *testing.T, configPath string, input string) string {
	t.Helper()

	scriptDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	tempDir := t.TempDir()
	cmd := exec.Command("sh", "./https-setup.sh")
	cmd.Dir = scriptDir
	cmd.Env = append(os.Environ(),
		"CONFIG_FILE="+configPath,
		"TMPDIR="+tempDir,
	)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("https-setup.sh failed: %v\n%s", err, output)
	}

	return string(output)
}
