package entity

import (
	"time"
)

type Feed struct {
	Pk       int    `json:"pk"`
	FeedUrl string	`json:"feed_url"`			
}

type Article struct {
	Pk         int  	 `json:"pk"`
	Title      string	 `json:"title"`
	Content    string	 `json:"content"`
	SourceUrl  string    `json:"source_url"`
	Updated    time.Time `json:"updated"`
	FeedPk     int	     `json:"feed_pk"`
}