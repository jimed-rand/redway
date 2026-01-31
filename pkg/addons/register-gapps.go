package addons

import (
	"fmt"
	"os/exec"
	"strings"
)

// RegisterGApps fetches the Android ID from the container for Google Play certification
func RegisterGApps(containerName string) error {
	runtime := DetectRuntime()

	fmt.Printf("Fetching Android ID for Google Play certification from container: %s\n", containerName)

	// Command to fetch Android ID from GSF database
	// Waydroid recommendation: sqlite3 /data/data/com.google.android.gsf/databases/gservices.db "select value from main where name = 'android_id';"

	// We use docker/podman exec to run this inside the container
	// Note: We use 'su -c' to ensure we have root permissions if needed,
	// though docker exec usually runs as root by default in redroid.
	sqlCmd := "sqlite3 /data/data/com.google.android.gsf/databases/gservices.db \"select value from main where name = 'android_id';\""

	cmd := exec.Command(runtime, "exec", containerName, "sh", "-c", sqlCmd)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Try alternative if sqlite3 is not in PATH but might be in /system/xbin or /system/bin
		sqlCmdAlt := "/system/xbin/sqlite3 /data/data/com.google.android.gsf/databases/gservices.db \"select value from main where name = 'android_id';\""
		cmd = exec.Command(runtime, "exec", containerName, "sh", "-c", sqlCmdAlt)
		output, err = cmd.CombinedOutput()

		if err != nil {
			return fmt.Errorf("Failed to fetch Android ID: %v\nOutput: %s\n\nMake sure GAPPS are installed and the container has been running for a few minutes.", err, string(output))
		}
	}

	androidID := strings.TrimSpace(string(output))
	if androidID == "" {
		return fmt.Errorf("Android ID not found. Make sure you have opened Google Play Store at least once in the container.")
	}

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Printf("Your Android ID (GSF): %s\n", androidID)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("\nPlease visit the following URL to register your device:")
	fmt.Println("https://www.google.com/android/uncertified/")
	fmt.Println("\nEnter the ID above in the 'Google Services Framework Android ID' field.")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("\nNote: It may take some time (usually a few minutes, but sometimes up to 2 hours) for the registration to take effect.")
	fmt.Println("After registration, you may need to restart the container.")

	return nil
}
