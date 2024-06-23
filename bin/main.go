package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gocarina/gocsv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/victorfernandesraton/lazydin/adapters"
	"github.com/victorfernandesraton/lazydin/browser"
	"github.com/victorfernandesraton/lazydin/config"
	"github.com/victorfernandesraton/lazydin/storage"
	"github.com/victorfernandesraton/lazydin/workflow"
)

// Constants for flag names
const (
	appName                = "lazydin"
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
	configUsername         = "username"
	configPassword         = "password"
)

var (
	configPath      string
	credentialsFile string
	configs         *config.Config
	databse         *sql.DB
	postsStore      *storage.PostStorage
	authorStore     *storage.AuthorStorage
)

var rootCmd = &cobra.Command{
	Use:   "lazydin",
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

var createCredentials = &cobra.Command{
	Use:   "create-credentials",
	Short: "Start proccess to define credentials in config credentials file",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Print("Enter password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)
		return config.SetCredentials(username, password)

	},
}

var createStorage = &cobra.Command{
	Use:   "create-storage",
	Short: "Start proccess to define path to storage file",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Enter path: default %s", configs.SQlite)
		sqlitePath, _ := reader.ReadString('\n')
		return config.SetStorage(sqlitePath)

	},
}

func init() {
	var err error
	rootCmd.PersistentFlags().StringP(flagConfig, "c", configPath, "Configguration path")
	rootCmd.PersistentFlags().StringP(flagUser, "u", "", "Linkedin Username")
	rootCmd.PersistentFlags().StringP(flagPassword, "p", "", "Linkedin Password")
	rootCmd.PersistentFlags().String(flagCredentials, credentialsFile, "Credential file storage in toml")

	searchPostsCmd.Flags().StringP(flagQuery, "q", "", "Query for search post")
	searchPostsCmd.Flags().StringP(flagOutput, "o", "", "Output file as csv")
	searchPostsCmd.Flags().StringP(flagSeparator, "", ";", "Output file as csv separator")

	rootCmd.AddCommand(searchPostsCmd)
	rootCmd.AddCommand(commentPostCmd)
	rootCmd.AddCommand(createCredentials)
	rootCmd.AddCommand(createStorage)

	configs, err = config.LoadConfig()
	if err != nil {
		log.Fatalf(err.Error())

	}

}

func main() {
	var err error
	if _, err := os.Stat(configs.SQlite); os.IsNotExist(err) {
		if _, err := os.Create(configs.SQlite); err != nil {
			log.Fatalf(err.Error())
		}
	}
	databse, err = sql.Open("sqlite3", configs.SQlite)
	if err != nil {
		log.Fatalf(err.Error())
	}

	postsStore = storage.NewPostStorage(databse)
	authorStore = storage.NewAuthorStorage(databse)
	if err = authorStore.CreateTable(); err != nil {
		log.Fatalf(err.Error())

	}

	if err = postsStore.CreateTable(); err != nil {
		log.Fatalf(err.Error())
	}
	if err = rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
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

	usernameFlag := rootCmd.PersistentFlags().Lookup(flagUser).Value.String()
	passwordFlag := rootCmd.PersistentFlags().Lookup(flagPassword).Value.String()
	credentials, err := config.LoadCredentials(configs, usernameFlag, passwordFlag)
	if err != nil {
		return err
	}
	opts := browser.CreateBrowserOptions(browser.DefaultBrowserOptions())
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()

	ctx, cancel := chromedp.NewContext(actx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var htmlPost []string
	if err := chromedp.Run(ctx,
		workflow.Auth(credentials.Username, credentials.Password), workflow.SearchForPosts(query, &htmlPost),
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
		for _, v := range result {
			if _, err := postsStore.Upsert(&v.Post); err != nil {
				return err
			}
			if _, err := authorStore.Upsert(&v.Author); err != nil {
				return err
			}
		}
		return nil
	}
	if !strings.HasSuffix(outputFile, ".csv") {
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
