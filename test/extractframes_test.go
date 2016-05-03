package main

import (
	"github.com/Remcoman/gomovie"
	"image/png"
	"os"
	"testing"
)

func TestFrame2Png(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	video := gomovie.VideoSource{Path : path}
	video.Open()

	frame, err := video.ReadFrame()

	if err != nil {
		t.Fatal("Could not get frame from video")
	}

	img := frame.ToNRGBAImage()

	file, err := os.Create("frame.png")
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		t.Fatal(err)
	}
}
