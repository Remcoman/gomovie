package gomovie

import (
	"os"
	"testing"
)

func TestTransformFrames(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	video, err := FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}

	video.SampleReader = nil

	frameTransformer := NewFrameTransformer(video)

	frameTransformer.AddTransform(FrameTransform{
		Transform: func(f *Frame) {
			t.Logf("Transforming frame: %v", f.Index)
		},
	})

	config := WriteConfig{
		VideoCodec: "libx264",
	}

	err = FfmpegWrite("../videos/transformed.mp4", frameTransformer, config)
	if err != nil {
		t.Fatal("Can not write video!")
	}
}

func TestTransformSamples(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	video, err := FfmpegOpen(path)
	if err != nil {
		t.Fatal("Could not open video")
	}

	video.FrameReader = nil

	sampleTransformer := NewSampleTransformer(video)

	sampleTransformer.AddTransform(SampleTransform{
		Transform: func(s *SampleBlock, depth int) {
			s.ConvertFloats(func(i int, f float32) float32 {
				return f * .5
			})
		},
	})

	config := WriteConfig{
		AudioCodec: "mp3",
	}

	err = FfmpegWrite("../videos/transformed.mp3", sampleTransformer, config)
	if err != nil {
		t.Fatal("Can not write audio!")
	}
}
