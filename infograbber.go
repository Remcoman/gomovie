package gomovie

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Info struct {
	Width     int
	Height    int
	Duration  float32
	FrameRate float32
	Rotation  int
}

func (i Info) String() string {
	return fmt.Sprintf("Width: %d, Height : %d, Duration : %f, Framerate : %f", i.Width, i.Height, i.Duration, i.FrameRate)
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
}

func (f FFProbeStream) getRotation() int {
	if len(f.Side_data_list) > 0 {
		//find display matrix element
		for _, v := range f.Side_data_list {
			side_data_el := v.(map[string]interface{})

			side_data_type := side_data_el["side_data_type"].(string)

			if side_data_type != "Display Matrix" {
				continue
			}

			return side_data_el["rotation"].(int)
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

	info = &Info{
		Width:     int(videoStream.Width),
		Height:    int(videoStream.Height),
		FrameRate: videoStream.getFrameRate(),
		Duration:  out.Format.getDuration(),
		Rotation:  videoStream.getRotation(),
	}

	return
}
