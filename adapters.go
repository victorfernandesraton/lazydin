package linkedisney

import (
	"errors"

	"github.com/PuerkitoBio/goquery"
)

const (
	author_name        = "li div.update-components-actor div .update-components-actor__title span span span"
	author_description = "li div.update-components-actor div .update-components-actor__description"
	autor_avatar       = "li div.update-components-actor div  a.app-aware-link"
	post               = "li div.update-components-text span.break-words"
	post_link          = "li div.feed-shared-update-v2"
)

func ExtractAuthor(dom *goquery.Document) (*Author, error) {
	url, hasUrl := dom.Find(autor_avatar).Attr("href")
	if !hasUrl {
		return nil, errors.New("Not found author picture url")
	}
	author := &Author{
		Name:        dom.Find(author_name).First().Text(),
		Description: dom.Find(author_description).First().Text(),
		Url:         url,
	}
	return author, nil
}

func ExtractPost(dom *goquery.Document) (*Post, error) {
	urn, hasUrn := dom.Find(post_link).First().Attr("data-urn")
	if !hasUrn {
		return nil, errors.New("Not found urn in user")
	}

	post := &Post{
		Url:     urn,
		Content: dom.Find(post).First().Text(),
	}
	return post, nil
}
