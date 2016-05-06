package gomovie

import (
	"io"
)

type PCM16Sample int16

func (p PCM16Sample) ToFloat() float32 {
	return float32(p) / float32(32768.) 
}

type PCMReader interface {
	AudioInfo() *AudioInfo	
	io.Reader
}