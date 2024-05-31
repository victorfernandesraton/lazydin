package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/chromedp/chromedp"
	"github.com/victorfernandesraton/lazydin"
	"github.com/victorfernandesraton/lazydin/workflow"

	"errors"

	"github.com/spf13/cobra"
)

func GetUserAndPassword(cmd *cobra.Command) (string, string) {
	username := os.Getenv("LINKEDIN_USERNAME")
	password := os.Getenv("LINKEDIN_PASSWORD")
	user, _ := cmd.Flags().GetString("user")
	pass, _ := cmd.Flags().GetString("password")

	if user != "" {
		username = user
	}
	if pass != "" {
		password = pass
	}

	return username, password
}

var configFile string

// Cobra command structure
var rootCmd = &cobra.Command{
	Use:   "vagabot",
	Short: "CLI for interacting with Linkedin",
}

var searchPostsCmd = &cobra.Command{
	Use:   "search-posts",
	Short: "Search for posts on Linkedin",
	RunE: func(cmd *cobra.Command, args []string) error {
		query, err := cmd.Flags().GetString("query")
		if err != nil {
			return err
		}
		username, password := GetUserAndPassword(cmd)
		fmt.Println(username, password)
		opts := CreateBrowserOptions()
		actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer acancel()
		ctx, cancel := chromedp.NewContext(actx, chromedp.WithDebugf(log.Printf))
		defer cancel()

		var htmlPost []string
		if err := chromedp.Run(ctx,
			workflow.Auth(username, password),
			workflow.SearchForPosts(query, &htmlPost),
		); err != nil {
			log.Fatal("Error when execute", err)
			return err
		}
		content, err := workflow.ExtractOuterHTML(ctx)
		if err != nil {
			log.Fatal("Error when extract content", err)
			return err
		}

		result, err := lazydin.ExtractContent(content)
		if err != nil {
			log.Fatal("Error when parse content", err)

			return err
		}
		log.Println(len(result))
		return nil
	},
}

var commentPostCmd = &cobra.Command{
	Use:   "post-comment",
	Short: "Post a comment on a Linkedin post",
	Run: func(cmd *cobra.Command, args []string) {
		panic(errors.New("not implemented yed"))
	},
}

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	rootCmd.Flags().StringP("user", "u", "", "Linkedin Username")
	rootCmd.Flags().StringP("password", "p", "", "Linkedin Password")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "lazydin.sqlite", "Path to configuration file (optional)")
	searchPostsCmd.Flags().StringP("query", "q", "", "Query for search post")

	rootCmd.AddCommand(searchPostsCmd)
	rootCmd.AddCommand(commentPostCmd)
}

func CreateBrowserOptions() []chromedp.ExecAllocatorOption {

	headless := os.Getenv("HEADLESS")
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless == "true"),
		chromedp.Flag("start-maximized", true),
	)
	return opts
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
