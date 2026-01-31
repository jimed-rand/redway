package addons

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"reddock/pkg/config"
)

type AddonInjector struct {
	availableAddons map[string]Addon
	workDir         string
}

func NewAddonInjector() *AddonInjector {
	// Check for required dependencies
	if err := CheckDependencies(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	addons := map[string]Addon{
		"houdini":      NewHoudiniAddon(),
		"ndk":          NewNDKAddon(),
		"litegapps":    NewLiteGappsAddon(),
		"mindthegapps": NewMindTheGappsAddon(),
		"opengapps":    NewOpenGappsAddon(),
	}

	return &AddonInjector{
		availableAddons: addons,
		workDir:         "/tmp/reddock-addons",
	}
}

func (ai *AddonInjector) GetAddon(name string) (Addon, error) {
	addon, ok := ai.availableAddons[name]
	if !ok {
		return nil, fmt.Errorf("addon '%s' not found", name)
	}
	return addon, nil
}

func (ai *AddonInjector) ListAddons() []string {
	var names []string
	for name := range ai.availableAddons {
		names = append(names, name)
	}
	return names
}

func (ai *AddonInjector) InjectToContainer(containerName, addonName, version, arch string) error {
	addon, err := ai.GetAddon(addonName)
	if err != nil {
		return err
	}

	if !addon.IsSupported(version) {
		return fmt.Errorf("%s does not support Android %s", addon.Name(), version)
	}

	if err := ensureDir(ai.workDir); err != nil {
		return err
	}

	fmt.Printf("\n=== Injecting %s to container %s ===\n", addon.Name(), containerName)

	if err := ai.checkContainerRunning(containerName); err != nil {
		return err
	}

	fmt.Printf("Downloading %s...\n", addon.Name())
	if err := addon.Download(version, arch); err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	fmt.Printf("Extracting %s...\n", addon.Name())
	if err := addon.Extract(version, arch); err != nil {
		return fmt.Errorf("extract failed: %v", err)
	}

	fmt.Printf("Preparing files for injection...\n")
	addonDir := filepath.Join(ai.workDir, addonName)
	if err := addon.Copy(version, arch, ai.workDir); err != nil {
		return fmt.Errorf("copy failed: %v", err)
	}

	fmt.Printf("Injecting files into container...\n")
	if err := ai.copyToContainer(containerName, addonDir); err != nil {
		return fmt.Errorf("injection failed: %v", err)
	}

	fmt.Printf("Setting permissions...\n")
	if err := ai.setPermissions(containerName, addonName); err != nil {
		fmt.Printf("Warning: Failed to set some permissions: %v\n", err)
	}

	// Update container config
	cfg, err := config.Load()
	if err == nil {
		container := cfg.GetContainer(containerName)
		if container != nil {
			alreadyExists := false
			for _, a := range container.Addons {
				if a == addonName {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				container.Addons = append(container.Addons, addonName)
				config.Save(cfg)
			}
		}
	}

	fmt.Printf("\n✓ Successfully injected %s into %s\n", addon.Name(), containerName)
	fmt.Printf("\nNote: You may need to restart the container for changes to take effect:\n")
	fmt.Printf("  sudo reddock restart %s\n", containerName)

	return nil
}

func (ai *AddonInjector) checkContainerRunning(containerName string) error {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("podman", "inspect", "-f", "{{.State.Status}}", containerName)
		output, err = cmd.Output()
		if err != nil {
			return fmt.Errorf("container '%s' not found", containerName)
		}
	}

	status := strings.TrimSpace(string(output))
	if status != "running" {
		return fmt.Errorf("container '%s' is not running (status: %s). Please start it first", containerName, status)
	}

	return nil
}

func (ai *AddonInjector) copyToContainer(containerName, addonDir string) error {
	runtime := ai.detectRuntime()

	if _, err := os.Stat(addonDir); os.IsNotExist(err) {
		return fmt.Errorf("addon directory not found: %s", addonDir)
	}

	entries, err := os.ReadDir(addonDir)
	if err != nil {
		return fmt.Errorf("failed to read addon directory: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(addonDir, entry.Name())

		cmd := exec.Command(runtime, "cp", srcPath, fmt.Sprintf("%s:/", containerName))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy %s: %v", entry.Name(), err)
		}
		fmt.Printf("  ✓ Copied %s\n", entry.Name())
	}

	return nil
}

func (ai *AddonInjector) setPermissions(containerName, addonName string) error {
	runtime := ai.detectRuntime()

	var commands [][]string

	switch addonName {
	case "houdini":
		commands = [][]string{
			{"chmod", "644", "/system/etc/init/houdini.rc"},
			{"chmod", "-R", "755", "/system/bin/houdini*"},
			{"chmod", "-R", "644", "/system/lib*/libhoudini*"},
		}
	case "ndk":
		commands = [][]string{
			{"chmod", "644", "/system/etc/init/ndk_translation.rc"},
			{"chmod", "-R", "644", "/system/lib*/libndk*"},
		}
	case "litegapps", "mindthegapps", "opengapps":
		commands = [][]string{
			{"chmod", "-R", "755", "/system/priv-app"},
			{"chmod", "-R", "644", "/system/etc/permissions"},
			{"chmod", "-R", "644", "/system/framework"},
		}
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(runtime, "exec", containerName)
		cmd.Args = append(cmd.Args, cmdArgs...)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: chmod failed for %v: %v\n", cmdArgs, err)
		}
	}

	return nil
}

func (ai *AddonInjector) detectRuntime() string {
	if _, err := exec.LookPath("podman"); err == nil {
		cmd := exec.Command("podman", "ps")
		if err := cmd.Run(); err == nil {
			return "podman"
		}
	}
	return "docker"
}

func (ai *AddonInjector) InjectMultiple(containerName string, addons []AddonRequest) error {
	fmt.Printf("\n=== Injecting %d addons to container %s ===\n", len(addons), containerName)

	if err := ai.checkContainerRunning(containerName); err != nil {
		return err
	}

	for i, req := range addons {
		fmt.Printf("\n[%d/%d] Processing %s...\n", i+1, len(addons), req.Name)
		if err := ai.InjectToContainer(containerName, req.Name, req.Version, req.Arch); err != nil {
			fmt.Printf("✗ Failed to inject %s: %v\n", req.Name, err)
			fmt.Printf("  Continuing with remaining addons...\n")
			continue
		}
	}

	fmt.Printf("\n=== Injection complete ===\n")
	fmt.Printf("\nRestart the container to apply changes:\n")
	fmt.Printf("  sudo reddock restart %s\n", containerName)

	return nil
}

func (ai *AddonInjector) GetSupportedVersions(addonName string) ([]string, error) {
	addon, err := ai.GetAddon(addonName)
	if err != nil {
		return nil, err
	}
	return addon.SupportedVersions(), nil
}

func (ai *AddonInjector) Cleanup() error {
	return os.RemoveAll(ai.workDir)
}

type AddonRequest struct {
	Name    string
	Version string
	Arch    string
}
