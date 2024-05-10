package browser

import (
	"github.com/tebeka/selenium"
)

func Local(driver string, port int) (*selenium.Service, error) {
	service, err := selenium.NewChromeDriverService(driver, port)
	if err != nil {
		return nil, err
	}
	return service, nil
}
