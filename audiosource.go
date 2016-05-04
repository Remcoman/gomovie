package gomovie

import (
	"os/exec"
	"io"
	"errors"
	"fmt"
)

type AudioSource struct {
	Path   string
	Start float64
	Duration float64
	
	Info *AudioInfo
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (a *AudioSource) AudioInfo() *AudioInfo {
	return a.Info
}

func (g *AudioSource) Close() (err error) {
	err = g.cmd.Process.Kill()
	return
}

func (g *AudioSource) Open() (err error) {
	var audioInfo *AudioInfo
	var stderr io.ReadCloser
	var stdout io.ReadCloser
	
	if _, audioInfo, err = ExtractInfo(g.Path); err != nil {
		return
	}
	
	if audioInfo == nil {
		return errors.New("No audio found in file")
	}
	
	g.Info = audioInfo

	g.cmd = exec.Command(
		FfmpegPath,

		"-i", g.Path,
		"-loglevel", "error",
		"-ss", formatTime(g.Start),
		"-acodec", "pcm_s16le",
		"-ar", "44100",
		"-ac", "2",
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

func (g *AudioSource) Read(p []byte) (int, error) {
	return g.stdout.Read(p)
}