package gomovie

import (
	"io"
)

type SampleInt16 int16

func (p SampleInt16) ToFloat() float32 {
	return float32(p) / float32(32768.)
}

type SampleReader interface {
	ReadSample() (*[]SampleInt16, error)
	AudioInfo() *AudioInfo
	io.Reader
}
