package cliconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	configPath := writeConfig(t, "base_url: http://127.0.0.1:8000/\ntoken: file-token\n")
	t.Setenv(configEnv, configPath)
	t.Setenv(baseURLEnv, "")
	t.Setenv(tokenEnv, "")

	configuration, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.BaseURL != "http://127.0.0.1:8000" {
		t.Fatalf("BaseURL = %q", configuration.BaseURL)
	}
	if configuration.Token != "file-token" {
		t.Fatalf("Token = %q", configuration.Token)
	}
}

func TestEnvironmentOverridesFile(t *testing.T) {
	configPath := writeConfig(t, "base_url: http://file.example\ntoken: file-token\n")
	t.Setenv(configEnv, configPath)
	t.Setenv(baseURLEnv, "https://environment.example/")
	t.Setenv(tokenEnv, "environment-token")

	configuration, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.BaseURL != "https://environment.example" {
		t.Fatalf("BaseURL = %q", configuration.BaseURL)
	}
	if configuration.Token != "environment-token" {
		t.Fatalf("Token = %q", configuration.Token)
	}
}

func TestLoadRequiresConnectionSettings(t *testing.T) {
	t.Setenv(configEnv, filepath.Join(t.TempDir(), "missing.yaml"))
	t.Setenv(baseURLEnv, "")
	t.Setenv(tokenEnv, "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want a validation error")
	}
}

func TestSaveCreatesSecureConfigAndSupportsPartialUpdates(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "config.yaml")
	t.Setenv(configEnv, configPath)
	t.Setenv(baseURLEnv, "https://environment.example")
	t.Setenv(tokenEnv, "environment-token")

	baseURL := "http://127.0.0.1:8000/"
	token := "file-token"
	path, err := Save(Updates{BaseURL: &baseURL, Token: &token})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if path != configPath {
		t.Fatalf("Save() path = %q", path)
	}

	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if permissions := fileInfo.Mode().Perm(); permissions != 0o600 {
		t.Fatalf("config permissions = %o", permissions)
	}

	updatedToken := "updated-token"
	if _, err := Save(Updates{Token: &updatedToken}); err != nil {
		t.Fatalf("partial Save() error = %v", err)
	}

	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	configuration, err := loadFile(configPath)
	if err != nil {
		t.Fatalf("loadFile() error = %v", err)
	}
	if configuration.BaseURL != "http://127.0.0.1:8000" {
		t.Fatalf("BaseURL = %q", configuration.BaseURL)
	}
	if configuration.Token != updatedToken {
		t.Fatalf("Token = %q", configuration.Token)
	}
	if string(contents) == "" || configuration.Token == os.Getenv(tokenEnv) {
		t.Fatal("environment token was unexpectedly persisted")
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
