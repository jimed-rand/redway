package container

import (
	"fmt"
	"reddock/pkg/ui"
)

type Pruner struct {
	runtime Runtime
}

func NewPruner() *Pruner {
	return &Pruner{
		runtime: NewRuntime(),
	}
}

func (p *Pruner) Prune() error {
	if err := CheckRoot(); err != nil {
		return err
	}

	msg := fmt.Sprintf("Pruning unused images using %s...", p.runtime.Name())
	s := ui.NewSpinner(msg)
	s.Start()
	output, err := p.runtime.PruneImages()
	if err != nil {
		return fmt.Errorf("Failed to prune images: %v", err)
	}
	s.Finish("Unused images have been pruned successfully")

	if output != "" {
		fmt.Println(output)
	}

	return nil
}
