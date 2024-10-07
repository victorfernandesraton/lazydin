package storage

import (
	"database/sql"

	"github.com/victorfernandesraton/lazydin/domain"
)

const (
	createAuthorTableQuery = `
		CREATE TABLE IF NOT EXISTS authors (
			url TEXT PRIMARY KEY UNIQUE,
			name TEXT,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	upsertAuthorQuery = `
		INSERT INTO authors (url, name, description) 
		VALUES (?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET 
			name=excluded.name,  
			updated_at=CURRENT_TIMESTAMP,
			description=excluded.description
		RETURNING url;
	`

	selectAuthorByUrlQuery = `
		SELECT url, name, description, created_at, updated_at FROM authors WHERE url = ?;
	`
)

type AuthorStorage struct {
	db *sql.DB
}

func NewAuthorStorage(db *sql.DB) *AuthorStorage {
	return &AuthorStorage{db: db}
}

func (as *AuthorStorage) CreateTable() error {
	_, err := as.db.Exec(createAuthorTableQuery)
	return err
}

func (as *AuthorStorage) Upsert(author *domain.Author) (*domain.Author, error) {
	err := as.db.QueryRow(upsertAuthorQuery, author.Url, author.Name, author.Description).
		Scan(&author.Url)
	if err != nil {
		return nil, err
	}
	return as.GetByUrl(author.Url)
}

func (as *AuthorStorage) GetByUrl(url string) (*domain.Author, error) {
	var author domain.Author
	err := as.db.QueryRow(selectAuthorByUrlQuery, url).
		Scan(&author.Url, &author.Name, &author.Description, &author.CreatedAt, &author.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &author, nil
}
