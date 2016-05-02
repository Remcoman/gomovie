package gomovie

import (
	"image"
	"fmt"
	
	"errors"
	"os/exec"
	"io"
)

type Encoder struct {
	Path string
	FfmpegArgs []string
	
	cmd *exec.Cmd
	stdin io.WriteCloser
}

func (w *Encoder) Close() {
	if err := w.cmd.Process.Kill(); err != nil {
		panic(err)
	}
}

func (w *Encoder) Open() error {
	w.cmd = exec.Command(FfmpegPath, "yo")
	
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

func (w *Encoder) WriteNRGBABytes(bytes []byte) {
	//TODO add validation of byte data
	w.stdin.Write(bytes)
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
	w.stdin.Write(fr.Bytes)
}
