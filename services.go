package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/net/html"
)

type ServiceResponse struct {
	name      string
	available bool
	url       string
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

	available := true
	url := ""
	if len(results.Repositories) > 0 {
		available = false
		url = results.Repositories[0].GetHTMLURL()
	}

	return ServiceResponse{
		name:      "github",
		available: available,
		url:       url,
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

	available := false
	if g.existsPackage(tokenizer, className) {
		available = true
	}

	return ServiceResponse{
		name:      "gopkg",
		available: available,
		url:       url,
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

type Homebrew struct{}

func (h Homebrew) Check(name string) (ServiceResponse, error) {
	url := fmt.Sprintf("https://formulae.brew.sh/formula/%s", name)
	available, err := checkStatus404(url)
	if err != nil {
		return ServiceResponse{}, err
	}

	return ServiceResponse{
		name:      "homebrew",
		available: available,
		url:       url,
	}, nil
}

type Npm struct{}

func (n Npm) Check(name string) (ServiceResponse, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s", name)
	available, err := checkStatus404(url)
	if err != nil {
		return ServiceResponse{}, err
	}

	return ServiceResponse{
		name:      "npm",
		available: available,
		url:       url,
	}, nil
}

type Pypi struct{}

func (p Pypi) Check(name string) (ServiceResponse, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", name)
	available, err := checkStatus404(url)
	if err != nil {
		return ServiceResponse{}, err
	}

	return ServiceResponse{
		name:      "pypi",
		available: available,
		url:       url,
	}, nil

}

type RubyGems struct{}

func (r RubyGems) Check(name string) (ServiceResponse, error) {
	url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", name)
	available, err := checkStatus404(url)
	if err != nil {
		return ServiceResponse{}, err
	}

	return ServiceResponse{
		name:      "rubygems",
		available: available,
		url:       url,
	}, nil
}

type Crate struct{}

func (c Crate) Check(name string) (ServiceResponse, error) {
	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", name)
	available, err := checkStatus404(url)
	if err != nil {
		return ServiceResponse{}, err
	}

	return ServiceResponse{
		name:      "crates",
		available: available,
		url:       url,
	}, nil
}

type PackageResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Repository  string `json:"repository"`
	Downloads   int    `json:"downloads"`
	Favers      int    `json:"favers"`
}

type APIResponse struct {
	Results []PackageResult `json:"results"`
	Total   int             `json:"total"`
	Next    string          `json:"next"`
}

type Packagist struct{}

func (p Packagist) Check(name string) (ServiceResponse, error) {
	url := fmt.Sprintf("https://packagist.org/search.json?q=%s", name)
	response, err := http.Get(url)
	if err != nil {
		return ServiceResponse{}, nil
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ServiceResponse{}, err
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return ServiceResponse{}, err
	}

	available := false
	if len(apiResponse.Results) == 0 {
		available = true
	}

	return ServiceResponse{
		name:      "packagist",
		available: available,
		url:       url,
	}, nil
}

func checkStatus404(url string) (bool, error) {
	response, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return true, nil
	}

	return false, nil
}
