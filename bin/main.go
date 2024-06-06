package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/victorfernandesraton/lazydin/adapters"
	"github.com/victorfernandesraton/lazydin/workflow"
)

// Constants for flag names
const (
	flagUser               = "user"
	flagPassword           = "password"
	flagConfig             = "config"
	flagQuery              = "query"
	flagOutput             = "output"
	flagSeparator          = "sep"
	flagCredentials        = "credentials"
	flagDatabase           = "database"
	defaultDatabaseFile    = "lazydin.sqlite"
	defaultCredentialsFile = "credentials.toml"
)

var (
	configPath      string
	credentialsFile string
	username        string
	password        string
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

func init() {
	dir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configPath = path.Join(dir, "lazydin")
	credentialsFile = path.Join(configPath, defaultCredentialsFile)
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	rootCmd.PersistentFlags().StringP(flagConfig, "c", configPath, "Configguration path")
	rootCmd.PersistentFlags().StringP(flagUser, "u", "", "Linkedin Username")
	rootCmd.PersistentFlags().StringP(flagPassword, "p", "", "Linkedin Password")
	rootCmd.PersistentFlags().String(flagCredentials, credentialsFile, "Credential file storage in toml")

	searchPostsCmd.Flags().StringP(flagQuery, "q", "", "Query for search post")
	searchPostsCmd.Flags().StringP(flagOutput, "o", "", "Output file as csv")
	searchPostsCmd.Flags().StringP(flagSeparator, "", ";", "Output file as csv separator")

	rootCmd.AddCommand(searchPostsCmd)
	rootCmd.AddCommand(commentPostCmd)

	if err := loadCredentials(); err != nil {
		log.Fatalf("Error loading credentials: %v", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
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
	outputFile, err := cmd.Flags().GetString(flagOutput)
	if err != nil {
		return fmt.Errorf("failed to get output flag: %w", err)
	}

	separator, err := cmd.Flags().GetString(flagSeparator)
	if err != nil {
		return fmt.Errorf("failed to get csv separator: %w", err)
	}
	opts := createBrowserOptions()
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	ctx, cancel := chromedp.NewContext(actx, chromedp.WithDebugf(log.Printf))
	defer cancel()

	var htmlPost []string
	if err := chromedp.Run(ctx,
		workflow.Auth(username, password), workflow.SearchForPosts(query, &htmlPost),
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
	if outputFile == "" {
		log.Printf("Number of posts found: %d", len(result))
	} else if !strings.HasSuffix(outputFile, ".csv") {
		return fmt.Errorf("invalid file format for output, got %v, but only supported is .csv", outputFile)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	runeSeparator := []rune(separator)
	csvWriter.Comma = runeSeparator[0]
	err = gocsv.MarshalCSV(&result, csvWriter)
	if err != nil {
		return err
	}
	csvWriter.Flush()

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
		return errors.New(
			"username and password must be set either via flags or environment variables",
		)
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
