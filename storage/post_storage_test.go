package storage_test

import (
	"database/sql"
	"testing"

	"github.com/victorfernandesraton/lazydin/domain"
	"github.com/victorfernandesraton/lazydin/storage"
)

func TestPostStorage(t *testing.T) {
	databse, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf(err.Error())
	}
	postStorage := storage.NewPostStorage(databse)
	authorStorage := storage.NewAuthorStorage(databse)
	author := &domain.Author{
		Url: "some-url", Name: "Victor Raton",
	}
	other_author := &domain.Author{
		Url: "some-other-url", Name: "Captain Jack Sparrow",
	}
	t.Run("create table", func(t *testing.T) {
		if err := authorStorage.CreateTable(); err != nil {
			t.Fatalf(err.Error())
		}
		if err := postStorage.CreateTable(); err != nil {
			t.Fatalf(err.Error())
		}

		author, err = authorStorage.Upsert(author)
		if err != nil {
			t.Fatalf(err.Error())
		}

		other_author, err = authorStorage.Upsert(other_author)
		if err != nil {
			t.Fatalf(err.Error())
		}
	})
	t.Run("create post", func(t *testing.T) {
		post := &domain.Post{
			Url: "some-url", Content: "some-content", AuthorUrl: author.Url,
		}
		post, err := postStorage.Upsert(post)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if post.Url != "some-url" {
			t.Fatalf("Post url error, expect %s, got %s", "some-url", post.Url)
		}
		if post.Content != "some-content" {
			t.Fatalf("Post content error, expect %s, got %s", "some-content", post.Content)
		}
		if post.AuthorUrl != "some-url" {
			t.Fatalf("Post author url error, expect %s, got %s", "some-url", post.AuthorUrl)
		}
	})

	t.Run("create another post", func(t *testing.T) {
		post := &domain.Post{
			Url: "some-other-url", Content: "some-other-content", AuthorUrl: other_author.Url,
		}
		post, err := postStorage.Upsert(post)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if post.Url != "some-other-url" {
			t.Fatalf("Post url error, expect %s, got %s", "some-url", post.Url)
		}
		if post.Content != "some-other-content" {
			t.Fatalf("Post content error, expect %s, got %s", "some-content", post.Content)
		}
		if post.AuthorUrl != "some-other-url" {
			t.Fatalf("Post author url error, expect %s, got %s", "some-other-url", post.AuthorUrl)
		}
	})

	t.Run("get all posts", func(t *testing.T) {
		posts, err := postStorage.GetAllPosts()
		if err != nil {
			t.Fatalf(err.Error())
		}
		if len(posts) != 2 {
			t.Fatalf("Expected 2 posts, got %d", len(posts))
		}
	})
	t.Run("get all posts by author url", func(t *testing.T) {
		posts, err := postStorage.GetAllPostsByAuthorUrl(author.Url)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if len(posts) != 1 {
			t.Fatalf("Expected 1 post, got %d", len(posts))
		}
	})

	t.Run("get all posts by author name", func(t *testing.T) {
		posts, err := postStorage.GetAllPostsByAuthorName("ack")
		if err != nil {
			t.Fatalf(err.Error())
		}
		if len(posts) != 1 {
			t.Fatalf("Expected 1 post, got %d", len(posts))
		}
	})
}
