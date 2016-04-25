package main

import "fmt"
import "image/png"
import "os"

import "github.com/Remcoman/gomovie"

func main() {
	grabber := gomovie.CreateGrabber("../app.webm")
	grabber.Open()
	
	frame := grabber.GetFrame()
	
	img := frame.ToNRGBAImage()
	
	file, err := os.Create("app.png")
	if err != nil {
		fmt.Println(err)
	}
	
	defer file.Close()
	
	err = png.Encode(file, img)
	if err != nil {
		fmt.Println(err)
	}
}