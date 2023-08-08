package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/net/html"
)

type ServiceResponse struct {
	name   string
	exists bool
	url    string
}

type NameChecker interface {
	Check(name string) (ServiceResponse, error)
}

type Github struct{}

func (g Github) Check(name string) (ServiceResponse, error) {
	client := github.NewClient(nil)

	results, _, err := client.Search.Repositories(context.Background(), name, nil)
	if err != nil {
		return ServiceResponse{}, err
	}

	exists := false
	url := ""
	if len(results.Repositories) > 0 {
		exists = true
		url = results.Repositories[0].GetHTMLURL()
	}

	return ServiceResponse{
		name:   "github",
		exists: exists,
		url:    url,
	}, nil
}

type GoPkg struct{}

func (g GoPkg) Check(name string) (ServiceResponse, error) {
	url := fmt.Sprintf("https://pkg.go.dev/search?q=%s&m=package", name)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return ServiceResponse{}, err
	}
	defer response.Body.Close()

	className := "go-GopherMessage"
	tokenizer := html.NewTokenizer(response.Body)

	exists := true
	if g.existsPackage(tokenizer, className) {
		exists = false
	}

	return ServiceResponse{
		name:   "gopkg",
		exists: exists,
		url:    url,
	}, nil
}

func (g GoPkg) existsPackage(tokenizer *html.Tokenizer, className string) bool {
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return false
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			for _, attr := range token.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, className) {
					return true
				}
			}
		}
	}
}
