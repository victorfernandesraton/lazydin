package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/chromedp/chromedp"
	"github.com/victorfernandesraton/lazydin/adapters"
	"github.com/victorfernandesraton/lazydin/browser"
	"github.com/victorfernandesraton/lazydin/config"
	"github.com/victorfernandesraton/lazydin/workflow"
)

type FindJobPost struct {
	Query string `json:"query"`
}

func searchPosts(w http.ResponseWriter, r *http.Request) {
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
	}

	content, err := workflow.ExtractOuterHTML(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	result, err := adapters.ExtractContent(content)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	for _, v := range result {
		if _, err := postsStore.Upsert(&v.Post); err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if _, err := authorStore.Upsert(&v.Author); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func startServer() {

	configs, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/scrape", scrapeHandler)
	http.ListenAndServe(":8080", nil)

}
