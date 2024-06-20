package domain

type Post struct {
	ID        uint64 `csv:-`
	Url       string `csv:"url"`
	Content   string `csv:"content"`
	AuthorUrl string `csv:"author_url"`
}

type Author struct {
	ID          uint64 `csv:-`
	Url         string `csv:"url"`
	Name        string `csv:"name"`
	Description string `csv:"description"`
}

type Content struct {
	Post   Post   `csv:"post"`
	Author Author `csv:"author"`
}
