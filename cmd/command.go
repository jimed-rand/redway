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
	case "prepare-lxc":
		return c.executePrepareLXC()
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

func (c *Command) executePrepareLXC() error {
	preparer := container.NewLXCPreparer()
	return preparer.PrepareLXC()
}

func (c *Command) executeInit() error {
	var containerName string
	var image string

	// Parse arguments: init [container-name] [image-url]
	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	if len(c.Args) > 1 {
		image = c.Args[1]
	} else {
		image = config.DefaultImageURL
	}

	init := container.NewInitializer(containerName, image)
	return init.Initialize()
}

func (c *Command) executeStart() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Start()
}

func (c *Command) executeStop() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Stop()
}

func (c *Command) executeRestart() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Restart()
}

func (c *Command) executeStatus() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	status := utils.NewStatusManager(containerName)
	return status.Show()
}

func (c *Command) executeShell() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	shell := utils.NewShellManager(containerName)
	return shell.Enter()
}

func (c *Command) executeAdbConnect() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	adb := utils.NewAdbManager(containerName)
	return adb.ShowConnection()
}

func (c *Command) executeRemove() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Remove()
}

func (c *Command) executeList() error {
	lister := container.NewLister()
	return lister.ListRedwayContainers()
}

func (c *Command) executeLog() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		containerName = config.DefaultContainerName
	}

	logger := utils.NewLogManager(containerName)
	return logger.Show()
}
