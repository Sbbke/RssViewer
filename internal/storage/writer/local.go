package storage

import (
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

func (w *LocalWriter) CreatePostSlide(ID int64, images [][]byte) error {
	folder := strconv.FormatInt(ID, 10)
	targetDir := filepath.Join(w.Path, "summary", "post", folder, "slide")

	if err := w.checkDirExist(targetDir); err != nil {
		return err
	}

	for i, img := range images {

		fp := filepath.Join(targetDir, fmt.Sprintf("%d.png", i))

		if err := os.WriteFile(fp, img, 0644); err != nil {
			return fmt.Errorf("error saving image to: %s, %w", fp, err)
		}
	}
	return nil
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

func (w *LocalWriter) CreateTopicSlide(ID int64, rssHash string, images [][]byte) error {

	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	targetDir := filepath.Join(w.Path, "summary", "topic", topicDirName, "slide")

	if err := w.checkDirExist(targetDir); err != nil {
		return err
	}

	for i, img := range images {

		fp := filepath.Join(targetDir, fmt.Sprintf("%d.png", i))

		if err := os.WriteFile(fp, img, 0644); err != nil {
			return fmt.Errorf("error saving image to: %s, %w", fp, err)
		}
	}
	return nil
}

// Update() implementation
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

func (w *LocalWriter) UpdatePostSlide(ID int64, images [][]byte) error {
	folder := strconv.FormatInt(ID, 10)
	finalDir := filepath.Join(w.Path, "summary", "post", folder, "slide")
	if err := w.atomicDirSwap(finalDir, images); err != nil {
		return err
	}
	return nil
}

func (w *LocalWriter) UpdateTopicSlide(ID int64, rssHash string, images [][]byte) error {

	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	targetDir := filepath.Join(w.Path, "summary", "topic", topicDirName, "slide")
	if err := w.atomicDirSwap(targetDir, images); err != nil {
		return err
	}
	return nil
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

func (w *LocalWriter) DeletePostSlide(ID int64) error {
	folder := strconv.FormatInt(ID, 10)
	target := filepath.Join(w.Path, "summary", "post", folder, "slide")
	return w.removePath(target)
}

func (w *LocalWriter) DeleteTopicSummary(ID int64, rssHash string) error {
	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	target := filepath.Join(w.Path, "summary", "topic", topicDirName, "summary.txt")
	return w.removePath(target)
}

func (w *LocalWriter) DeleteTopicSlide(ID int64, rssHash string) error {
	topicDirName := fmt.Sprintf("%d_%s", ID, rssHash)
	target := filepath.Join(w.Path, "summary", "topic", topicDirName, "slide")
	return w.removePath(target)
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

	closed := false
	renamed := false
	dir := filepath.Dir(path)
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

	for i, img := range images {
		fp := filepath.Join(tmpPath, fmt.Sprintf("%d.png", i))
		if err := os.WriteFile(fp, img, 0600); err != nil {
			if err := os.RemoveAll(tmpPath); err != nil {
				log.Printf("error removing path :%v", err)
			}
			return fmt.Errorf("failed to write slide image: %w", err)
		}
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
