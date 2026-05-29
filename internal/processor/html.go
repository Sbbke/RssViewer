package processor

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
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
