package gomovie

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"
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
		v = &FfmpegRGBAStream{Path: path, i: frameInfo}
	}

	if audioInfo != nil {
		a = &FfmpegPCMStream{Path: path, o: NewSampleFormat(), i: audioInfo}
	}

	vid = &Video{v, a}

	return
}

type FfmpegPCMStream struct {
	Path       string
	Start      float64
	Duration   float64
	Channels   int
	SampleRate int

	i *SampleReaderInfo
	o *SampleFormat
	r *Range

	sampleDepth int
	cmd         *exec.Cmd
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	offset      int
	opened      bool
}

func (src *FfmpegPCMStream) Range() *Range           { return src.r }
func (src *FfmpegPCMStream) Info() *SampleReaderInfo { return src.i }

func (src *FfmpegPCMStream) Slice(r *Range) SampleReader {
	var dur float32
	if src.r != nil {
		dur = src.r.Duration
	} else {
		dur = src.Info().Duration
	}

	r = r.Intersection(&Range{Start: 0, Duration: dur})

	if src.r != nil {
		r.parent = src.r.parent
	}

	return &FfmpegPCMStream{Path: src.Path, i: src.i, r: r}
}

func (src *FfmpegPCMStream) Close() (err error) {
	err = src.cmd.Process.Kill()
	return
}

func (src *FfmpegPCMStream) open() (err error) {
	if src.opened {
		panic("Reader is already opened!")
	}

	var stderr, stdout io.ReadCloser

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
		"-f", fmt.Sprintf("s%vle", src.sampleDepth),
	)

	if src.Channels != 0 {
		args = append(args, "-ac", strconv.FormatInt(int64(src.Channels), 10))
	}

	if src.SampleRate != 0 {
		args = append(args, "-ar", strconv.FormatInt(int64(src.SampleRate), 10))
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

	src.opened = true

	return nil
}

func (src *FfmpegPCMStream) ReadSampleBlock() (*SampleBlock, error) {
	if !src.opened {
		if err := src.open(); err != nil {
			return nil, err
		}
	}

	bytesPerSample := src.sampleDepth / 8

	//TODO there might not be enough data to fill the whole sample block
	//so we need to update the duration to reflect the cropped data

	b := make([]byte, src.o.BlockSize) //todo reuse this buffer
	br, _ := io.ReadFull(src.stdout, b)
	b = b[:br]

	var sampleData interface{}

	switch src.sampleDepth {
	case 32:
		sampleData = make([]SampleInt32, br/bytesPerSample)
	default:
		sampleData = make([]SampleInt16, br/bytesPerSample)
	}

	if err := binary.Read(bytes.NewBuffer(b), binary.LittleEndian, sampleData); err != nil {

		//process was already closed but we are still trying to read from it
		if src.cmd.ProcessState != nil {
			return nil, io.EOF
		}

		return nil, err
	}

	a := float32(bytesPerSample*src.i.SampleRate) / float32(src.i.Channels)

	//44100 samples per second. 2 bytes (16 bit) per sample
	time := float32(src.offset) / a
	duration := float32(br) / a

	src.offset += br

	return &SampleBlock{src.o, sampleData, time, duration}, nil
}

func (src *FfmpegPCMStream) SampleFormat() *SampleFormat {
	return src.o
}

func (src *FfmpegPCMStream) Read(p []byte) (int, error) {
	if !src.opened {
		if err := src.open(); err != nil {
			return 0, err
		}
	}
	return src.stdout.Read(p)
}

type FfmpegRGBAStream struct {
	Path string

	i *FrameReaderInfo
	r *Range

	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	offset int64
	opened bool
}

func (g *FfmpegRGBAStream) Range() *Range {
	return g.r
}

func (g *FfmpegRGBAStream) Slice(r *Range) FrameReader {
	var dur float32
	if g.r != nil {
		dur = g.r.Duration
	} else {
		dur = g.Info().Duration
	}

	r = r.Intersection(&Range{Start: 0, Duration: dur})

	if g.r != nil {
		r.parent = g.r.parent
	}

	return &FfmpegRGBAStream{Path: g.Path, i: g.i, r: r}
}

func (g *FfmpegRGBAStream) Close() (err error) {
	err = g.cmd.Process.Kill()
	return
}

func (g *FfmpegRGBAStream) open() (err error) {
	if g.opened {
		panic("Reader is already opened!")
	}

	var stdout, stderr io.ReadCloser

	args := []string{
		"-loglevel", "error",

		"-i", g.Path,

		"-f", "image2pipe",
		"-pix_fmt", "rgba",
		"-vcodec", "rawvideo",
	}

	if g.r != nil {
		if g.r.Start > 0 {
			args = append(args,
				"-ss",
				strconv.FormatFloat(float64(g.r.Start), 'f', -1, 32),
			)
		}

		if g.r.Duration > 0 {
			args = append(args,
				"-t",
				strconv.FormatFloat(float64(g.r.Duration), 'f', -1, 32),
			)
		}
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

	g.opened = true

	return nil
}

func (g *FfmpegRGBAStream) Read(p []byte) (n int, err error) {
	if g.r != nil && g.r.Duration == 0 {
		return 0, io.EOF
	}

	if !g.opened {
		if err = g.open(); err != nil {
			return 0, err
		}
	}

	n, err = g.stdout.Read(p)
	g.offset += int64(n)
	return
}

func (g *FfmpegRGBAStream) Info() *FrameReaderInfo {
	return g.i
}

func (g *FfmpegRGBAStream) ReadFrame() (*Frame, error) {
	if g.r != nil && g.r.Duration == 0 {
		return nil, io.EOF
	}

	if !g.opened {
		if err := g.open(); err != nil {
			return nil, err
		}
	}

	frameBytes := make([]byte, GlobalConfig.FramePixelDepth*g.i.Width*g.i.Height)
	frameIndex := int(g.offset / int64(len(frameBytes)))

	time := float32(frameIndex) * float32(1./g.i.FrameRate)

	if time > g.r.Duration {
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
		Width:  g.i.Width,
		Height: g.i.Height,
		Index:  frameIndex,
		Time:   time,
	}, nil
}
