package service

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTMLProcessor interface {
	Run(html []byte) (string, error)
}

type PostService struct {
	orch      *storage.DataOrch
	processor HTMLProcessor
	client    *http.Client
}

func NewPostService(orch *storage.DataOrch, processor HTMLProcessor) *PostService {
	return &PostService{
		orch:      orch,
		processor: processor,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

// GetContent assembles the full detail view for a single post: SQL metadata
// (title, publish date) plus freshly fetched and processed article content.
//
// The `content` column populated at ingestion time is the RSS feed's own
// (often truncated) summary; GetPostByID intentionally does not return it as
// the canonical body. This method re-fetches the original article HTML via
// GetPostURL and runs it through the HTMLProcessor to get the full text.
func (s *PostService) GetContent(postID int64) (*dto.PostDetailResponse, error) {
	meta, err := s.orch.GetReader().GetPostByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post service: getContent: postID=%d: %w", postID, err)
	}

	url, err := s.orch.GetReader().GetPostURL(postID)
	if err != nil {
		return nil, fmt.Errorf("post service: getContent: getPostURL: postID=%d: %w", postID, err)
	}

	html, err := s.fetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("post service: getContent: fetch article: postID=%d: %w", postID, err)
	}

	content, err := s.processor.Run(html)
	if err != nil {
		return nil, fmt.Errorf("post service: getContent: process article: postID=%d: %w", postID, err)
	}

	return &dto.PostDetailResponse{
		ID:          meta.ID,
		Title:       meta.Title,
		PublishedAt: meta.PublishedAt,
		Content:     content,
		// Summary and Slide are populated by dedicated summary/slide flows,
		// not part of this method's responsibility.
	}, nil
}

func (s *PostService) fetchHTML(url string) ([]byte, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("network fetch failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return body, nil
}
