package workflows

import (
	"errors"
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

const (
	USERNAME_INPUT_XPATH = "//input[@id='session_key']"
	PASSWORD_INPUT_XPATH = "//input[@id='session_password']"
	BUTTON_SUBMIT_XPATH  = "//*[@id='main-content']/section/div/div/form/div/button"
)

type Auth struct {
	Driver selenium.WebDriver
}

func (service *Auth) Run(args map[string]string) error {

	err := service.Driver.Get("https://linkedin.com")
	if err != nil {
		return err
	}
	if err := service.Driver.WaitWithTimeout(waitForInput, time.Duration(10)*time.Second); err != nil {
		return err
	}

	inputs_and_keys := map[string]string{
		"username": USERNAME_INPUT_XPATH,
		"password": PASSWORD_INPUT_XPATH,
	}

	for key, xpath := range inputs_and_keys {
		value, ok := args[key]
		if !ok {
			return errors.New(fmt.Sprintln("Not found expected key: %v", key))
		}
		input, err := service.Driver.FindElement(selenium.ByXPATH, xpath)
		if err != nil {
			return nil
		}
		HumanInputSimulate(input, 1, value)

	}
	submit_button, err := service.Driver.FindElement(selenium.ByXPATH, BUTTON_SUBMIT_XPATH)
	if err != nil {
		return err
	}
	submit_button.Click()
	return nil
}

func waitForInput(wd selenium.WebDriver) (bool, error) {
	inputName, err := wd.FindElement(selenium.ByXPATH, USERNAME_INPUT_XPATH)
	if err != nil {
		return false, err
	}

	displayed, err := inputName.IsDisplayed()
	return displayed, err

}
