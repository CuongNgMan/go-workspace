package domain

import "time"

type Article struct {
	ID          string
	Title       string
	Description string
	Link        string
	Published   *time.Time
	FeedTitle   string
}


