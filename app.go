package main

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/service"
	"RssViewer/internal/storage"
	"context"
	"fmt"
	"log"
	"os"
)

// App struct
type App struct {
	ctx          context.Context
	orch         *storage.DataOrch
	topicService *service.TopicService
	rssService *service.RssService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.orch = setupOrch()
	a.topicService = service.NewTopicService(a.orch)
	a.rssService = service.NewRssService(a.orch)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetTopics returns all topics. Rejects the JS promise on error.
func (a *App) GetTopics() ([]dto.TopicResponse, error) {
	topics, err := a.topicService.GetTopics()
	if err != nil {
		return nil, err
	}
	return topics, nil
}

// GetTopic returns a single topic (with its RSS feeds) by ID.
func (a *App) GetTopic(topicID int64) (dto.TopicResponse, error) {
	return a.topicService.GetTopicResponse(topicID)
}

// CreateTopic creates a new topic and returns it.
func (a *App) CreateTopic(name string) (dto.TopicResponse, error) {
	return a.topicService.CreateTopic(name)
}

// DeleteTopic deletes a topic by ID.
func (a *App) DeleteTopic(id int64) error {
	return a.topicService.DeleteTopic(id)
}

// CheckRssUpdate checks all subscribed feeds for new posts.
func (a *App) CheckRssUpdate() error {
	return a.rssService.CheckUpdate()
}

// RemoveRss deletes an RSS feed by ID.
func (a *App) RemoveRss(id int64) error {
	return a.rssService.RemoveRss(id)
}
func (a *App) SubmitRssUrl(rssURL string, topicID *int64) (dto.RssItem, error) {
	return a.rssService.SubmitRssUrl(a.ctx, rssURL, topicID)
}

// LinkRssToTopic associates an existing feed with a topic.
func (a *App) LinkRssToTopic(rssID, topicID int64) error {
	return a.rssService.LinkRssToTopic(rssID, topicID)
}
func (a *App) UnlinkRssFromTopic(rssID, topicID int64) error {
	return a.rssService.UnlinkRssFromTopic(rssID, topicID)
}
// GetAllRss returns every RSS feed, regardless of topic linkage.
func (a *App) GetAllRss() ([]dto.RssItem, error) {
	return a.rssService.GetAllRss()
}
func (a *App) GetRssDetail(rssID int64) (dto.RssResponse, error) {
	return a.rssService.GetRssDetail(rssID)
}
// GetStandaloneRss returns feeds not linked to any topic.
func (a *App) GetStandaloneRss() ([]dto.RssItem, error) {
	return a.rssService.GetStandaloneRss()
}
func setupOrch() *storage.DataOrch {
	if err := os.MkdirAll("temp", 0755); err != nil {
		log.Fatalf("error creating temp dir: %v", err)
	}

	da, err := storage.NewSqliteDB("temp/db")
	if err != nil {
		log.Fatalf("error initilizing sqlite db: %v", err)
	}
	orch, err := storage.NewDataOrch(da, "temp/local")
	if err != nil {
		log.Fatalf("new orch: %v", err)
	}
	return orch
}
