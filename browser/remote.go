package browser

import (
	"github.com/tebeka/selenium"
)

func Remote(url string, caps selenium.Capabilities) (selenium.WebDriver, error) {
	wd, err := selenium.NewRemote(caps, url)
	return wd, err

}
