package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const sampleConfig = `FilePaths:
  WebDomain: "play.example.com"
  HttpsCertFile: "server.crt"
  HttpsKeyFile: "server.key"
  HttpsEmail: "admin@example.com"
  HttpsCacheDir: "_datafiles/tls"
Network:
  HttpPort: 8080
  HttpsPort: 8443
  HttpsRedirect: false
`

func TestHTTPSSetupManualModePreservesExistingPortsByDefault(t *testing.T) {
	configPath := writeTempConfig(t)

	input := strings.Join([]string{
		"2",
		"",
		"",
		"",
		"",
		"",
		"",
		"Y",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input)
	if !strings.Contains(output, "HttpPort: 8080") {
		t.Fatalf("https-setup output did not preserve HTTP port:\n%s", output)
	}
	if !strings.Contains(output, "HttpsPort: 8443") {
		t.Fatalf("https-setup output did not preserve HTTPS port:\n%s", output)
	}

	updated := readConfig(t, configPath)
	if !strings.Contains(updated, "  HttpPort: 8080\n") {
		t.Fatalf("https-setup config did not preserve HTTP port:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsPort: 8443\n") {
		t.Fatalf("https-setup config did not preserve HTTPS port:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsEmail: \"\"\n") {
		t.Fatalf("https-setup config did not clear HttpsEmail in manual mode:\n%s", updated)
	}
}

func TestHTTPSSetupManualModeAllowsCustomPorts(t *testing.T) {
	configPath := writeTempConfig(t)

	input := strings.Join([]string{
		"2",
		"",
		"",
		"",
		"true",
		"9080",
		"9443",
		"Y",
		"",
	}, "\n")

	runHTTPSSetup(t, configPath, input)

	updated := readConfig(t, configPath)
	if !strings.Contains(updated, "  HttpPort: 9080\n") {
		t.Fatalf("https-setup config did not update HTTP port:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsPort: 9443\n") {
		t.Fatalf("https-setup config did not update HTTPS port:\n%s", updated)
	}
	if !strings.Contains(updated, "  HttpsRedirect: true\n") {
		t.Fatalf("https-setup config did not update redirect setting:\n%s", updated)
	}
}

func writeTempConfig(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(sampleConfig), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return configPath
}

func readConfig(t *testing.T, configPath string) string {
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
