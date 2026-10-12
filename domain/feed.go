package domain

type Feed struct {
	Title       string
	Description string
	Link        string
	Articles    []*Article
}
