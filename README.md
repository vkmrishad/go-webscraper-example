# go-webscraper-example

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/vkmrishad/go-webscraper-example.svg)](https://github.com/vkmrishad/go-webscraper-example)
[![Lint Status](https://github.com/vkmrishad/go-webscraper-example/workflows/test/badge.svg)](https://github.com/vkmrishad/go-webscraper-example/actions)

A web application which takes a website URL as an input and provides general information
about the contents of the page
- HTML Version
- Page Title
- Headings count by level
- Amount of internal and external links
- Amount of inaccessible links
- If a page contains a login form (NB: Login is not consistent)

## Clone the project

```
git clone https://github.com/vkmrishad/go-webscraper-example.git
$ cd go-webscraper-example
```

## Project tree

```
.
├── Dockerfile
├── README.md
├── controllers
│   ├── common.go
│   └── scraper.go
├── go-webscraper-example
├── go.mod
├── go.sum
├── main.go
├── models
│   ├── common.go
│   └── scraper.go
├── test
│   └── endpoint_test.go
└── utility
    ├── common.go
    └── scraper.go
```

## Runserver using run command
```
$ go run .
or 
$ go run main.go
```

## Runserver using build

```
$ go build .
$ ./go-webscraper-example
```

## Runserver using Docker

```
$ docker build -t go-webscraper-example . --rm
$ docker run -p 8080:8080 --name go-webscraper-example --rm go-webscraper-example
```

## Run tests

```
$ go test ./test -v
```

Server URL: http://127.0.0.1:8080

## Example Request and Response

Endpoint: [POST]http://127.0.0.1:8080/scraper/

Request
```
POST /scraper/ HTTP/1.1
Host: 127.0.0.1:8080
Content-Type: application/json
Cookie: csrftoken=WhdzUpxwGJm5UBpiSb2fxavIQ8zFGcHKdTWJkUOWRG4993AevH1rYe0QS3b8QYzr
Content-Length: 43

{
    "url": "https://mohammedrishad.com"
}
```

Response
```
{
    "url": "https://mohammedrishad.com",
    "html_version": "HTML 5",
    "page_title": "Mohammed Rishad",
    "heading_count": {
        "h1": 1,
        "h2": 0,
        "h3": 1,
        "h4": 0,
        "h5": 0,
        "h6": 0
    },
    "links": {
        "links": {
            "internal": {
                "urls": [
                    "https://www.mohammedrishad.com/#body",
                    "https://www.mohammedrishad.com/#skill_sec",
                    "https://www.mohammedrishad.com/#work_sec",
                    "https://www.mohammedrishad.com/#edu_sec",
                    "https://www.mohammedrishad.com/#exp_sec",
                    "https://www.mohammedrishad.com/#achivement_sec",
                    "https://www.mohammedrishad.com/#interest_sec",
                    "https://www.mohammedrishad.com/#contact_sec",
                    "https://www.mohammedrishad.com/#body",
                    "https://www.mohammedrishad.com/MohammedRishadResume.pdf",
                    "https://www.mohammedrishad.com/#body"
                ],
                "count": 11
            },
            "external": {
                "urls": [
                    "https://studysmarter.de",
                    "https://uoc.ac.in/",
                    "https://www.nuflights.com/",
                    "https://www.cliphire.co/",
                    "https://moonllight.com/",
                    "https://transportsimple.com/",
                    "https://www.etailpet.com/",
                    "https://www.feedmyfurbaby.co.nz/",
                    "mailto:mohammedrishad.vk@gmail.com",
                    "https://www.linkedin.com/in/vkmrishad",
                    "https://github.com/vkmrishad",
                    "https://stackoverflow.com/users/6874947/mohammed-rishad",
                    "https://www.facebook.com/vkmrishad",
                    "https://www.twitter.com/vkmrishad",
                    "https://www.youtube.com/channel/UCvnIlVDRk46_xpnrrAiYrTg"
                ],
                "count": 15
            },
            "failed": {
                "urls": [
                    "mailto:mohammedrishad.vk@gmail.com",
                    "https://moonllight.com/",
                    "https://uoc.ac.in/",
                    "https://www.linkedin.com/in/vkmrishad"
                ],
                "count": 4
            }
        }
    },
    "page_contains_login_form": false
}
```
