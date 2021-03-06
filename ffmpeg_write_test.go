package gomovie_test

import (
	"image/png"
	"os"
	"testing"

	"github.com/Remcoman/gomovie"
)

func TestAudioToMP3(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	video, err := gomovie.FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}
	video.FrameReader = nil //<-- no video

	config := gomovie.WriteConfig{AudioCodec: "mp3"}

	err = gomovie.FfmpegWrite("videos/test.mp3", video, config)
	if err != nil {
		t.Fatal("Could not write audio")
	}
}

func TestFrame2Png(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	video, err := gomovie.FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open the video")
	}

	frame, err := video.ReadFrame()
	if err != nil {
		t.Fatal("Could not get frame from video")
	}

	img := frame.ToNRGBAImage()

	file, err := os.Create("videos/frame.png")
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteFrame(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	vid, err := gomovie.FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}

	t.Log("Writing video to ../videos/test.mp4")

	config := gomovie.WriteConfig{
		VideoCodec: "libx264",
		AudioCodec: "mp3",
		ProgressCallback: func(progress float32) {
			t.Log(progress)
		},
	}

	if err := gomovie.FfmpegWrite("../videos/test.mp4", vid, config); err != nil {
		t.Fatal(err)
	}
}
