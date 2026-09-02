package dto

import (
	images "RssViewer/internal/storage/images/local"
	"time"
)

type BriefingSlideResponse struct {
	Slides    [][]byte `json:"slides"` // image slides
	CreatedAt string   `json:"createdAt"`
}

type BriefingTextResponse struct {
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}
type TopicAllInOne struct {
	TopicID   int64                  `json:"topicId"`
	Rss       []RssDetailResponse    `json:"rss"`
	Summary   *BriefingTextResponse  `json:"summary,omitempty"`
	Slide     *BriefingSlideResponse `json:"slide,omitempty"`
	CreatedAt string                 `json:"createdAt"`
}

type TopicResponse struct {
	TopicID   int64                 `json:"topicId"`
	Rss       []RssItem             `json:"rss"`
	Name      string                `json:"name"`
	Summary   *BriefingTextResponse `json:"summary,omitempty"` // Inline pointer: nil means "not generated yet"
	SummaryID int64                 `json:"summaryId"`
	CreatedAt string                `json:"createdAt"`
}

type RssItem struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	SubscribedAt time.Time `json:"subscribedAt"`
}

type RssResponse struct {
	Info  RssItem    `json:"info"`
	Posts []PostItem `json:"posts"`
}

type RssDetailResponse struct {
	Info  RssItem              `json:"info"`
	Posts []PostDetailResponse `json:"posts"`
}

type PostItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	PublishedAt string `json:"publishedAt"`
}

type PostSummaryResponse struct {
	Meta    PostItem              `json:"meta"`
	Summary *BriefingTextResponse `json:"summary,omitempty"`
}

type PostDetailResponse struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	PublishedAt string                 `json:"publishedAt"`
	Content     string                 `json:"content"`
	Summary     *BriefingTextResponse  `json:"summary,omitempty"`
	Slide       *BriefingSlideResponse `json:"slide,omitempty"`
}


// ==========================================
// REQUEST PAYLOADS (Frontend -> Go)
// ==========================================

type SummaryPayload struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
}

type TopicPayload struct {
	Name string `json:"name"`
}

type TopicSlidePayload struct {
	TopicID int64            `json:"topicId"`
	Slide   [][]byte         `json:"slide"`
	Meta    images.BriefingMeta `json:"meta"`
}

type PostSlidePayload struct {
	PostID int64            `json:"postId"`
	Slide  [][]byte         `json:"slide"`
	Meta   images.BriefingMeta `json:"meta"`
}

// SlideDeletePayload identifies one specific briefing (topic or post) to
// remove — a bare ID is no longer enough since slides are keyed by hash.
type SlideDeletePayload struct {
	ID   int64  `json:"id"`
	Hash string `json:"hash"`
}

type RssPayload struct {
	Title string `json:"title"`
	Url   string `json:"url"`
	Xml   []byte `json:"xml"`
}

type RssUpdatePayload struct {
	ID   int64      `json:"id"`
	Body RssPayload `json:"body"`
}

type PostPayload struct {
	RssID       int64  `json:"rssId"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	Content     string `json:"content"`
	PublishedAt string `json:"publishedAt"`
}



type MutationResult struct {
	GeneratedID int64  `json:"generatedId"`
	Err         string `json:"err,omitempty"` // String error message for JS compatibility
}

type LinkRssTopicPayload struct {
	RssID   int64
	TopicID int64
}
