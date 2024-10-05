package storage

import (
	"database/sql"
	"log"

	"github.com/victorfernandesraton/lazydin/domain"
)

const (
	createPostTableQuery = `
		CREATE TABLE IF NOT EXISTS posts (
			url TEXT UNIQUE PRIMARY KEY,
			content TEXT,
                        author_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(author_url) REFERENCES authors(url)
		);
	`

	upsertPostQuery = `
		INSERT INTO posts (url, content, author_url) 
		VALUES (?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET updated_at=CURRENT_TIMESTAMP
		RETURNING url;
	`

	selectPostsQuery = `
            SELECT url, content, author_url, created_at, updated_at FROM posts;
        `

	selectPostByIdQuery = `
		SELECT url, content, author_url, created_at, updated_at FROM posts WHERE id = ?;
	`

	selectPostByUrlQuery = `
		SELECT url, content, author_url, created_at, updated_at FROM posts WHERE url = ?;
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
	err := ps.db.QueryRow(upsertPostQuery, post.Url, post.Content, post.AuthorUrl).
		Scan(&post.Url)
	if err != nil {
		return nil, err
	}
	return ps.GetByUrl(post.Url)
}

func (ps *PostStorage) GetAllPosts() ([]domain.Post, error) {
	var posts []domain.Post
	rows, err := ps.db.Query(selectPostsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var post domain.Post
		err = rows.Scan(&post.Url, &post.Content, &post.AuthorUrl, &post.CreatedAt, &post.UpdatedAt)
		log.Println(err)
		if err != nil {
			return nil, err
		}
		log.Println(post)
		posts = append(posts, post)
	}
	return posts, nil
}

func (ps *PostStorage) GetByUrl(url string) (*domain.Post, error) {
	var post domain.Post
	err := ps.db.QueryRow(selectPostByUrlQuery, url).
		Scan(&post.Url, &post.Content, &post.AuthorUrl, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}
