package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadManifest loads a manifest file by tool name.
func LoadManifest(dir, tool string) (*Manifest, error) {
	path := filepath.Join(dir, tool+".yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	if m.Name == "" {
		m.Name = tool
	}
	if m.Prefix == "" {
		m.Prefix = filepath.Join("/opt/cdaprod", m.Name)
	}
	if m.BinaryName == "" {
		m.BinaryName = m.Name
	}
	return &m, nil
}

// LoadAllManifests returns all manifests within dir.
func LoadAllManifests(dir string) ([]*Manifest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*Manifest
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		m, err := LoadManifest(dir, e.Name()[:len(e.Name())-5])
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
