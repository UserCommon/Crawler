package html_utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func FetchHtml(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil

}

func ParseHtml(html_input string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader(html_input))
	if err != nil {
		return nil, err
	}
	return doc, err
}

func ExtractLinks(root *html.Node, link string) ([]string, error) {
	/// Take html page and link where we reading it, get link

	links := make([]string, 0, 10)

	if root == nil {
		return links, nil
	}

	queue := []*html.Node{}
	queue = append(queue, root)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		switch node.Type {
		case html.ElementNode:
			if node.Data == "a" {
				for _, attr := range node.Attr {
					if attr.Key == "href" {
						if isRelative(attr.Val) {
							links = append(links, fmt.Sprintf("%s%s", link, attr.Val))
						} else {
							links = append(links, attr.Val)
						}
					}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			queue = append(queue, child)
		}
	}
	return links, nil
}

func isRelative(url string) bool {
	return strings.HasPrefix(url, "/")
}
