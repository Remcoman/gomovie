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
	
	if info, err := gomovie.ExtractInfo(video); err != nil {
		t.Fatal(err)
	} else {
		t.Log(info.String())
	}
	
} 
