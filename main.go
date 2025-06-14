package main

import (
	"fmt"
	"os"

	"github.com/npayetteraynauld/Blog-Aggregator/internal/config"
)

func main() {
	//Read JSON to config struct
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)	
	}
	
	//initialize state struct
	s := state{
		cfg: &cfg,
	}

	//initialize commands struct
	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}

	//register login handler in commands
	cmds.register("login", handlerLogin)

	//parsing arguments
	arguments := os.Args
	if len(arguments) < 2 {
		fmt.Println("No arguments provided")
		os.Exit(1)
	}
	funcName := arguments[1]
	args := arguments [2:]

	//creating command
	cmd := command{
		name: funcName,
		args: args,
	}

	//run command
	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}


	
	cfg, err = config.Read()
	fmt.Println(cfg)
}
