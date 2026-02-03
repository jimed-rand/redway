package addons

import (
	"fmt"
	"os"
	"path/filepath"
	"reddock/pkg/container"
	"reddock/pkg/ui"
	"strings"
)

type AddonManager struct {
	availableAddons map[string]Addon
	workDir         string
	cacheDir        string
	runtime         container.Runtime
}

func NewAddonManager() *AddonManager {
	addons := map[string]Addon{
		"houdini":      NewHoudiniAddon(),
		"ndk":          NewNDKAddon(),
		"litegapps":    NewLiteGappsAddon(),
		"mindthegapps": NewMindTheGappsAddon(),
		"opengapps":    NewOpenGappsAddon(),
	}

	return &AddonManager{
		availableAddons: addons,
		workDir:         "/tmp/reddock-addons",
		runtime:         container.NewRuntime(),
	}
}

func (am *AddonManager) GetAddon(name string) (Addon, error) {
	addon, ok := am.availableAddons[name]
	if !ok {
		return nil, fmt.Errorf("addon '%s' not found", name)
	}
	return addon, nil
}

func (am *AddonManager) ListAddons() []string {
	var names []string
	for name := range am.availableAddons {
		names = append(names, name)
	}
	return names
}

func (am *AddonManager) GetAddonsByType(t AddonType, version string) []Addon {
	var addons []Addon
	for _, addon := range am.availableAddons {
		if addon.Type() == t {
			if version == "" || addon.IsSupported(version) {
				addons = append(addons, addon)
			}
		}
	}
	return addons
}

func (am *AddonManager) GetAddonNamesByType(t AddonType, version string) []string {
	var names []string
	for name, addon := range am.availableAddons {
		if addon.Type() == t {
			if version == "" || addon.IsSupported(version) {
				names = append(names, name)
			}
		}
	}
	return names
}

func (am *AddonManager) PrepareAddon(addonName, version, arch string) error {
	addon, err := am.GetAddon(addonName)
	if err != nil {
		return err
	}

	if err := ensureDir(am.workDir); err != nil {
		return err
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Preparing %s...", addon.Name()))
	spinner.Start()
	defer func() {
		if !spinner.IsDone() {
			spinner.Finish("Preparation interrupted")
		}
	}()

	err = addon.Install(version, arch, am.workDir, func(msg string) {
		spinner.SetMessage(msg)
	})
	if err != nil {
		spinner.Finish(fmt.Sprintf("Failed to prepare %s", addon.Name()))
		return err
	}
	spinner.Finish(fmt.Sprintf("Successfully prepared %s", addon.Name()))
	return nil
}

func (am *AddonManager) BuildDockerfile(baseImage string, addons []string) (string, error) {
	var dockerfile strings.Builder

	dockerfile.WriteString(fmt.Sprintf("FROM %s\n\n", baseImage))

	for _, addonName := range addons {
		addon, err := am.GetAddon(addonName)
		if err != nil {
			return "", err
		}
		instructions := addon.DockerfileInstructions()
		if instructions != "" {
			dockerfile.WriteString(instructions)
		}
	}

	// Add Redroid boot arguments
	dockerfile.WriteString("\n# Redroid boot arguments\n")
	dockerfile.WriteString("CMD [\"androidboot.redroid_gpu_mode=auto\"]\n")

	return dockerfile.String(), nil
}

func (am *AddonManager) BuildCustomImage(baseImage, targetImage, version, arch string, addonNames []string) error {
	if err := ensureDir(am.workDir); err != nil {
		return err
	}

	fmt.Println("\n=== Building custom Redroid Image ===")
	fmt.Printf("Base Image: %s\n", baseImage)
	fmt.Printf("Target Image: %s\n", targetImage)
	fmt.Printf("Addons: %v\n\n", addonNames)

	// Pull base image if it's official redroid image
	if strings.HasPrefix(baseImage, "redroid/redroid:") {
		fmt.Printf("Pulling official Redroid image %s...\n", baseImage)
		if err := am.runtime.PullImage(baseImage); err != nil {
			return fmt.Errorf("Failed to pull official image: %v", err)
		}
	} else if !strings.HasPrefix(baseImage, "reddock-custom:") && !strings.HasPrefix(baseImage, "reddock/") {
		// For other images, try to pull if not local
		fmt.Printf("Pulling base image %s...\n", baseImage)
		if err := am.runtime.PullImage(baseImage); err != nil {
			// If pull fails, might be a local image, just warn
			fmt.Printf("Warning: Could not pull image %s, hoping it exists locally: %v\n", baseImage, err)
		}
	}

	for _, addonName := range addonNames {
		if err := am.PrepareAddon(addonName, version, arch); err != nil {
			return fmt.Errorf("Failed to prepare %s: %v", addonName, err)
		}
	}

	dockerfileContent, err := am.BuildDockerfile(baseImage, addonNames)
	if err != nil {
		return err
	}

	dockerfilePath := filepath.Join(am.workDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("Failed to write Dockerfile: %v", err)
	}

	fmt.Println("Dockerfile created:")
	fmt.Println(dockerfileContent)

	spinner := ui.NewSpinner("Building Docker image...")
	spinner.Start()

	buildArgs := []string{"build", "-t", targetImage, am.workDir}
	if am.runtime.Name() == "docker" {
		buildArgs = []string{"buildx", "build", "--load", "-t", targetImage, am.workDir}
	}
	cmd := am.runtime.Command(buildArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		spinner.Finish("Failed to build Docker image")
		fmt.Println(string(output))
		return fmt.Errorf("Failed to build Docker image: %v", err)
	}
	spinner.Finish(fmt.Sprintf("Successfully built %s", targetImage))

	return nil
}

func (am *AddonManager) GetSupportedVersions(addonName string) ([]string, error) {
	addon, err := am.GetAddon(addonName)
	if err != nil {
		return nil, err
	}
	return addon.SupportedVersions(), nil
}

func (am *AddonManager) Cleanup() error {
	return os.RemoveAll(am.workDir)
}
