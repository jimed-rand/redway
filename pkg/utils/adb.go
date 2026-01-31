package utils

import (
	"fmt"
	"os/exec"
	"time"

	"redway/pkg/container"
)

type AdbManager struct {
	manager       *container.Manager
	containerName string
}

func NewAdbManager(containerName string) *AdbManager {
	return &AdbManager{
		manager:       container.NewManagerForContainer(containerName),
		containerName: containerName,
	}
}

func (a *AdbManager) ShowConnection() error {
	if err := container.CheckRoot(); err != nil {
		return err
	}
	if !a.manager.IsRunning() {
		return fmt.Errorf("The container '%s' is not running. Start it with 'redway start %s'", a.containerName, a.containerName)
	}

	fmt.Printf("Waiting for container '%s' to get an IP address...\n", a.containerName)

	var ip string
	var err error
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		ip, err = a.manager.GetIP()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		if i%5 == 0 && i > 0 {
			fmt.Printf("  still waiting (%ds/30s)...\n", i)
		}
	}

	if err != nil {
		return fmt.Errorf("Failed to get container IP after 30 seconds: %v\n\nHint: Android might still be booting, or your LXC networking (lxc-net) might not be providing DHCP", err)
	}

	fmt.Println("\nADB Connection Information:")
	fmt.Println("===========================")
	fmt.Printf("IP Address: %s\n", ip)

	fmt.Printf("\nAttempting to connect via ADB...\n")
	cmd := exec.Command("adb", "connect", fmt.Sprintf("%s:5555", ip))
	output, _ := cmd.CombinedOutput()
	fmt.Printf("ADB Output: %s", string(output))

	fmt.Printf("\nYou can now use:\n")
	fmt.Printf("  adb shell              # Access Android shell\n")
	fmt.Printf("  adb install app.apk    # Install APK\n")
	fmt.Printf("  adb logcat             # View logs\n")
	fmt.Printf("  adb devices            # List connected devices\n")

	return nil
}
