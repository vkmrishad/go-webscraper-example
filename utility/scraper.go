package utility

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/vkmrishad/go-webscraper-example/models"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

var internalUrlsExceptions = []string{"mailto:", "tel:"}
var loginPageSlugList = []string{"login", "log-in", "signin", "sign-in"}
var doctypes = make(map[string]string)

func init() {
	doctypes["HTML 4.01 Strict"] = `"-//W3C//DTD HTML 4.01//EN"`
	doctypes["HTML 4.01 Transitional"] = `"-//W3C//DTD HTML 4.01 Transitional//EN"`
	doctypes["HTML 4.01 Frameset"] = `"-//W3C//DTD HTML 4.01 Frameset//EN"`
	doctypes["XHTML 1.0 Strict"] = `"-//W3C//DTD XHTML 1.0 Strict//EN"`
	doctypes["XHTML 1.0 Transitional"] = `"-//W3C//DTD XHTML 1.0 Transitional//EN"`
	doctypes["XHTML 1.0 Frameset"] = `"-//W3C//DTD XHTML 1.0 Frameset//EN"`
	doctypes["XHTML 1.1"] = `"-//W3C//DTD XHTML 1.1//EN"`
	doctypes["HTML 5"] = `<!DOCTYPE html>`
}

// Check doctypes
func checkDoctype(html string) string {
	var version = "Unknown"

	for doctype, matcher := range doctypes {
		match := strings.Contains(html, matcher)

		if match == true {
			version = doctype
		}
	}
	return version
}

// Detect HTML Document Version From a string
func DetectFromString(htmlDoc string) string {
	htmlVersion := checkDoctype(htmlDoc)
	return htmlVersion
}

func fetchUrl(url string, chFailedUrls chan string, chIsFinished chan bool) {

	// Open url.
	// Need to use http.Client in order to set a custom user agent:
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)

	// Inform the channel chIsFinished that url fetching is done (no
	// matter whether successful or not). Defer triggers only once
	// we leave fetchUrl():
	defer func() {
		chIsFinished <- true
	}()

	// If url could not be opened, we inform the channel chFailedUrls:
	if err != nil || (
		resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != 999
		) {
		chFailedUrls <- url
		return
	}

}

func getHtmlVersion(doc goquery.Document) string {
	// HTML version
	var htmlVersion string
	html, err := doc.Html()
	if err != nil {
		log.Fatal(err)
	}
	htmlVersion = checkDoctype(html)
	return htmlVersion
}

func getPageTitle(doc goquery.Document) string {
	// Page title
	title := doc.Find("title").Text()
	return title
}

func getHeadingCount(doc goquery.Document) models.HeadingCount {
	// Get all headings count
	headings := models.HeadingCount{
		H1: doc.Find("h1").Length(),
		H2: doc.Find("h2").Length(),
		H3: doc.Find("h3").Length(),
		H4: doc.Find("h4").Length(),
		H5: doc.Find("h5").Length(),
		H6: doc.Find("h16").Length(),
	}
	return headings
}

func getAllLinks(doc goquery.Document) models.Links {
	var internalUrls []string
	var externalUrls []string
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")

		u, err := url.Parse(href)
		if err != nil {
			log.Fatal(err)
		}

		if doc.Url.Host == u.Host || (u.Host == "" && !Contains(internalUrlsExceptions, href)) {
			if u.Host == "" {
				// Host added for missing hosts
				internalUrls = append(internalUrls, fmt.Sprintf("%s://%s/%s", doc.Url.Scheme, doc.Url.Host, href))
			} else {
				internalUrls = append(internalUrls, href)
			}

		} else {
			externalUrls = append(externalUrls, href)
		}

	})

	// Create 2 channels, 1 to track urls we could not open
	// and 1 to inform url fetching is done:
	chFailedUrls := make(chan string)
	chIsFinished := make(chan bool)

	// Open all urls concurrently using the 'go' keyword:
	for _, url := range externalUrls {
		go fetchUrl(url, chFailedUrls, chIsFinished)
	}

	// Receive messages from every concurrent goroutine. If
	// an url fails, we log it to failedUrls array:
	failedUrls := make([]string, 0)
	for i := 0; i < len(externalUrls); {
		select {
		case url := <-chFailedUrls:
			failedUrls = append(failedUrls, url)
		case <-chIsFinished:
			i++
		}
	}

	links := models.Links{
		Internal: models.LinkDetails{
			Urls:  internalUrls,
			Count: len(internalUrls),
		},
		External: models.LinkDetails{
			Urls:  externalUrls,
			Count: len(externalUrls),
		},
		Failed: models.LinkDetails{
			Urls:  failedUrls,
			Count: len(failedUrls),
		},
	}

	return links
}

func getPageContainsLoginForm(link string, doc goquery.Document) bool {
	// Attempt to find login form, but not a consistent logic
	// Check page contains login form
	is_login_form := false
	is_login_url := false
	email := false
	password := false
	other := false

	// Check login slug terms in URL
	for _, b := range loginPageSlugList {
		if strings.Contains(link, b) {
			is_login_url = true
		}
	}

	doc.Find("input[type]").Each(func(index int, item *goquery.Selection) {
		_type, _ := item.Attr("type")

		if _type == "password" {
			password = true
		}

		if _type == "email" {
			email = true
		}

		if _type != "text" {
			other = true
		}
	})

	if is_login_url || (password && (email || other) && (other && !email)) {
		is_login_form = true
	}
	return is_login_form
}

func WebScraper(link string, doc goquery.Document) models.ScraperResponse {
	// All scraper functions response merged
	jsonData := models.ScraperResponse{
		Url:                   link,
		HtmlVersion:           getHtmlVersion(doc),
		PageTitle:             getPageTitle(doc),
		HeadingCount:          getHeadingCount(doc),
		Links:                 getAllLinks(doc),
		PageContainsLoginForm: getPageContainsLoginForm(link, doc),
	}
	return jsonData
}
