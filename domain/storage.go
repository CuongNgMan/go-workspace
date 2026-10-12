package domain

type ArticleReader interface {
	GetRecent(n int) []*Article
	GetByID(id string) (*Article, error)
}

type Storage interface {
	ArticleReader
	AddArticles(articles []*Article) error
}

