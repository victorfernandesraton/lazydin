// Command text is a chromedp example demonstrating how to extract text from a
// specific element.
package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/chromedp/chromedp"
	"github.com/victorfernandesraton/vagabot2/workflow"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
}

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-maximized", true),
		// other options below
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// create context
	defer acancel()
	ctx, cancel := chromedp.NewContext(actx, chromedp.WithDebugf(log.Printf))
	defer cancel()

	// run task list
	var res []string

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

	log.Println(res)
}
