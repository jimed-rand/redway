package addons

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
	Download(version, arch string, onStatus func(string)) error
	Extract(version, arch string, onStatus func(string)) error
	Copy(version, arch, outputDir string, onStatus func(string)) error
	Install(version, arch, outputDir string, onStatus func(string)) error
	DockerfileInstructions() string
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

func (b *BaseAddon) DockerfileInstructions() string {
	// Default behavior: COPY <addon_name> /
	// This assumes the addon extracts its content into a directory named after the addon type/name
	// in the build context, and we want to merge it into root.
	dirName := ""
	switch b.addonType {
	case AddonTypeHoudini:
		dirName = "houdini"
	case AddonTypeNDK:
		dirName = "ndk"
	case AddonTypeGapps:
		dirName = "gapps" // Assuming gapps addons use 'gapps' or similar.
		// Actually gapps might differ (litegapps, opengapps etc).
		// But BaseAddon doesn't know the specific implementation details easily unless we standardize.
		// For now, let's just return empty string in BaseAddon and force specific addons to implement/override it if they want non-empty.
		// Or better: BaseAddon shouldn't guess.
		return ""
	}
	if dirName != "" {
		return fmt.Sprintf("COPY %s /\n", dirName)
	}
	return ""
}

func getDownloadDir() string {
	home := os.Getenv("HOME")
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "reddock", "downloads")
	}
	return filepath.Join(home, ".cache", "reddock", "downloads")
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func copyDir2(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	// Use cp -a to preserve permissions, symlinks, etc.
	// This is much more robust for Android system files.
	cmd := exec.Command("cp", "-a", filepath.Clean(src)+"/.", filepath.Clean(dst)+"/")
	return cmd.Run()
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	cmd := exec.Command("cp", "-a", src, dst)
	return cmd.Run()
}

func extractZip(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
