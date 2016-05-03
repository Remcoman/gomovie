package main

import (
	"github.com/Remcoman/gomovie"
	
	"os"
	"testing"
	"io"
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
	
	out := &gomovie.VideoOutput{
		Path : "test.mp4",
	}
	
	t.Log("Writing video to test.mp4")
	
	totalRead, err := io.Copy(out, vid)
	if err != nil {
		t.Log(err)
		t.Fatal(totalRead)
	}
	
	out.Close()
}