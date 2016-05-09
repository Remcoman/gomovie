package gomovie

import (
	"io"
)

type Sample16 int16

func (p Sample16) ToFloat() float32 {
	return float32(p) / float32(32768.)
}

type SampleReader interface {
	ReadSample() (*[]Sample16, error)
	AudioInfo() *AudioInfo
	io.Reader
}
