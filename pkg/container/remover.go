package container

import (
	"fmt"
	"os"
	"reddock/pkg/config"
	"reddock/pkg/ui"
)

type Remover struct {
	config        *config.Config
	containerName string
	runtime       Runtime
}

func NewRemover(containerName string) *Remover {
	cfg, _ := config.Load()
	return &Remover{
		config:        cfg,
		containerName: containerName,
		runtime:       NewRuntime(),
	}
}

func (r *Remover) Remove(removeImage bool) error {
	if err := CheckRoot(); err != nil {
		return err
	}

	container := r.config.GetContainer(r.containerName)
	if container == nil {
		return fmt.Errorf("Container '%s' not found", r.containerName)
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{
			name: fmt.Sprintf("Stopping and removing container '%s'", container.Name),
			fn: func() error {
				if r.runtime.IsRunning(container.Name) {
					r.runtime.Stop(container.Name)
				}
				if err := r.runtime.Remove(container.Name, true); err != nil {
					fmt.Printf("\nWarning: Failed to remove container: %v\n", err)
				}
				return nil
			},
		},
		{
			name: fmt.Sprintf("Removing data directory: %s", container.GetDataPath()),
			fn: func() error {
				if err := os.RemoveAll(container.GetDataPath()); err != nil {
					fmt.Printf("\nWarning: Could not remove data directory: %v\n", err)
				}
				return nil
			},
		},
	}

	if removeImage {
		steps = append(steps, struct {
			name string
			fn   func() error
		}{
			name: fmt.Sprintf("Removing image: %s", container.ImageURL),
			fn: func() error {
				if err := r.runtime.RemoveImage(container.ImageURL); err != nil {
					fmt.Printf("\nWarning: Could not remove image: %v\n", err)
				}
				return nil
			},
		})
	}

	steps = append(steps, struct {
		name string
		fn   func() error
	}{
		name: "Updating configuration",
		fn: func() error {
			r.config.RemoveContainer(container.Name)
			if err := config.Save(r.config); err != nil {
				return fmt.Errorf("Failed to save config: %v", err)
			}
			return nil
		},
	})

	bar := ui.NewProgressBar(len(steps), "Removing...")
	bar.Start()

	for _, step := range steps {
		bar.SetMessage(step.name)
		if err := step.fn(); err != nil {
			return err
		}
		bar.Increment()
	}
	bar.Finish(fmt.Sprintf("Container '%s' removed successfully", container.Name))

	return nil
}
