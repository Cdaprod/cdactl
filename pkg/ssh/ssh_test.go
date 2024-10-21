// pkg/ssh/ssh_test.go

package ssh

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "testing"
)

func TestLoadConfig(t *testing.T) {
    // Setup temporary config file
    tempDir := os.TempDir()
    configPath := filepath.Join(tempDir, "config.yaml")
    configContent := `
ssh:
  default_user: testuser
  hosts:
    - name: Test Server
      host: test.example.com
      user: testuser
      port: 22
      key_path: ~/.ssh/test_id_rsa
`
    if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
        t.Fatalf("Failed to write temp config file: %v", err)
    }
    defer os.Remove(configPath)

    // Override Viper config path
    viper.SetConfigFile(configPath)

    config, err := LoadConfig()
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }

    if config.DefaultUser != "testuser" {
        t.Errorf("Expected default_user to be 'testuser', got '%s'", config.DefaultUser)
    }

    if len(config.Hosts) != 1 {
        t.Errorf("Expected 1 host, got %d", len(config.Hosts))
    }

    host := config.Hosts[0]
    if host.Name != "Test Server" || host.Host != "test.example.com" || host.User != "testuser" || host.Port != 22 || host.KeyPath != "~/.ssh/test_id_rsa" {
        t.Errorf("Host details do not match expected values")
    }
}

func TestDetermineHost(t *testing.T) {
    hosts := []SSHHost{
        {
            Name:    "Server A",
            Host:    "servera.example.com",
            User:    "usera",
            Port:    22,
            KeyPath: "~/.ssh/servera_id_rsa",
        },
        {
            Name:    "Server B",
            Host:    "serverb.example.com",
            User:    "userb",
            Port:    2222,
            KeyPath: "~/.ssh/serverb_id_rsa",
        },
    }

    repoURL := "git@servera.example.com:username/repo.git"
    host := determineHost(repoURL, hosts)
    if host == nil || host.Name != "Server A" {
        t.Errorf("Failed to determine correct host for repo URL")
    }

    repoURL = "git@unknown.example.com:username/repo.git"
    host = determineHost(repoURL, hosts)
    if host != nil {
        t.Errorf("Expected no host match, but got one")
    }
}

func TestPublicKeyAuth(t *testing.T) {
    // Setup temporary SSH key file
    tempDir := os.TempDir()
    keyPath := filepath.Join(tempDir, "test_id_rsa")
    keyContent := `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQC3+......
-----END RSA PRIVATE KEY-----`
    if err := ioutil.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
        t.Fatalf("Failed to write temp SSH key file: %v", err)
    }
    defer os.Remove(keyPath)

    auth, err := publicKeyAuth(keyPath)
    if err != nil {
        t.Fatalf("Failed to create public key auth: %v", err)
    }

    if auth.Username != "git" {
        t.Errorf("Expected username 'git', got '%s'", auth.Username)
    }

    if auth.AuthMethod == nil {
        t.Errorf("Expected AuthMethod to be set, got nil")
    }
}