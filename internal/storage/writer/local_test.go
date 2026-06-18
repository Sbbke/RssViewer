package writer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestLocalRss validates the LocalWriter asset generation pipeline
func TestLocalRss(t *testing.T) {
	// 1. Utilize t.TempDir() to securely provision an isolated test workspace
	testTmpDir := t.TempDir()

	// 2. Initialize your system under test using the sandboxed environment path
	w := NewLocalWriter(testTmpDir)

	// 3. Define structured mock binary payload allocations
	mockImages := [][]byte{
		[]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01"), // Mock PNG header data
		[]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x02"),
	}

	var mockTopicID int64 = 42
	mockRssHash := "8a9f3b1c2d"

	// 4. Execute the targeted functional write routine
	err := w.UpdateTopicSlide(mockTopicID, mockRssHash, mockImages)
	if err != nil {
		t.Fatalf("UpdateTopicSlide execution failed: %v", err)
	}

	// 5. Conduct state verification assertions against target files
	for i := range mockImages {
		expectedPath := filepath.Join(testTmpDir, "summary", "topic", "42_8a9f3b1c2d", "slide", fmt.Sprintf("%d.png", i))
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected slide asset %d not found at destination path: %s", i, expectedPath)
		}
	}

	targetFilePath := filepath.Join(testTmpDir, "summary", "topic", "42_8a9f3b1c2d", "slide", "0.png")
	got, err := os.ReadFile(targetFilePath)
	if err != nil {
		t.Fatalf("could not read written target asset file: %v", err)
	}
	if !bytes.Equal(got, mockImages[0]) {
		t.Errorf("content payload mismatch: got %x, want %x", got, mockImages[0])
	}

	// 6. Execute an overwrite update to confirm old directory layout state is fully replaced
	updatedImages := [][]byte{
		[]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x03"),
	}
	err = w.UpdateTopicSlide(mockTopicID, mockRssHash, updatedImages)
	if err != nil {
		t.Fatalf("second execution of UpdateTopicSlide failed: %v", err)
	}

	// Confirm old slide 1.png is completely removed from the filesystem view
	stalePath := filepath.Join(testTmpDir, "summary", "topic", "42_8a9f3b1c2d", "slide", "1.png")
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Errorf("stale asset from previous write loop remains; directory atomic swap was not completely clean")
	}
}

// TestLocalRss_UnwritableDir validates failure path handling under constrained filesystem security contexts
func TestLocalRss_UnwritableDir(t *testing.T) {
	testTmpDir := t.TempDir()
	lockedDir := filepath.Join(testTmpDir, "summary")

	if err := os.MkdirAll(lockedDir, 0755); err != nil {
		t.Fatalf("failed to establish test directory tree layout: %v", err)
	}

	// Enforce structural read-only bitmasks to inhibit write allocations
	if err := os.Chmod(lockedDir, 0444); err != nil {
		t.Fatalf("failed to adjust filesystem security mode flags: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(lockedDir, 0755)
	})

	w := NewLocalWriter(testTmpDir)
	err := w.UpdateTopicSlide(42, "hash", [][]byte{[]byte("data")})
	if err == nil {
		t.Error("expected access denial or permission configuration failure block, but execution returned nil error context")
	}
}
