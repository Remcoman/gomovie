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
	"io"
)

type Config struct {
	VideoCodec         string
	AudioCodec		   string
	ExtraArgs          []string
	ProgressCallback   func(progress float32)
	DebugFFmpegOutput  bool
	
	audioResultChan    chan bool
}

func getFifoName() (string, error) {
	for id := 1; id < 1000; id++ {
		name := fmt.Sprintf("gomovie_audio_%v", id)
		if err := syscall.Mknod(name, syscall.S_IFIFO|0666, 0); err == nil {
			return name, nil
		}
	}
	return "", errors.New("Could not create a fifo name!")
}

func Encode(path string, src interface{}, config Config) (err error) {
	var (
		videoSrc VideoReader
		audioSrc AudioReader
		stdinSource io.Reader
		totalFrames float32
		progressBuffer *bytes.Buffer
		audioResultChan chan bool
		fifoName string
	)

	args := make([]string, 0, 25 + len(config.ExtraArgs))
	
	args = append(args, "-y")
	
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
			
		case *Video :
			videoSrc = t.VideoReader
			audioSrc = t.AudioReader
			
		default :
			return errors.New("Invalid type!")
	}

	//video has been specified
	if videoSrc != nil {
		videoInfo := videoSrc.Info()
		totalFrames = videoInfo.Duration * videoInfo.FrameRate
		
		args = append(args,
			"-s", FormatSize(videoInfo.Width, videoInfo.Height), //size
			"-r", strconv.FormatFloat(float64(videoInfo.FrameRate), 'g', 8, 32), //framerate
			"-pix_fmt", "rgba",
			"-f", "rawvideo",
			"-i", "pipe:0",
		)
	}
	
	if audioSrc != nil {
		
		if videoSrc != nil {
			audioResultChan = make(chan bool)
			
			if fifoName, err = getFifoName(); err != nil {
				return
			}
			
			//TODO find a unique fifo name
			if err = syscall.Mknod(fifoName, syscall.S_IFIFO|0666, 0); err != nil {
				return
			}
			
			defer os.Remove(fifoName)

			go Encode(fifoName, audioSrc, Config{AudioCodec : "s16le", audioResultChan : audioResultChan})

			//audio input
			args = append(args,
				"-f", "s16le",
				"-ar", "44100",
				"-ac", "2",
				"-i", fifoName,
			)
			
		} else {
			args = append(args,
				"-f", "s16le", 
				"-ar", "44100",
				"-ac", "2",
				"-i", "pipe:0",
			)
		}
		
	}

	if videoSrc != nil {
		args = append(args, 
			"-pix_fmt", "yuv420p",
			"-f", config.VideoCodec,
		)
		
		stdinSource = videoSrc
		
	} else {
		args = append(args, 
			"-f", config.AudioCodec,
		)
		
		stdinSource = audioSrc
	}
	
	args = append(args, config.ExtraArgs...)
	args = append(args, path)
	
	cmd := exec.Command(FfmpegPath, args...)
	cmd.Stdin = stdinSource

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
	
	//wait for signal from parent thread
	if config.audioResultChan != nil {
		<- config.audioResultChan
	}

	err = cmd.Wait()
	
	if audioResultChan != nil {
		audioResultChan <- true
	}
	
	return
}
