package gomovie

import (
	"fmt"
	"image"

	"errors"
	"io"
	"os/exec"
)

type Encoder struct {
	Path       string
	FfmpegArgs []string

	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func (w *Encoder) Close() {
	if err := w.cmd.Process.Kill(); err != nil {
		panic(err)
	}
}

func (w *Encoder) Open() error {
	w.cmd = exec.Command(FfmpegPath, 
		"-y",
		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		//"-s", formatSize(width, height)
		"-pix_fmt", "rgb24",
		//"-r", "24",
		"-i", "-",
		"-an",
		"vcodec", "mpeg",
		w.Path
	)

	w.cmd.Start()

	stdin, err := w.cmd.StdinPipe()
	if err != nil {
		return err
	}

	w.stdin = stdin

	go func() {
		err := w.cmd.Wait()
		fmt.Println(err)
	}()

	return nil
}

func (w *Encoder) Write(p []byte) (int, error) {
	//TODO add validation of byte data
	return w.stdin.Write(p)
}

func (w *Encoder) WriteImage(img image.Image) error {
	var bytes *[]byte

	switch t := img.(type) {
	case *image.NRGBA:
		bytes = &t.Pix
	case *image.RGBA:
		bytes = &t.Pix
	default:
		return errors.New("Unsupported image type!")
	}

	w.stdin.Write(*bytes)

	return nil
}

func (w *Encoder) WriteFrame(fr *Frame) {
	w.Write(fr.Bytes)
}
