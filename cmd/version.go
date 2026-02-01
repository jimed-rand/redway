package cmd

import "fmt"

var Version = "2.22.4"

func (c *Command) executeVersion() error {
	fmt.Printf("Reddock version %s\n", Version)
	return nil
}
