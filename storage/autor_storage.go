package storage

import (
	"database/sql"
	"time"

	"github.com/victorfernandesraton/lazydin/domain"
)

const (
	createAuthorTableQuery = `
		CREATE TABLE IF NOT EXISTS authors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT UNIQUE,
			name TEXT,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	upsertAuthorQuery = `
		INSERT INTO authors (url, name, description, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET name=excluded.name, description=excluded.description, updated_at=excluded.updated_at
		RETURNING id;
	`

	selectAuthorByIdQuery = `
		SELECT id, url, name, description, created_at, updated_at FROM authors WHERE id = ?;
	`

	selectAuthorByUrlQuery = `
		SELECT id, url, name, description, created_at, updated_at FROM authors WHERE url = ?;
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
	now := time.Now()
	err := as.db.QueryRow(upsertAuthorQuery, author.Url, author.Name, author.Description, now, now).
		Scan(&author.ID)
	if err != nil {
		return nil, err
	}
	return as.GetById(author.ID)
}

func (as *AuthorStorage) GetById(id uint64) (*domain.Author, error) {
	var author domain.Author
	err := as.db.QueryRow(selectAuthorByIdQuery, id).
		Scan(&author.ID, &author.Url, &author.Name, &author.Description, &author.CreatedAt, &author.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &author, nil
}

func (as *AuthorStorage) GetByUrl(url string) (*domain.Author, error) {
	var author domain.Author
	err := as.db.QueryRow(selectAuthorByUrlQuery, url).
		Scan(&author.ID, &author.Url, &author.Name, &author.Description, &author.CreatedAt, &author.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &author, nil
}
