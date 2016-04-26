package main

import (
	"fmt"
	"image/png"
	"os"
	"log"
	"flag"
	
	"github.com/Remcoman/gomovie"
)

func main() {
	video := flag.String("file", "", "The video file")
	flag.Parse()
	
	grabber := gomovie.CreateGrabber(*video)
	grabber.Open()
	
	frame, err := grabber.GetFrame()
	
	if err != nil {
		log.Fatal("Could not get frame from video")
	}
	
	img := frame.ToNRGBAImage()
	
	file, err := os.Create("frame.png")
	if err != nil {
		fmt.Println(err)
	}
	
	defer file.Close()
	
	err = png.Encode(file, img)
	if err != nil {
		fmt.Println(err)
	}
}