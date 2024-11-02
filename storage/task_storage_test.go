package storage_test

import (
	"database/sql"
	"testing"

	"github.com/victorfernandesraton/lazydin/domain"
	"github.com/victorfernandesraton/lazydin/storage"
)

func TestTaskStorage(t *testing.T) {
	databse, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf(err.Error())
	}

	taskStorage := storage.NewTaskStorage(databse)
	postStorage := storage.NewPostStorage(databse)
	authorStorage := storage.NewAuthorStorage(databse)
	var post *domain.Post
	var task *domain.Task
	t.Run("create table", func(t *testing.T) {
		if err := authorStorage.CreateTable(); err != nil {
			t.Fatalf(err.Error())
		}
		if err := postStorage.CreateTable(); err != nil {
			t.Fatalf(err.Error())
		}
		if err := taskStorage.CreateTable(); err != nil {
			t.Fatalf(err.Error())
		}
	})

	t.Run("create post", func(t *testing.T) {
		author, err := authorStorage.Upsert(&domain.Author{
			Url: "some-author-url", Name: "Victor Raton", Description: "some-description",
		})
		if err != nil {
			t.Fatalf(err.Error())
		}
		post, err = postStorage.Upsert(
			&domain.Post{
				Url: "some-url", Content: "some-content", AuthorUrl: author.Url,
			},
		)
		if err != nil {
			t.Fatalf(err.Error())
		}
	})

	t.Run("create task", func(t *testing.T) {
		task, err = taskStorage.CreateTask()
		if err != nil {
			t.Fatalf(err.Error())
		}
	})

	t.Run("create task post", func(t *testing.T) {
		taskPost, err := taskStorage.CreateTaskPost(task.Id, post.Url)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if taskPost.TaskId != task.Id {
			t.Fatalf("expected task id %d, got %d", task.Id, taskPost.TaskId)
		}
		if taskPost.PostUrl != post.Url {
			t.Fatalf("expected post url %s, got %s", post.Url, taskPost.PostUrl)
		}
	})
}
