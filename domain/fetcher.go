package domain

import "context"

type Fetcher interface {
	FetchFeed(ctx context.Context, url string) (*Feed, error)
}
