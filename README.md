# Blog-Aggregator

A simple command-line blog aggregator written in Go. Fetches RSS feeds, parses posts, and stores 
them in a database using Postgres 

--

## Requirements

You will need to have the following installed:
- Go
- PostgreSQL

--

## Installation

To use this program, run: 

```bash
go install ./...
```

While in the repo directory

You will also need to create a configuration file at: 

~./gatorconfig.json

--

## Usage

Once install, you can run the program:

Blog-Aggregator [command] [flags]

Commands:

- reset (reset all data from database)

- register (register user to database)

- login (login into registered user, flag = user)

- users (list of all registered users)

- addfeed (add a feed to database linked to current user, flag1 = feed name, flag2 = feed url)

- feeds (list of all registered feeds)

- agg (Aggregate posts from all registered feeds, flag = interval (ex: 1s, 1m, 1h))

- follow (follow specified feed with current user, flag = feed url)

- following (list of all followed feeds)

- unfollow (unfollow specified feed, flag = feed url)

- browse (prints most recent posts to stdout, flag = limit to query)

