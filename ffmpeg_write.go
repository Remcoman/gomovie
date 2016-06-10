package gomovie

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type WriteConfig struct {
	VideoCodec        string
	AudioCodec        string
	ExtraArgs         []string
	ProgressCallback  func(progress float32)
	DebugFFmpegOutput bool

	audioResultChan chan bool
}

func createUniqueFifo() (string, error) {
	for id := 1; id < 1000; id++ {
		name := fmt.Sprintf("gomovie_audio_%v", id)
		if err := syscall.Mknod(name, syscall.S_IFIFO|0666, 0); err == nil {
			return name, nil
		}
	}
	return "", errors.New("Could not create a fifo name!")
}

func FfmpegWrite(path string, src interface{}, config WriteConfig) (err error) {
	var (
		frameReader     FrameReader
		sampleReader    SampleReader
		stdinSource     io.Reader
		totalFrames     float32
		progressBuffer  *bytes.Buffer
		audioResultChan chan bool
		fifoName        string
	)

	args := make([]string, 0, 25+len(config.ExtraArgs))

	//-y means force overwrite
	args = append(args, "-y")

	if !config.DebugFFmpegOutput && config.ProgressCallback != nil {
		args = append(args,
			"-progress", "pipe:2",
			"-v", "panic",
		)
	}

	switch t := src.(type) {
	case FrameReader:
		frameReader = t

	case SampleReader:
		sampleReader = t

	case *Video:
		frameReader = t.FrameReader
		sampleReader = t.SampleReader

	default:
		return errors.New("Can't write given object. It should implement FrameReader or SampleReader. Or it should be of type *Video")
	}

	//video has been specified
	if frameReader != nil {
		frameInfo := frameReader.Info()
		totalFrames = frameInfo.Duration * frameInfo.FrameRate

		args = append(args,
			"-s", fmt.Sprintf("%dx%d", frameInfo.Width, frameInfo.Height), //size
			"-r", strconv.FormatFloat(float64(frameInfo.FrameRate), 'g', 8, 32), //framerate
			"-pix_fmt", "rgba",
			"-f", "rawvideo",
			"-i", "pipe:0",
		)
	}

	//audio has been specified
	if sampleReader != nil {
		audioInfo := sampleReader.Info()
		sampleFormat := fmt.Sprintf("s%vle", sampleReader.SampleFormat().Depth)

		args = append(args,
			"-f", sampleFormat,
			"-ar", strconv.FormatInt(int64(audioInfo.SampleRate), 10),
			"-ac", strconv.FormatInt(int64(audioInfo.Channels), 2),
		)

		if frameReader != nil {
			audioResultChan = make(chan bool)

			if fifoName, err = createUniqueFifo(); err != nil {
				return
			}

			defer os.Remove(fifoName)

			go FfmpegWrite(fifoName, sampleReader, WriteConfig{ExtraArgs: []string{"-f", sampleFormat}, audioResultChan: audioResultChan})

			args = append(args, "-i", fifoName)

		} else {

			args = append(args, "-i", "pipe:0")

		}

	}

	if frameReader != nil {
		//output format
		args = append(args, "-pix_fmt", "yuv420p")
		stdinSource = frameReader

	} else {
		//output format
		stdinSource = sampleReader
	}

	if config.VideoCodec != "" {
		args = append(args, "-vcodec", config.VideoCodec)
	}

	if config.AudioCodec != "" {
		args = append(args, "-acodec", config.AudioCodec)
	}

	//extra output formats
	args = append(args, config.ExtraArgs...)

	args = append(args, path)

	cmd := exec.Command(GlobalConfig.FfmpegPath, args...)
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
		<-config.audioResultChan
	}

	err = cmd.Wait()

	//signal the audio channel to close
	if audioResultChan != nil {
		audioResultChan <- true
	}

	return
}
