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
		ON CONFLICT(url) DO UPDATE SET 
		 	content=excluded.content,
			author_url=excluded.author_url,
			updated_at=CURRENT_TIMESTAMP
		RETURNING url;
	`

	selectPostsQuery = `
            SELECT url, content, author_url, created_at, updated_at FROM posts;
        `

	selectPostsQueryByAuthorUrl = `
            SELECT url, content, author_url, created_at, updated_at FROM posts WHERE author_url = ?;
        `
	selectPostsQueryByAuthorNameLike = `
            SELECT
                    p.url,
                    p.content,
                    p.author_url,
                    p.created_at,
                    p.updated_at
            FROM
                    posts p
            INNER JOIN authors a ON
                    a.url = p.author_url
            WHERE a.name LIKE $1`
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

func (ps *PostStorage) GetAllPostsByAuthorUrl(url string) ([]domain.Post, error) {

	var posts []domain.Post
	rows, err := ps.db.Query(selectPostsQueryByAuthorUrl, url)
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

func (ps *PostStorage) GetAllPostsByAuthorName(name string) ([]domain.Post, error) {
	var posts []domain.Post
	rows, err := ps.db.Query(selectPostsQueryByAuthorNameLike, "%"+name+"%")

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
