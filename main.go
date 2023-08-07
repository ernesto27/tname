package main

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// homebrew
// https://formulae.brew.sh/formula/dagger

func main() {
	url := "https://pkg.go.dev/search?q=docker&m=package"

	// Make the HTTP GET request
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	className := "go-GopherMessage"
	tokenizer := html.NewTokenizer(response.Body)

	if checkClass(tokenizer, className) {
		fmt.Println("package not exists")
	} else {
		fmt.Println("package exists")
	}
}

func checkClass(tokenizer *html.Tokenizer, className string) bool {
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
