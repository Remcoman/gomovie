package gomovie

type TransformerFunc func(*Frame) *Frame

type FrameTransformer struct {
	src FrameReader
	transforms []TransformerFunc
}

func (tr *FrameTransformer) Info() *Info {
	return tr.Info()
}

func (tr *FrameTransformer) AddTransform(f TransformerFunc) *FrameTransformer {
	tr.transforms = append(tr.transforms, f)
	return tr
}