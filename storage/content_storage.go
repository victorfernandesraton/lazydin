package storage

import (
	"database/sql"
	"fmt"

	"github.com/victorfernandesraton/lazydin/domain"
)

type ContentStorage struct {
	db *sql.DB
}

// Upsert both authors and posts in a single transaction
func (store *ContentStorage) UpsertPostsAndAuthors(content []domain.Content) error {
	authors := make(map[string]*domain.Author)
	for _, item := range content {
		author, exists := authors[item.Author.Url]
		if !exists {
			authors[author.Url] = &item.Author
		}
	}
	tx, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Map to store author IDs by URL
	authorIDMap := make(map[string]int64)

	// Upsert authors
	for _, author := range authors {
		upsertAuthorStmt := `
        INSERT INTO authors (url, name, description)
        VALUES (?, ?, ?)
        ON CONFLICT(url) DO UPDATE SET
        updated_at = CURRENT_TIMESTAMP;
        `
		res, err := tx.Exec(upsertAuthorStmt, author.Url, author.Name, author.Description)
		if err != nil {
			return fmt.Errorf("error upserting author %s: %w", author.Url, err)
		}

		// Get the author ID
		id, err := res.LastInsertId()
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error getting author ID for %s: %w", author.Url, err)
		}

		// If no new ID, get the existing ID
		if err == sql.ErrNoRows {
			var existingID int64
			err = tx.QueryRow(`SELECT id FROM authors WHERE url = ?`, author.Url).Scan(&existingID)
			if err != nil {
				return fmt.Errorf("error retrieving existing author ID for %s: %w", author.Url, err)
			}
			id = existingID
		}

		authorIDMap[author.Url] = id
	}

	// Upsert posts
	for _, item := range content {
		authorID, exists := authorIDMap[item.Post.AuthorUrl]
		if !exists {
			return fmt.Errorf("author ID not found for URL: %s", item.Post.AuthorUrl)
		}

		upsertPostStmt := `
        INSERT INTO posts (url, content, author_id)
        VALUES (?, ?, ?)
        ON CONFLICT(url) DO UPDATE SET
        updated_at = CURRENT_TIMESTAMP;
        `
		_, err := tx.Exec(upsertPostStmt, item.Post.Url, item.Post.Content, authorID)
		if err != nil {
			return fmt.Errorf("error upserting post %s: %w", item.Post.Url, err)
		}
	}

	return nil
}
