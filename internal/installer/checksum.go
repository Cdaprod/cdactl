package installer

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := bufio.NewReader(f).WriteTo(h); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func verifySha256(filePath, checksumText string) error {
	want := ""
	for _, line := range strings.Split(checksumText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			if len(fields[0]) == 64 {
				want = fields[0]
				break
			}
			if len(fields[len(fields)-1]) == 64 {
				want = fields[len(fields)-1]
				break
			}
		}
	}
	if want == "" {
		return fmt.Errorf("could not parse checksum file")
	}
	got, err := sha256File(filePath)
	if err != nil {
		return err
	}
	if strings.ToLower(want) != strings.ToLower(got) {
		return fmt.Errorf("checksum mismatch: have %s want %s", got, want)
	}
	return nil
}

func guessInstalledBinary(verDir, binaryName string) (string, error) {
	cand := filepath.Join(verDir, binaryName)
	if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
		return cand, nil
	}
	var found string
	filepath.WalkDir(verDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && filepath.Base(path) == binaryName {
			found = path
			return fmt.Errorf("stop")
		}
		return nil
	})
	if found == "" {
		return "", fmt.Errorf("binary %s not found in %s", binaryName, verDir)
	}
	return found, nil
}
