package cmd

import (
	"fmt"
	
	"redway/pkg/config"
	"redway/pkg/container"
	"redway/pkg/utils"
)

type Command struct {
	Name string
	Args []string
}

func NewCommand(name string, args []string) *Command {
	return &Command{
		Name: name,
		Args: args,
	}
}

func (c *Command) Execute() error {
	switch c.Name {
	case "init":
		return c.executeInit()
	case "start":
		return c.executeStart()
	case "stop":
		return c.executeStop()
	case "restart":
		return c.executeRestart()
	case "status":
		return c.executeStatus()
	case "shell":
		return c.executeShell()
	case "adb-connect":
		return c.executeAdbConnect()
	case "remove":
		return c.executeRemove()
	case "list":
		return c.executeList()
	case "log":
		return c.executeLog()
	default:
		return fmt.Errorf("unknown command: %s", c.Name)
	}
}

func (c *Command) executeInit() error {
	var image string
	if len(c.Args) > 0 {
		image = c.Args[0]
	} else {
		image = config.DefaultImageURL
	}
	
	init := container.NewInitializer(image)
	return init.Initialize()
}

func (c *Command) executeStart() error {
	mgr := container.NewManager()
	return mgr.Start()
}

func (c *Command) executeStop() error {
	mgr := container.NewManager()
	return mgr.Stop()
}

func (c *Command) executeRestart() error {
	mgr := container.NewManager()
	return mgr.Restart()
}

func (c *Command) executeStatus() error {
	status := utils.NewStatusManager()
	return status.Show()
}

func (c *Command) executeShell() error {
	shell := utils.NewShellManager()
	return shell.Enter()
}

func (c *Command) executeAdbConnect() error {
	adb := utils.NewAdbManager()
	return adb.ShowConnection()
}

func (c *Command) executeRemove() error {
	mgr := container.NewManager()
	return mgr.Remove()
}

func (c *Command) executeList() error {
	lister := container.NewLister()
	return lister.List()
}

func (c *Command) executeLog() error {
	logger := utils.NewLogManager()
	return logger.Show()
}
