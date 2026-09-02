package images

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type BriefingMeta struct {
	Hash        string  `json:"hash"`
	Timeframe   string  `json:"timeframe"`   // "weekly", "monthly", "custom"
	PeriodStart string  `json:"periodStart"`
	PeriodEnd   string  `json:"periodEnd"`
	RssIDs      []int64 `json:"rssIds"`
	NumSlides   int     `json:"numSlides"`
	CreatedAt   string  `json:"createdAt"`
}

type LocalReader struct {
	basePath string
}

func NewLocalReader(path string) *LocalReader {
	return &LocalReader{
		basePath: path,
	}
}

func (r *LocalReader) ReadRss(ID int64) (string, error) {
	targetPath := filepath.Join(r.basePath, "rss", fmt.Sprintf("%d.xml", ID))
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read raw rss payload configuration: %w", err)
	}
	return string(data), nil
}
func (r *LocalReader) ReadTopicSummary(ID int64, rssHash string) (string, error) {
	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	targetPath := filepath.Join(r.basePath, "summary", "topic", topicDirName, "summary.txt")

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read target topic summary payload: %w", err)
	}
	return string(data), nil

}

func (r *LocalReader) ReadTopicSlide(ID int64, rssHash string) ([][]byte, error) {
	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	slideDir := filepath.Join(r.basePath, "summary", "topic", topicDirName, "slide")

	return r.readOrderedImagesFromDir(slideDir)
}

func (r *LocalReader) GetTopicSlideMeta(ID int64) ([]BriefingMeta, error){
	metaPath := filepath.Join(r.basePath, "summary", "topic", fmt.Sprintf("%d", ID), "meta.json")
	return r.readBriefingMeta(metaPath)
}

func (r *LocalReader) ReadPostSummary(ID int64) (string, error) {
	targetPath := filepath.Join(r.basePath, "summary", "post", fmt.Sprintf("%d", ID), "summary.txt")

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read historical target post summary payload: %w", err)
	}
	return string(data), nil

}

func (r *LocalReader) ReadPostSlide(ID int64) ([][]byte, error) {
	slideDir := filepath.Join(r.basePath, "summary", "post", fmt.Sprintf("%d", ID), "slide")

	return r.readOrderedImagesFromDir(slideDir)
}



func (r *LocalReader) GetPostMeta(ID int64) ([]BriefingMeta, error){
	metaPath := filepath.Join(r.basePath, "summary", "post", fmt.Sprintf("%d", ID),"meta.json")
	return r.readBriefingMeta(metaPath)
}


func (r *LocalReader) readBriefingMeta(dirPath string) ([]BriefingMeta, error) {
	var res []BriefingMeta

	data, err := os.ReadFile(dirPath)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("unmarshal json failed: %w", err)
	}

	return res, nil
}

func (r *LocalReader) readOrderedImagesFromDir(dirPath string) ([][]byte, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan structural asset directory boundary: %w", err)
	}

	type indexedImage struct {
		index int
		data  []byte
	}
	var collection []indexedImage

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
			continue
		}

		// Strip string extensions to isolate the precise sequence number
		nameNoExt := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		idx, err := strconv.Atoi(nameNoExt)
		if err != nil {
			// Skip unindexed or descriptive asset filenames seamlessly
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract file data byte context from %s: %w", entry.Name(), err)
		}

		collection = append(collection, indexedImage{index: idx, data: fileData})
	}

	// Enforce numerical sorting constraints explicitly
	sort.Slice(collection, func(i, j int) bool {
		return collection[i].index < collection[j].index
	})

	// Construct clean flat arrays matching interface declarations
	result := make([][]byte, len(collection))
	for i, item := range collection {
		result[i] = item.data
	}

	return result, nil
}
