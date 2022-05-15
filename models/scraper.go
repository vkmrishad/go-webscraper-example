package models

type HeadingCount struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
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

type ScraperRequestBody struct {
	Url string `validate:"required" json:"Url"`
}

type ScraperResponse struct {
	Url                   string       `json:"url"`
	HtmlVersion           string       `json:"html_version"`
	PageTitle             string       `json:"page_title"`
	HeadingCount          HeadingCount `json:"heading_count"`
	Links                 Links        `json:"links"`
	PageContainsLoginForm bool         `json:"page_contains_login_form"`
}
