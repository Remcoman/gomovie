package gomovie

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

const (
	PCMSampleSize = 16
	BufferSize    = 512
)

type AudioSource struct {
	Path     string
	Start    float64
	Duration float64

	Info   *AudioInfo
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (src *AudioSource) AudioInfo() *AudioInfo {
	return src.Info
}

func (src *AudioSource) Close() (err error) {
	err = src.cmd.Process.Kill()
	return
}

func (src *AudioSource) Open() (err error) {
	var audioInfo *AudioInfo
	var stderr io.ReadCloser
	var stdout io.ReadCloser

	if _, audioInfo, err = ExtractInfo(src.Path); err != nil {
		return
	}

	if audioInfo == nil {
		return errors.New("No audio found in file")
	}

	src.Info = audioInfo

	src.cmd = exec.Command(
		FfmpegPath,

		"-i", src.Path,
		"-loglevel", "error",
		"-ss", FormatTime(src.Start),
		"-acodec", "pcm_s"+string(PCMSampleSize)+"le",
		"-ar", "44100",
		"-ac", "2",
		"-",
	)

	if stderr, err = src.cmd.StderrPipe(); err != nil {
		return
	}

	src.stderr = stderr

	if stdout, err = src.cmd.StdoutPipe(); err != nil {
		return
	}

	src.stdout = stdout

	if err := src.cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := src.cmd.Wait(); err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func (src *AudioSource) ReadSample() (sample []SampleInt16, err error) {
	o := make([]byte, BufferSize)
	if _, err = io.ReadFull(src.stdout, o); err != nil {
		return
	}
	byteReader := bytes.NewReader(o)

	sample = make([]SampleInt16, PCMSampleSize)
	binary.Read(byteReader, binary.LittleEndian, sample)

	return
}

func (src *AudioSource) Read(p []byte) (int, error) {
	return src.stdout.Read(p)
}
