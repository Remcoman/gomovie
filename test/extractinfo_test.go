package main

import (
	"testing"
	"os"
	
	"github.com/Remcoman/gomovie"
)

func TestInfo(t *testing.T) {
	video := os.Getenv("GOMOVIE_VIDEO")
	if video == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}
	
	if videoInfo, audioInfo, err := gomovie.ExtractInfo(video); err != nil {
		t.Fatal(err)
	} else {
		
		if videoInfo != nil {
			t.Log(videoInfo.String())
		}
		
		if audioInfo != nil {
			t.Log(audioInfo.String())	
		}
		
	}
	
} 
