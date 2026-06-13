package processor

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
	"bytes"
	"regexp"
	"strings"
	"fmt"
)

type HtmlProcessor struct{
}

func NewHtmlProcessor() *HtmlProcessor{
	return &HtmlProcessor{}
}

func (hp *HtmlProcessor) Run(url string) error{

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil{
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK{
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.Body{
			fmt.Println(n.Attr)
			for _, a := range n.Attr {
				fmt.Println(a.Val)
			}
		}
	}

	return nil
}

func CleanHTML(rawHTML string) (string, error) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	
	// Define a recursive function to traverse the HTML node tree
	var f func(*html.Node)
	f = func(n *html.Node) {
		// If the current node is a script, style, nav, svg, etc., skip parsing its children
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			if tag == "script" || tag == "style" || tag == "svg" || tag == "nav" || tag == "header" || tag == "footer" || tag == "noscript" {
				return 
			}
		}

		// If it's a plain text node, append its content
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
			buf.WriteString(" ")
		}

		// Keep traversing down the tree
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	
	f(doc)

	// Regex to collapse multiple spaces, tabs, and newlines down to a single space
	re := regexp.MustCompile(`\s+`)
	cleanText := re.ReplaceAllString(buf.String(), " ")

	return strings.TrimSpace(cleanText), nil
}
