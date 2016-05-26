package gomovie

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sort"
	"sync"
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
func NewFrameTransformer(src interface{}) FrameTransformer {
	switch t := src.(type) {
	case VideoReader:
		return FrameTransformer{FrameReader: t.FrameReader}
	case FrameReader:
		return FrameTransformer{FrameReader: t}
	default:
		panic("Can not create a FrameTransformer!")
	}
}

// FrameTransformer applies a transform to each frame. Great for video editing. implements the FrameReader interface.
type FrameTransformer struct {
	FrameReader
	transforms []FrameTransform
}

// AddTransform appends a transform to the frame transform list
func (ft *FrameTransformer) AddTransform(f FrameTransform) *FrameTransformer {
	ft.transforms = append(ft.transforms, f)
	return ft
}

func (ft *FrameTransformer) applyResizes(f *Frame) {
	for _, transform := range ft.transforms {
		transform.Resize(f)
	}
}

func (ft *FrameTransformer) applyTransforms(f *Frame) {
	for _, transform := range ft.transforms {
		transform.Transform(f)
	}
}

func (ft *FrameTransformer) Read(p []byte) (int, error) {
	todo := make([]*Frame, 0, 5)

	//sequentally read some frames
	for i := 0; i < 4; i++ {
		f, err := ft.FrameReader.ReadFrame()
		if err != nil {
			break
		}
		todo = append(todo, f)
	}

	if len(todo) == 0 {
		return 0, errors.New("Could not read any frames!")
	}

	var wait sync.WaitGroup
	wait.Add(len(todo))

	done := make(chan *Frame)
	for _, f := range todo {
		go func(f Frame) {
			defer wait.Done()
			fp := &f
			ft.applyResizes(fp)
			ft.applyTransforms(fp)
			done <- fp
		}(*f)
	}

	wait.Wait()
	close(done)

	sorted := make(sortedFrames, 0, 5)
	for x := range done {
		sorted = append(sorted, x)
	}
	sort.Sort(sorted)

	total := 0
	for _, f := range sorted {
		total += copy(p, f.Data)
	}

	return total, nil
}

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

type SampleTransform struct {
	Transform func(s *SampleBlock, depth int)
}

func NewSampleTransformer(src interface{}) SampleTransformer {
	switch t := src.(type) {
	case VideoReader:
		return SampleTransformer{SampleReader: t.SampleReader}
	case SampleReader:
		return SampleTransformer{SampleReader: t}
	default:
		panic("Can not create sample transformer")
	}
}

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
