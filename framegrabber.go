package gomovie

import (
	"fmt"
	"os/exec"
	"io"
	"image"
)

type Frame struct {
	Bytes []byte
	Width int
	Height int
	Index int
}

func (f *Frame) ToNRGBAImage() *image.NRGBA {
	return &image.NRGBA{
		Pix : f.Bytes,
		Stride : f.Width * 4,
		Rect : image.Rect(0, 0, f.Width, f.Height),
	}
}

type Grabber struct {
	Path string
	Info *Info
	
	cmd *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	index int
}

func (g *Grabber) Open() error {
	if info, err := ExtractInfo(g.Path); err != nil {
		return err
	} else {
		g.Info = info
	}
	
	g.cmd = exec.Command(
		FfmpegPath,
		
		"-i", g.Path, 
		"-loglevel", "error", 
		"-f", "image2pipe",
		"-pix_fmt", "rgba",
		"-vcodec", "rawvideo", 
		"-",
	)
	
	if stderr, err := g.cmd.StderrPipe(); err != nil {
		return err
	} else {
		g.stderr = stderr
	}
	
	if stdout, err := g.cmd.StdoutPipe(); err != nil {
		return err
	} else {
		g.stdout = stdout
	}
	
	if err := g.cmd.Start(); err != nil {
		return err
	}
	
	go func() {
		err := g.cmd.Wait()
		fmt.Println(err)
	}()
	
	return nil
}

func (g *Grabber) GetFrame() (*Frame, error) {
	bytes := make([]byte, 4*g.Info.Width*g.Info.Height)
	
	if _, err := io.ReadFull(g.stdout, bytes); err != nil {
		return nil, err
	}
	
	g.index = g.index + 1
	
	return &Frame{
		Bytes : bytes,
		Width : g.Info.Width,
		Height : g.Info.Height,
		Index : g.index,
	}, nil
}

func CreateGrabber(path string) Grabber {
	return Grabber{Path : path, index : 0}
}