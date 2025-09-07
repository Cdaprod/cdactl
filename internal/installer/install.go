package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Install installs or updates a tool.
func Install(ctx context.Context, cfg Config, tool string, explicitTag string) error {
	m, err := LoadManifest(cfg.ToolsDir, tool)
	if err != nil {
		return err
	}
	cli := newHTTPClient(cfg.HTTPTimeout)
	tag := explicitTag
	if tag == "" {
		tag, err = resolveLatestTag(ctx, cli, m.Owner, m.Repo, cfg.Channel, cfg.GHToken)
		if err != nil {
			return err
		}
	}
	fmt.Printf("[cdactl] %s: tag=%s channel=%s\n", m.Name, tag, cfg.Channel)
	verDir := filepath.Join(m.Prefix, tag)
	curLink := filepath.Join(m.Prefix, "current")
	if _, err := os.Stat(verDir); err == nil && !cfg.Force {
		fmt.Printf("[cdactl] %s: version already present at %s\n", m.Name, verDir)
	} else {
		if err := os.MkdirAll(verDir, 0o755); err != nil {
			return err
		}
		osVal, archVal := osID(), archID()
		asset := m.AssetTemplate
		asset = strings.ReplaceAll(asset, "{os}", osVal)
		asset = strings.ReplaceAll(asset, "{arch}", archVal)
		asset = strings.ReplaceAll(asset, "{tag}", tag)
		url := ""
		if m.Mirror != "" {
			url = m.Mirror
			url = strings.ReplaceAll(url, "{tag}", tag)
			url = strings.ReplaceAll(url, "{asset}", asset)
		} else {
			url = fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", m.Owner, m.Repo, tag, asset)
		}
		tmp := filepath.Join(os.TempDir(), m.Name+"-"+tag+".dl")
		fmt.Println("[cdactl] fetching:", url)
		if err := downloadToFile(ctx, cli, url, cfg.GHToken, tmp); err != nil {
			return err
		}
		csURL := url + ".sha256"
		csTmp := tmp + ".sha256"
		if err := downloadToFile(ctx, cli, csURL, cfg.GHToken, csTmp); err == nil {
			if b, err := os.ReadFile(csTmp); err == nil {
				if err := verifySha256(tmp, string(b)); err != nil {
					return err
				}
				fmt.Println("[cdactl] checksum OK")
			}
			_ = os.Remove(csTmp)
		}
		dstBin := filepath.Join(verDir, m.BinaryName)
		if isArchive(asset) {
			if err := extractTo(tmp, verDir); err != nil {
				return err
			}
			found, err := guessInstalledBinary(verDir, m.BinaryName)
			if err != nil {
				return err
			}
			if err := os.Chmod(found, 0o755); err != nil {
				return err
			}
			if found != dstBin {
				if err := os.Rename(found, dstBin); err != nil {
					return err
				}
			}
		} else {
			if err := copyFileMode(tmp, dstBin, 0o755); err != nil {
				return err
			}
		}
		_ = os.Remove(tmp)
		fmt.Println("[cdactl] installed into", verDir)
	}
	if err := atomicSymlink(verDir, curLink); err != nil {
		return err
	}
	fmt.Printf("[cdactl] current -> %s\n", verDir)
	link := filepath.Join(cfg.BinDirLink, m.Name)
	target := filepath.Join(curLink, m.BinaryName)
	if err := ensureSymlink(link, target); err != nil {
		fmt.Printf("[cdactl] WARN: linking %s -> %s failed: %v\n", link, target, err)
		fmt.Println("         Try: sudo ln -sfn", target, link)
		return nil
	}
	fmt.Printf("[cdactl] linked %s -> %s\n", link, target)
	return nil
}

func osID() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	default:
		return runtime.GOOS
	}
}

func archID() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}
