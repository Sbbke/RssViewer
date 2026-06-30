package service

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage"
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

func (s *PostService) GetContent(postID int64) (*dto.PostDetailResponse, error) {

	return nil, nil

}
