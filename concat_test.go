package gomovie_test

import (
	"testing"

	"github.com/Remcoman/gomovie"
)

func TestConcatVideo(t *testing.T) {
	vid1, _ := gomovie.FfmpegOpen("videos/IMG_0273.mov")
	vid1.SampleReader = nil

	vid2, _ := gomovie.FfmpegOpen("videos/IMG_0273 copy.mov")
	vid2.SampleReader = nil
	concatenated := gomovie.Concat(vid1, vid2)

	cfg := gomovie.WriteConfig{
		VideoCodec:        "libx264",
		AudioCodec:        "mp3",
		DebugFFmpegOutput: true,
	}

	gomovie.FfmpegWrite("videos/both.mp4", concatenated, cfg)
}
