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
	
	//-y means force overwrite
	args = append(args, "-y")
	
	if !config.DebugFFmpegOutput && config.ProgressCallback != nil {
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
	
	//audio has been specified
	if audioSrc != nil {
		audioInfo := audioSrc.Info()
		
		args = append(args,
			"-f",  sampleFormat, 
			"-ar", strconv.FormatInt(int64(audioInfo.SampleRate), 10),
			"-ac", strconv.FormatInt(int64(audioInfo.Channels), 2),
		)
		
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
			
			sampleFormat := fmt.Sprintf("s%vle", audioSrc.SampleDepth())

			go Encode(fifoName, audioSrc, Config{AudioCodec : sampleFormat, audioResultChan : audioResultChan})

			args = append(args, "-i",  fifoName)
			
		} else {
			
			args = append(args, "-i",  "pipe:0")
		
		}
		
	}

	if videoSrc != nil {
		//output format
		args = append(args, 
			"-pix_fmt", "yuv420p",
			"-f", config.VideoCodec,
		)
		
		stdinSource = videoSrc
		
	} else {
		//output format
		args = append(args, 
			"-f", config.AudioCodec,
		)
		
		stdinSource = audioSrc
	}
	
	//extra output formats
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
