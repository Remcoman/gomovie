package main

import (
	"github.com/Remcoman/gomovie"
	"os"
	"testing"
)

func TestAudioToMP3(t *testing.T) {
	path := os.Getenv("GOMOVIE_VIDEO")
	if path == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}
	
	audio := gomovie.AudioSource{Path : path}
	audio.Open()
}