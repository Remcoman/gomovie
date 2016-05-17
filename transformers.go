package gomovie

type FrameTransformerFunc func(*[]byte, *VideoInfo)

type FrameTransformer struct {
	Src VideoReader
	transforms []FrameTransformerFunc
}

func (tr *FrameTransformer) AddTransform(f FrameTransformerFunc) *FrameTransformer {
	tr.transforms = append(tr.transforms, f)
	return tr
}

func (tr *FrameTransformer) applyTransform(fr *Frame) {
	bytes := &fr.Bytes
	
	for _, transform := range tr.transforms {
		transform(bytes, tr.Info())
	}
}

func (tr *FrameTransformer) Info() *VideoInfo {
	return tr.Src.Info()
}

func (tr *FrameTransformer) Read(p []byte) (int, error) {
	fr, err := tr.ReadFrame()
	if err != nil {
		return 0, err
	}
	return copy(p, fr.Bytes), nil
}

func (tr *FrameTransformer) ReadFrame() (*Frame, error) {
	fr, err := tr.Src.ReadFrame()
	if err != nil {
		return nil, err
	}
	tr.applyTransform(fr)
	return fr, nil
}