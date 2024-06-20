package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/victorfernandesraton/lazydin/domain"
)

const (
	sqlInsertPost  = "INSERT INTO posts (url, content, author_url) VALUES (?, ?, ?)"
	sqlUpdatePost  = "UPDATE posts SET url = ?, content = ?, author_url = ? WHERE id = ?"
	sqlSelectByID  = "SELECT id, url, content, author_url FROM posts WHERE id = ?"
	sqlSelectByUrl = "SELECT id, url, content, author_url FROM posts WHERE url = ?"
)

type SQLitePostStore struct {
	db *sql.DB
}

func NewSQLitePostStore(db *sql.DB) *SQLitePostStore {
	return &SQLitePostStore{db: db}
}

func (s *SQLitePostStore) Upsert(post *domain.Post) error {
	existingPost, err := s.GetByID(post.ID)
	if err != nil {
		return fmt.Errorf("error checking existing post by ID: %w", err)
	}

	if existingPost == nil {
		existingPost, err = s.GetByUrl(post.Url)
		if err != nil {
			return fmt.Errorf("error checking existing post by URL: %w", err)
		}
	}

	var execErr error
	if existingPost == nil {
		_, execErr = s.db.Exec(sqlInsertPost, post.Url, post.Content, post.AuthorUrl)
	} else {
		_, execErr = s.db.Exec(sqlUpdatePost, post.Url, post.Content, post.AuthorUrl, existingPost.ID)
	}

	if execErr != nil {
		return fmt.Errorf("error executing upsert: %w", execErr)
	}

	return nil
}

func (s *SQLitePostStore) GetByID(id uint64) (*domain.Post, error) {
	post := &domain.Post{}
	err := s.db.QueryRow(sqlSelectByID, id).Scan(&post.ID, &post.Url, &post.Content, &post.AuthorUrl)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error querying post by ID: %w", err)
	}
	return post, nil
}

func (s *SQLitePostStore) GetByUrl(url string) (*domain.Post, error) {
	post := &domain.Post{}
	err := s.db.QueryRow(sqlSelectByUrl, url).Scan(&post.ID, &post.Url, &post.Content, &post.AuthorUrl)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error querying post by URL: %w", err)
	}
	return post, nil
}
