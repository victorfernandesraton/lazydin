package linkedisney

type Post struct {
	Url       string
	Content   string
	AuthorUrl string
}

type Author struct {
	Url         string
	Name        string
	Description string
}
