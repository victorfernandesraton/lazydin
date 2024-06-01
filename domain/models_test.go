package domain_test

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/gocarina/gocsv"
	"github.com/victorfernandesraton/lazydin/domain"
)

const CSVSeparator = ';'

func TestCreateCSV(t *testing.T) {

	var builder strings.Builder
	csvWriter := csv.NewWriter(&builder)
	csvWriter.Comma = CSVSeparator
	t.Run("should create author csv", func(t *testing.T) {
		author := []domain.Author{{
			Url:         "somme_url",
			Name:        "Victor\n Raton",
			Description: "Some description",
		}}
		err := gocsv.MarshalCSV(&author, csvWriter)
		if err != nil {
			t.Errorf(err.Error())
		}
		csvWriter.Flush()
		res := builder.String()
		defer builder.Reset()
		if res == "" {
			t.Errorf("result can not be empty")
		}

	})

	t.Run("should create post csv", func(t *testing.T) {
		posts := []domain.Post{{
			Url:       "somme_url",
			Content:   "some content",
			AuthorUrl: "author/url",
		}}
		err := gocsv.MarshalCSV(&posts, csvWriter)
		if err != nil {
			t.Errorf(err.Error())
		}
		csvWriter.Flush()
		res := builder.String()
		defer builder.Reset()
		if res == "" {
			t.Errorf("result can not be empty")
		}

	})

	t.Run("should create content csv", func(t *testing.T) {
		author := &domain.Author{
			Url:         "somme_url",
			Name:        "Victor\n Raton",
			Description: "Some description",
		}
		post := &domain.Post{
			Url:       "somme_url",
			Content:   "some content",
			AuthorUrl: "author/url",
		}
		contents := []domain.Content{{
			Author: *author, Post: *post,
		}}
		err := gocsv.MarshalCSV(&contents, csvWriter)
		if err != nil {
			t.Errorf(err.Error())
		}
		csvWriter.Flush()
		res := builder.String()
		defer builder.Reset()
		if res == "" {
			t.Errorf("result can not be empty")
		}

	})
}
