package workflows

import (
	"time"

	"github.com/tebeka/selenium"
)

type Workflow interface {
	Run(args map[string]string) error
}

func HumanInputSimulate(element selenium.WebElement, delay int, text string) error {
	element.Clear()
	for _, char := range text {
		time.Sleep(time.Duration(delay) * time.Second)
		if err := element.SendKeys(string(char)); err != nil {
			return err
		}

	}

	return nil
}
