package gomovie

import (
	"fmt"
	"math"
	"strconv"
)

var FfmpegPath string = "/usr/bin/ffmpeg"
var FfprobePath string = "/usr/bin/ffprobe"

func FormatSize(width int, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

func FormatTime(time float64) string {
	
	hour := strconv.FormatFloat(math.Floor(time / 3600.), 'f', 0, 32)
	if len(hour) < 2 {
		hour = "0" + hour
	}

	min := strconv.FormatFloat(math.Mod(math.Floor(time/60.), 60.), 'f', 0, 32)
	if len(min) < 2 {
		min = "0" + min
	}

	seconds := strconv.FormatFloat(math.Mod(time, 60), 'f', 0, 32)
	if len(seconds) < 2 {
		seconds = "0" + seconds
	}

	return hour + ":" + min + ":" + seconds
}
