package linkedisney_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	linkedisney "github.com/victorfernandesraton/vagabot2"
)

var dom *goquery.Document

func TestMain(m *testing.M) {
	filePath := filepath.Join("testdata", "output.html")

	content, err := os.ReadFile(filePath)
	if err != nil {
		panic("failed to read test data file: " + err.Error())
	}

	dom, err = goquery.NewDocumentFromReader(bytes.NewReader(content))

	if err != nil {
		panic("failed to parse document for goquery" + err.Error())
	}

	m.Run()
}

func TestParseAuthor(t *testing.T) {
	res, err := linkedisney.ExtractAuthor(dom)

	if err != nil || res == nil {
		t.Fail()
	}

	if !strings.Contains(res.Name, "Tammy") || !strings.Contains(res.Name, "Silva") {
		t.Log(res.Name)
		t.Fail()
	}

}
func TestParsePost(t *testing.T) {
	res, err := linkedisney.ExtractPost(dom)

	if err != nil || res == nil {
		t.Log(err)
		t.Fail()
	}

	if res.Url == "" {
		t.Fatalf("Not found url")
	}

}
