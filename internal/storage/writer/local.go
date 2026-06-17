package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type LocalWriter struct{
	Path string
}

func NewLocalWriter(folder_path string) *LocalWriter{
	return &LocalWriter{
		Path : folder_path,
	}	
}

func (w *LocalWriter) checkDirExist(path string) error{
	err := os.MkdirAll(path,0755) //from golang doc:  If path is already a directory, MkdirAll does nothing and returns nil.

	if err !=nil {
		return fmt.Errorf("error creating directory: %s, %w", path, err)
	}

	return nil
}
func (w *LocalWriter) CreateRss( ID int64, content string) error{

	folder := strconv.FormatInt(ID, 10)
	rp := filepath.Join(w.Path, "rss", folder)
	
	if err := w.checkDirExist(rp); err != nil{
		return err
	}

	r := filepath.Join(rp, "rss.xml")	
	err := os.WriteFile(r, []byte(content), 0644)
	if err != nil{
		return fmt.Errorf("error saving raw rss to: %s, %w", rp, err)
	}

	return nil
}

func (w *LocalWriter) CreatePostSummary( ID int64, content string) error{
	folder := strconv.FormatInt(ID, 10)
	targetDir := filepath.Join(w.Path, "summary","post", folder)
	
	if err := w.checkDirExist(targetDir); err != nil{
		return err
	}


	fp := filepath.Join(targetDir, "summary.txt")

	if err := os.WriteFile(fp, []byte(content), 0644); err != nil{
		return fmt.Errorf("error saving summary to: %s, %w", fp, err)	
	}
	
	return nil
}

func (w *LocalWriter) CreateTopicSummary( ID int64, rssHash string, content string) error{

	folder := strconv.FormatInt(ID, 10)
	topicDirName := fmt.Sprintf("%s_%s", folder, rssHash)
	targetDir := filepath.Join(w.Path, "summary","topic",topicDirName)	
	if err := w.checkDirExist(targetDir); err != nil{
		return err
	}


	fp := filepath.Join(targetDir, "summary.txt")

	if err := os.WriteFile(fp, []byte(content), 0644); err != nil{
		return fmt.Errorf("error saving summary to: %s, %w", fp, err)	
	}
	
	return nil
}

func (w *LocalWriter) CreateTopicSlide( ID int64, rssHash string, images [][]byte) error{


	folder := strconv.FormatInt(ID, 10)
	topicDirName := fmt.Sprintf("%s_%s", folder, rssHash)
	targetDir := filepath.Join(w.Path, "summary","topic",topicDirName,"slide")	
		
	if err := w.checkDirExist(targetDir); err != nil{
		return err
	}

	
	for i, img := range images{
	
		fp := filepath.Join(targetDir, fmt.Sprintf("%d.png", i))
	
		if err := os.WriteFile(fp, img, 0644); err != nil{
			return fmt.Errorf("error saving image to: %s, %w", fp, err)	
		}
	}	
	return nil
}


func (w *LocalWriter) Update() error{
	return nil
}

func (w *LocalWriter) Delete() error{
	return nil
}



