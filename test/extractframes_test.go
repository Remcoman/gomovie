package main

import (
	"github.com/Remcoman/gomovie"
	"image/png"
	"os"
	"testing"
)

func TestFrame2Png(t *testing.T) {
	video := os.Getenv("GOMOVIE_VIDEO")
	if video == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	grabber := gomovie.CreateGrabber(video)
	grabber.Open()

	frame, err := grabber.GetFrame()

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
