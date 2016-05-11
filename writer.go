package gomovie

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

type Config struct {
	Codec            string
	ExtraArgs        []string
	ProgressCallback func(progress float32)
}

func Encode(path string, src interface{}, config Config) (err error) {
	var videoSrc FrameReader
	//var audioSrc SampleReader
	var totalFrames float32
	var ok bool

	args := []string{
		"-y",
		"-progress", "pipe:2",
		"-v", "panic",
	}

	if videoSrc, ok = src.(FrameReader); ok {
		videoInfo := videoSrc.VideoInfo()
		totalFrames = videoInfo.Duration * videoInfo.FrameRate

		args = append(args, 
			"-s", FormatSize(videoInfo.Width, videoInfo.Height),
			"-r", strconv.FormatFloat(float64(videoInfo.FrameRate), 'g', 8, 32),

			"-pix_fmt", "rgba",
			"-f", "rawvideo",
			"-i", "pipe:0",

			"-pix_fmt", "yuv420p",
			"-vcodec", config.Codec,
		)
	}
	
	//todo also write from audio if presetn
	if _, ok = src.(SampleReader); ok {
		
		//both audio and video so lets write audio to temp file
		if videoSrc != nil {
			//mkfifo /tmp/audio
			//output to pipe /tmp/audio
			//set input to /tmp/audio

			//go Encode("temp.wav", audioSrc, Config{})

			//args = append(args,
			//	"-i", "temp.wav",
			//)
		} else {
			//audio only
			// fmt.Println("yo")
			// args = append(args, 
			// 	"-i", "pipe:0",
			// 	"-f", "pcm_s16le",
			// 	"-acodec", config.Codec,
			// )
		}
	}

	args = append(args, config.ExtraArgs...)
	args = append(args, path)

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
				case <-quit:
					return

				default:
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
