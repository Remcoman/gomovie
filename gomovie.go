package gomovie

import (
	"image"
	"io"
)

var GlobalConfig = struct {
	FfmpegPath  string
	FfprobePath string
}{
	FfmpegPath:  "/usr/bin/ffmpeg",
	FfprobePath: "/usr/bin/ffprobe",
}

//SampleInt16 describes a single 16 bit sample
type SampleInt16 int16

//ToFloat Normalizes the sample to a value between -1 and 1
func (p SampleInt16) Float() float32 {
	return float32(p) / float32(32768.)
}

//SampleInt32 describes a single 32 bit sample
type SampleInt32 int32

//ToFloat Normalizes the sample to a value between -1 and 1
func (p SampleInt32) Float() float32 {
	return float32(p) / float32(2147483648.)
}

type Sample interface {
	Float() float32
}

type SampleBlock struct {
	Data     interface{}
	Time     float32
	Duration float32
}

func (sb *SampleBlock) ConvertFloats(fn func(i int, f float32) float32) {
	switch t := sb.Data.(type) {
	case []SampleInt16:
		for i, v := range t {
			t[i] = SampleInt16(fn(i, v.Float()) * 32768.)
		}
	case []SampleInt32:
		for i, v := range t {
			t[i] = SampleInt32(fn(i, v.Float()) * 2147483648.)
		}
	}
}

//SampleReader describes an interface to read audio sample blocks
type SampleReader interface {

	//read a single sample block in the format SampleInt16 or SampleInt32 (depending on SampleDepth)
	ReadSampleBlock() (*SampleBlock, error)

	//the sample bit depth
	SampleDepth() int

	//information about the sample format and src
	Info() *SampleSrcInfo

	io.ReadCloser
}

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

//FrameSrcInfo contains information about an Video stream in a video file
type FrameSrcInfo struct {
	CodecName string
	Width     int
	Height    int
	Rotation  int
	FrameRate float32
	Duration  float32
}

//SampleSrcInfo contains information about an Audio stream in a video file
type SampleSrcInfo struct {
	CodecName  string
	Duration   float32
	SampleRate int
	Channels   int
}

//FrameReader describes an interface to read frames from a video
type FrameReader interface {
	//Read a single frame from the framereader
	ReadFrame() (*Frame, error)

	//Get information about the frame format and src
	Info() *FrameSrcInfo

	io.ReadCloser
}

//Video wraps the FrameReader and SampleReader
type VideoReader struct {
	FrameReader
	SampleReader
}

//Info returns information about the FrameReader and SampleReader
func (v *VideoReader) Info() (frameInfo *FrameSrcInfo, audioInfo *SampleSrcInfo) {
	if v.FrameReader != nil {
		frameInfo = v.FrameReader.Info()
	}

	if v.SampleReader != nil {
		audioInfo = v.SampleReader.Info()
	}

	return
}

//Close closes both the FrameReader and SampleReader
func (v *VideoReader) Close() (err error) {
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
