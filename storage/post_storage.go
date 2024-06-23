package storage

import (
	"database/sql"
	"time"

	"github.com/victorfernandesraton/lazydin/domain"
)

const (
	createPostTableQuery = `
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT UNIQUE,
			content TEXT,
			author_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(author_id) REFERENCES authors(id)
		);
	`

	upsertPostQuery = `
		INSERT INTO posts (url, content, author_id, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET content=excluded.content, author_id=excluded.author_id, updated_at=excluded.updated_at
		RETURNING id;
	`

	selectPostByIdQuery = `
		SELECT id, url, content, author_id, created_at, updated_at FROM posts WHERE id = ?;
	`

	selectPostByUrlQuery = `
		SELECT id, url, content, author_id, created_at, updated_at FROM posts WHERE url = ?;
	`
)

type PostStorage struct {
	db *sql.DB
}

func NewPostStorage(db *sql.DB) *PostStorage {
	return &PostStorage{db: db}
}

func (ps *PostStorage) CreateTable() error {
	_, err := ps.db.Exec(createPostTableQuery)
	return err
}

func (ps *PostStorage) Upsert(post *domain.Post) (*domain.Post, error) {
	now := time.Now()
	err := ps.db.QueryRow(upsertPostQuery, post.Url, post.Content, post.AuthorId, now, now).
		Scan(&post.ID)
	if err != nil {
		return nil, err
	}
	return ps.GetById(post.ID)
}

func (ps *PostStorage) GetById(id uint64) (*domain.Post, error) {
	var post domain.Post
	err := ps.db.QueryRow(selectPostByIdQuery, id).
		Scan(&post.ID, &post.Url, &post.Content, &post.AuthorId, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (ps *PostStorage) GetByUrl(url string) (*domain.Post, error) {
	var post domain.Post
	err := ps.db.QueryRow(selectPostByUrlQuery, url).
		Scan(&post.ID, &post.Url, &post.Content, &post.AuthorId, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}
