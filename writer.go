package gomovie

import (
	"strconv"
	"os/exec"
	"bytes"
	"strings"
)

type Config struct {
	Codec string
	ExtraArgs []string
	ProgressCallback func(progress float32)
}

func Encode(path string, src interface{}, config Config) (err error) {
	videoSrc := src.(FrameReader)
	
	videoInfo := videoSrc.VideoInfo()
	totalFrames := videoInfo.Duration * videoInfo.FrameRate

	args := []string{		
		"-y",
		
		"-s", FormatSize(videoInfo.Width, videoInfo.Height),
		"-r", strconv.FormatFloat(float64(videoInfo.FrameRate), 'g', 8, 32),
		
		"-pix_fmt", "rgba",
		"-f", "rawvideo",
		"-i", "pipe:0",
		"-progress", "pipe:2",
		"-v", "panic",
		
		"-pix_fmt", "yuv420p",
		"-vcodec", config.Codec,
		
		path,
	}
	
	args = append(args, config.ExtraArgs...)
	
	cmd := exec.Command(FfmpegPath, args...)
	cmd.Stdin = videoSrc
	
	buf := new(bytes.Buffer)
	
	cmd.Stderr = buf
	
	if err = cmd.Start(); err != nil {
		return
	}
	
	if config.ProgressCallback != nil {
		quit := make(chan bool)
	
		defer func() {
			quit <- true
		}()
		
		go func(buf *bytes.Buffer) {
			
			for {
				select {
					case <- quit :
						return
						
					default :
						line, err := buf.ReadString('\n')
						if err == nil {
							parts := strings.Split(line, "=")
							if len(parts) > 1 && parts[0] == "frame" && len(parts[1]) > 0 {
								frameNum, _ := strconv.ParseInt(parts[1][:len(parts[1])-1], 10, 32)
								config.ProgressCallback(float32(frameNum) / totalFrames)
							}
						}
				}
			}

		}(buf)
	}

	return cmd.Wait()
}