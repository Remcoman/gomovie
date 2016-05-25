package gomovie

import (
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

const (
	BufferSize = 512

	PixelDepth = 4
)

func FfmpegOpen(path string) (vid *Video, err error) {
	frameInfo, audioInfo, err := ExtractInfo(path)
	if err != nil {
		return
	}

	var (
		v *FfmpegRGBAStream
		a *FfmpegPCMStream
	)

	if frameInfo != nil {
		v = &FfmpegRGBAStream{Path: path, I: frameInfo}
		if err = v.Open(); err != nil {
			return
		}
	}

	if audioInfo != nil {
		a = &FfmpegPCMStream{Path: path, ReadDepth: 16, I: audioInfo}
		if err = a.Open(); err != nil {
			return
		}
	}

	vid = &Video{v, a}

	return
}

type FfmpegPCMStream struct {
	Path     string
	Start    float64
	Duration float64

	ReadDepth      int
	ReadSampleRate int
	ReadChannels   int

	I *SampleSrcInfo

	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	offset int
}

func (src *FfmpegPCMStream) Info() *SampleSrcInfo {
	return src.I
}

func (src *FfmpegPCMStream) Close() (err error) {
	err = src.cmd.Process.Kill()
	return
}

func (src *FfmpegPCMStream) Open() (err error) {
	var stderr io.ReadCloser
	var stdout io.ReadCloser

	args := []string{
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
		args = append(args, "-ac", strconv.FormatInt(int64(src.ReadChannels), 10))
	}

	if src.ReadSampleRate != 0 {
		args = append(args, "-ar", strconv.FormatInt(int64(src.ReadSampleRate), 10))
	}

	args = append(args, "-")

	src.cmd = exec.Command(
		GlobalConfig.FfmpegPath,
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

func (src *FfmpegPCMStream) ReadSampleBlock() (*SampleBlock, error) {
	bytesPerSample := src.ReadDepth / 8

	var sampleData interface{}

	switch src.ReadDepth {
	case 16:
		sampleData = make([]SampleInt16, BufferSize/bytesPerSample)
	case 32:
		sampleData = make([]SampleInt32, BufferSize/bytesPerSample)
	}

	if err := binary.Read(src.stdout, binary.LittleEndian, sampleData); err != nil {
		return nil, err
	}

	a := float32(bytesPerSample*src.I.SampleRate) / float32(src.I.Channels)

	//44100 samples per second. 2 bytes (16 bit) per sample
	time := float32(src.offset) / a
	duration := float32(BufferSize) / a

	src.offset += BufferSize

	return &SampleBlock{
		Data:     sampleData,
		Time:     time,
		Duration: duration,
	}, nil
}

func (src *FfmpegPCMStream) SampleDepth() int {
	return src.ReadDepth
}

func (src *FfmpegPCMStream) Read(p []byte) (int, error) {
	return src.stdout.Read(p)
}

type FfmpegRGBAStream struct {
	Path     string
	Start    float32
	Duration float32
	I        *FrameSrcInfo

	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	offset int64
}

func (g *FfmpegRGBAStream) Close() (err error) {
	err = g.cmd.Process.Kill()
	return
}

func (g *FfmpegRGBAStream) Open() (err error) {
	var (
		stdout, stderr io.ReadCloser
	)

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
			strconv.FormatFloat(float64(g.Start), 'f', -1, 32),
		)
	}

	if g.Duration > 0 {
		args = append(args,
			"-t",
			strconv.FormatFloat(float64(g.Duration), 'f', -1, 32),
		)
	}

	args = append(args, "-")

	g.cmd = exec.Command(GlobalConfig.FfmpegPath, args...)

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

func (g *FfmpegRGBAStream) Read(p []byte) (n int, err error) {
	r, err := g.stdout.Read(p)
	g.offset += int64(r)
	return
}

func (g *FfmpegRGBAStream) Info() *FrameSrcInfo {
	return g.I
}

func (g *FfmpegRGBAStream) ReadFrame() (*Frame, error) {
	frameBytes := make([]byte, PixelDepth*g.I.Width*g.I.Height)
	frameIndex := int(g.offset / int64(len(frameBytes)))

	time := float32(frameIndex) * float32(1./g.I.FrameRate)

	if g.Duration != 0 && time > g.Duration {
		return nil, io.EOF
	}

	r, err := io.ReadFull(g.stdout, frameBytes)

	if err != nil {
		if r == 0 { //got invalid file descriptor (ffmpeg autocloses stdout?)
			err = io.EOF
		}

		return nil, err
	}

	g.offset += int64(r)

	return &Frame{
		Data:   frameBytes,
		Width:  g.I.Width,
		Height: g.I.Height,
		Index:  frameIndex,
		Time:   time,
	}, nil
}
