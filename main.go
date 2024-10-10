package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gocarina/gocsv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/victorfernandesraton/lazydin/adapters"
	"github.com/victorfernandesraton/lazydin/browser"
	"github.com/victorfernandesraton/lazydin/config"
	"github.com/victorfernandesraton/lazydin/domain"
	"github.com/victorfernandesraton/lazydin/server"
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
	flagUrl                = "url"
	flagAction             = "action"
	flagPort               = "port"
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

var commands = []cobra.Command{
	{
		Use:   "search",
		Short: "Search for posts on Linkedin",
		RunE:  searchPosts,
	}, {
		Use:     "follow",
		Short:   "Follow specific user By id or url",
		Example: "follow [--id integer | --url user linkedin profile urls]",
		RunE:    followUser,
	},
	{
		Use:   "prospect",
		Short: "UNDER CONSTRUCTION Prospect about some post/job with the author",
		Run: func(cmd *cobra.Command, args []string) {
			log.Fatal(errors.New("not implemented yet"))
		},
	},

	{
		Use:   "comment",
		Short: "Post a comment on a Linkedin post",
		Run: func(cmd *cobra.Command, args []string) {
			log.Fatal(errors.New("not implemented yet"))
		},
	},
	{
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
	},
	{
		Use:   "create-storage",
		Short: "Start proccess to define path to storage file",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			fmt.Printf("Enter path: default %s", configs.SQlite)
			sqlitePath, _ := reader.ReadString('\n')
			return config.SetStorage(sqlitePath)

		},
	},
	{
		Use:     "server",
		Short:   "Start server",
		Example: "server [--port integer ]",
		Run: func(cmd *cobra.Command, args []string) {

			port, err := cmd.Flags().GetInt(flagPort)
			if err != nil {
				panic(err)
			}
			configs, err := config.LoadConfig()
			if err != nil {
				panic(err)
			}

			if _, err := os.Stat(configs.SQlite); os.IsNotExist(err) {
				if _, err := os.Create(configs.SQlite); err != nil {
					panic(err)
				}
			}
			databse, err = sql.Open("sqlite3", configs.SQlite)
			if err != nil {
				panic(err)
			}
			authorStore = storage.NewAuthorStorage(databse)
			if err = authorStore.CreateTable(); err != nil {
				panic(err)

			}

			postsStore = storage.NewPostStorage(databse)
			if err = postsStore.CreateTable(); err != nil {
				panic(err)
			}

			http.HandleFunc("/search", server.SearchPostsInLinkedin)
			http.HandleFunc("/posts", server.GetPosts)
			http.HandleFunc("/config/credentials", server.UpdateUserConfig)
			http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		},
	},
}

func init() {
	rootCmd.PersistentFlags().StringP(flagConfig, "c", configPath, "Configguration path")
	rootCmd.PersistentFlags().StringP(flagUser, "u", "", "Linkedin Username")
	rootCmd.PersistentFlags().StringP(flagPassword, "p", "", "Linkedin Password")
	rootCmd.PersistentFlags().String(flagCredentials, credentialsFile, "Credential file storage in toml")

	commands[0].Flags().StringP(flagQuery, "q", "", "Query for search post")
	commands[0].Flags().StringP(flagOutput, "o", "", "Output file as csv")
	commands[0].Flags().StringP(flagSeparator, "", ";", "Output file as csv separator")

	commands[1].Flags().StringP(flagUrl, "", "", "valid profile url")
	commands[1].Flags().StringP(flagAction, "a", "Follow", "Action to execute")
	commands[6].Flags().IntP(flagPort, "", 8081, "Port for server")

	for _, cmd := range commands {
		rootCmd.AddCommand(&cmd)
	}

}

func main() {
	var err error

	configs, err = config.LoadConfig()
	if err != nil {
		log.Fatalf(err.Error())

	}
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

	if !strings.HasSuffix(outputFile, ".csv") && outputFile != "" {
		return fmt.Errorf("invalid file format for output, got %v, but only supported is .csv", outputFile)
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

	if err := chromedp.Run(ctx,
		workflow.Auth(credentials.Username, credentials.Password), workflow.SearchForPosts(query),
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
	} else {
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
		defer csvWriter.Flush()

	}

	return nil
}

// followUser is a function to using id from database or url to follow a linkedin user
// this function handle for follow-user command
func followUser(cmd *cobra.Command, args []string) error {
	var user *domain.Author
	url, err := cmd.Flags().GetString(flagUrl)
	if err != nil {
		return fmt.Errorf("failed to get url flag: %w", err)
	}

	selectedAction, err := cmd.Flags().GetString(flagAction)
	if err != nil {

		return fmt.Errorf("failed to get action flag: %w", err)
	}

	user = &domain.Author{Url: url}
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

	if err := chromedp.Run(ctx,
		workflow.Auth(credentials.Username, credentials.Password), workflow.GoToUserPage(*user),
	); err != nil {
		return fmt.Errorf("failed to execute chromedp tasks: %w", err)
	}
	if err := workflow.ExecuteFollowAction(ctx, selectedAction); err != nil {
		return err
	}

	return nil
}
