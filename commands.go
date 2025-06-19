package main

import (
	"fmt"
	"time"
	"os"
	"log"
	"context"
	"strconv"
	"strings"
	
	"golang.org/x/net/html"
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

func handlerAgg(s *state, cmd command) error {
	//check for args
	if len(cmd.args) == 0 {
		return fmt.Errorf("No arguments provided, need time between requests (ex: 1s, 1m, 1h)")
	} else if len(cmd.args) > 1 {
		return fmt.Errorf("Too many arguments provided, only need time between requests (ex: 1s, 1m, 1h)")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error parsing duration %v", err)
	}

	fmt.Println("Collecting feeds every " + cmd.args[0])
	
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("Need to provide name and url")
	}

	//Create feed
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.args[0],
		Url: cmd.args[1],
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed: %v", err)
	}

	//Create Feed_Follow record
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed_follow record: %v", err)
	}

	fmt.Println("New feed:")
	fmt.Printf("  - ID: %v\n", feed.ID)
	fmt.Printf("  - CreatedAt: %v\n", feed.CreatedAt)
	fmt.Printf("  - UpdatedAt: %v\n", feed.UpdatedAt)
	fmt.Printf("  - Name: %v\n", feed.Name)
	fmt.Printf("  - Url: %v\n", feed.Url)
	fmt.Printf("  - UserID: %v\n", feed.UserID)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting feeds: %v", err)
	}
	
	fmt.Println("Feeds:")
	for _, feed := range feeds {
		name, err := s.db.GetUserNameFromID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("Error getting user name from id: %v", err)
		}

		fmt.Printf("  - Name: %v\n", feed.Name)
		fmt.Printf("  - Url: %v\n", feed.Url)
		fmt.Printf("  - Added By: %v\n\n", name)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	//check for args
	if len(cmd.args) == 0 {
		return fmt.Errorf("No arguments provided, need url")
	} else if len(cmd.args) > 1 {
		return fmt.Errorf("Too many arguments provided, only need url")
	}

	//Get FeedID
	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error getting feed: %v", err)
	}
	feedID := feed.ID

	//Create Feed_Follow record
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feedID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed_follow record: %v", err)
	}

	fmt.Printf("Feed name: %v\n", feed.Name)
	fmt.Printf("Current user: %v\n", user.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	//get following feeds for current user
	followingFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Error getting following feeds: %v", err)
	}

	fmt.Println("Following:")
	for _, feedfollow := range followingFeeds {
		fmt.Println("  - " + feedfollow.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	//check for args
	if len(cmd.args) == 0 {
		return fmt.Errorf("No arguments provided, need url")
	} else if len(cmd.args) > 1 {
		return fmt.Errorf("Too many arguments provided, only need url")
	}

	//Get FeedID
	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error getting feed: %v", err)
	}

	//delete record of feed_follow
	err = s.db.Unfollow(context.Background(), database.UnfollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error unfollowing: %v", err)
	}
	
	fmt.Printf("Unfollowing %v\n", feed.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	//check for limit arg
	var limit int32
	if len(cmd.args) == 0 {
		limit = 2
	} else if len(cmd.args) > 0 {
		v, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			//couldn't parse
			limit = 2
		} else {
			limit = int32(v)
		}
	}

	//Print posts for user with provided limit arg
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: limit,
	})
	if err != nil {
		return fmt.Errorf("Error getting posts for user: %v", err)
	}

	fmt.Printf("Most recent %v posts:\n", limit)
	for _, post := range posts {
		fmt.Println()
		fmt.Printf("  - Title: %v\n", post.Title)
		fmt.Printf("  - Published at: %v\n", post.PublishedAt)
		fmt.Printf("  - Link: %v\n", post.Url)
		fmt.Printf("  - Description: %v\n", stripHTML(post.Description.String))
	}

	return nil
}

func stripHTML(input string) string {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return ""
	}
	var text string
	recursivestripHTML(doc, &text)
	return text
}

func recursivestripHTML(node *html.Node, text *string) {
	if node == nil {
		return
	}

	if node.Type == html.TextNode {
		*text = *text + node.Data 
	}

	if node.FirstChild != nil {
		recursivestripHTML(node.FirstChild, text)
	}

	if node.NextSibling != nil {
		recursivestripHTML(node.NextSibling, text)
	}

	return
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		//get user
		u, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Error getting user: %v", err)
		}
		
		//pass in user to logged in handlers
		err = handler(s, cmd, u)
		if err != nil {
			return err
		}

		return nil
	}
}
