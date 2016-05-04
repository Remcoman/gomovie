package gomovie

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"io"
)

func formatSize(width int, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

func formatTime(time float64) string {
	
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

type PCMReader interface {
	AudioInfo() *AudioInfo
	
	io.Reader
}

type FrameReader interface {
	ReadFrame() (*Frame, error)
	VideoInfo() *VideoInfo
	
	io.Reader
}

type ConcatenatedVideo struct {
	readers []FrameReader

	index int
	videoInfo *VideoInfo
	audioInfo *AudioInfo
}

func (f *ConcatenatedVideo) VideoInfo() *VideoInfo {
	return f.videoInfo
}

func (f *ConcatenatedVideo) AudioInfo() *AudioInfo {
	return f.audioInfo
}

func (f *ConcatenatedVideo) ReadFrame() (fr *Frame, err error) {
	for {
		if fr, err = f.readers[f.index].ReadFrame(); err == nil {
			//TODO if fr size is not the size as the computed size then we need to pad the frame data
			return
		}
		f.index++
		if f.index >= len(f.readers) {
			return nil, io.EOF
		}
	}
}

func Concat(readers ...FrameReader) *ConcatenatedVideo {
	sumInfo := new(VideoInfo)

	for _, reader := range readers {
		info := reader.VideoInfo()
		sumInfo.Duration += info.Duration
		sumInfo.Width = int(math.Max(float64(info.Width), float64(sumInfo.Width)))
		sumInfo.Height = int(math.Max(float64(info.Height), float64(sumInfo.Height)))
		sumInfo.FrameRate = info.FrameRate //throw error if different framerates?
		sumInfo.Rotation = 0
	}
	
	return &ConcatenatedVideo{readers, 0, sumInfo, nil}
}
