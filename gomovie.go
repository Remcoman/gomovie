package gomovie

import (
	"fmt"
	"math"
	"strconv"
	"image"
	"io"
)

var FfmpegPath string = "/usr/bin/ffmpeg"
var FfprobePath string = "/usr/bin/ffprobe"

type SampleInt16 int16

func (p SampleInt16) ToFloat() float32 {
	return float32(p) / float32(32768.)
}

type AudioReader interface {
	ReadSample() ([]SampleInt16, error)
	Info() *AudioInfo
	io.Reader
}

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

type VideoReader interface {
	ReadFrame() (*Frame, error)
	Info() *VideoInfo
	
	io.Reader
}

type Video struct {
	VideoReader
	AudioReader
}

func (v *Video) Info() (videoInfo *VideoInfo, audioInfo *AudioInfo) {
	if v.VideoReader != nil {
		videoInfo = v.VideoReader.Info()
	}
	
	if v.AudioReader != nil {
		audioInfo = v.AudioReader.Info()
	}
	
	return
}

func FormatSize(width int, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

func FormatTime(time float64) string {
	
	hour := strconv.FormatFloat(math.Floor(time / 3600.), 'f', 0, 32)
	if len(hour) < 2 {
		hour = "0" + hour
	}

	min := strconv.FormatFloat(math.Mod(math.Floor(time/60.), 60.), 'f', 0, 32)
	if len(min) < 2 {
		min = "0" + min
	}

	seconds := strconv.FormatFloat(math.Mod(time, 60), 'f', 0, 32)
	if len(seconds) < 2 {
		seconds = "0" + seconds
	}

	return hour + ":" + min + ":" + seconds
}
