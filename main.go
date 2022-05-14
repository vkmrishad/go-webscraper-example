package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i > 0
}

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

func getHeadingCount(doc goquery.Document) HeadingCount {
	// Get all headings count
	headings := HeadingCount{
		H1: doc.Find("h1").Length(),
		H2: doc.Find("h2").Length(),
		H3: doc.Find("h3").Length(),
		H4: doc.Find("h4").Length(),
		H5: doc.Find("h5").Length(),
		H6: doc.Find("h16").Length(),
	}
	return headings
}

type link struct {
	links string `json:"links"`
	count string `json:"count"`
}

type links struct {
	internal []link `json:"internal"`
	external []link `json:"external"`
	failed   []link `json:"failed"`
}

type LinkDetails struct {
	Urls  []string `json:"urls"`
	Count int      `json:"count"`
}

type Links struct {
	Internal LinkDetails `json:"internal"`
	External LinkDetails `json:"external"`
	Failed   LinkDetails `json:"failed"`
}

type AllLinks struct {
	Links Links `json:"links"`
}

func getAllLinks(doc goquery.Document) AllLinks {
	var internalUrls []string
	var externalUrls []string
	doc.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")
		// fmt.Printf("link: %s - anchor text: %s\n", href, item.Text())

		u, err := url.Parse(href)
		if err != nil {
			log.Fatal(err)
		}

		if doc.Url.Host == u.Host || (u.Host == "" && !contains(internalUrlsExceptions, href)) {
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

	// // Print all urls we could not open:
	// fmt.Println("Internal Links: ", internalLinks)
	// fmt.Println("External Links: ", externalLinks)
	// fmt.Println("Could not fetch these urls: ", failedUrls)

	links := AllLinks{
		Links: Links{
			Internal: LinkDetails{
				Urls:  internalUrls,
				Count: len(internalUrls),
			},
			External: LinkDetails{
				Urls:  externalUrls,
				Count: len(externalUrls),
			},
			Failed: LinkDetails{
				Urls:  failedUrls,
				Count: len(failedUrls),
			},
		},
	}

	return links
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Health check endpoint
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "API is up and running")
}

type HeadingCount struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

type ScraperResponse struct {
	Url          string       `json:"url"`
	HtmlVersion  string       `json:"html_version"`
	PageTitle    string       `json:"page_title"`
	HeadingCount HeadingCount `json:"heading_count"`
	Links        AllLinks     `json:"links"`
}

type ScraperRequestBody struct {
	Url string `validate:"required", json:"Url"`
}

func (a *ScraperRequestBody) validate() url.Values {
	err := url.Values{}

	// Write your own validation rules
	if a.Url == "" {
		err.Add("Url", "This field is required")
	}

	_, e := url.ParseRequestURI(a.Url)
	if e != nil && a.Url != "" {
		err.Add("Url", "Not a valid URL")
	}

	return err
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorMessage struct {
	Error Error `json:"errors"`
}

func Scraper(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	requestBody := &ScraperRequestBody{}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(requestBody); err != nil {
		msg := ErrorMessage{
			Error: Error{Code: http.StatusBadRequest, Message: "Request body is empty"},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msg)
		return
	}

	if errors := requestBody.validate(); len(errors) > 0 {
		err := map[string]interface{}{"validationError": errors}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	} else {
		_url := requestBody.Url

		doc, err := goquery.NewDocument(_url)

		if err != nil {
			msg := ErrorMessage{
				Error: Error{Code: http.StatusBadRequest, Message: "URL is not reachable"},
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(msg)
			return
		}

		// HTML version
		fmt.Printf("\nVersion: %s", getHtmlVersion(*doc))

		// Get title
		fmt.Printf("\ntitle: %s", getPageTitle(*doc))

		// Get all headings count
		jsonStr, err := json.Marshal(getHeadingCount(*doc))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\nheadingCount: %s", string(jsonStr))

		// All Links
		a := getAllLinks(*doc)

		fmt.Print(a)

		response := ScraperResponse{
			Url:          _url,
			HtmlVersion:  getHtmlVersion(*doc),
			PageTitle:    getPageTitle(*doc),
			HeadingCount: getHeadingCount(*doc),
			Links:        a,
		}

		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Println("Unable to encode JSON")
		}

		fmt.Println(string(jsonResponse))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/health-check", HealthCheck).Methods("GET")
	router.HandleFunc("/scraper", Scraper).Methods("POST")
	http.Handle("/", router)

	//start and listen to requests
	fmt.Println("Starting API server")
	http.ListenAndServe(":8080", router)

}
