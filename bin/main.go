package main

import (
	"os"

	"github.com/tebeka/selenium"
	"github.com/victorfernandesraton/linkedisney/browser"
	"github.com/victorfernandesraton/linkedisney/workflows"
)

func main() {
	driver, err := browser.Remote("http://localhost:4444/wd/hub", selenium.Capabilities{"browserName": "chrome"})
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	// configure the browser options

	auth_workflow := workflows.Auth{Driver: driver}
	args := map[string]string{"username": os.Getenv("LINKEDIN_USERNAME"), "password": os.Getenv("LINKEDIN_PASSWORD")}
	if err := auth_workflow.Run(args); err != nil {
		panic(err)
	}

}
