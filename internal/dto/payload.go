package dto

import "time"

type BriefingSlideResponse struct {
	Slides    []string  `json:"slides"` // Generated asset paths
	CreatedAt time.Time `json:"createdAt"`
}

type BriefingTextResponse struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type TopicAllInOne struct {
    TopicID   int64               `json:"topicId"`
    Rss       []RssDetailResponse `json:"rss"`
    Summary   *BriefingTextResponse `json:"summary"`
    Slide     *BriefingSlideResponse `json:"slide"`
    CreatedAt time.Time           `json:"createdAt"`
}

type TopicResponse struct {
	TopicID   int64                 `json:"topicId"`
	Rss       []RssItem      `json:"rss"`
	Name string `json:"name"`
	Summary   *BriefingTextResponse `json:"summary"`   // Inline pointer: nil means "not generated yet"
	SummaryID int64                 `json:"summaryId"` 
	CreatedAt time.Time             `json:"createdAt"`
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

type RssDetailResponse struct{
	Info RssItem
	Posts []PostDetailResponse
}

type PostItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"publishedAt"`
}

type PostSummaryResponse struct {
	Meta    PostItem             `json:"meta"`
	Summary *BriefingTextResponse `json:"summary"`
}

type PostDetailResponse struct {
    ID          int64                 `json:"id"`
    Title       string                `json:"title"`
    PublishedAt time.Time             `json:"publishedAt"`
    Content     string                `json:"content"`
    Summary     *BriefingTextResponse `json:"summary,omitempty"`
    Slide       *BriefingSlideResponse `json:"slide,omitempty"`
}

type TopicPayload struct {
	Name string
}

type TopicSummaryPayload struct{
	TopicID int64
	Body string
}

type TopicSlidePayload struct{
	TopicID int64
	Slide [][]byte
}


type RssPayload struct {
	TopicID int64
	Title   string
	Url     string
	Content string
}

type PostPayload struct {
	RssID       int64
	Title       string
	Url         string
	PublishedAt time.Time
}

type PostSummaryPayload struct{
	PostID int64
	Body string
}

type PostSlidePayload struct{
	PostID int64
	Slide [][]byte
}

type MutationResult struct {
	GeneratedID int64
	Err         error
}
