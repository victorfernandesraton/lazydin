package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/chromedp/chromedp"
	"github.com/victorfernandesraton/lazydin/adapters"
	"github.com/victorfernandesraton/lazydin/browser"
	"github.com/victorfernandesraton/lazydin/config"
	"github.com/victorfernandesraton/lazydin/domain"
	"github.com/victorfernandesraton/lazydin/storage"
	"github.com/victorfernandesraton/lazydin/workflow"

	_ "github.com/mattn/go-sqlite3"
)

var configs *config.Config
var postsStore *storage.PostStorage
var authorStore *storage.AuthorStorage

var databse *sql.DB

type FindJobPost struct {
	Query string `json:"query"`
}

func SearchPostsInLinkedin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req FindJobPost
	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if req.Query == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	opts := browser.CreateBrowserOptions(browser.DefaultBrowserOptions())
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	ctx, cancel := chromedp.NewContext(actx, chromedp.WithLogf(log.Printf))
	defer cancel()

	configs, err := config.LoadConfig()
	credentials := config.GetCredentials(configs)
	if err := chromedp.Run(ctx,
		workflow.Auth(credentials.Username, credentials.Password), workflow.SearchForPosts(req.Query),
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	content, err := workflow.ExtractOuterHTML(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := adapters.ExtractContent(content)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, v := range result {
		if _, err := postsStore.Upsert(&v.Post); err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := authorStore.Upsert(&v.Author); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
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
		posts, err = postsStore.GetAllPosts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if authorName != "" {
		posts, err = postsStore.GetAllPostsByAuthorName(authorName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if authorUrl != "" {
		posts, err = postsStore.GetAllPostsByAuthorUrl(authorUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(posts)
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
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
}
