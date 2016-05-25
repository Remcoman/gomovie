package gomovie

import (
	"os"
	"testing"
)

func TestOpenVideo(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	_, err := FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}
}
