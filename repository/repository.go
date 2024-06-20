package repository

import "github.com/victorfernandesraton/lazydin/domain"

type PostStore interface {
	Upsert(post *domain.Post) *domain.Post
	GetByID(id uint64) (*domain.Post, error)
	GetByUrl(url string) (*domain.Post, error)
}
