package gomovie

import (
	"errors"
)

var GlobalConfig = struct {
	FfmpegPath      string
	FfprobePath     string
	SampleBlockSize int
	FramePixelDepth int
}{
	FfmpegPath:      "/usr/bin/ffmpeg",
	FfprobePath:     "/usr/bin/ffprobe",
	SampleBlockSize: 512,
	FramePixelDepth: 4,
}

var emptyReaderError = errors.New("Duration of 0 is not allowed")

type Range struct {
	parent *Range

	Start    float32
	Duration float32
}

func (r *Range) Intersection(r2 *Range) *Range {
	e1 := r.Start + r.Duration
	e2 := r2.Start + r2.Duration

	e := e1
	if e2 < e1 {
		e = e2
	}

	s := r.Start
	if r2.Start > r.Start {
		s = r2.Start
		if s > e {
			s = e
		}
	}

	return &Range{Start: s, Duration: e - s}
}

func (r *Range) AbsStart() float32 {
	i := r
	var s float32
	for i != nil {
		s += i.Start
		i = i.parent
	}
	return s
}

//Video wraps the FrameReader and SampleReader
type Video struct {
	FrameReader
	SampleReader
}

func (v *Video) Slice(r *Range) (*Video, error) {
	return &Video{v.FrameReader.Slice(r), v.SampleReader.Slice(r)}, nil
}

//Info returns information about the FrameReader and SampleReader
func (v *Video) Info() (frameReaderInfo *FrameReaderInfo, sampleReaderInfo *SampleReaderInfo) {
	if v.FrameReader != nil {
		frameReaderInfo = v.FrameReader.Info()
	}

	if v.SampleReader != nil {
		sampleReaderInfo = v.SampleReader.Info()
	}

	return
}

//Close closes both the FrameReader and SampleReader
func (v *Video) Close() (err error) {
	if v.FrameReader != nil {
		if err = v.FrameReader.Close(); err != nil {
			return
		}
	}

	if v.SampleReader != nil {
		err = v.SampleReader.Close()
	}

	return
}
