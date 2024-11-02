package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/angelofallars/htmx-go"
	"github.com/chromedp/chromedp"
	"github.com/victorfernandesraton/lazydin/adapters"
	"github.com/victorfernandesraton/lazydin/browser"
	"github.com/victorfernandesraton/lazydin/config"
	"github.com/victorfernandesraton/lazydin/domain"
	"github.com/victorfernandesraton/lazydin/storage"
	"github.com/victorfernandesraton/lazydin/workflow"

	_ "github.com/mattn/go-sqlite3"
)

var Configs *config.Config
var PostsStore *storage.PostStorage
var AuthorStore *storage.AuthorStorage
var Tmpl *template.Template

// Implement

type FindJobPost struct {
	Query string `json:"query"`
}

func ExecuteSearchPostInLinkedin(query string) ([]domain.Content, error) {

	opts := browser.CreateBrowserOptions(browser.DefaultBrowserOptions())
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	ctx, cancel := chromedp.NewContext(actx, chromedp.WithLogf(log.Printf))
	defer cancel()

	Configs, err := config.LoadConfig()
	credentials := config.GetCredentials(Configs)
	if err := chromedp.Run(ctx,
		workflow.Auth(credentials.Username, credentials.Password), workflow.SearchForPosts(query),
	); err != nil {
		return nil, err
	}

	content, err := workflow.ExtractOuterHTML(ctx)
	if err != nil {
		return nil, err
	}
	return adapters.ExtractContent(content)
}

func SearchPostsInLinkedin(w http.ResponseWriter, r *http.Request) {
	fmt.Println(fmt.Sprintf("searching posts in linkedin method %s", r.Method))
	switch r.Method {
	case "GET":
		err := Tmpl.ExecuteTemplate(w, "search.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case "POST":
		if !htmx.IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var posts []domain.Post
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query := r.Form.Get("query")
		if query == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		result, err := ExecuteSearchPostInLinkedin(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, v := range result {
			posts = append(posts, v.Post)
			if _, err := PostsStore.Upsert(&v.Post); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := AuthorStore.Upsert(&v.Author); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		err = Tmpl.ExecuteTemplate(w, "posts-list.html", posts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func GetPostByUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	url := r.PathValue("post_url")
	if url == "" {
		http.Error(w, "Invalid request slug", http.StatusBadRequest)
		return
	}

	post, err := PostsStore.GetByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = Tmpl.ExecuteTemplate(w, "post-item.html", post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var posts []domain.Post
	var err error
	authorName := r.URL.Query().Get("author_name")
	authorUrl := r.URL.Query().Get("author_url")
	if authorName != "" && authorUrl != "" {
		http.Error(w, "Invalid query params, using author_name or author_url", http.StatusBadRequest)
		return
	}

	if authorName == "" && authorUrl == "" {
		posts, err = PostsStore.GetAllPosts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if authorName != "" {
		posts, err = PostsStore.GetAllPostsByAuthorName(authorName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if authorUrl != "" {
		posts, err = PostsStore.GetAllPostsByAuthorUrl(authorUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if htmx.IsHTMX(r) {
		err = Tmpl.ExecuteTemplate(w, "posts-list.html", posts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		err = Tmpl.ExecuteTemplate(w, "posts.html", posts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func GetAuthors(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	var authors []domain.Author
	var err error
	if name != "" {

		authors, err = AuthorStore.GetByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		authors, err = AuthorStore.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if htmx.IsHTMX(r) {
		err = Tmpl.ExecuteTemplate(w, "authors-list.html", authors)
	} else {
		err = Tmpl.ExecuteTemplate(w, "authors.html", authors)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetAuthorByUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	url := r.PathValue("author_url")
	if url == "" {
		http.Error(w, "Invalid request slug", http.StatusBadRequest)
		return
	}

	author, err := AuthorStore.GetByUrl(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = Tmpl.ExecuteTemplate(w, "author-single.html", author)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateUserConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req config.CredentialsConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := config.SetCredentials(req.Username, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := Tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
