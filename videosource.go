package gomovie

import (
	"fmt"
	"io"
	"os/exec"
)

type VideoSource struct {
	Path   string
	Start float64
	Duration float64

	info   *Info
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	index  int
}

func (g *VideoSource) Close() (err error) {
	err = g.cmd.Process.Kill()
	return
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
		"-ss", formatTime(g.Start),
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
		if err := g.cmd.Wait(); err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func (g *VideoSource) Read(p []byte) (int, error) {
	return g.stdout.Read(p)
}

func (g *VideoSource) Info() *Info {
	return g.info
}

func (g *VideoSource) Time() float64 {
	return float64(g.index) * float64(1./g.info.FrameRate)
}

func (g *VideoSource) ReadFrame() (*Frame, error) {
	bytes := make([]byte, 4*g.info.Width*g.info.Height)
	
	if g.Duration != 0 && g.Time() > g.Duration {
		return nil, io.EOF
	}

	if n, err := io.ReadFull(g.stdout, bytes); err != nil {
		if n == 0 { //got invalid file descriptor (ffmpeg autocloses stdout?)
			err = io.EOF
		}
		return nil, err
	}

	g.index = g.index + 1

	return &Frame{
		Bytes:  bytes,
		Width:  g.info.Width,
		Height: g.info.Height,
		Index:  g.index,
	}, nil
}