package gomovie

import (
	"image"
	"image/draw"
	"io"
)

type frameReaderList struct {
	readers []FrameReader

	r *Range
	i *FrameReaderInfo

	index int
	l     []byte
}

func (src *frameReaderList) Info() *FrameReaderInfo { return src.i }
func (src *frameReaderList) Range() *Range          { return src.r }

func (src *frameReaderList) fitInImg(f *Frame) []byte {
	srcImg := f.ToNRGBAImage()
	dstImg := image.NewNRGBA(image.Rect(0, 0, src.i.Width, src.i.Height))
	draw.Draw(dstImg, srcImg.Bounds(), srcImg, image.Point{X: 0, Y: 0}, draw.Over)
	return dstImg.Pix
}

func (src *frameReaderList) Slice(r *Range) FrameReader {
	readers := src.readers

	var (
		e int
		t float32
	)

	for len(readers) > 0 {
		if readers[e].Range() != nil {
			t += readers[e].Range().Duration
		} else {
			t += readers[e].Info().Duration
		}

		if r.Start > t { //range does NOT start in current item
			readers = readers[1:]
		} else {
			if t > r.Duration {
				break
			}
			e++
		}
	}

	readers = readers[:e]

	return &frameReaderList{
		readers: readers,

		//range
		r: &Range{parent: src.r, Start: r.Start, Duration: 0.},
	}
}

//Close closes all the readers
func (src *frameReaderList) Close() (err error) {
	for _, v := range src.readers {
		if err = v.Close(); err != nil {
			return
		}
	}
	return
}

func (src *frameReaderList) Read(p []byte) (n int, err error) {
	if len(src.l) > 0 {
		n = copy(p, src.l)
		src.l = src.l[n:]
	} else {
		if src.l != nil {
			src.l = nil
		}

		var f *Frame
		f, err = src.ReadFrame()
		if err != nil {
			return
		}

		n = copy(p, f.Data)
		if n < len(f.Data) {
			src.l = f.Data[n:]
		}
	}

	return
}

func (src *frameReaderList) ReadFrame() (f *Frame, err error) {
	for len(src.readers) > 0 {
		var rf *Frame

		rf, err = src.readers[0].ReadFrame()

		if err == io.EOF { //try the next reader
			src.readers = src.readers[1:]
			continue
		}

		if err != nil {
			return
		}

		f = &Frame{
			Width:  src.i.Width,
			Height: src.i.Height,
			Index:  src.index,
			Time:   float32(src.index) * (1. / src.i.FrameRate),
		}

		//if frame size is not the same as info size
		//align the frame within containing frame
		if rf.Width != src.i.Width || rf.Height != src.i.Height {
			f.Data = src.fitInImg(rf)
		} else {
			f.Data = rf.Data
		}

		src.index++

		return
	}
	return nil, io.EOF
}

func concatFrameReaders(readers ...FrameReader) FrameReader {
	sumInfo := new(FrameReaderInfo)

	for _, reader := range readers {
		info := reader.Info()
		sumInfo.Duration += info.Duration
		sumInfo.Width = intMax(info.Width, sumInfo.Width)
		sumInfo.Height = intMax(info.Height, sumInfo.Height)
		sumInfo.FrameRate = float32Max(info.FrameRate, sumInfo.FrameRate)
	}

	return &frameReaderList{readers: readers, i: sumInfo}
}

type sampleReaderList struct {
	readers []SampleReader

	r *Range
	o *SampleFormat
	i *SampleReaderInfo

	l []byte
}

func (src *sampleReaderList) Info() *SampleReaderInfo     { return src.i }
func (src *sampleReaderList) SampleFormat() *SampleFormat { return src.o }
func (src *sampleReaderList) Range() *Range               { return src.r }

func (src *sampleReaderList) Slice(r *Range) SampleReader {
	readers := src.readers

	var (
		e int
		t float32
	)

	for len(readers) > 0 {
		if readers[e].Range() != nil {
			t += readers[e].Range().Duration
		} else {
			t += readers[e].Info().Duration
		}

		if r.Start > t { //range does NOT start in current item
			readers = readers[1:]
		} else {
			if t > r.Duration {
				break
			}
			e++
		}
	}

	readers = readers[:e]

	return &sampleReaderList{
		readers: readers,

		//range
		r: &Range{parent: src.r, Start: r.Start, Duration: 0.},

		i: src.i,
	}
}

//Close closes all the readers
func (src *sampleReaderList) Close() (err error) {
	for _, v := range src.readers {
		if err = v.Close(); err != nil {
			return
		}
	}
	return
}

func (src *sampleReaderList) Read(p []byte) (n int, err error) {
	if src.l == nil { //no leftover data to read from so get new sample block
		var bl *SampleBlock

		bl, err = src.ReadSampleBlock()
		if err != nil {
			return
		}
		src.l = bl.Bytes()
	}

	n = copy(p, src.l) //copy some data into p
	src.l = src.l[n:]  //save remaining data in slice of l

	if len(src.l) == 0 { //no data left in l? Set it to null
		src.l = nil
	}

	return
}

func (src *sampleReaderList) ReadSampleBlock() (b *SampleBlock, err error) {
	for len(src.readers) > 0 {
		r := src.readers[0]

		so := r.SampleFormat()
		*so = *src.o //pass the sample format to the sub reader

		b, err = r.ReadSampleBlock()
		if err == nil || err != io.EOF {
			return
		}
		src.readers = src.readers[1:]
	}
	return nil, io.EOF
}

func concatSampleReaders(readers ...SampleReader) SampleReader {
	sumInfo := new(SampleReaderInfo)

	for _, reader := range readers {
		info := reader.Info()
		sumInfo.Duration += info.Duration
		sumInfo.SampleRate = intMax(info.SampleRate, sumInfo.SampleRate)
		sumInfo.Channels = intMax(info.Channels, sumInfo.Channels)
	}

	return &sampleReaderList{readers: readers, o: NewSampleFormat(), i: sumInfo}
}

func Concat(readers ...interface{}) *Video {

	frameReaders := make([]FrameReader, len(readers))
	sampleReaders := make([]SampleReader, len(readers))

	var (
		sumFrameInfo          FrameReaderInfo
		sumSampleInfo         SampleReaderInfo
		hasSamples, hasFrames bool
	)

	addSampleReader := func(index int, reader SampleReader) {
		sampleReaders[index] = reader
		hasSamples = true

		i := reader.Info()
		sumSampleInfo.SampleRate = intMax(i.SampleRate, sumSampleInfo.SampleRate)
		sumSampleInfo.Channels = intMax(i.Channels, sumSampleInfo.Channels)
	}

	addFrameReader := func(index int, reader FrameReader) {
		frameReaders[index] = reader
		hasFrames = true

		i := reader.Info()
		sumFrameInfo.Width = intMax(i.Width, sumFrameInfo.Width)
		sumFrameInfo.Height = intMax(i.Height, sumFrameInfo.Height)
	}

	for index, reader := range readers {
		switch r := reader.(type) {
		case FrameReader:
			addFrameReader(index, r)
		case SampleReader:
			addSampleReader(index, r)
		case *Video:
			if r.SampleReader != nil {
				addSampleReader(index, r.SampleReader)
			}
			if r.FrameReader != nil {
				addFrameReader(index, r.FrameReader)
			}
		}
	}

	for x := 0; x < len(readers); x++ {
		if hasFrames && frameReaders[x] == nil {
			i := sumFrameInfo //create a copy
			i.Duration = sampleReaders[x].Info().Duration
			frameReaders[x] = NewNullFrameReader(&i)
		} else if hasSamples && sampleReaders[x] == nil {
			i := sumSampleInfo //create a copy
			i.Duration = frameReaders[x].Info().Duration
			sampleReaders[x] = NewNullSampleReader(&i)
		}
	}

	vid := new(Video)

	if hasFrames {
		vid.FrameReader = concatFrameReaders(frameReaders...)
	}

	if hasSamples {
		vid.SampleReader = concatSampleReaders(sampleReaders...)
	}

	return vid
}

func intMax(i1, i2 int) int {
	if i2 > i1 {
		return i2
	}
	return i1
}

func float32Max(f1, f2 float32) float32 {
	if f2 > f1 {
		return f2
	}
	return f1
}
