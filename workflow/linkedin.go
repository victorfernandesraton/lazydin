package workflow

import (
	"context"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/victorfernandesraton/lazydin/domain"
)

const (
	linkedinDomain          = "https://linkedin.com"
	login                   = "https://linkedin.com/login"
	linkedinFeed            = "https://www.linkedin.com/feed"
	username_xpath          = "//input[@id='username']"
	password_xpath          = "//input[@id='password']"
	submit_xpath            = "//button[@type='submit']"
	search_xpath            = "//input[@placeholder='Search']"
	search_qs               = "#global-nav-typeahead > input"
	button_posts            = "//nav/*/ul/li/button[text() = 'Posts']"
	post_xpath              = "//ul[@role='list' and contains(@class, 'reusable-search__entity-result-list')]/li"
	profileActionButtons_qs = "main button.pvs-profile-actions__action"
)

func Auth(username, password string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(linkedinDomain),
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
		chromedp.Navigate(linkedinFeed),
		chromedp.SendKeys(search_xpath, query),
		chromedp.KeyEvent(kb.Enter),
		chromedp.WaitVisible(button_posts),
		chromedp.Click(button_posts),
		chromedp.WaitVisible(post_xpath),
	}
}
func ExtractOuterHTML(ctx context.Context) (outerHTML []string, err error) {
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(post_xpath, &nodes, chromedp.BySearch)); err != nil {
		return nil, err
	}

	for _, node := range nodes {
		var html string
		if err := chromedp.Run(ctx, chromedp.OuterHTML(node.FullXPath(), &html)); err != nil {
			return nil, err
		}
		outerHTML = append(outerHTML, html)
	}
	return outerHTML, nil
}

func GoToUserPage(user domain.Author) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(user.Url),
	}
}

func ExtractPriofileActions(ctx context.Context) ([]*cdp.Node, error) {
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(profileActionButtons_qs, &nodes, chromedp.BySearch)); err != nil {
		return nil, err
	}
	return nodes, nil
}
