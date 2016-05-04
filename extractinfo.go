package gomovie

import (
	"encoding/json"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"math"
	"fmt"
)

type VideoInfo struct {
	CodecName string
	Width     int
	Height    int
	Rotation  int
	FrameRate float32
	Duration   float32
}

func (i VideoInfo) String() string {
	return fmt.Sprintf("CodecName: %s, Duration: %f, Width: %d, Height : %d, Framerate : %f, Rotation: %d",
		i.CodecName,
		i.Duration,
		i.Width, 
		i.Height,  
		i.FrameRate,
		i.Rotation,
	)
}

type AudioInfo struct {
	CodecName string
	Duration  float32
	SampleRate int
	Channels int
}

func (i AudioInfo) String() string {
	return fmt.Sprintf("CodecName: %s, Duration: %f, SampleRate: %d, Channels: %d",
		i.CodecName,
		i.Duration,
		i.SampleRate,
		i.Channels,
	)
}



type FFProbeFormat struct {
	Filename string
	Duration string
}

func (f FFProbeFormat) FloatDuration() float32 {
	dur, _ := strconv.ParseFloat(f.Duration, 32)
	return float32(dur)
}

type FFProbeStream struct {
	Codec_name	   string
	
	Sample_rate	   string
	Channels	   int
	
	Width          int
	Height         int
	Avg_frame_rate string
	Codec_type     string
	Side_data_list []interface{}
	Tags		   map[string]interface{}
}

func (f FFProbeStream) IntSampleRate() int {
	rate, _ := strconv.ParseInt(f.Sample_rate, 10, 32)
	return int(rate)
}

func (f FFProbeStream) Rotation() int {
	if len(f.Side_data_list) > 0 {
		//find display matrix element
		for _, v := range f.Side_data_list {
			sideDataEl := v.(map[string]interface{})

			sideDataType := sideDataEl["side_data_type"].(string)
			
			if sideDataType != "Display Matrix" {
				continue
			}
			
			return int(sideDataEl["rotation"].(float64))
		}
	} else {
		if f.Tags != nil {
			if rotate, err := strconv.ParseInt(f.Tags["rotate"].(string), 10, 32); err == nil && rotate != 0 {
				log.Print("Rotate tag found but display matrix seems missing. It seems you are using a old ffmpeg version?")
			}
		}
	}

	return 0
}

func (f FFProbeStream) FloatFrameRate() float32 {
	parts := strings.Split(f.Avg_frame_rate, "/")
	first, _ := strconv.ParseFloat(parts[0], 32)
	last, _ := strconv.ParseFloat(parts[1], 32)
	return float32(first / last)
}

type FFProbeOutput struct {
	Format  FFProbeFormat
	Streams []FFProbeStream
}

func (o *FFProbeOutput) StreamByType(typeId string) *FFProbeStream {
	for _, stream := range o.Streams {
		if stream.Codec_type == typeId {
			return &stream
		}
	}
	return nil
}

func ExtractInfo(path string) (videoInfo *VideoInfo, audioInfo *AudioInfo, err error) {
	cmd := exec.Command(
		FfprobePath,

		"-i", path,

		"-print_format", "json",

		"-show_streams",
		"-show_format",

		"-v", "quiet",
	)

	bytes, err := cmd.Output()
	
	if err != nil {
		return
	}

	var out FFProbeOutput
	if err = json.Unmarshal(bytes, &out); err != nil {
		return
	}
	
	if videoStream := out.StreamByType("video"); videoStream != nil {
		rotation := videoStream.Rotation()
		width := int(videoStream.Width)
		height := int(videoStream.Height)

		if math.Abs(math.Mod(float64(rotation), 180.)) == 90 {
			oldWidth := width
			width = height
			height = oldWidth 	
		}
		
		videoInfo = &VideoInfo{
			CodecName: videoStream.Codec_name,
			FrameRate: videoStream.FloatFrameRate(),
			Width:     width,
			Height:    height,
			Rotation:  rotation,
			Duration:  out.Format.FloatDuration(),
		}
	}
	
	if audioStream := out.StreamByType("audio"); audioStream != nil {
		audioInfo = &AudioInfo{
			CodecName: audioStream.Codec_name,
			SampleRate: audioStream.IntSampleRate(), 
			Duration:  out.Format.FloatDuration(),
			Channels: audioStream.Channels,
		}
	}

	return
}
