package images

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type LocalWriter struct {
	Path string
}

func NewLocalWriter(folder_path string) *LocalWriter {
	return &LocalWriter{
		Path: folder_path,
	}
}

func (w *LocalWriter) checkDirExist(path string) error {
	err := os.MkdirAll(path, 0755) //from golang doc:  If path is already a directory, MkdirAll does nothing and returns nil.

	if err != nil {
		return fmt.Errorf("error creating directory: %s, %w", path, err)
	}

	return nil
}
func (w *LocalWriter) CreateRss(ID int64, content string) error {

	folder := strconv.FormatInt(ID, 10)
	rp := filepath.Join(w.Path, "rss", folder)

	if err := w.checkDirExist(rp); err != nil {
		return err
	}

	r := filepath.Join(rp, "rss.xml")
	err := os.WriteFile(r, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error saving raw rss to: %s, %w", rp, err)
	}

	return nil
}

func (w *LocalWriter) CreatePostSummary(ID int64, content string) error {
	folder := strconv.FormatInt(ID, 10)
	targetDir := filepath.Join(w.Path, "summary", "post", folder)

	if err := w.checkDirExist(targetDir); err != nil {
		return err
	}

	fp := filepath.Join(targetDir, "summary.txt")

	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		return fmt.Errorf("error saving summary to: %s, %w", fp, err)
	}

	return nil
}



func (w *LocalWriter) CreatePostSlide(postID int64, meta BriefingMeta, images [][]byte) error {
	if !validHash(meta.Hash) {
		return fmt.Errorf("invalid briefing hash %q", meta.Hash)
	}
	meta.NumSlides = len(images)

	postDir := filepath.Join(w.Path, "summary", "post", strconv.FormatInt(postID, 10))
	slideDir := filepath.Join(postDir, "slide_"+meta.Hash)

	if err := w.checkDirExist(slideDir); err != nil {
		return err
	}
	for i, img := range images {
		fp := filepath.Join(slideDir, fmt.Sprintf("%d.png", i))
		if err := os.WriteFile(fp, img, 0644); err != nil {
			return fmt.Errorf("error saving image to: %s, %w", fp, err)
		}
	}

	metaPath := filepath.Join(postDir, "meta.json")
	return w.upsertBriefingMeta(metaPath, meta, false)
}
func (w *LocalWriter) CreateTopicSummary(ID int64, rssHash string, content string) error {

	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	targetDir := filepath.Join(w.Path, "summary", "topic", topicDirName)
	if err := w.checkDirExist(targetDir); err != nil {
		return err
	}

	fp := filepath.Join(targetDir, "summary.txt")

	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		return fmt.Errorf("error saving summary to: %s, %w", fp, err)
	}

	return nil
}

func (w *LocalWriter) CreateTopicSlide(topicID int64, meta BriefingMeta, images [][]byte) error {
	if !validHash(meta.Hash) {
		return fmt.Errorf("invalid briefing hash %q", meta.Hash)
	}
	meta.NumSlides = len(images)

	topicDir := filepath.Join(w.Path, "summary", "topic", strconv.FormatInt(topicID, 10))
	slideDir := filepath.Join(topicDir, "slide_"+meta.Hash)

	if err := w.checkDirExist(slideDir); err != nil {
		return err
	}
	for i, img := range images {
		fp := filepath.Join(slideDir, fmt.Sprintf("%d.png", i))
		if err := os.WriteFile(fp, img, 0644); err != nil {
			return fmt.Errorf("error saving image to: %s, %w", fp, err)
		}
	}

	metaPath := filepath.Join(topicDir, "meta.json")
	return w.upsertBriefingMeta(metaPath, meta, false /* mustExist */)
}

// first atomicWrite() endpoitn, then Create()
func (w *LocalWriter) UpdateRss(ID int64, content string) error {

	folder := strconv.FormatInt(ID, 10)
	targetDir := filepath.Join(w.Path, "rss", folder, "rss.xml")

	if err := w.atomicWrite(targetDir, []byte(content)); err != nil {
		return err
	}

	return nil
}

func (w *LocalWriter) UpdatePostSummary(ID int64, content string) error {
	folder := strconv.FormatInt(ID, 10)
	targetDir := filepath.Join(w.Path, "summary", "post", folder, "summary.txt")

	if err := w.atomicWrite(targetDir, []byte(content)); err != nil {
		return err
	}
	return nil
}

func (w *LocalWriter) UpdateTopicSummary(ID int64, rssHash string, content string) error {
	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	targetDir := filepath.Join(w.Path, "summary", "topic", topicDirName, "summary.txt")
	if err := w.atomicWrite(targetDir, []byte(content)); err != nil {
		return err
	}
	return nil
}

func (w *LocalWriter) UpdatePostSlide(postID int64, meta BriefingMeta, images [][]byte) error {
	if !validHash(meta.Hash) {
		return fmt.Errorf("invalid briefing hash %q", meta.Hash)
	}
	meta.NumSlides = len(images)

	postDir := filepath.Join(w.Path, "summary", "post", strconv.FormatInt(postID, 10))
	slideDir := filepath.Join(postDir, "slide_"+meta.Hash)

	if err := w.atomicDirSwap(slideDir, images); err != nil {
		return err
	}
	metaPath := filepath.Join(postDir, "meta.json")
	return w.upsertBriefingMeta(metaPath, meta, true)
}

func (w *LocalWriter) UpdateTopicSlide(topicID int64, meta BriefingMeta, images [][]byte) error {
	if !validHash(meta.Hash) {
		return fmt.Errorf("invalid briefing hash %q", meta.Hash)
	}
	meta.NumSlides = len(images)

	topicDir := filepath.Join(w.Path, "summary", "topic", strconv.FormatInt(topicID, 10))
	slideDir := filepath.Join(topicDir, "slide_"+meta.Hash)

	if err := w.atomicDirSwap(slideDir, images); err != nil {
		return err
	}
	metaPath := filepath.Join(topicDir, "meta.json")
	return w.upsertBriefingMeta(metaPath, meta, true /* mustExist */)
}


// Delete() implementation
func (w *LocalWriter) DeleteRss(ID int64) error {
	folder := strconv.FormatInt(ID, 10)
	target := filepath.Join(w.Path, "rss", folder)
	return w.removePath(target)
}

func (w *LocalWriter) DeletePostSummary(ID int64) error {
	folder := strconv.FormatInt(ID, 10)
	target := filepath.Join(w.Path, "summary", "post", folder, "summary.txt")
	return w.removePath(target)
}


func (w *LocalWriter) DeleteTopicSummary(ID int64, rssHash string) error {
	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	target := filepath.Join(w.Path, "summary", "topic", topicDirName, "summary.txt")
	return w.removePath(target)
}

func (w *LocalWriter) DeleteTopicSlide(topicID int64, hash string) error {
	if !validHash(hash) {
		return fmt.Errorf("invalid composition hash %q", hash)
	}
	topicDir := filepath.Join(w.Path, "summary", "topic", strconv.FormatInt(topicID, 10))
	if err := w.removePath(filepath.Join(topicDir, "slide_"+hash)); err != nil {
		return err
	}
	return w.removeBriefingMeta(filepath.Join(topicDir, "meta.json"), hash)
}

func (w *LocalWriter) DeletePostSlide(postID int64, hash string) error {
	if !validHash(hash) {
		return fmt.Errorf("invalid composition hash %q", hash)
	}
	postDir := filepath.Join(w.Path, "summary", "post", strconv.FormatInt(postID, 10))
	if err := w.removePath(filepath.Join(postDir, "slide_"+hash)); err != nil {
		return err
	}
	return w.removeBriefingMeta(filepath.Join(postDir, "meta.json"), hash)
}


// Helper wrapper to safely isolate and remove file artifacts
func (w *LocalWriter) removePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Avoid errors if the data was already cleared
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to clear storage target: %s, %w", path, err)
	}
	return nil
}

// Atomic update: write to temp file, then rename into place
func (w *LocalWriter) atomicWrite(path string, content []byte) error {
	closed := false
	renamed := false
	dir := filepath.Dir(path)

	// Ensure parent directory exists before creating temp file
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to ensure parent dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		if !closed {
			if err := tmp.Close(); err != nil {
				log.Printf("failed to close tmp file securely: %v", err)
			}
		}
		if !renamed {
			if err := os.Remove(tmpPath); err != nil {
				log.Printf("faild to remove path: %v", err)
			}
		}
	}()

	if _, err = tmp.Write(content); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	closed = true

	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file to target: %w", err)
	}
	renamed = true
	return nil
}

// Helper to atomically swap asset slide directories
func (w *LocalWriter) atomicDirSwap(path string, images [][]byte) error {

	// targetPath represents the final destination directory (e.g., "storage/slides/topic_1")
	parentDir := filepath.Dir(path)

	// Ensure the parent directory exists before creating a temp directory inside it
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory structure: %w", err)
	}

	// Create a secure temporary directory in the same filesystem partition to ensure atomic renaming
	tmpDir, err := os.MkdirTemp(parentDir, ".tmp-assets-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	swapped := false
	// Deferred safety cleanup loop
	defer func() {
		if !swapped {
			if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
				log.Printf("critical: failed to clean up temporary directory %s: %v", tmpDir, removeErr)
			}
		}
	}()

	// Write individual payload images sequentially into the temp folder boundary
	for i, img := range images {
		filename := fmt.Sprintf("%d.png", i)
		filePath := filepath.Join(tmpDir, filename)

		if err := os.WriteFile(filePath, img, 0600); err != nil {
			return fmt.Errorf("failed to write individual slide asset %s: %w", filename, err)
		}
	}

	// Remove the stale target directory if it exists to allow a clean folder overwrite
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing target asset directory: %w", err)
	}

	// Execute the POSIX atomic rename migration
	if err := os.Rename(tmpDir, path); err != nil {
		return fmt.Errorf("failed to atomically swap directory to target destination: %w", err)
	}

	swapped = true
	return nil

}
func (w *LocalWriter) upsertBriefingMeta(metaPath string, entry BriefingMeta, mustExist bool) error {
	list, err := readBriefingMetaListAllowMissing(metaPath)
	if err != nil {
		return err
	}

	found := false
	for i, b := range list {
		if b.Hash == entry.Hash {
			list[i] = entry
			found = true
			break
		}
	}
	if !found {
		if mustExist {
			return fmt.Errorf("no existing briefing with hash %q to update in %s", entry.Hash, metaPath)
		}
		list = append(list, entry)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal meta.json for %s: %w", metaPath, err)
	}
	return w.atomicWrite(metaPath, data)
}

func (w *LocalWriter) removeBriefingMeta(metaPath string, hash string) error {
	list, err := readBriefingMetaListAllowMissing(metaPath)
	if err != nil {
		return err
	}
	out := list[:0]
	for _, b := range list {
		if b.Hash != hash {
			out = append(out, b)
		}
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal meta.json for %s: %w", metaPath, err)
	}
	return w.atomicWrite(metaPath, data)
}

func readBriefingMetaListAllowMissing(metaPath string) ([]BriefingMeta, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []BriefingMeta{}, nil
		}
		return nil, fmt.Errorf("failed to read meta.json at %s: %w", metaPath, err)
	}
	var list []BriefingMeta
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta.json at %s: %w", metaPath, err)
	}
	return list, nil
}
