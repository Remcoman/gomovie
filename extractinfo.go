package gomovie

import (
	"encoding/json"
	"errors"
	"log"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"math"
)

type Info struct {
	Width     int
	Height    int
	Duration  float32
	FrameRate float32
	Rotation  int
}

func (i Info) String() string {
	return fmt.Sprintf("Width: %d, Height : %d, Duration : %f, Framerate : %f, Rotation: %d", 
		i.Width, 
		i.Height, 
		i.Duration, 
		i.FrameRate,
		i.Rotation,
	)
}

type FFProbeFormat struct {
	Filename string
	Duration string
}

func (f FFProbeFormat) getDuration() float32 {
	dur, _ := strconv.ParseFloat(f.Duration, 32)
	return float32(dur)
}

type FFProbeStream struct {
	Width          int
	Height         int
	Avg_frame_rate string
	Codec_type     string
	Side_data_list []interface{}
	Tags		   map[string]interface{}
}

func (f FFProbeStream) getRotation() int {
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

func (f FFProbeStream) getFrameRate() float32 {
	parts := strings.Split(f.Avg_frame_rate, "/")
	first, _ := strconv.ParseFloat(parts[0], 32)
	last, _ := strconv.ParseFloat(parts[1], 32)
	return float32(first / last)
}

type FFProbeOutput struct {
	Format  FFProbeFormat
	Streams []FFProbeStream
}

func (o *FFProbeOutput) getVideoStream() *FFProbeStream {
	for _, stream := range o.Streams {
		if stream.Codec_type == "video" {
			return &stream
		}
	}
	return nil
}

func ExtractInfo(path string) (info *Info, err error) {
	cmd := exec.Command(
		FfprobePath,

		"-i", path,

		"-print_format", "json",

		"-show_streams",
		"-show_format",

		"-select_streams", "v",

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

	videoStream := out.getVideoStream()
	if videoStream == nil {
		err = errors.New("No video stream found!")
		return
	}
	
	rotation := videoStream.getRotation()
	width := int(videoStream.Width)
	height := int(videoStream.Height)

	if math.Abs(math.Mod(float64(rotation), 180.)) == 90 {
		oldWidth := width
		width = height
		height = oldWidth 	
	} 

	info = &Info{
		Width:     width,
		Height:    height,
		FrameRate: videoStream.getFrameRate(),
		Duration:  out.Format.getDuration(),
		Rotation:  rotation,
	}

	return
}
