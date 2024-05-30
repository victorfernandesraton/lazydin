package workflow

import (
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

const (
	domain         = "https://linkedin.com"
	login          = "https://linkedin.com/login"
	username_xpath = "//input[@id='username']"
	password_xpath = "//input[@id='password']"
	submit_xpath   = "//button[@type='submit']"
	search_xpath   = "//input[@placeholder='Search']"
	search_qs      = "#global-nav-typeahead > input"
	button_posts   = "//nav/*/ul/li/button[text() = 'Posts']"
	post_xpath     = "//ul[@role='list' and contains(@class, 'reusable-search__entity-result-list ')]/li"
)

func Auth(username, password string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(domain),
		chromedp.Navigate(login),
		chromedp.WaitVisible(username_xpath),
		chromedp.SendKeys(username_xpath, username),
		chromedp.WaitVisible(password_xpath),
		chromedp.SendKeys(password_xpath, password),
		chromedp.Click(submit_xpath),
		chromedp.WaitVisible(search_xpath),
	}
}

func SearchForPosts(query string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.SendKeys(search_xpath, query),
		chromedp.KeyEvent(kb.Enter),
		chromedp.WaitVisible(button_posts),
		chromedp.Click(button_posts),
		chromedp.WaitVisible(post_xpath),
	}
}
