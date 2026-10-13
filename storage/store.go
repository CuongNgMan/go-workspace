package storage

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/CuongNgMan/go-workspace/domain"
	"go.etcd.io/bbolt"
)

const articlesBucketName = "articles"

// Compile-time checking to make BoltStore implement domain.Storage interface
var _ domain.Storage = (*BoltStore)(nil)

type BoltStore struct {
	db *bbolt.DB
}

func NewBoltStore(dbPath string, readOnly bool) (*BoltStore, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: readOnly,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if !readOnly {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(articlesBucketName))
			return err
		})

		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return &BoltStore{db: db}, nil
}

func (b *BoltStore) Close() error {
	return b.db.Close()
}

// AddArticles implements [domain.Storage].
func (b *BoltStore) AddArticles(articles []*domain.Article) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucketName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}
		for _, article := range articles {
			data, err := json.Marshal(article)
			if err != nil {
				return fmt.Errorf("failed to marshal article %s: %w", article.ID, err)
			}

			if err := bucket.Put([]byte(article.ID), data); err != nil {
				return fmt.Errorf("faild to store article %s: %w", article.ID, err)
			}
		}
		return nil
	})
}

// GetByID implements [domain.Storage].
func (b *BoltStore) GetByID(id string) (*domain.Article, error) {
	var article *domain.Article
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucketName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}
		data := bucket.Get([]byte(id))
		if data == nil {
			return nil
		}
		article = &domain.Article{}
		if err := json.Unmarshal(data, article); err != nil {
			return fmt.Errorf("failed to unmarshal article: %w", err)
		}
		return nil
	})
	return article, err
}

// GetRecent implements [domain.Storage].
func (b *BoltStore) GetRecent(n int) []*domain.Article {
	var articles []*domain.Article

	b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucketName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}
		bucket.ForEach(func(k, v []byte) error {
			var article *domain.Article
			if err := json.Unmarshal(v, &article); err != nil {
				return fmt.Errorf("failed to unmarshal article: %w", err)
			}
			articles = append(articles, article)
			return nil
		})
		return nil
	})

	slices.SortFunc(articles, func(a, b *domain.Article) int {
		if a.Published == nil && b.Published == nil {
			return 0
		}
		if a.Published == nil {
			return 1
		}
		if b.Published == nil {
			return -1
		}
		return b.Published.Compare(*a.Published)
	})
	if n > len(articles) {
		n = len(articles)
	}
	return articles[:n]
}
