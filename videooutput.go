package gomovie

import (
	"image"

	"errors"
	"io"
	"os/exec"
	"strconv"
)

var errInvalidReader = errors.New("Invalid reader")

type VideoOutput struct {
	Path       string
	FfmpegArgs []string
	Info Info
	
	opened bool
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func (w *VideoOutput) Close() (err error) {
	err = w.stdin.Close()
	return
}

func (w *VideoOutput) ReadFrom(r io.Reader) (read int64, err error) {
	frameReader, ok := r.(FrameReader)
	if !ok {
		err = errInvalidReader
		return
	}
	
	if !w.opened {
		w.Info = *frameReader.Info()
		if err = w.Open(); err != nil {
			return
		}
	}
	
	var frame *Frame
	var n int
	
	for {
		
		if frame, err = frameReader.ReadFrame(); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		
		if n, err = w.Write(frame.Bytes); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		
		read += int64(n)
	}
}

func (w *VideoOutput) Open() (err error) {
	if w.opened {
		return errors.New("Already opened")
	}
	
	args := []string{
		"-y",
		"-s", formatSize(w.Info.Width, w.Info.Height),
		"-pix_fmt", "rgba",
		"-f", "rawvideo",
		"-r", strconv.FormatFloat(float64(w.Info.FrameRate), 'g', 8, 32),
		"-i", "-",
		"-an",
		w.Path,
	}
	
	args = append(args, w.FfmpegArgs...)
	
	w.cmd = exec.Command(FfmpegPath, args...) 
	
	var stdin io.WriteCloser
	
	if stdin, err = w.cmd.StdinPipe(); err != nil {
		return
	}
	
	w.stdin = stdin
	w.opened = true

	if err = w.cmd.Start(); err != nil {
		return
	}
	
	go func() {
		w.cmd.Wait()
		w.opened = false
	}()

	return nil
}

func (w *VideoOutput) Write(p []byte) (int, error) {
	if !w.opened {
		return 0, errors.New("VideoOutput not opened")
	}
	
	return w.stdin.Write(p)
}

func (w *VideoOutput) WriteImage(img image.Image) error {	
	var bytes *[]byte

	switch t := img.(type) {
		case *image.NRGBA:
			bytes = &t.Pix
		case *image.RGBA:
			bytes = &t.Pix
		default:
			return errors.New("Unsupported image type!")
	}

	_, err := w.Write(*bytes)
	return err
}

func (w *VideoOutput) WriteFrame(fr *Frame) error {
	_, err := w.Write(fr.Bytes)
	return err
}
