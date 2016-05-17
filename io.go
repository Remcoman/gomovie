package gomovie

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"fmt"
	"errors"
	"os"
)

type Config struct {
	Codec            string
	ExtraArgs        []string
	ProgressCallback func(progress float32)
	DebugFFmpegOutput bool
}

func OpenVideo(path string) VideoAudio {
	videoInfo, audioInfo, _ := ExtractInfo(path)
	
	v := &FfmpegRGBAStream{Path : path, I : videoInfo}
	v.Open()
	
	a := &FfmpegPCMStream{Path : path, I : audioInfo}
	a.Open()
	
	return VideoAudio{v, a}
}

func Encode(path string, src interface{}, config Config) (err error) {
	var videoSrc VideoReader
	var audioSrc AudioReader
	var totalFrames float32
	var progressBuffer *bytes.Buffer

	args := []string{"-y"}
	
	if !config.DebugFFmpegOutput {
		args = append(args,
			"-progress", "pipe:2",
			"-v", "panic",
		)
	}
	
	switch t := src.(type) {
		case VideoReader :
			videoSrc = t
			
		case AudioReader :
			audioSrc = t
			
		case VideoAudio :
			videoSrc = t.V
			audioSrc = t.A
			
		default :
			return errors.New("Invalid type!")
	}

	if videoSrc != nil {
		videoInfo := videoSrc.Info()
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
	if audioSrc != nil {
		
		//both audio and video so lets write audio to temp file
		if videoSrc != nil {
			//mkfifo /tmp/audio
			//output to pipe /tmp/audio
			//set input to /tmp/audio
			
			fmt.Println(audioSrc.Info())

			syscall.Mknod("audio", syscall.S_IFIFO|0666, 0)

			go Encode("audio", audioSrc, Config{})

			args = append(args,
				"-f", "u16le",
				"-i", "audio",
			)
		} else {
			args = append(args,
				"-f", "pcm_s16le", 
				"-i", "pipe:0",
				"-acodec", "wav",
			)
		}
	}

	args = append(args, config.ExtraArgs...)
	args = append(args, path)
	
	fmt.Println(args)

	cmd := exec.Command(FfmpegPath, args...)
	
	cmd.Stdin = videoSrc

	if !config.DebugFFmpegOutput && config.ProgressCallback != nil {
		progressBuffer = new(bytes.Buffer)
		cmd.Stderr = progressBuffer
		
		quit := make(chan bool)

		defer func() {
			quit <- true
		}()

		go func() {

			for {
				select {
				case <-quit:
					return

				default:
					line, err := progressBuffer.ReadString('\n')
					if err == nil {
						parts := strings.Split(line, "=")
						if len(parts) > 1 && parts[0] == "frame" && len(parts[1]) > 0 {
							frameNum, _ := strconv.ParseInt(parts[1][:len(parts[1])-1], 10, 32)
							config.ProgressCallback(float32(frameNum) / totalFrames)
						}
					}
				}
			}

		}()
	} else if config.DebugFFmpegOutput {
		cmd.Stderr = os.Stdout
	}

	if err = cmd.Start(); err != nil {
		return
	}

	return cmd.Wait()
}
