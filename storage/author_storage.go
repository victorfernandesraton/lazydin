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

	selectAuthorByNameQuery = `
		SELECT url, name, description, created_at, updated_at FROM authors WHERE name = ?;
	`

	selectAllAuthors = `
		SELECT url, name, description, created_at, updated_at FROM authors;
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

func (as *AuthorStorage) GetByName(name string) ([]domain.Author, error) {
	var authors []domain.Author
	rows, err := as.db.Query(selectAuthorByNameQuery, "'%"+name+"%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var author domain.Author
		err = rows.Scan(&author.Url, &author.Name, &author.Description, &author.CreatedAt, &author.UpdatedAt)

		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}
	return authors, nil
}

func (as *AuthorStorage) GetAll() ([]domain.Author, error) {
	var authors []domain.Author
	rows, err := as.db.Query(selectAllAuthors)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var author domain.Author
		err = rows.Scan(&author.Url, &author.Name, &author.Description, &author.CreatedAt, &author.UpdatedAt)

		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}
	return authors, nil
}
