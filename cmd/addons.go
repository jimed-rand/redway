package cmd

import (
	"fmt"
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
	case "inject":
		return c.executeAddonsInject(subArgs)
	case "inject-multi":
		return c.executeAddonsInjectMulti(subArgs)
	default:
		return fmt.Errorf("unknown addons subcommand: %s", subCommand)
	}
}

func (c *Command) showAddonsHelp() error {
	fmt.Println("Addons Management")
	fmt.Println("\nUsage: reddock addons [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  list                                List available addons")
	fmt.Println("  inject <container> <addon>          Inject addon to running container")
	fmt.Println("  inject-multi <container> <addons>   Inject multiple addons")
	fmt.Println("\nAvailable Addons:")
	fmt.Println("  houdini       - Intel Houdini ARM translation (x86/x86_64 only)")
	fmt.Println("  ndk           - NDK ARM translation (x86/x86_64 only)")
	fmt.Println("  litegapps     - LiteGapps (Google Apps)")
	fmt.Println("  mindthegapps  - MindTheGapps (Google Apps)")
	fmt.Println("  opengapps     - OpenGapps (Google Apps, Android 11 only)")
	fmt.Println("\nExamples:")
	fmt.Println("  # List available addons")
	fmt.Println("  sudo reddock addons list")
	fmt.Println("")
	fmt.Println("  # Inject single addon to running container")
	fmt.Println("  sudo reddock addons inject my-android13 litegapps")
	fmt.Println("  sudo reddock addons inject my-android11 opengapps")
	fmt.Println("")
	fmt.Println("  # Inject multiple addons at once")
	fmt.Println("  sudo reddock addons inject-multi my-android13 litegapps ndk")
	fmt.Println("  sudo reddock addons inject-multi my-android14 mindthegapps houdini")
	fmt.Println("")
	fmt.Println("Note: Container must be running before injecting addons")
	return nil
}

func (c *Command) executeAddonsList() error {
	injector := addons.NewAddonInjector()
	addonNames := injector.ListAddons()

	fmt.Println("Available Addons:")
	fmt.Println(strings.Repeat("-", 70))

	for _, name := range addonNames {
		addon, _ := injector.GetAddon(name)
		versions := addon.SupportedVersions()
		fmt.Printf("%-15s - %s\n", name, addon.Name())
		fmt.Printf("                Supported versions: %v\n", versions)
		fmt.Println()
	}

	return nil
}

func (c *Command) executeAddonsInject(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: reddock addons inject <container-name> <addon-name>")
	}

	containerName := args[0]
	addonName := args[1]
	arch := addons.GetHostArch()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	container := cfg.GetContainer(containerName)
	if container == nil {
		return fmt.Errorf("container '%s' not found", containerName)
	}

	version := extractVersionFromImage(container.ImageURL)
	if version == "" {
		return fmt.Errorf("cannot determine Android version from image: %s", container.ImageURL)
	}

	injector := addons.NewAddonInjector()
	defer injector.Cleanup()

	return injector.InjectToContainer(containerName, addonName, version, arch)
}

func (c *Command) executeAddonsInjectMulti(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: reddock addons inject-multi <container-name> <addon1> [addon2] ...")
	}

	containerName := args[0]
	addonNames := args[1:]
	arch := addons.GetHostArch()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	container := cfg.GetContainer(containerName)
	if container == nil {
		return fmt.Errorf("container '%s' not found", containerName)
	}

	version := extractVersionFromImage(container.ImageURL)
	if version == "" {
		return fmt.Errorf("cannot determine Android version from image: %s", container.ImageURL)
	}

	injector := addons.NewAddonInjector()
	defer injector.Cleanup()

	var requests []addons.AddonRequest
	for _, addonName := range addonNames {
		requests = append(requests, addons.AddonRequest{
			Name:    addonName,
			Version: version,
			Arch:    arch,
		})
	}

	return injector.InjectMultiple(containerName, requests)
}

func extractVersionFromImage(imageURL string) string {
	parts := strings.Split(imageURL, ":")
	if len(parts) < 2 {
		return ""
	}

	versionPart := parts[1]

	versionPart = strings.TrimSuffix(versionPart, "-latest")
	versionPart = strings.TrimSuffix(versionPart, "-gapps")
	versionPart = strings.TrimSuffix(versionPart, "-ndk")
	versionPart = strings.TrimSuffix(versionPart, "-houdini")
	versionPart = strings.TrimSuffix(versionPart, "-magisk")
	versionPart = strings.TrimSuffix(versionPart, "-widevine")

	if strings.Contains(versionPart, "_") {
		return versionPart
	}

	knownVersions := []string{
		"16.0.0", "15.0.0", "14.0.0", "13.0.0", "12.0.0",
		"11.0.0", "10.0.0", "9.0.0", "8.1.0",
		"16.0.0_64only", "15.0.0_64only", "14.0.0_64only",
		"13.0.0_64only", "12.0.0_64only",
	}

	for _, v := range knownVersions {
		if strings.HasPrefix(versionPart, v) {
			return v
		}
	}

	return versionPart
}
