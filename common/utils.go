package common

import (
	"net/url"
	"sort"

	"github.com/vkmrishad/go-webscraper-example/models"
)

func Contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i > 0
}

func Validate(a *models.ScraperRequestBody) url.Values {
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
