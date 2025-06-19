package main

import (
	"net/http"
	"time"
	"log"
	"io"
	"encoding/xml"
	"context"
	"html"
	"fmt"
	"database/sql"

	"github.com/google/uuid"
	"github.com/npayetteraynauld/Blog-Aggregator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	//define client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	//Create request
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	//Set header
	req.Header.Set("User-Agent", "gator")

	//Make the request
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making the request: %v", err)
	}
	defer res.Body.Close()

	//Read the response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading the response: %v", err)
	}

	//Unmarshal into structs
	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		log.Fatalf("Error unmarshalling: %v", err)
	}

	//unescaping strings
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func scrapeFeeds(s *state) error {
	//Get next feed to fetch
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting next feed to fetch: %v", err)
	}

	//Mark feed fetched
	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
		ID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error marking feed fetched: %v", err)
	}

	//Fetch feed
	rssfeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("Error fetching feed: %v", err)
	}

	//saving posts from feeds
	for _, item := range rssfeed.Channel.Item {
		//parsing PubDate into time.Time
		t, err := time.Parse(time.RFC1123, item.PubDate)
		
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: item.Title,
			Url: item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid: true,
			},
			PublishedAt: t,
			FeedID: feed.ID,
		})
		if err != nil && err.Error() != `pq: duplicate key value violates unique constraint "posts_url_key"`{
			return fmt.Errorf("Error creating post: %v", err)
		}
	}
	return nil
}
