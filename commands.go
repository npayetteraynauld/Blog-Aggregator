package main

import (
	"fmt"

	"github.com/npayetteraynauld/Blog-Aggregator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	//checking existence
	fn, exists := c.commands[cmd.name]
	if !exists {
		return fmt.Errorf("No function with specified name")
	}

	return fn(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	_, exists := c.commands[name]
	if exists {
		fmt.Println("Function already registered")
	}

	c.commands[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No arguments provided, need username")
	} else if len(cmd.args) > 1 {
		return fmt.Errorf("Too many arguments provided, only need username")
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Println("User has been set")
	return nil
}
