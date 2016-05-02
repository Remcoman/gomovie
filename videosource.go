package gomovie

import (
	"fmt"
	"os/exec"
	"io"
)

type VideoSource struct {
	Path string
	
	info *Info
	cmd *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	index int
}

func (g *VideoSource) Close() {
	if err := g.cmd.Process.Kill(); err != nil {
		panic(err)
	}
}

func (g *VideoSource) Open() error {
	if info, err := ExtractInfo(g.Path); err != nil {
		return err
	} else {
		g.info = info
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

func (g *VideoSource) Info() *Info {
	return g.info
}

func (g *VideoSource) ReadFrame() (*Frame, error) {
	bytes := make([]byte, 4*g.info.Width*g.info.Height)
	
	if _, err := io.ReadFull(g.stdout, bytes); err != nil {
		return nil, err
	}
	
	g.index = g.index + 1
	
	return &Frame{
		Bytes : bytes,
		Width : g.info.Width,
		Height : g.info.Height,
		Index : g.index,
	}, nil
}

func OpenVideo(path string) FrameReader {
	source := &VideoSource{Path : path}
	source.Open()
	return source
}