package gomovie_test

import (
	"os"
	"testing"

	"github.com/Remcoman/gomovie"
)

func TestOpenVideo(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	_, err := gomovie.FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}
}
