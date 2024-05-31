package adapters

import (
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/victorfernandesraton/lazydin"
)

const (
	author_name        = "li div.update-components-actor div .update-components-actor__title span span span"
	author_description = "li div.update-components-actor div .update-components-actor__description"
	autor_avatar       = "li div.update-components-actor div  a.app-aware-link"
	post               = "li div.update-components-text span.break-words"
	post_link          = "li div.feed-shared-update-v2"
)

func ExtractAuthor(dom *goquery.Document) (*lazydin.Author, error) {
	url, hasUrl := dom.Find(autor_avatar).Attr("href")
	if !hasUrl {
		return nil, nil
	}
	author := &lazydin.Author{
		Name:        dom.Find(author_name).First().Text(),
		Description: dom.Find(author_description).First().Text(),
		Url:         url,
	}
	return author, nil
}

func ExtractPost(dom *goquery.Document) (*lazydin.Post, error) {
	urn, hasUrn := dom.Find(post_link).First().Attr("data-urn")
	if !hasUrn {
		return nil, errors.New("Not found urn in user")
	}

	post := &lazydin.Post{
		Url:     urn,
		Content: dom.Find(post).First().Text(),
	}
	return post, nil
}

func ExtractContent(results []string) (contents []lazydin.Content, err error) {
	for _, v := range results {
		dom, err := goquery.NewDocumentFromReader(strings.NewReader(v))
		if err != nil {
			return nil, err
		}
		author, err := ExtractAuthor(dom)
		if err != nil {
			return nil, err
		}
		if author != nil {
			post, err := ExtractPost(dom)
			if err != nil {
				return nil, err
			}
			post.AuthorUrl = author.Url
			contents = append(contents, lazydin.Content{
				Author: *author,
				Post:   *post,
			})
		}

	}

	return contents, nil

}
