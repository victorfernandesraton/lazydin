package storage_test

import (
	"database/sql"
	"testing"

	"github.com/victorfernandesraton/lazydin/domain"
	"github.com/victorfernandesraton/lazydin/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestAuthorStorage(t *testing.T) {
	databse, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf(err.Error())
	}
	authorStorage := storage.NewAuthorStorage(databse)
	t.Run("create table", func(t *testing.T) {
		if err := authorStorage.CreateTable(); err != nil {
			t.Fatalf(err.Error())
		}
	})
	t.Run("create author", func(t *testing.T) {
		author := &domain.Author{
			Url: "some-url", Name: "Victor Raton",
		}
		author, err := authorStorage.Upsert(author)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if author.Url != "some-url" {
			t.Fatalf("Author url error, expect %s, got %s", "some-url", author.Url)
		}
		if author.Name != "Victor Raton" {
			t.Fatalf("Author url error, expect %s, got %s", "Victor Raton", author.Name)
		}
	})

	t.Run("create other author", func(t *testing.T) {
		author := &domain.Author{
			Url: "some-other", Name: "Victor Raton",
		}
		author, err := authorStorage.Upsert(author)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if author.Url != "some-other" {
			t.Fatalf("Author url error, expect %s, got %s", "some-url", author.Url)
		}
		if author.Name != "Victor Raton" {
			t.Fatalf("Author url error, expect %s, got %s", "Victor Raton", author.Name)
		}
	})

	t.Run("upsert author", func(t *testing.T) {
		author := &domain.Author{
			Url: "some-url", Name: "Captain Jack Sparrow",
		}
		author, err := authorStorage.Upsert(author)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if author.Url != "some-url" {
			t.Fatalf("Author url error, expect %s, got %s", "some-url", author.Url)
		}
		if author.Name != "Captain Jack Sparrow" {
			t.Fatalf("Author url error, expect %s, got %s", "Captain Jack Sparrow", author.Name)
		}
	})

	t.Run("get by url", func(t *testing.T) {
		author, err := authorStorage.GetByUrl("some-url")
		if err != nil {
			t.Fatalf(err.Error())
		}
		if author.Url != "some-url" {
			t.Fatalf("Author shoud be using url some-url")
		}
	})

}
