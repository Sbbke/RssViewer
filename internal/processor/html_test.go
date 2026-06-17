package processor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHtml_Success(t *testing.T) {
	// 1. Create the test server and verify the request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Optional: Assert that the method is GET
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><h1>Hello, Go!</h1></body></html>`))
	}))
	defer server.Close()

	hp := NewHtmlProcessor()
	
	err := hp.Run(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestHtml_ServerError(t *testing.T) {
	// Simulate a failing remote server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	hp := NewHtmlProcessor()
	
	err := hp.Run(server.URL)
	if err == nil {
		t.Fatal("expected an error due to 500 Status Code, but got nil")
	}
}

func TestCleanHtml(t *testing.T) {
	rawHtml := "<html><header>some header</header><body>Hello World!</body><footer>some footer</footer></html><script>here is some script</script>"			
	
	expected := "Hello World!"

	hp := NewHtmlProcessor()
	result, err := hp.CleanHTML(rawHtml)
	if err !=nil{
		t.Fatalf("something went wrong")
	}

	if result != expected{
		t.Fatalf("Did not process html correctly")
	}

	filepath := "/home/user/RssViewer/test/post1.html"
	file, err := os.ReadFile(filepath)
	if err != nil{
		t.Fatalf("error reading test html")
	}
	content, err := hp.CleanHTMLToMarkdown(string(file))
	if err != nil{
		t.Fatalf("error processing test html")
	}

	fmt.Println(content)
}

