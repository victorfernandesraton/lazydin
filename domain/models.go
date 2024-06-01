package domain

type Post struct {
	Url       string `csv:"url"`
	Content   string `csv:"content"`
	AuthorUrl string `csv:"author_url"`
}

type Author struct {
	Url         string `csv:"url"`
	Name        string `csv:"name"`
	Description string `csv:"description"`
}

type Content struct {
	Post   Post   `csv:"post"`
	Author Author `csv:"author"`
}
