package entity

import (
	"time"
)

type Feed struct {
	Pk      int    `json:"pk"`
	FeedUrl string `json:"feed_url"`
}

type Article struct {
	Pk        int       `json:"pk"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	SourceUrl string    `json:"source_url"`
	Published time.Time `json:"published"`
	FeedPk    int       `json:"feed_pk"`
}
