package addons

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type OpenGappsAddon struct {
	*BaseAddon
	dlLinks   map[string][]string
	extractTo string
	nonApks   []string
	skip      []string
}

func NewOpenGappsAddon() *OpenGappsAddon {
	versions := []string{"11.0.0"}
	
	dlLinks := map[string][]string{
		"x86_64": {"https://sourceforge.net/projects/opengapps/files/x86_64/20220503/open_gapps-x86_64-11.0-pico-20220503.zip", "5a6d242be34ad1acf92899c7732afa1b"},
		"x86":    {"https://sourceforge.net/projects/opengapps/files/x86/20220503/open_gapps-x86-11.0-pico-20220503.zip", "efda4943076016d00b40e0874b12ddd3"},
		"arm64":  {"https://sourceforge.net/projects/opengapps/files/arm64/20220503/open_gapps-arm64-11.0-pico-20220503.zip", "67e927e4943757f418e4f934825cf987"},
		"arm":    {"https://sourceforge.net/projects/opengapps/files/arm/20220215/open_gapps-arm-11.0-pico-20220215.zip", "8719519fa32ae83a62621c6056d32814"},
	}

	nonApks := []string{
		"defaultetc-common.tar.lz",
		"defaultframework-common.tar.lz",
		"googlepixelconfig-common.tar.lz",
		"vending-common.tar.lz",
	}

	skip := []string{
		"setupwizarddefault-x86_64.tar.lz",
		"setupwizardtablet-x86_64.tar.lz",
	}

	baseAddon := NewBaseAddon("OpenGapps", AddonTypeGapps, versions)
	return &OpenGappsAddon{
		BaseAddon: baseAddon,
		dlLinks:   dlLinks,
		extractTo: "/tmp/ogapps/extract",
		nonApks:   nonApks,
		skip:      skip,
	}
}

func (o *OpenGappsAddon) Download(version, arch string) error {
	if !o.IsSupported(version) {
		return fmt.Errorf("OpenGapps only supports Android 11.0.0")
	}

	archLinks, ok := o.dlLinks[arch]
	if !ok {
		return fmt.Errorf("OpenGapps not available for architecture %s", arch)
	}

	url := archLinks[0]
	filename := filepath.Join(o.downloadDir, "open_gapps.zip")

	if err := ensureDir(o.downloadDir); err != nil {
		return err
	}

	fmt.Printf("Downloading OpenGapps for Android %s (%s)...\n", version, arch)
	return downloadFile(url, filename)
}

func (o *OpenGappsAddon) Extract(version, arch string) error {
	filename := filepath.Join(o.downloadDir, "open_gapps.zip")
	
	if err := ensureDir(o.extractTo); err != nil {
		return err
	}

	fmt.Println("Extracting OpenGapps archive...")
	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(o.extractTo, f.Name)

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

func (o *OpenGappsAddon) Copy(version, arch, outputDir string) error {
	copyDir := filepath.Join(outputDir, "gapps")
	
	if err := os.RemoveAll(copyDir); err != nil {
		return err
	}

	if err := ensureDir(copyDir); err != nil {
		return err
	}

	appUnpackDir := filepath.Join(o.extractTo, "appunpack")
	if err := ensureDir(appUnpackDir); err != nil {
		return err
	}

	coreDir := filepath.Join(o.extractTo, "Core")
	files, err := os.ReadDir(coreDir)
	if err != nil {
		return fmt.Errorf("failed to read Core directory: %v", err)
	}

	for _, file := range files {
		lzFile := file.Name()
		
		if o.shouldSkip(lzFile) {
			continue
		}

		if err := os.RemoveAll(appUnpackDir); err != nil {
			return err
		}
		if err := ensureDir(appUnpackDir); err != nil {
			return err
		}

		lzPath := filepath.Join(coreDir, lzFile)

		if !o.isNonApk(lzFile) {
			fmt.Printf("Processing app package: %s\n", lzFile)
			
			cmd := exec.Command("tar", "--lzip", "-xvf", lzPath, "-C", appUnpackDir)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to extract %s: %v", lzFile, err)
			}

			dirs, err := os.ReadDir(appUnpackDir)
			if err != nil || len(dirs) == 0 {
				continue
			}
			appName := dirs[0].Name()

			dpiDirs, err := os.ReadDir(filepath.Join(appUnpackDir, appName))
			if err != nil || len(dpiDirs) == 0 {
				continue
			}
			xxDpi := dpiDirs[0].Name()

			privDirs, err := os.ReadDir(filepath.Join(appUnpackDir, appName, xxDpi))
			if err != nil || len(privDirs) == 0 {
				continue
			}
			appPriv := privDirs[0].Name()

			appSrcDir := filepath.Join(appUnpackDir, appName, xxDpi, appPriv)
			apps, err := os.ReadDir(appSrcDir)
			if err != nil {
				continue
			}

			for _, app := range apps {
				srcPath := filepath.Join(appSrcDir, app.Name())
				dstPath := filepath.Join(copyDir, "system", "priv-app", app.Name())
				if err := copyDir2(srcPath, dstPath); err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("Processing extra package: %s\n", lzFile)
			
			cmd := exec.Command("tar", "--lzip", "-xvf", lzPath, "-C", appUnpackDir)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to extract %s: %v", lzFile, err)
			}

			dirs, err := os.ReadDir(appUnpackDir)
			if err != nil || len(dirs) == 0 {
				continue
			}
			appName := dirs[0].Name()

			commonDir := filepath.Join(appUnpackDir, appName, "common")
			commonDirs, err := os.ReadDir(commonDir)
			if err != nil {
				continue
			}

			for _, ccdir := range commonDirs {
				srcPath := filepath.Join(commonDir, ccdir.Name())
				dstPath := filepath.Join(copyDir, "system", ccdir.Name())
				if err := copyDir2(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (o *OpenGappsAddon) Install(version, arch, outputDir string) error {
	if err := o.Download(version, arch); err != nil {
		return err
	}
	if err := o.Extract(version, arch); err != nil {
		return err
	}
	return o.Copy(version, arch, outputDir)
}

func (o *OpenGappsAddon) isNonApk(filename string) bool {
	for _, nonApk := range o.nonApks {
		if filename == nonApk {
			return true
		}
	}
	return false
}

func (o *OpenGappsAddon) shouldSkip(filename string) bool {
	for _, skip := range o.skip {
		if strings.Contains(filename, skip) {
			return true
		}
	}
	return false
}
