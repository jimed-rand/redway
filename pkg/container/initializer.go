package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"redway/pkg/config"
)

type Initializer struct {
	config *config.Config
	image  string
}

func NewInitializer(image string) *Initializer {
	cfg, _ := config.Load()
	cfg.ImageURL = image

	return &Initializer{
		config: cfg,
		image:  image,
	}
}

func (i *Initializer) Initialize() error {
	fmt.Println("Initializing the container...")
	fmt.Printf("Image: %s\n", i.image)
	fmt.Printf("Container: %s\n\n", i.config.ContainerName)

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Checking kernel modules", i.checkKernelModules},
		{"Checking LXC tools", i.checkLXCTools},
		{"Setting up LXC networking", i.setupLXCNetworking},
		{"Adjusting OCI template", i.adjustOCITemplate},
		{"Checking required tools", i.checkRequiredTools},
		{"Creating container", i.createContainer},
		{"Fixing container filesystem", i.fixContainerFilesystem},
		{"Creating data directory", i.createDataDirectory},
		{"Adjusting container config", i.adjustContainerConfig},
		{"Applying networking workaround", i.applyNetworkingWorkaround},
	}

	for _, step := range steps {
		fmt.Printf("[*] %s...\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s failed: %v", step.name, err)
		}
	}

	i.config.Initialized = true
	if err := config.Save(i.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Println("\nThe container has been initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("  redway start        # Start the container\n")
	fmt.Printf("  redway adb-connect  # Get ADB connection info\n")
	fmt.Printf("  redway shell        # Access container shell\n")

	return nil
}

func (i *Initializer) checkKernelModules() error {
	binderFound := false
	binderPaths := []string{
		"/sys/module/binder_linux",
		"/sys/module/binder",
		"/dev/binderfs",
		"/dev/binder",
	}

	for _, path := range binderPaths {
		if _, err := os.Stat(path); err == nil {
			binderFound = true
			break
		}
	}

	if binderFound {
		fmt.Println("Binder support (binderfs/binder module) detected")
		return nil
	}

	fmt.Println("Binder support (binderfs/binder module) not detected")
	fmt.Println("You need to enable the binderfs in your kernel or install binder module.")
	fmt.Println("                                               ")
	fmt.Println("If container fails to start, try loading modules:")
	fmt.Println("                                               ")
	fmt.Println("sudo modprobe binder_linux devices=\"binder,hwbinder,vndbinder\"")
	fmt.Println("                                               ")
	fmt.Println("See README for more information.")

	return nil
}

func (i *Initializer) checkLXCTools() error {
	tools := []string{"lxc-create", "lxc-start", "lxc-stop", "lxc-info"}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found. Please install lxc-utils", tool)
		}
	}

	fmt.Println("LXC tools available")
	return nil
}

func (i *Initializer) setupLXCNetworking() error {
	if _, err := os.Stat("/sys/class/net/lxcbr0"); err == nil {
		fmt.Println("LXC bridge (lxcbr0) already exists")
		return i.setupNAT()
	}

	fmt.Println("Setting up LXC networking...")

	cmd := exec.Command("systemctl", "is-enabled", "--quiet", "lxc-net")
	if err := cmd.Run(); err != nil {
		fmt.Println("Enabling lxc-net service...")
		enableCmd := exec.Command("systemctl", "enable", "lxc-net")
		if err := enableCmd.Run(); err != nil {
			fmt.Printf("Warning: Could not enable lxc-net: %v\n", err)
		}
	}

	cmd = exec.Command("systemctl", "is-active", "--quiet", "lxc-net")
	if err := cmd.Run(); err != nil {
		fmt.Println("Starting lxc-net service...")
		startCmd := exec.Command("systemctl", "start", "lxc-net")
		if err := startCmd.Run(); err != nil {
			return fmt.Errorf("failed to start lxc-net: %v. Try manually: sudo systemctl start lxc-net", err)
		}
	}

	if _, err := os.Stat("/sys/class/net/lxcbr0"); err != nil {
		return fmt.Errorf("lxcbr0 bridge still not available after starting lxc-net")
	}

	if err := i.setupNAT(); err != nil {
		fmt.Printf("Warning: NAT setup failed: %v\n", err)
	}

	fmt.Println("LXC networking configured successfully")
	return nil
}

func (i *Initializer) setupNAT() error {
	fmt.Println("Setting up NAT for container networking...")

	checkCmd := exec.Command("sh", "-c", "iptables -t nat -C POSTROUTING -s 10.0.3.0/24 ! -d 10.0.3.0/24 -j MASQUERADE 2>/dev/null")
	if err := checkCmd.Run(); err == nil {
		fmt.Println("NAT rule already exists")
		return i.ensureIPForwarding()
	}

	natCmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.0.3.0/24", "!", "-d", "10.0.3.0/24", "-j", "MASQUERADE")
	if err := natCmd.Run(); err != nil {
		return fmt.Errorf("failed to add NAT rule: %v", err)
	}

	fmt.Println("NAT rule added successfully")
	return i.ensureIPForwarding()
}

func (i *Initializer) ensureIPForwarding() error {
	content, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err == nil && strings.TrimSpace(string(content)) == "1" {
		fmt.Println("IP forwarding already enabled")
		return nil
	}

	fmt.Println("Enabling IP forwarding...")
	fwdCmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := fwdCmd.Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}

	sysctlConf := "/etc/sysctl.d/99-lxc-ip-forward.conf"
	if _, err := os.Stat(sysctlConf); os.IsNotExist(err) {
		if err := os.WriteFile(sysctlConf, []byte("net.ipv4.ip_forward=1\n"), 0644); err != nil {
			fmt.Printf("Warning: Could not persist IP forwarding setting: %v\n", err)
		} else {
			fmt.Println("IP forwarding persisted to sysctl")
		}
	}

	return nil
}

func (i *Initializer) adjustOCITemplate() error {
	templatePath := "/usr/share/lxc/templates/lxc-oci"

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Println("Note: lxc-oci template not found, skipping adjustment")
		return nil
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	modified := strings.ReplaceAll(string(content), "set -eu", "set -u")

	if err := os.WriteFile(templatePath, []byte(modified), 0755); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	fmt.Println("OCI template adjusted")
	return nil
}

func (i *Initializer) checkRequiredTools() error {
	tools := []string{"skopeo", "umoci", "jq"}
	missing := []string{}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required tools: %s. Please install them", strings.Join(missing, ", "))
	}

	fmt.Println("Required tools installed")
	return nil
}

func (i *Initializer) createContainer() error {
	containerPath := i.config.GetContainerPath()

	if _, err := os.Stat(containerPath); err == nil {
		fmt.Println("Container already exists")
		return nil
	}

	fmt.Printf("Creating LXC container from %s...\n", i.image)

	cmd := exec.Command("lxc-create",
		"-n", i.config.ContainerName,
		"-t", "oci",
		"--",
		"-u", i.image)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	fmt.Println("Container created")
	return nil
}

func (i *Initializer) fixContainerFilesystem() error {
	rootfs := i.config.GetRootfsPath()

	etcDir := filepath.Join(rootfs, "etc")
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create /etc directory: %v", err)
	}

	hostnamePath := filepath.Join(etcDir, "hostname")
	if _, err := os.Stat(hostnamePath); os.IsNotExist(err) {
		if err := os.WriteFile(hostnamePath, []byte(i.config.ContainerName+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to create /etc/hostname: %v", err)
		}
		fmt.Println("Created /etc/hostname")
	}

	hostsPath := filepath.Join(etcDir, "hosts")
	if _, err := os.Stat(hostsPath); os.IsNotExist(err) {
		hostsContent := fmt.Sprintf("127.0.0.1 localhost\n127.0.1.1 %s\n::1 localhost ip6-localhost ip6-loopback\n", i.config.ContainerName)
		if err := os.WriteFile(hostsPath, []byte(hostsContent), 0644); err != nil {
			return fmt.Errorf("failed to create /etc/hosts: %v", err)
		}
		fmt.Println("Created /etc/hosts")
	}

	fmt.Println("Container filesystem fixed")
	return nil
}

func (i *Initializer) createDataDirectory() error {
	if err := os.MkdirAll(i.config.DataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	fmt.Printf("Data directory: %s\n", i.config.DataPath)
	return nil
}

func (i *Initializer) adjustContainerConfig() error {
	configPath := i.config.GetConfigFilePath()

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		if !strings.Contains(line, "lxc.include") {
			newLines = append(newLines, line)
		}
	}

	additionalConfig := fmt.Sprintf(`
### Redway Configuration
lxc.init.cmd = /init androidboot.hardware=redroid androidboot.redroid_gpu_mode=%s
lxc.apparmor.profile = unconfined
lxc.autodev = 1
lxc.autodev.tmpfs.size = 25000000
lxc.mount.entry = %s data none bind 0 0
`, i.config.GPUMode, i.config.DataPath)

	newLines = append(newLines, additionalConfig)

	finalContent := strings.Join(newLines, "\n")

	if err := os.WriteFile(configPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	fmt.Println("Container config adjusted")
	return nil
}

func (i *Initializer) applyNetworkingWorkaround() error {
	workaroundPath := filepath.Join(i.config.GetRootfsPath(), "vendor", "bin", "ipconfigstore")

	if _, err := os.Stat(workaroundPath); err == nil {
		if err := os.Remove(workaroundPath); err != nil {
			fmt.Printf("Warning: Could not remove ipconfigstore: %v\n", err)
		} else {
			fmt.Println("Networking workaround applied")
		}
	} else {
		fmt.Println("Networking workaround not needed")
	}

	return nil
}
