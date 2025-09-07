package installer

import "time"

// Config describes installer behaviour.
// Provided by CLI layer.
type Config struct {
	ToolsDir    string        // directory with manifests
	PrefixBase  string        // base directory for installs
	BinDirLink  string        // directory for symlinks (/usr/local/bin)
	Channel     string        // release channel
	Force       bool          // reinstall if exists
	GHToken     string        // optional GitHub token
	HTTPTimeout time.Duration // network timeout
}

// Manifest describes a tool release.
type Manifest struct {
	Name          string `yaml:"name"`
	Owner         string `yaml:"owner"`
	Repo          string `yaml:"repo"`
	AssetTemplate string `yaml:"asset_template"`
	BinaryName    string `yaml:"binary_name"`
	Prefix        string `yaml:"prefix"`
	Mirror        string `yaml:"mirror"`
}
