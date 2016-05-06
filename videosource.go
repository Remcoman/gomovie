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

	Audio  *AudioSource
	Info   *VideoInfo
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	index  int
}

func (g *VideoSource) Close() (err error) {
	err = g.cmd.Process.Kill()
	return
}

func (g *VideoSource) Open() (err error) {
	var videoInfo *VideoInfo
	var audioInfo *AudioInfo
	var stderr io.ReadCloser
	var stdout io.ReadCloser
	
	if videoInfo, _, err = ExtractInfo(g.Path); err != nil {
		return
	}
	
	g.Info = videoInfo
	
	if audioInfo != nil {
		g.Audio = &AudioSource{Path : g.Path, Info : audioInfo}
	}

	g.cmd = exec.Command(
		FfmpegPath,

		"-i", g.Path,
		"-loglevel", "error",
		"-ss", FormatTime(g.Start),
		"-f", "image2pipe",
		"-pix_fmt", "rgba",
		"-vcodec", "rawvideo",
		"-",
	)

	if stderr, err = g.cmd.StderrPipe(); err != nil {
		return
	} 
	
	g.stderr = stderr

	if stdout, err = g.cmd.StdoutPipe(); err != nil {
		return
	}
	
	g.stdout = stdout

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

func (g *VideoSource) VideoInfo() *VideoInfo {
	return g.Info
}

func (g *VideoSource) AudioInfo() *AudioInfo {
	if g.Audio != nil {
		return g.Audio.AudioInfo()
	}
	return nil
}

func (g *VideoSource) ReadFrame() (*Frame, error) {
	bytes := make([]byte, 4*g.Info.Width*g.Info.Height)
	time := float64(g.index) * float64(1./g.Info.FrameRate)
	
	if g.Duration != 0 && time > g.Duration {
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
		Width:  g.Info.Width,
		Height: g.Info.Height,
		Index:  g.index,
	}, nil
}