package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/vkmrishad/go-webscraper-example/common"
	"github.com/vkmrishad/go-webscraper-example/models"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

var internalUrlsExceptions = []string{"mailto:", "tel:"}
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
	if err != nil || (resp.StatusCode != 200 && resp.StatusCode != 999) {
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

func getAllLinks(doc goquery.Document) models.AllLinks {
	var internalUrls []string
	var externalUrls []string
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")
		// fmt.Printf("link: %s - anchor text: %s\n", href, item.Text())

		u, err := url.Parse(href)
		if err != nil {
			log.Fatal(err)
		}

		if doc.Url.Host == u.Host || (u.Host == "" && !common.Contains(internalUrlsExceptions, href)) {
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

	links := models.AllLinks{
		Links: models.Links{
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
		},
	}
	return links
}

func Scraper(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	requestBody := &models.ScraperRequestBody{}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(requestBody); err != nil {
		msg := models.ErrorMessage{
			Error: models.Error{Code: http.StatusBadRequest, Message: "Request body is empty"},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msg)
		return
	}

	if errors := common.Validate(requestBody); len(errors) > 0 {
		err := map[string]interface{}{"validationError": errors}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	} else {
		_url := requestBody.Url

		doc, err := goquery.NewDocument(_url)

		if err != nil {
			msg := models.ErrorMessage{
				Error: models.Error{Code: http.StatusBadRequest, Message: "URL is not reachable"},
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(msg)
			return
		}

		response := models.ScraperResponse{
			Url:          _url,
			HtmlVersion:  getHtmlVersion(*doc),
			PageTitle:    getPageTitle(*doc),
			HeadingCount: getHeadingCount(*doc),
			Links:        getAllLinks(*doc),
		}

		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Println("Unable to encode JSON")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}
