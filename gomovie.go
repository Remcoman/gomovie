package gomovie

import (
	"fmt"
	"image"
	"math"
	"strings"

	"io"
)

func formatSize(width int, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

func formatTime(time float64) string {
	hour := string(math.Floor(time / 3600.))
	if len(hour) < 2 {
		hour = "0" + hour
	}

	min := string(math.Mod(math.Floor(time/60.), 60.))
	if len(min) < 2 {
		min = "0" + min
	}

	seconds := string(math.Mod(time, 60))
	if len(seconds) < 2 {
		seconds = "0" + seconds
	}

	return hour + ":" + min + ":" + seconds
}

type EOV struct {
	msg string
}

func (e EOV) Error() string {
	return e.msg
}

type Frame struct {
	Bytes  []byte
	Width  int
	Height int
	Index  int
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
	Info() *Info

	io.WriterTo
}

type ConcatenatedVideo struct {
	readers []FrameReader

	index int
}

func (f *ConcatenatedVideo) Info() *Info {
	sumInfo := new(Info)

	for reader := range f.readers {
		info := reader.Info()
		sumInfo.Duration += info.Duration
		sumInfo.Width = math.Max(float64(info.Width), float64(sumInfo.Width))
		sumInfo.Height = math.Max(float64(info.Height), float64(sumInfo.Height))
	}

	sumInfo.FrameRate = f.readers[0].Info().FrameRate
	sumInfo.Rotation = 0

	return sumInfo
}

func (f *ConcatenatedVideo) ReadFrame() (fr *Frame, err error) {
	for {
		if fr, err = f.readers[f.index].ReadFrame(); err == nil {
			return
		} else {
			f.index = f.index + 1
			if f.index >= len(f.readers) {
				return nil, EOV{msg: "No more frames"}
			}
		}
	}
}

func (f *ConcatenatedVideo) WriteTo(w io.Writer) (n int64, err error) {
	var fr *Frame

	n = 0

	for {
		fr, err = f.ReadFrame()

		if err != nil {
			return
		}

		w, err = w.Write(fr.Bytes)
		n += w

		if err != nil {
			return
		}
	}

	return
}

func Concat(readers ...FrameReader) *ConcatenatedVideo {
	return &ConcatenatedVideo{readers, 0}
}
