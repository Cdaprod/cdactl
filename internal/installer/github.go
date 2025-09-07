package installer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ghRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

func resolveLatestTag(ctx context.Context, cli *http.Client, owner, repo, channel, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=30", owner, repo)
	resp, err := httpGET(ctx, cli, url, token)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("GitHub API %s: %s", url, resp.Status)
	}
	var rels []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rels); err != nil {
		return "", err
	}
	if strings.ToLower(channel) == "canary" {
		for _, r := range rels {
			if !r.Draft && r.Prerelease {
				return r.TagName, nil
			}
		}
	}
	for _, r := range rels {
		if !r.Draft && !r.Prerelease {
			return r.TagName, nil
		}
	}
	for _, r := range rels {
		if !r.Draft {
			return r.TagName, nil
		}
	}
	return "", fmt.Errorf("no releases found for %s/%s", owner, repo)
}
