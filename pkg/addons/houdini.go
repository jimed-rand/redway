package addons

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

type HoudiniAddon struct {
	*BaseAddon
	dlLinks   map[string]map[string][]string
	extractTo string
}

func NewHoudiniAddon() *HoudiniAddon {
	versions := []string{"8.1.0", "9.0.0", "10.0.0", "11.0.0", "12.0.0", "13.0.0", "14.0.0", "15.0.0", "16.0.0"}

	dlLinks := map[string]map[string][]string{
		"8.1.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/46682f423b8497db3f96222f2669d770eff764c3.zip"},
			"md5": {"cd4dd2891aa18e7699d33dcc3fe3ffd4"},
		},
		"9.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/46682f423b8497db3f96222f2669d770eff764c3.zip"},
			"md5": {"cd4dd2891aa18e7699d33dcc3fe3ffd4"},
		},
		"10.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"11.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"12.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"13.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"14.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"15.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"16.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
	}

	baseAddon := NewBaseAddon("Houdini", AddonTypeHoudini, versions)
	return &HoudiniAddon{
		BaseAddon: baseAddon,
		dlLinks:   dlLinks,
		extractTo: "/tmp/houdiniunpack",
	}
}

func (h *HoudiniAddon) Download(version, arch string) error {
	if !h.IsSupported(version) {
		return fmt.Errorf("Houdini not supported for Android %s", version)
	}

	if arch != "x86" && arch != "x86_64" {
		return fmt.Errorf("Houdini only supports x86/x86_64 architecture")
	}

	url := h.dlLinks[version]["url"][0]
	filename := filepath.Join(h.downloadDir, "libhoudini.zip")

	if err := ensureDir(h.downloadDir); err != nil {
		return err
	}

	fmt.Printf("Downloading Houdini for Android %s...\n", version)
	return downloadFile(url, filename)
}

func (h *HoudiniAddon) Extract(version, arch string) error {
	filename := filepath.Join(h.downloadDir, "libhoudini.zip")

	if err := ensureDir(h.extractTo); err != nil {
		return err
	}

	fmt.Println("Extracting Houdini archive...")
	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(h.extractTo, f.Name)

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

func (h *HoudiniAddon) Copy(version, arch, outputDir string) error {
	copyDir := filepath.Join(outputDir, "houdini")

	if err := os.RemoveAll(copyDir); err != nil {
		return err
	}

	url := h.dlLinks[version]["url"][0]
	re := regexp.MustCompile(`([a-zA-Z0-9]+)\.zip`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return fmt.Errorf("Failed to extract name from URL")
	}
	name := matches[1]

	srcDir := filepath.Join(h.extractTo, "vendor_intel_proprietary_houdini-"+name, "prebuilts")
	dstDir := filepath.Join(copyDir, "system")

	fmt.Println("Copying Houdini library files...")
	if err := copyDir2(srcDir, dstDir); err != nil {
		return err
	}

	initRC := h.getInitRC()
	initPath := filepath.Join(copyDir, "system", "etc", "init", "houdini.rc")
	if err := ensureDir(filepath.Dir(initPath)); err != nil {
		return err
	}

	if err := os.WriteFile(initPath, []byte(initRC), 0644); err != nil {
		return err
	}

	return nil
}

func (h *HoudiniAddon) Install(version, arch, outputDir string) error {
	if err := h.Download(version, arch); err != nil {
		return err
	}
	if err := h.Extract(version, arch); err != nil {
		return err
	}
	return h.Copy(version, arch, outputDir)
}

func (h *HoudiniAddon) GetBootArgs(version, arch string) []string {
	// For 64-only images, we only set 64-bit properties
	if version == "12.0.0_64only" || version == "13.0.0_64only" || version == "14.0.0_64only" || version == "15.0.0_64only" || version == "16.0.0_64only" {
		return []string{
			"androidboot.use_memfd=1",
			"ro.product.cpu.abilist=x86_64,arm64-v8a",
			"ro.product.cpu.abilist64=x86_64,arm64-v8a",
			"ro.dalvik.vm.isa.arm64=x86_64",
			"ro.enable.native.bridge.exec=1",
			"ro.dalvik.vm.native.bridge=libhoudini.so",
		}
	}

	return []string{
		"ro.product.cpu.abilist=x86_64,arm64-v8a,x86,armeabi-v7a,armeabi",
		"ro.product.cpu.abilist64=x86_64,arm64-v8a",
		"ro.product.cpu.abilist32=x86,armeabi-v7a,armeabi",
		"ro.dalvik.vm.isa.arm=x86",
		"ro.dalvik.vm.isa.arm64=x86_64",
		"ro.enable.native.bridge.exec=1",
		"ro.vendor.enable.native.bridge.exec=1",
		"ro.vendor.enable.native.bridge.exec64=1",
		"ro.dalvik.vm.native.bridge=libhoudini.so",
	}
}

func (h *HoudiniAddon) getInitRC() string {
	return `
on early-init
    mount binfmt_misc binfmt_misc /proc/sys/fs/binfmt_misc

on property:ro.enable.native.bridge.exec=1
    copy /system/etc/binfmt_misc/arm_exe /proc/sys/fs/binfmt_misc/register
    copy /system/etc/binfmt_misc/arm_dyn /proc/sys/fs/binfmt_misc/register

on property:ro.enable.native.bridge.exec64=1
    copy /system/etc/binfmt_misc/arm64_exe /proc/sys/fs/binfmt_misc/register
    copy /system/etc/binfmt_misc/arm64_dyn /proc/sys/fs/binfmt_misc/register

on property:sys.boot_completed=1
    exec -- /system/bin/sh -c "echo ':arm_exe:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x01\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x02\\\\x00\\\\x28::/system/bin/houdini:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm_dyn:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x01\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x03\\\\x00\\\\x28::/system/bin/houdini:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm64_exe:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x02\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x02\\\\x00\\\\xb7::/system/bin/houdini64:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm64_dyn:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x02\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x03\\\\x00\\\\xb7::/system/bin/houdini64:P' >> /proc/sys/fs/binfmt_misc/register"

`
}
