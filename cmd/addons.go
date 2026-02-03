package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"reddock/pkg/addons"
	"reddock/pkg/config"
)

func (c *Command) executeAddons() error {
	if len(c.Args) == 0 {
		return c.showAddonsHelp()
	}

	subCommand := c.Args[0]
	subArgs := c.Args[1:]

	switch subCommand {
	case "list":
		return c.executeAddonsList()
	case "build":
		return c.executeAddonsBuild(subArgs)
	case "prepare":
		return c.executeAddonsPrepare(subArgs)
	default:
		return fmt.Errorf("unknown addons subcommand: %s", subCommand)
	}
}

func (c *Command) showAddonsHelp() error {
	fmt.Println("Addons Management")
	fmt.Println("\nUsage: reddock addons [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  list                                         	List available addons")
	fmt.Println("  prepare <addon> <version>                    	Prepare addon files (for runtime install)")
	fmt.Println("  build <n> <version> <addons...>            	Build custom image with addons")
	fmt.Println("\nAvailable Addons:")
	fmt.Println("  houdini       		- Intel Houdini ARM translation (x86/x86_64 only)")
	fmt.Println("  ndk           		- NDK ARM translation (x86/x86_64 only)")
	fmt.Println("  litegapps     		- LiteGapps (Google Apps)")
	fmt.Println("  mindthegapps  		- MindTheGapps (Google Apps)")
	fmt.Println("  opengapps     		- OpenGapps (Google Apps, Android 11 only)")
	fmt.Println("\nExamples:")
	fmt.Println("  reddock addons list")
	fmt.Println("  reddock addons prepare houdini 13.0.0")
	fmt.Println("  reddock addons build custom-android13 13.0.0 litegapps ndk")
	fmt.Println("\nRuntime Installation (redroid-script approach):")
	fmt.Println("  1. Prepare the addon:    reddock addons prepare houdini 13.0.0")
	fmt.Println("  2. Start the container:  sudo reddock start android13")
	fmt.Println("  3. Install to running:   sudo reddock dockerfile install android13 houdini")
	fmt.Println("  4. Save changes:         sudo reddock dockerfile commit android13 myimage:latest")
	return nil
}

func (c *Command) executeAddonsList() error {
	manager := addons.NewAddonManager()
	addonNames := manager.ListAddons()

	fmt.Println("Available Addons:")
	fmt.Println(strings.Repeat("-", 50))

	for _, name := range addonNames {
		addon, _ := manager.GetAddon(name)
		versions := addon.SupportedVersions()
		fmt.Printf("%-15s - %s\n", name, addon.Name())
		fmt.Printf("Supported versions: %v\n", versions)
		fmt.Println()
	}

	return nil
}

func (c *Command) executeAddonsBuild(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("Command invalid!\nUsage: reddock addons build <image-name> <android-version> <addon1> [addon2] ...\nFormat: Use NAMESPACE/REPOSITORY[:TAG] (Avoid HOST[:PORT]/ for local images)")
	}

	imageName := args[0]
	if err := config.ValidateImageName(imageName); err != nil {
		return fmt.Errorf("Invalid image name: %v", err)
	}
	version := args[1]
	addonNames := args[2:]
	arch := getHostArch()

	manager := addons.NewAddonManager()
	defer manager.Cleanup()

	for _, addonName := range addonNames {
		if _, err := manager.GetAddon(addonName); err != nil {
			return fmt.Errorf("Invalid addon: %s", addonName)
		}
	}

	baseImage := fmt.Sprintf("redroid/redroid:%s-latest", version)
	return manager.BuildCustomImage(baseImage, imageName, version, arch, addonNames)
}

// executeAddonsPrepare prepares addon files without building an image
// This is useful for installing addons to a running container
func (c *Command) executeAddonsPrepare(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Usage: reddock addons prepare <addon-name> <android-version>\n\n" +
			"This prepares addon files without building an image.\n" +
			"After preparation, use 'reddock dockerfile install <container> <addon>' to install to a running container.\n\n" +
			"Example:\n" +
			"  reddock addons prepare houdini 13.0.0\n" +
			"  reddock dockerfile install android13 houdini")
	}

	addonName := args[0]
	version := args[1]
	arch := getHostArch()

	manager := addons.NewAddonManager()

	addon, err := manager.GetAddon(addonName)
	if err != nil {
		return fmt.Errorf("Invalid addon: %s", addonName)
	}

	fmt.Printf("\n=== Preparing %s addon for Android %s ===\n", addon.Name(), version)
	fmt.Printf("Architecture: %s\n\n", arch)

	if err := manager.PrepareAddon(addonName, version, arch); err != nil {
		return err
	}

	fmt.Println("\nAddon prepared successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  To install to a running container:")
	fmt.Printf("    sudo reddock dockerfile install <container-name> %s\n", addonName)
	fmt.Println("\n  To build a Docker image with this addon:")
	fmt.Printf("    sudo reddock addons build myimage:latest %s %s\n", version, addonName)

	return nil
}

func getHostArch() string {
	machine := runtime.GOARCH

	mapping := map[string]string{
		"386":   "x86",
		"amd64": "x86_64",
		"arm64": "arm64",
		"arm":   "arm",
	}

	if arch, ok := mapping[machine]; ok {
		return arch
	}
	return "x86_64"
}
