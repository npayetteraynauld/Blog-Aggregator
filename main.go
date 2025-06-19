package main

import (
	"fmt"
	"os"
	"database/sql"

	_"github.com/lib/pq"
	"github.com/npayetteraynauld/Blog-Aggregator/internal/config"
	"github.com/npayetteraynauld/Blog-Aggregator/internal/database"
)

func main() {
	//Read JSON to config struct
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)	
	}

	//open connection to database
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Errorf("Error opening database: %v", err)
	}
	
	dbQueries := database.New(db)

	//initialize state struct
	s := state{
		db: dbQueries,
		cfg: &cfg,
	}

	//initialize commands struct
	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}

	//register handlers in commands
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

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
}
