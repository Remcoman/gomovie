package main

import (
	// "github.com/Remcoman/gomovie"
	
	// "os"
	// "testing"
	// "io"
	
	// "image"
)

//func TestTransformFrames(t *testing.T) {
	// path := os.Getenv("GOMOVIE_VIDEO")
	// if path == "" {
	// 	t.Fatal("GOMOVIE_VIDEO not set!")
	// }
	
	// vid := &gomovie.VideoSource{
	// 	Path : path,
	// }
	// vid.Open()
	
	// tr := &gomovie.FrameTransformer{Src : vid}
	
	// tr.AddTransform(func (bytes *[]byte, info *gomovie.VideoInfo) {
	// 	img := image.NRGBA{Pix : *bytes, Stride : 4 * info.Width, Rect : image.Rect(0,0,info.Width, info.Height)}
		
	// 	part := img.SubImage(image.Rect(0,0,200,200)).(*image.NRGBA)
		
	// 	var (x,y int)

	// 	for y < part.Rect.Dy() {
	// 		x = 0
	// 		for x < part.Rect.Dx() {
	// 			px := part.PixOffset(x,y)
	// 			part.Pix[px] = 255
	// 			part.Pix[px+1] = 255
	// 			part.Pix[px+2] = 255
	// 			part.Pix[px+3] = 255
	// 			x++
	// 		}
	// 		y++
	// 	}
		
	// 	*bytes = img.Pix
	// })
	
	// out := &gomovie.VideoOutput{
	// 	Path : "../videos/test2.mp4",
	// }
	
	// defer out.Close()
	
	// t.Log("Writing video to ../videos/test2.mp4")
	
	// _, err := io.Copy(out, tr)
	// if err != nil {
	// 	t.Fatal(err)
	// }
//}