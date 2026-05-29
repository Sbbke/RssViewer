package processor

import (
	"testing"
	"os"
	"path/filepath"
)


func TestRSS(t *testing.T){
	dir , err:= os.Getwd()
	if err != nil{
		t.Fatalf("failed to get current directory")
	}
		

	path := filepath.Join(dir, "test", "research_google.xml")
	rp := NewRssProcessor()
	feed, err := os.ReadFile(path)
	if err != nil{
		t.Fatalf("error while reading file from path %s", path)
	}

	err =	rp.Run(string(feed))
	if err != nil{
		t.Fatalf("Expected no error, but got: %v", err)
	}

}

