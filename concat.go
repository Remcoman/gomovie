package gomovie

import (
	"io"
	"math"
)

type FrameReaderList struct {
	readers   []FrameReader
	frameInfo *FrameSrcInfo
}

func (src *FrameReaderList) Info() *FrameSrcInfo {
	return src.frameInfo
}

func (src *FrameReaderList) Close() (err error) {
	for _, v := range src.readers {
		if err = v.Close(); err != nil {
			return
		}
	}
	return
}

func (src *FrameReaderList) Read(p []byte) (n int, err error) {
	for len(src.readers) > 0 {
		n, err = src.readers[0].Read(p)
		if n > 0 || err != io.EOF {
			if err == io.EOF {
				// Don't return EOF yet. There may be more bytes
				// in the remaining readers.
				err = nil
			}
			return
		}
		src.readers = src.readers[1:]
	}
	return 0, io.EOF
}

func (src *FrameReaderList) ReadFrame() (fr *Frame, err error) {
	for len(src.readers) > 0 {
		fr, err = src.readers[0].ReadFrame()
		if err == nil || err != io.EOF {
			return
		}
		src.readers = src.readers[1:]
	}
	return
}

func ConcatVideo(readers ...FrameReader) *FrameReaderList {
	sumInfo := new(FrameSrcInfo)

	for _, reader := range readers {
		info := reader.Info()
		sumInfo.Duration += info.Duration
		sumInfo.Width = int(math.Max(float64(info.Width), float64(sumInfo.Width)))
		sumInfo.Height = int(math.Max(float64(info.Height), float64(sumInfo.Height)))
		sumInfo.FrameRate = info.FrameRate //throw error if different framerates?
		sumInfo.Rotation = 0
	}

	return &FrameReaderList{readers, sumInfo}
}
