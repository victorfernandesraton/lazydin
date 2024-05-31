package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/chromedp/chromedp"
	"github.com/victorfernandesraton/lazydin"
	"github.com/victorfernandesraton/lazydin/workflow"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
}

var res []string

func main() {
	headless := os.Getenv("HEADLESS")
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless == "true"),
		chromedp.Flag("start-maximized", true),
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)

	defer acancel()
	ctx, cancel := chromedp.NewContext(actx, chromedp.WithDebugf(log.Printf))
	defer cancel()

	if err := chromedp.Run(ctx,
		workflow.Auth(os.Getenv("LINKEDIN_USERNAME"), os.Getenv("LINKEDIN_PASSWORD")),
		workflow.SearchForPosts("GOLANG + REMOTO", &res),
	); err != nil {
		log.Fatal("Error when execute", err)
	}
	res, err := workflow.ExtractOuterHTML(ctx)
	if err != nil {
		log.Fatal("Error when extract content", err)
	}

	content, err := lazydin.ExtractContent(res)
	if err != nil {
		log.Fatal("Error when parse content", err)
	}

	log.Println(len(content))
}
