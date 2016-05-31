package gomovie

import (
	"bytes"
	"encoding/binary"
	"io"
	"sort"
	"sync"
)

const (
	DefaultParallel = 5
)

type sortedFrames []*Frame

func (s sortedFrames) Len() int {
	return len(s)
}

func (s sortedFrames) Less(i int, j int) bool {
	return s[i].Index < s[j].Index
}

func (s sortedFrames) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}

// FrameTransform Describes the frame transform operation. Each transform should modify the Bytes field.
// The resize operation is optional and is called before the transform. The resize operation should modify the Width and Height of the frame.
type FrameTransform struct {
	Transform func(f *Frame)
	Resize    func(f *Frame)
}

// NewFrameTransformer convenience constructor to create a new FrameTransformer from a Video or an FrameReader
func NewFrameTransformer(src interface{}) *FrameTransformer {
	switch t := src.(type) {
	case *Video:
		return &FrameTransformer{FrameReader: t.FrameReader}
	case FrameReader:
		return &FrameTransformer{FrameReader: t}
	default:
		panic("Can not create a FrameTransformer!")
	}
}

// FrameTransformer applies a transform to each frame. Great for video editing. implements the FrameReader interface.
type FrameTransformer struct {
	FrameReader

	transforms    []FrameTransform
	ParallelCount int

	processing bool
	todo       chan *Frame
	done       chan *Frame
	buffer     sortedFrames
	current    *Frame
	nextIndex  int
	quit       chan bool
}

// AddTransform appends a transform to the frame transform list
func (ft *FrameTransformer) AddTransform(f FrameTransform) *FrameTransformer {
	ft.transforms = append(ft.transforms, f)
	return ft
}

func (ft *FrameTransformer) Close() error {
	ft.quit <- true
	close(ft.todo)
	return nil
}

func (ft *FrameTransformer) Read(p []byte) (int, error) {
	if !ft.processing {
		parallel := ft.ParallelCount
		if parallel == 0 {
			parallel = DefaultParallel
		}

		ft.todo = make(chan *Frame, parallel)
		ft.done = make(chan *Frame)
		ft.processing = true

		go ft.process(parallel)
	}

	var i int

	for i = 0; i < 100; i++ {

		if ft.current == nil && len(ft.buffer) > 0 {
			sort.Sort(ft.buffer) //todo don't always sort

			if ft.buffer[0].Index == ft.nextIndex {
				ft.current = ft.buffer[0]
			}
		}

		if ft.current != nil { //found something
			c := copy(p, ft.current.Data)

			if c == 0 { //nothing could be copied so we should go to the next frame
				ft.nextIndex++
				ft.buffer = ft.buffer[1:]
				ft.current = nil
			} else {
				ft.current.Data = ft.current.Data[c:]
				return c, nil
			}

		} else {

			f, ok := <-ft.done
			if !ok { //closed channel!
				//if done is closed but we still have data in the buffer something is really wrong and we are missing some frames
				break
			}

			ft.buffer = append(ft.buffer, f)
		}

	}

	err := io.EOF

	if i == 1000 {
		err = io.ErrNoProgress //don't know what happened here
	}

	return 0, err
}

//read a single frame and apply the transforms to it
func (ft *FrameTransformer) ReadFrame() (*Frame, error) {
	f, err := ft.FrameReader.ReadFrame()
	if err != nil {
		return nil, err
	}

	fc := *f
	fp := &fc
	ft.applyResizes(fp)
	ft.applyTransforms(fp)
	return fp, nil
}

func (ft *FrameTransformer) applyResizes(f *Frame) {
	for _, transform := range ft.transforms {
		if transform.Resize != nil {
			transform.Resize(f)
		}
	}
}

func (ft *FrameTransformer) applyTransforms(f *Frame) {
	for _, transform := range ft.transforms {
		transform.Transform(f)
	}
}

func (ft *FrameTransformer) process(parallel int) {
	//read until there are no more frames
	go func() {
		for {
			f, err := ft.FrameReader.ReadFrame()
			if err != nil { //no more frames! (or a strange error)
				close(ft.todo)
				break
			}

			select {
			case ft.todo <- f:
			case <-ft.quit:
				return
			}
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(parallel)

	//create x number of frame transformers which monitor the todo list
	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			for f := range ft.todo {
				ft.applyResizes(f)
				ft.applyTransforms(f)
				ft.done <- f
			}
		}()
	}

	wg.Wait()

	//all frame transformers were done so we can signal that the done channel needs to be closed
	close(ft.done)
}

// SampleTransform Describes the sample transform operation. Each transform should modify the SampleBlock.
type SampleTransform struct {
	Transform func(s *SampleBlock, depth int)
}

// NewSampleTransformer convenience constructor to create a new SampleTransformer from a Video or an SampleReader
func NewSampleTransformer(src interface{}) *SampleTransformer {
	switch t := src.(type) {
	case *Video:
		return &SampleTransformer{SampleReader: t.SampleReader}
	case SampleReader:
		return &SampleTransformer{SampleReader: t}
	default:
		panic("Can not create sample transformer")
	}
}

// SampleTransformer applies a transform to each sample block. Great for audio editing. implements the SampleReader interface.
type SampleTransformer struct {
	SampleReader
	transforms []SampleTransform
}

// AddTransform appends a transform to the frame transform list
func (ft *SampleTransformer) AddTransform(f SampleTransform) *SampleTransformer {
	ft.transforms = append(ft.transforms, f)
	return ft
}

func (ft *SampleTransformer) applyTransforms(s *SampleBlock) {
	depth := ft.SampleDepth()
	for _, transform := range ft.transforms {
		transform.Transform(s, depth)
	}
}

func (ft *SampleTransformer) Read(p []byte) (int, error) {
	sb, err := ft.ReadSampleBlock()
	if err != nil {
		return 0, err
	}

	//write int16 data to byte array
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, sb.Data)

	return copy(p, buf.Bytes()), nil
}

//Read a single sampleblock which contains an array of int16 or int32 values (depending on the SampleFormat)
func (ft *SampleTransformer) ReadSampleBlock() (*SampleBlock, error) {
	sb, err := ft.SampleReader.ReadSampleBlock()
	if err != nil {
		return nil, err
	}

	sc := *sb
	sp := &sc
	ft.applyTransforms(sp)
	return sp, nil
}
