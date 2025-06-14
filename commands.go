package main

import (
	"fmt"
	"time"
	"os"
	"log"
	"context"

	"github.com/google/uuid"
	"github.com/npayetteraynauld/Blog-Aggregator/internal/config"
	"github.com/npayetteraynauld/Blog-Aggregator/internal/database"
)

type state struct {
	db *database.Queries
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
	//check for args
	if len(cmd.args) == 0 {
		return fmt.Errorf("No arguments provided, need username")
	} else if len(cmd.args) > 1 {
		return fmt.Errorf("Too many arguments provided, only need username")
	}

	//check if user exists in database
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("No users with specified name")
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Println("User has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	//check for args
	if len(cmd.args) == 0 {
		return fmt.Errorf("No arguments provided, need name")
	} else if len(cmd.args) > 1 {
		return fmt.Errorf("Too many arguments provided, only need name")
	}
	
	//check if user already exists
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		fmt.Println("User already exists")
		os.Exit(1)
	}

	//create User
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.args[0],
	})
	if err != nil {
		return fmt.Errorf("Error creating User: %v", err)
	}

	//change current user
	s.cfg.SetUser(cmd.args[0])

	fmt.Println("User successfully created")
	log.Println(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteRecords(context.Background())
	if err != nil {
		fmt.Errorf("Error Resetting data: %v", err)
		os.Exit(1)
	}

	if err == nil {
		fmt.Println("successfully reset data")
		os.Exit(0)
	}

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error returning users: %v", err)
	}

	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}

	return nil
}
