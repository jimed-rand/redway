package addons

import (
	"os"
	"path/filepath"
)

type AddonType string

const (
	AddonTypeGapps    AddonType = "gapps"
	AddonTypeHoudini  AddonType = "houdini"
	AddonTypeNDK      AddonType = "ndk"
	AddonTypeMagisk   AddonType = "magisk"
	AddonTypeWidevine AddonType = "widevine"
)

type Addon interface {
	Name() string
	Type() AddonType
	SupportedVersions() []string
	IsSupported(version string) bool
	Download(version, arch string) error
	Extract(version, arch string) error
	Copy(version, arch, outputDir string) error
	Install(version, arch, outputDir string) error
}

type BaseAddon struct {
	name              string
	addonType         AddonType
	supportedVersions []string
	downloadDir       string
}

func NewBaseAddon(name string, addonType AddonType, versions []string) *BaseAddon {
	downloadDir := getDownloadDir()
	return &BaseAddon{
		name:              name,
		addonType:         addonType,
		supportedVersions: versions,
		downloadDir:       downloadDir,
	}
}

func (b *BaseAddon) Name() string {
	return b.name
}

func (b *BaseAddon) Type() AddonType {
	return b.addonType
}

func (b *BaseAddon) SupportedVersions() []string {
	return b.supportedVersions
}

func (b *BaseAddon) IsSupported(version string) bool {
	for _, v := range b.supportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

func getDownloadDir() string {
	home := os.Getenv("HOME")
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "reddock", "downloads")
	}
	return filepath.Join(home, ".cache", "reddock", "downloads")
}
