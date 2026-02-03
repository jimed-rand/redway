package utils

import (
	"fmt"
	"os"
	"os/exec"

	"reddock/pkg/config"
	"reddock/pkg/container"
)

type LogManager struct {
	config        *config.Config
	containerName string
	runtime       container.Runtime
}

func NewLogManager(containerName string) *LogManager {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Failed to load config: %v\n", err)
		cfg = config.GetDefault()
	}
	return &LogManager{
		config:        cfg,
		containerName: containerName,
		runtime:       container.NewRuntime(),
	}
}

func (l *LogManager) Show() error {
	if err := container.CheckRoot(); err != nil {
		return err
	}
	cont := l.config.GetContainer(l.containerName)
	if cont == nil {
		return fmt.Errorf("container '%s' not found", l.containerName)
	}

	fmt.Printf("Showing the logs for container: %s\n", l.containerName)
	fmt.Println("Press Ctrl+C to exit")

	cmd := exec.Command(l.runtime.Name(), "logs", "-f", l.containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
