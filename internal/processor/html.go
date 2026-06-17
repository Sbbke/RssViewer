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

func (hp *HtmlProcessor) CleanHTML(rawHTML string) (string, error) {
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

func (hp *HtmlProcessor) CleanHTMLToMarkdown(rawHTML string) (string, error) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	// Define a recursive function to traverse the HTML node tree and apply Markdown rules
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			switch tag {
			// Skip non-content metadata and layout blocks entirely
			case "script", "style", "svg", "nav", "header", "footer", "noscript", "form", "iframe":
				return

			// Headers
			case "h1":
				buf.WriteString("\n\n# ")
			case "h2":
				buf.WriteString("\n\n## ")
			case "h3":
				buf.WriteString("\n\n### ")

			// Paragraphs and Line Breaks
			case "p", "div":
				buf.WriteString("\n\n")
			case "br":
				buf.WriteString("\n")

			// Lists
			case "li":
				buf.WriteString("\n* ")

			// Inline Text Formatting
			case "strong", "b":
				buf.WriteString("**")
			case "em", "i":
				buf.WriteString("*")

			// Links: Capture href attribute if present
			case "a":
				var href string
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						href = attr.Val
						break
					}
				}
				buf.WriteString("[")
				// Recurse children to evaluate text inside the link anchor
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
				buf.WriteString(fmt.Sprintf("](%s)", href))
				return // Skip default child traversal since it was handled manually for the anchor text
			}
		}

		// Append plain text nodes
		if n.Type == html.TextNode {
			// Trim native newlines from the raw text node to prevent markdown fragmentation
			text := strings.ReplaceAll(n.Data, "\n", " ")
			buf.WriteString(text)
		}

		// Continue standard depth-first traversal for child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}

		// Add closing syntax for inline formatting blocks after visiting children
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			switch tag {
			case "strong", "b":
				buf.WriteString("**")
			case "em", "i":
				buf.WriteString("*")
			case "h1", "h2", "h3", "p", "div":
				buf.WriteString("\n\n")
			}
		}
	}

	f(doc)

	result := buf.String()

	// Replace tabs and multiple horizontal spaces with a single space
	spaceRe := regexp.MustCompile(`[ \t]+`)
	result = spaceRe.ReplaceAllString(result, " ")

	// Strip spaces trailing at the end of lines or preceding a newline to prevent markdown corruption
	lineSpaceRe := regexp.MustCompile(`(?m)[ \t]+\n`)
	result = lineSpaceRe.ReplaceAllString(result, "\n")

	// Collapse three or more consecutive newlines down to exactly two consecutive newlines
	newlineRe := regexp.MustCompile(`\n{3,}`)
	result = newlineRe.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result), nil
}
