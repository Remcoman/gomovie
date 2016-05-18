package gomovie

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"bytes"
	"encoding/binary"
)

const (
	BufferSize    = 512
)

func OpenVideo(path string) *Video {
	videoInfo, audioInfo, _ := ExtractInfo(path)
	
	v := &FfmpegRGBAStream{Path : path, I : videoInfo}
	v.Open()
	
	a := &FfmpegPCMStream{Path : path, ReadDepth : 16, I : audioInfo}
	a.Open()
	
	return &Video{v, a}
}

type FfmpegPCMStream struct {
	Path     string
	Start    float64
	Duration float64
	
	ReadDepth int
	ReadSampleRate int
	ReadChannels int
	
	I   *AudioInfo
	
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (src *FfmpegPCMStream) Info() *AudioInfo {
	return src.I
}

func (src *FfmpegPCMStream) Close() (err error) {
	err = src.cmd.Process.Kill()
	return
}

func (src *FfmpegPCMStream) Open() (err error) {
	var stderr io.ReadCloser
	var stdout io.ReadCloser
	
	args := []string {
		"-loglevel", "error",
		
		"-i", src.Path,
		
		"-vn",
	}
	
	if src.Start > 0 {
		args = append(args, 
			"-ss", 
			strconv.FormatFloat(src.Start, 'f', -1, 32),
		)
	}
	
	if src.Duration > 0 {
		args = append(args,
			"-t",
			strconv.FormatFloat(src.Duration, 'f', -1, 32),
		)
	}
	
	args = append(args,
		"-f", fmt.Sprintf("s%vle", src.ReadDepth),
	)
	
	if src.ReadChannels != 0 {
		args = append(args, "-ac", strconv.FormatInt(src.ReadChannels, 10))
	}
	
	if src.ReadSampleRate != 0 {
		args = append(args, "-ar", strconv.FormatInt(src.ReadSampleRate, 10))
	}
	
	args = append(args, "-")
	
	src.cmd = exec.Command(
		FfmpegPath,
		args...,	
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

func (src *FfmpegPCMStream) ReadSampleBlock() (sample []Sample, err error) {
	bytesPerSample := src.ReadDepth / 8

	switch src.ReadDepth {
		case 16 :
			sample = make([]SampleInt16, BufferSize / bytesPerSample)
		case 32 :
			sample = make([]SampleInt32, BufferSize / bytesPerSample)
	}
 
	err = binary.Read(src.stdout, binary.LittleEndian, sample)

	return
}

func (src *FfmpegPCMStream) SampleDepth() int {
	return src.ReadDepth
}

func (src *FfmpegPCMStream) Read(p []byte) (int, error) {
	return src.stdout.Read(p)
}


type FfmpegRGBAStream struct {
	Path   string
	Start float64
	Duration float64
	I   *VideoInfo
	
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	index  int
}

func (g *FfmpegRGBAStream) Close() (err error) {
	err = g.cmd.Process.Kill()
	return
}

func (g *FfmpegRGBAStream) Open() (err error) {
	var stderr io.ReadCloser
	var stdout io.ReadCloser
	
	args := []string{
		"-loglevel", "error",
		
		"-i", g.Path,
		
		"-f", "image2pipe",
		"-pix_fmt", "rgba",
		"-vcodec", "rawvideo",
	}
	
	if g.Start > 0 {
		args = append(args, 
			"-ss", 
			strconv.FormatFloat(g.Start, 'f', -1, 32),
		)
	}
	
	if g.Duration > 0 {
		args = append(args,
			"-t",
			strconv.FormatFloat(g.Duration, 'f', -1, 32),
		)
	}
	
	args = append(args, "-")

	g.cmd = exec.Command(FfmpegPath, args...)

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

func (g *FfmpegRGBAStream) Read(p []byte) (int, error) {
	return g.stdout.Read(p)
}

func (g *FfmpegRGBAStream) Info() *VideoInfo {
	return g.I
}

func (g *FfmpegRGBAStream) ReadFrame() (*Frame, error) {
	bytes := make([]byte, 4*g.I.Width*g.I.Height)
	time := float64(g.index) * float64(1./g.I.FrameRate)
	
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
		Width:  g.I.Width,
		Height: g.I.Height,
		Index:  g.index,
	}, nil
}