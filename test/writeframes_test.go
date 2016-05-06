package main

import (
	"github.com/Remcoman/gomovie"
	
	"os"
	"testing"
)

func TestWriteFrame(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}
	
	vid := &gomovie.VideoSource{
		Path : path,
	}
	vid.Open()
	
	t.Log("Writing video to ../videos/test.mp4")
	
	config := gomovie.Config{
		Codec : "libx264",
		ProgressCallback : func(progress float32) {
			t.Log(progress)
		},
	}
	
	if err := gomovie.Encode("../videos/test.mp4", vid, config); err != nil {
		t.Fatal(err)
	}
}