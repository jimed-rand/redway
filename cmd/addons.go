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
	case "register-gapps":
		return c.executeAddonsRegisterGApps(subArgs)
	default:
		return fmt.Errorf("Unknown addons subcommand: %s", subCommand)
	}
}

func (c *Command) showAddonsHelp() error {
	fmt.Println("Addons Management")
	fmt.Println("\nUsage: reddock addons [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  list                                List available addons")
	fmt.Println("  inject <container> <addon>          Inject addon to running container")
	fmt.Println("  inject-multi <container> <addons>   Inject multiple addons")
	fmt.Println("  register-gapps <container>          Fetch Android ID for Google Play registration")
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
		fmt.Printf("Supported versions: %v\n", versions)
		fmt.Println()
	}

	return nil
}

func (c *Command) executeAddonsInject(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Command invalid! usage: reddock addons inject <container-name> <addon-name>")
	}

	containerName := args[0]
	addonName := args[1]
	arch := addons.GetHostArch()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	container := cfg.GetContainer(containerName)
	if container == nil {
		return fmt.Errorf("Container '%s' not found", containerName)
	}

	version := config.ExtractVersionFromImage(container.ImageURL)
	if version == "" {
		return fmt.Errorf("Cannot determine Android version from image: %s", container.ImageURL)
	}

	injector := addons.NewAddonInjector()
	defer injector.Cleanup()

	return injector.InjectToContainer(containerName, addonName, version, arch)
}

func (c *Command) executeAddonsInjectMulti(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Command invalid! usage: reddock addons inject-multi <container-name> <addon1> [addon2] ...")
	}

	containerName := args[0]
	addonNames := args[1:]
	arch := addons.GetHostArch()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	container := cfg.GetContainer(containerName)
	if container == nil {
		return fmt.Errorf("Container '%s' not found", containerName)
	}

	version := config.ExtractVersionFromImage(container.ImageURL)
	if version == "" {
		return fmt.Errorf("Cannot determine Android version from image: %s", container.ImageURL)
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

func (c *Command) executeAddonsRegisterGApps(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Command invalid! usage: reddock addons register-gapps <container-name>")
	}

	containerName := args[0]
	return addons.RegisterGApps(containerName)
}

// Removed local extractVersionFromImage in favor of utils.ExtractVersionFromImage
