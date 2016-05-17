package main

import (
	"github.com/Remcoman/gomovie"
	
	"os"
	"testing"
)

func TestWriteAudio(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}
	
	vid := gomovie.OpenVideo(path)
	vid.Video = nil
	
	t.Log("Writing audio to ../videos/test.raw")
	
	config := gomovie.Config{
		Codec : "libx264",
		DebugFFmpegOutput : true,
	}
	
	if err := gomovie.Encode("../videos/test.raw", vid, config); err != nil {
		t.Fatal(err)
	}
}