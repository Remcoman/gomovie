package gomovie

import (
	"image"
	"io"
)

//Frame describes a single rgba frame
type Frame struct {
	Data   []byte
	Width  int
	Height int
	Index  int
	Time   float32
}

//ToNRGBAImage converts the frame to a image.NRGBA
func (f *Frame) ToNRGBAImage() *image.NRGBA {
	return &image.NRGBA{
		Pix:    f.Data,
		Stride: f.Width * 4,
		Rect:   image.Rect(0, 0, f.Width, f.Height),
	}
}

//FrameReaderInfo contains information about an Video stream in a video file
type FrameReaderInfo struct {
	CodecName string
	Width     int
	Height    int
	Rotation  int
	FrameRate float32
	Duration  float32
}

//FrameReader describes an interface to read frames from a video
type FrameReader interface {
	io.ReadCloser

	Slice(r *Range) FrameReader
	Range() *Range

	//Read a single frame from the framereader
	ReadFrame() (*Frame, error)

	//Get information about the frame format and src
	Info() *FrameReaderInfo
}
