package gomovie

import (
	"io"
	"math"
)

type ConcatenatedVideo struct {
	readers []FrameReader

	index int
	videoInfo *VideoInfo
	//audioInfo *AudioInfo
}

func (f *ConcatenatedVideo) VideoInfo() *VideoInfo {
	return f.videoInfo
}

// func (f *ConcatenatedVideo) AudioInfo() *AudioInfo {
// 	return f.audioInfo
// }

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
	
	return &ConcatenatedVideo{readers, 0, sumInfo}
}