package gomovie_test

import (
	"os"
	"testing"

	"github.com/Remcoman/gomovie"
)

func TestTransformFrames(t *testing.T) {
	return

	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	video, err := gomovie.FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}

	video.SampleReader = nil

	frameTransformer := gomovie.NewFrameTransformer(video)
	frameTransformer.ParallelCount = 5

	frameTransformer.AddTransform(gomovie.FrameTransform{
		Transform: func(f *gomovie.Frame) {
			t.Logf("Transforming frame: %v", f.Index)
		},
	})

	config := gomovie.WriteConfig{
		VideoCodec:        "libx264",
		DebugFFmpegOutput: true,
	}

	err = gomovie.FfmpegWrite("videos/test.mp4", frameTransformer, config)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransformSamples(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	video, err := gomovie.FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}

	video.FrameReader = nil

	sampleTransformer := gomovie.NewSampleTransformer(video)

	sampleTransformer.AddTransform(gomovie.SampleTransform{
		Transform: func(s *gomovie.SampleBlock, depth int) {
			s.ConvertFloats(func(i int, f float32) float32 {
				return f * .5
			})
		},
	})

	config := gomovie.WriteConfig{
		AudioCodec:        "mp3",
		DebugFFmpegOutput: true,
	}

	err = gomovie.FfmpegWrite("videos/transformed.mp3", sampleTransformer, config)
	if err != nil {
		t.Fatal(err)
	}
}
