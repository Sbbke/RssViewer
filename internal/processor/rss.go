package processor

import(
	"github.com/mmcdole/gofeed"
	"fmt"
)

type RssProcessor struct {
	parser *gofeed.Parser
}

func NewRssProcessor() *RssProcessor{
	return &RssProcessor{
		parser: gofeed.NewParser(),
	}
}

func (r *RssProcessor) Run(content string) error{
	feed , err := r.parser.ParseString(content)
	if err != nil{
		return fmt.Errorf("failed to parse %s: %w", content, err)
	}
	fmt.Println(feed.Title)
	return nil
}

