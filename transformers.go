package gomovie

import (
	"bytes"
	"encoding/binary"
)

// FrameTransform Describes the frame transform operation. Each transform should modify the Bytes field.
// The resize operation is optional and is called before the transform. The resize operation should modify the Width and Height of the frame.
type FrameTransform struct {
	Transform func(f *Frame)
	Resize    func(f *Frame)
}

// NewFrameTransformer convenience constructor to create a new FrameTransformer from a Video or an FrameReader
func NewFrameTransformer(src interface{}) FrameTransformer {
	switch t := src.(type) {
	case Video:
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
	// r := [5]*Frame

	// for i := 0; i < 5; i++ {
	// 	f, err := ft.FrameReader.ReadFrame()
	// 	if err != nil {
	// 		break
	// 	}
	// 	r[i] =
	// }

	//out := make(chan *Frame, 5)

	//stuff the channel with x frames & process the frames in parallel? & output the result ordered

	fr, err := ft.ReadFrame()
	if err != nil {
		return 0, err
	}
	return copy(p, fr.Data), nil
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
	case Video:
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
