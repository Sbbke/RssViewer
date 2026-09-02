package service

import (
	"context"
	"fmt"

	"RssViewer/internal/dto"
	"RssViewer/internal/storage"
)

type BriefingService struct {
	orch *storage.DataOrch
}

func NewBriefingService(orch *storage.DataOrch) *BriefingService {
	return &BriefingService{
		orch: orch,
	}
}

// GetLatestBriefing retrieves the latest briefing data for a Topic.
//
// Topic metadata and RSS/Post data are read from DBReader.
// Generated summary/slide assets are read from LocalReader.
//
func (s *BriefingService) GetLatestBriefingByTopic(
	ctx context.Context,
	topicID int64,
) (*dto.BriefingSlideResponse, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	if err := validateID(topicID, "topicID"); err != nil {
		return nil, err
	} 

	reader := s.orch.GetLocalReader()
	if reader == nil {
		return nil, fmt.Errorf("database reader is not available")
	}
	

	// [][]byte
	slide, err := reader.ReadTopicSlide(topicID, "")
	if err != nil{
		return nil, fmt.Errorf("Failed to get slide f")
	}
	result := dto.BriefingSlideResponse{
		Slides: slide,
	}

	return &result, nil
}

// GenerateBriefing generates a briefing for a Topic or Post.
//
// The actual LLM / generation component is not part of DataOrch's
// documented API, so this method validates the request and dispatches
// to the appropriate generation path. The generated result should
// then be persisted using DataOrch's summary/slide mutation APIs.
func (s *BriefingService) GenerateBriefing(
	ctx context.Context,
	targetID int64,
	targetType string,
) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateID(targetID, "targetID"); err != nil {
		return err
	}

	switch targetType {
	case "topic":
		return s.generateTopicBriefing(ctx, targetID)

	case "post":
		return s.generatePostBriefing(ctx, targetID)

	default:
		return fmt.Errorf(
			"unsupported targetType %q: expected %q or %q",
			targetType,
			"topic",
			"post",
		)
	}
}

// GetBriefingByTopic retrieves topic metadata together with generated
// briefing information.
//
// NOTE:
// Resolving the Topic summary currently requires rssHash, but the
// documented DBReader API does not expose that value.
func (s *BriefingService) GetBriefingByTopic(
	ctx context.Context,
	topicID int64,
	rssID []int64,
) (*dto.BriefingSlideResponse, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	if err := validateID(topicID, "topicID"); err != nil {
		return nil, err
	}
	
	for id := range rssID {
		if err := validateID(int64(id), "rssID"); err != nil{
			return nil, err
		}
	}
	reader := s.orch.GetLocalReader()
	if reader == nil {
		return nil, fmt.Errorf("database reader is not available")
	}
	// cal rssHash
	rssHash := ""	
	slides, err := reader.ReadTopicSlide(topicID, rssHash)
	if err != nil {
		return nil, err
	}

	// The TopicResponse can be returned directly for the SQL-backed
	// portion of the response.
	//
	// Summary cannot be populated because LocalReader requires:
	//
	//     ReadTopicSummary(topicID, rssHash)
	//
	// and rssHash is not exposed by DBReader.
	res := dto.BriefingSlideResponse{
		Slides: slides,
	}
	return &res, nil
}

// GetBriefingByPost retrieves a Post together with its generated
// summary and slide assets.
func (s *BriefingService) GetBriefingByPost(
	ctx context.Context,
	postID int64,
	rssID int64,
) (*dto.PostDetailResponse, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	if err := validateID(postID, "postID"); err != nil {
		return nil, err
	}

	if err := validateID(rssID, "rssID"); err != nil {
		return nil, err
	}

	reader := s.orch.GetReader()
	if reader == nil {
		return nil, fmt.Errorf("database reader is not available")
	}

	post, err := reader.GetPostByID(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post %d: %w", postID, err)
	}

	// Validate that the supplied RSS actually contains the Post.
	posts, err := reader.GetPostsByRss(rssID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to validate post %d against rss %d: %w",
			postID,
			rssID,
			err,
		)
	}

	found := false
	for _, item := range posts {
		if item.ID == postID {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf(
			"post %d does not belong to rss %d",
			postID,
			rssID,
		)
	}

	result := dto.PostDetailResponse{
		ID:          post.ID,
		Title:       post.Title,
		PublishedAt: post.PublishedAt,
		Content:     post.Content,
	}

	result, err = s.attachPostAssets(result)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to attach briefing assets for post %d: %w",
			postID,
			err,
		)
	}

	return &result, nil
}

func (s *BriefingService) attachPostAssets(
	post dto.PostDetailResponse,
) (dto.PostDetailResponse, error) {
	localReader := s.orch.GetLocalReader()
	if localReader == nil {
		return post, fmt.Errorf("local reader is not available")
	}

	// -----------------------------
	// Summary
	// -----------------------------
	summary, err := localReader.ReadPostSummary(post.ID)
	if err == nil {
		post.Summary = &dto.BriefingTextResponse{
			Body:      summary,
			CreatedAt: "",
		}
	}

	// -----------------------------
	// Slides
	// -----------------------------
	slides, err := localReader.ReadPostSlide(post.ID)
	if err != nil {
		return post, nil
	}

	if len(slides) == 0 {
		return post, nil
	}

	post.Slide = &dto.BriefingSlideResponse{
		Slides:    slides,
		CreatedAt: "",
	}

	return post, nil
}

func (s *BriefingService) generateTopicBriefing(
	ctx context.Context,
	topicID int64,
) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	reader := s.orch.GetReader()
	if reader == nil {
		return fmt.Errorf("database reader is not available")
	}

	// This is the actual inference input described by the DBReader API:
	// Topic -> Posts within the requested generation window.
	//
	// The generation window should eventually be determined by the
	// application's briefing policy (weekly/monthly/etc.).
	//
	// from := ...
	// to   := ...
	// posts, err := reader.GetPostsByTopicInWindow(topicID, from, to)

	_ = reader

	return fmt.Errorf("topic briefing generation pipeline is not configured")
}

func (s *BriefingService) generatePostBriefing(
	ctx context.Context,
	postID int64,
) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	reader := s.orch.GetReader()
	if reader == nil {
		return fmt.Errorf("database reader is not available")
	}

	post, err := reader.GetPostByID(postID)
	if err != nil {
		return fmt.Errorf("failed to get post %d: %w", postID, err)
	}

	// post is the source document passed to the briefing generator.
	//
	// After generation:
	//
	// s.orch.AddPostSummary(dto.SummaryPayload{
	//     ID:      postID,
	//     Content: generatedText,
	// })
	//
	// s.orch.AddPostSlide(dto.PostSlidePayload{
	//     PostID: postID,
	//     Slide:  generatedSlides,
	// })

	_ = post

	return fmt.Errorf("post briefing generation pipeline is not configured")
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context must not be nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func validateID(id int64, name string) error {
	if id <= 0 {
		return fmt.Errorf("%s must be greater than zero", name)
	}

	return nil
}
