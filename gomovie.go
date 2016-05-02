package gomovie

import (
	"image"
)

type EOV struct {
	msg string
}

func (e EOV) Error() string {
	return e.msg
}

type Frame struct {
	Bytes []byte
	Width int
	Height int
	Index int
}

func (f *Frame) ToNRGBAImage() *image.NRGBA {
	return &image.NRGBA{
		Pix : f.Bytes,
		Stride : f.Width * 4,
		Rect : image.Rect(0, 0, f.Width, f.Height),
	}
}

type FrameReader interface {
	ReadFrame() (*Frame, error)
	Info() *Info
}

type ConcatenatedVideo struct {
	readers []FrameReader
	
	index int
}

func (f *ConcatenatedVideo) Duration() *Info {
	return new(Info)
}

func (f *ConcatenatedVideo) ReadFrame() (fr *Frame, err error) {
	for {
		if fr, err = f.readers[f.index].ReadFrame(); err == nil {
			return
		} else {
			f.index = f.index + 1
			if f.index >= len(f.readers) {
				return nil, EOV{msg : "No more frames"}
			}
		}
	}
}

func Concat(readers ...FrameReader) *ConcatenatedVideo {
	return &ConcatenatedVideo{readers, 0}
}