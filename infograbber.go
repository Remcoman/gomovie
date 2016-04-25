
package gomovie

import (
	"fmt"
	"os/exec"
	"encoding/json"
	"strings"
	"strconv"
)

type Info struct {
	Width int
	Height int
	Filename string
	Duration float32
	FrameRate float32
}

func (this Info) String() string {
	return fmt.Sprintf("Info about %s -> Width: %d, Height : %d, Duration : %f, Framerate : %f", this.Filename, this.Width, this.Height, this.Duration, this.FrameRate)
}

type FFProbeFormat struct {
	Filename string
	Duration string
}

func (this FFProbeFormat) getDuration() float32 {
	dur, _ := strconv.ParseFloat(this.Duration, 32)
	return float32(dur)
}

type FFProbeStream struct {
	Width int
	Height int
	Avg_frame_rate string
}

func (this FFProbeStream) getFrameRate() float32 {
	parts := strings.Split(this.Avg_frame_rate, "/")
	first, _ := strconv.ParseFloat(parts[0], 32)
	last, _ := strconv.ParseFloat(parts[1], 32)
	return float32(first / last)
}

type FFProbeOutput struct {
	Format FFProbeFormat
	Streams []FFProbeStream
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
	
	info = &Info{
		Width : int(out.Streams[0].Width),
		Height : int(out.Streams[0].Height),
		FrameRate : out.Streams[0].getFrameRate(),
		Duration : out.Format.getDuration(),
		Filename : out.Format.Filename,
	}
	
	return
}