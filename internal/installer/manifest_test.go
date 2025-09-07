package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	dir := t.TempDir()
	data := []byte("name: demo\nowner: foo\nrepo: bar\nasset_template: bin_{os}_{arch}\nbinary_name: demo\n")
	if err := os.WriteFile(filepath.Join(dir, "demo.yaml"), data, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	m, err := LoadManifest(dir, "demo")
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m.Name != "demo" || m.Owner != "foo" || m.Repo != "bar" {
		t.Fatalf("unexpected manifest: %+v", m)
	}
	list, err := LoadAllManifests(dir)
	if err != nil || len(list) != 1 {
		t.Fatalf("LoadAllManifests: %v len=%d", err, len(list))
	}
}
