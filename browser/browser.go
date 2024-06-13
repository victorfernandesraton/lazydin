package browser

import (
	"github.com/chromedp/chromedp"
)

type BrowserOptions struct {
	Maximized bool
	Headless  bool
}

func DefaultBrowserOptions() BrowserOptions {
	return BrowserOptions{
		Maximized: true,
		Headless:  false,
	}
}

func CreateBrowserOptions(options BrowserOptions) []chromedp.ExecAllocatorOption {
	return append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", options.Headless),
		chromedp.Flag("start-maximized", options.Maximized),
	)
}
