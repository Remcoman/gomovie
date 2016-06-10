package gomovie

import "io"

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

//SampleFormat describes the format of a SampleBlock
type SampleFormat struct {
	Depth     int
	BlockSize int
}

//NewSampleFormat creates a SampleFormat with the default values
func NewSampleFormat() *SampleFormat {
	return &SampleFormat{Depth: 16, BlockSize: GlobalConfig.SampleBlockSize}
}

//SampleBlock describes a chunk of sample values
type SampleBlock struct {
	*SampleFormat

	Data     interface{}
	Time     float32
	Duration float32
}

func (sb *SampleBlock) Bytes() (bd []byte) {
	switch d := sb.Data.(type) {
	case []SampleInt16:
		bd = make([]byte, len(d)*2)
		for i, x := range d {
			v := uint16(x)
			bd[i] = byte(v)
			bd[i+1] = byte(v >> 8)
		}
	case []SampleInt32:
		bd = make([]byte, len(d)*4)
		for i, x := range d {
			v := uint32(x)
			bd[i] = byte(v)
			bd[i+1] = byte(v >> 8)
			bd[i+2] = byte(v >> 16)
			bd[i+3] = byte(v >> 24)
		}
	}
	return
}

//ConvertFloats converts each value to a float (normalized between 0 and 1) and passes it to the given callback. The callback is expected to return a modified float value.
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

//SampleReaderInfo contains information about an Audio stream in a video file
type SampleReaderInfo struct {
	CodecName  string
	Duration   float32
	SampleRate int
	Channels   int
}

//SampleReader describes an interface to read audio sample blocks
type SampleReader interface {
	io.ReadCloser

	Slice(r *Range) SampleReader
	Range() *Range

	//read a single sample block in the format SampleInt16 or SampleInt32 (depending on SampleDepth)
	ReadSampleBlock() (*SampleBlock, error)

	SampleFormat() *SampleFormat

	//information about the sample reader
	Info() *SampleReaderInfo
}
