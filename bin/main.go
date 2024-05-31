package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/victorfernandesraton/lazydin/adapters"
	"github.com/victorfernandesraton/lazydin/workflow"
)

// Constants for flag names
const (
	flagUser          = "user"
	flagPassword      = "password"
	flagConfig        = "config"
	flagQuery         = "query"
	defaultConfigFile = "lazydin.sqlite"
)

var (
	configFile string
	username   string
	password   string
)

var rootCmd = &cobra.Command{
	Use:   "vagabot",
	Short: "CLI for interacting with Linkedin",
}

var searchPostsCmd = &cobra.Command{
	Use:   "search-posts",
	Short: "Search for posts on Linkedin",
	RunE:  searchPosts,
}

var commentPostCmd = &cobra.Command{
	Use:   "post-comment",
	Short: "Post a comment on a Linkedin post",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal(errors.New("not implemented yet"))
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	rootCmd.PersistentFlags().StringP(flagUser, "u", "", "Linkedin Username")
	rootCmd.PersistentFlags().StringP(flagPassword, "p", "", "Linkedin Password")

	searchPostsCmd.Flags().StringP(flagQuery, "q", "", "Query for search post")

	rootCmd.AddCommand(searchPostsCmd)
	rootCmd.AddCommand(commentPostCmd)

	if err := loadCredentials(); err != nil {
		log.Fatalf("Error loading credentials: %v", err)
	}
}

// searchPosts handles the search-posts command
func searchPosts(cmd *cobra.Command, args []string) error {
	query, err := cmd.Flags().GetString(flagQuery)
	if err != nil {
		return fmt.Errorf("failed to get query flag: %w", err)
	}
	if query == "" {
		return errors.New("query flag is required")
	}

	opts := createBrowserOptions()
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	ctx, cancel := chromedp.NewContext(actx, chromedp.WithDebugf(log.Printf))
	defer cancel()

	var htmlPost []string
	if err := chromedp.Run(ctx,
		workflow.Auth(username, password),
		workflow.SearchForPosts(query, &htmlPost),
	); err != nil {
		return fmt.Errorf("failed to execute chromedp tasks: %w", err)
	}

	content, err := workflow.ExtractOuterHTML(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract outer HTML: %w", err)
	}

	result, err := adapters.ExtractContent(content)
	if err != nil {
		return fmt.Errorf("failed to extract content: %w", err)
	}

	log.Printf("Number of posts found: %d", len(result))
	return nil
}

// loadCredentials loads the Linkedin credentials from environment variables or flags
func loadCredentials() error {
	envUsername := os.Getenv("LINKEDIN_USERNAME")
	envPassword := os.Getenv("LINKEDIN_PASSWORD")

	usernameFlag := rootCmd.PersistentFlags().Lookup(flagUser).Value.String()
	passwordFlag := rootCmd.PersistentFlags().Lookup(flagPassword).Value.String()

	if usernameFlag != "" {
		username = usernameFlag
	} else {
		username = envUsername
	}

	if passwordFlag != "" {
		password = passwordFlag
	} else {
		password = envPassword
	}

	if username == "" || password == "" {
		return errors.New("username and password must be set either via flags or environment variables")
	}

	return nil
}

// createBrowserOptions creates the browser options for chromedp
func createBrowserOptions() []chromedp.ExecAllocatorOption {
	headless := os.Getenv("HEADLESS") == "true"
	return append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("start-maximized", true),
	)
}
