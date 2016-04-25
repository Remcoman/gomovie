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

func (this *Frame) ToNRGBAImage() *image.NRGBA {
	return &image.NRGBA{
		Pix : this.Bytes,
		Stride : this.Width * 4,
		Rect : image.Rect(0, 0, this.Width, this.Height),
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

func (this *Grabber) Open() error {
	if info, err := ExtractInfo(this.Path); err != nil {
		return err
	} else {
		this.Info = info
	}
	
	this.cmd = exec.Command(
		"/usr/bin/ffmpeg",
		
		"-i", this.Path, 
		"-loglevel", "error", 
		"-f", "image2pipe",
		"-pix_fmt", "rgba",
		"-vcodec", "rawvideo", 
		"-",
	)
	
	if stderr, err := this.cmd.StderrPipe(); err != nil {
		return err
	} else {
		this.stderr = stderr
	}
	
	if stdout, err := this.cmd.StdoutPipe(); err != nil {
		return err
	} else {
		this.stdout = stdout
	}
	
	if err := this.cmd.Start(); err != nil {
		return err
	}
	
	go func() {
		err := this.cmd.Wait()
		fmt.Println(err)
	}()
	
	return nil
}

func (this *Grabber) GetFrame() Frame {
	bytes := make([]byte, 4*this.Info.Width*this.Info.Height)
	
	io.ReadFull(this.stdout, bytes)
	
	this.index = this.index + 1
	
	return Frame {
		Bytes : bytes,
		Width : this.Info.Width,
		Height : this.Info.Height,
		Index : this.index,
	}
}

func CreateGrabber(path string) Grabber {
	return Grabber{Path : path, index : 0}
}