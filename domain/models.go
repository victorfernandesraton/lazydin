package domain

import "time"

type Post struct {
	ID        uint64    `csv:"-"`
	Url       string    `csv:"url"`
	Content   string    `csv:"content"`
	AuthorUrl string    `csv:"author_url"`
	AuthorId  uint64    `csv:"-"`
	CreatedAt time.Time `csv:"-"`
	UpdatedAt time.Time `csv:"-"`
}

type Author struct {
	ID          uint64    `csv:"-"`
	Url         string    `csv:"url"`
	Name        string    `csv:"name"`
	Description string    `csv:"description"`
	CreatedAt   time.Time `csv:"-"`
	UpdatedAt   time.Time `csv:"-"`
}

type Content struct {
	Post   Post   `csv:"post"`
	Author Author `csv:"author"`
}
