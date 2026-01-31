package addons

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type MindTheGappsAddon struct {
	*BaseAddon
	dlLinks   map[string]map[string][]string
	extractTo string
}

func NewMindTheGappsAddon() *MindTheGappsAddon {
	versions := []string{"12.0.0", "12.0.0_64only", "13.0.0", "13.0.0_64only", "14.0.0", "15.0.0", "16.0.0"}

	dlLinks := map[string]map[string][]string{
		"16.0.0": {
			"x86_64": {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20250330/MindTheGapps-15.0.0-x86_64-20250330.zip", "e54694828bd74e9066b2534a9675c31e"},
			"arm64":  {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20250330/MindTheGapps-15.0.0-arm64-20250330.zip", "79acb62f0f7c66b0f0bcadae5624f3d1"},
			"arm":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20250330/MindTheGapps-15.0.0-arm-20250330.zip", "4ced6a404a714e61831e16068c3642b3"},
		},
		"15.0.0": {
			"x86_64": {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20250330/MindTheGapps-15.0.0-x86_64-20250330.zip", "e54694828bd74e9066b2534a9675c31e"},
			"arm64":  {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20250330/MindTheGapps-15.0.0-arm64-20250330.zip", "79acb62f0f7c66b0f0bcadae5624f3d1"},
			"arm":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20250330/MindTheGapps-15.0.0-arm-20250330.zip", "4ced6a404a714e61831e16068c3642b3"},
		},
		"14.0.0": {
			"x86_64": {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-14.0.0-x86_64-20240226.zip", "a827a84ccb0cf5914756e8561257ed13"},
			"x86":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-14.0.0-x86-20240226.zip", "45736b21475464e4a45196b9aa9d3b7f"},
			"arm64":  {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-14.0.0-arm64-20240226.zip", "a0905cc7bf3f4f4f2e3f59a4e1fc789b"},
			"arm":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-14.0.0-arm-20240226.zip", "fa167a3b7a10c4d3e688a59cd794f75b"},
		},
		"13.0.0": {
			"x86_64": {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-13.0.0-x86_64-20240226.zip", "eee87a540b6e778f3a114fff29e133aa"},
			"x86":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-13.0.0-x86-20240226.zip", "d928c5eabb4394a97f2d7a5c663e7c2e"},
			"arm64":  {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-13.0.0-arm64-20240226.zip", "ebdf35e17bc1c22337762fcf15cd6e97"},
			"arm":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240619/MindTheGapps-13.0.0-arm-20240619.zip", "ec7aa5efc9e449b101bc2ee7448a49bf"},
		},
		"13.0.0_64only": {
			"x86_64": {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-13.0.0-x86_64-20240226.zip", "eee87a540b6e778f3a114fff29e133aa"},
			"arm64":  {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240226/MindTheGapps-13.0.0-arm64-20240226.zip", "ebdf35e17bc1c22337762fcf15cd6e97"},
		},
		"12.0.0_64only": {
			"x86_64": {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240619/MindTheGapps-12.1.0-x86_64-20240619.zip", "05d6e99b6e6567e66d43774559b15fbd"},
			"arm64":  {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240619/MindTheGapps-12.1.0-arm64-20240619.zip", "94dd174ff16c2f0006b66b25025efd04"},
		},
		"12.0.0": {
			"x86_64": {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240619/MindTheGapps-12.1.0-x86_64-20240619.zip", "05d6e99b6e6567e66d43774559b15fbd"},
			"x86":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240619/MindTheGapps-12.1.0-x86-20240619.zip", "ff2421a75afbdda8a003e4fd25e95050"},
			"arm64":  {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240619/MindTheGapps-12.1.0-arm64-20240619.zip", "94dd174ff16c2f0006b66b25025efd04"},
			"arm":    {"https://github.com/s1204IT/MindTheGappsBuilder/releases/download/20240619/MindTheGapps-12.1.0-arm-20240619.zip", "5af756b3b5776c2f6ee024a9f7f42a2f"},
		},
	}

	baseAddon := NewBaseAddon("MindTheGapps", AddonTypeGapps, versions)
	return &MindTheGappsAddon{
		BaseAddon: baseAddon,
		dlLinks:   dlLinks,
		extractTo: "/tmp/mindthegapps/extract",
	}
}

func (m *MindTheGappsAddon) Download(version, arch string) error {
	if !m.IsSupported(version) {
		return fmt.Errorf("MindTheGapps not supported for Android %s", version)
	}

	versionLinks, ok := m.dlLinks[version]
	if !ok {
		return fmt.Errorf("no download links for version %s", version)
	}

	archLinks, ok := versionLinks[arch]
	if !ok {
		return fmt.Errorf("MindTheGapps not available for architecture %s on Android %s", arch, version)
	}

	url := archLinks[0]
	filename := filepath.Join(m.downloadDir, "mindthegapps.zip")

	if err := ensureDir(m.downloadDir); err != nil {
		return err
	}

	fmt.Printf("Downloading MindTheGapps for Android %s (%s)...\n", version, arch)
	return downloadFile(url, filename)
}

func (m *MindTheGappsAddon) Extract(version, arch string) error {
	filename := filepath.Join(m.downloadDir, "mindthegapps.zip")

	if err := ensureDir(m.extractTo); err != nil {
		return err
	}

	fmt.Println("Extracting MindTheGapps archive...")
	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(m.extractTo, f.Name)

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

func (m *MindTheGappsAddon) Copy(version, arch, outputDir string) error {
	copyDir := filepath.Join(outputDir, "mindthegapps")

	if err := os.RemoveAll(copyDir); err != nil {
		return err
	}

	srcDir := filepath.Join(m.extractTo, "system")
	dstDir := filepath.Join(copyDir, "system")

	fmt.Println("Copying MindTheGapps files...")
	return copyDir2(srcDir, dstDir)
}

func (m *MindTheGappsAddon) Install(version, arch, outputDir string) error {
	if err := m.Download(version, arch); err != nil {
		return err
	}
	if err := m.Extract(version, arch); err != nil {
		return err
	}
	return m.Copy(version, arch, outputDir)
}

func (m *MindTheGappsAddon) GetBootArgs(version, arch string) []string {
	return []string{
		"ro.setupwizard.mode=DISABLED",
	}
}
