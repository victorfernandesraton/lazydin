package domain

import "time"

type Post struct {
	Url       string    `csv:"url"`
	Content   string    `csv:"content"`
	AuthorUrl string    `csv:"author_url"`
	CreatedAt time.Time `csv:"-"`
	UpdatedAt time.Time `csv:"-"`
}

type Author struct {
	Url         string    `csv:"url"`
	Name        string    `csv:"name"`
	Description string    `csv:"description"`
	CreatedAt   time.Time `csv:"-"`
	UpdatedAt   time.Time `csv:"-"`
}

type Relationship struct {
	AuthorUrl string `csv:"author_url"`
	Relation  string `csv:"relation"`
	Mutuals   bool   `csv:"mutuals"`
}

type Content struct {
	Post   Post   `csv:"post"`
	Author Author `csv:"author"`
}

type Task struct {
	Id        int64     `csv:"id"`
	Status    int       `csv:"status"`
	CreatedAt time.Time `csv:"created_at"`
	UpdatedAt time.Time `csv:"updated_at"`
}

type TaskPost struct {
	TaskId    int64     `csv:"task_id"`
	PostUrl   string    `csv:"post_url"`
	CreatedAt time.Time `csv:"created_at"`
	UpdatedAt time.Time `csv:"updated_at"`
}
