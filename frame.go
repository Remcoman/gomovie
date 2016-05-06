package gomovie

import (
	"fmt"
	"image"
	"io"
)

type Frame struct {
	Bytes  []byte
	Width  int
	Height int
	Index  int
}

func (f *Frame) String() string {
	return fmt.Sprintf("Index: %d, Width: %d, Height: %d", f.Index, f.Width, f.Height)
}

func (f *Frame) ToNRGBAImage() *image.NRGBA {
	return &image.NRGBA{
		Pix:    f.Bytes,
		Stride: f.Width * 4,
		Rect:   image.Rect(0, 0, f.Width, f.Height),
	}
}

type FrameReader interface {
	ReadFrame() (*Frame, error)
	VideoInfo() *VideoInfo
	
	io.Reader
}