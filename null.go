package gomovie

import (
	"io"
	"math"
)

type nullFrameReader struct {
	i *FrameReaderInfo
	r *Range

	initialized bool
	buf         []byte
	bufSlice    []byte
	frameIndex  int
	frameCount  int
}

func (src *nullFrameReader) Range() *Range          { return src.r }
func (src *nullFrameReader) Info() *FrameReaderInfo { return src.i }

func (src *nullFrameReader) Slice(r *Range) FrameReader {
	if src.r != nil {
		r = r.Intersection(src.r)
		r.parent = src.r
	}
	return &nullFrameReader{i: src.i, r: r}
}

func (src *nullFrameReader) init() {
	src.buf = make([]byte, src.i.Width*src.i.Height*4) //all zero values
	src.bufSlice = src.buf[:]

	tpf := 1. / src.i.FrameRate
	src.frameCount = int(math.Floor(float64(src.i.Duration / tpf)))

	src.initialized = true
}

func (src *nullFrameReader) lastFrame() (n int) {
	n = src.frameCount - 1
	if src.r != nil {
		n = int(src.r.Duration * src.i.FrameRate) //todo check rounding
	}
	return
}

func (src *nullFrameReader) frameIndexToTime(index int) float32 {
	var s float32
	if src.r != nil {
		s = src.r.Start
	}
	return s + float32(index)*float32(1./src.i.FrameRate)
}

func (src *nullFrameReader) ReadFrame() (*Frame, error) {
	if !src.initialized {
		src.init()
	}

	if src.frameIndex > src.lastFrame()-1 {
		return nil, io.EOF
	}

	fr := &Frame{
		Data:   src.buf[:],
		Index:  src.frameIndex,
		Time:   src.frameIndexToTime(src.frameIndex),
		Width:  src.i.Width,
		Height: src.i.Height,
	}

	src.frameIndex++

	return fr, nil
}

func (src *nullFrameReader) Read(p []byte) (n int, err error) {
	if !src.initialized {
		src.init()
	}

	if src.frameIndex > src.lastFrame()-1 {
		return 0, io.EOF
	}

	n = copy(p, src.bufSlice)
	src.bufSlice = src.bufSlice[n:]

	if len(src.bufSlice) == 0 { //no more data left so restart
		src.bufSlice = src.buf[:]
		src.frameIndex++
	}

	return
}

func (src *nullFrameReader) Close() error {
	return nil
}

func NewNullFrameReader(info *FrameReaderInfo) FrameReader {
	return &nullFrameReader{i: info}
}

type nullSampleReader struct {
	i *SampleReaderInfo
	o *SampleFormat
	r *Range

	initialized bool
	totalBytes  int
	sampleData  interface{}
	buf         []byte
	offset      int
}

func (src *nullSampleReader) init() {
	bytesPerSample := src.o.Depth / 8

	src.totalBytes = int(src.i.Duration * float32(bytesPerSample*src.i.Channels*src.i.SampleRate))

	//only used for ReadSampleBlock
	switch src.o.Depth {
	case 16:
		src.sampleData = make([]SampleInt16, src.o.BlockSize/bytesPerSample)
	case 32:
		src.sampleData = make([]SampleInt32, src.o.BlockSize/bytesPerSample)
	}

	//only used for Read
	src.buf = make([]byte, src.totalBytes) //all zero values

	src.initialized = true
}

func (src *nullSampleReader) Info() *SampleReaderInfo     { return src.i }
func (src *nullSampleReader) SampleFormat() *SampleFormat { return src.o }
func (src *nullSampleReader) Close() error                { return nil }
func (src *nullSampleReader) Range() *Range               { return src.r }

func (src *nullSampleReader) Slice(r *Range) SampleReader {
	if src.r != nil {
		r = r.Intersection(src.r)
		r.parent = src.r
	}
	return &nullSampleReader{i: src.i, r: r}
}

func (src *nullSampleReader) Read(p []byte) (n int, err error) {
	if !src.initialized {
		src.init()
	}

	if src.offset >= src.totalBytes {
		return 0, io.EOF
	}

	//loop buffer. Read from bufSlice
	b := src.buf
	if src.offset+len(b) > src.totalBytes { //can't read beyond totalBytes
		b = b[:src.totalBytes-src.offset]
	}

	n = copy(p, b)
	src.offset += n

	return
}

func (src *nullSampleReader) ReadSampleBlock() (*SampleBlock, error) {
	if !src.initialized {
		src.init()
	}

	//get bytes for total duraton
	bytesPerSample := src.o.Depth / 8

	//we should not read beyond allowed range
	sd := src.sampleData
	br := src.o.BlockSize // normal range
	if src.offset+br >= src.totalBytes {
		br = src.totalBytes - br

		switch src.o.Depth {
		case 16:
			sd = sd.([]SampleInt16)[:(br / bytesPerSample)]
		case 32:
			sd = sd.([]SampleInt32)[:(br / bytesPerSample)]
		}
	}

	if br == 0 {
		return nil, io.EOF
	}

	//convert the byte shizzle to actual time values
	a := float32(bytesPerSample*src.i.SampleRate) / float32(src.i.Channels)
	time := float32(src.offset) / a
	duration := float32(br) / a

	src.offset += br

	return &SampleBlock{
		Data:     sd,
		Time:     time,
		Duration: duration,
	}, nil
}

func NewNullSampleReader(info *SampleReaderInfo) SampleReader {
	return &nullSampleReader{i: info}
}
