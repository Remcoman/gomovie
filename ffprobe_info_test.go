package gomovie_test

import (
	"os"
	"testing"

	"github.com/Remcoman/gomovie"
)

func TestInfo(t *testing.T) {
	video := os.Getenv("GOMOVIE_VIDEO")
	if video == "" {
		t.Fatal("GOMOVIE_VIDEO not set!")
	}

	if frameInfo, audioInfo, err := gomovie.ExtractInfo(video); err != nil {
		t.Fatal(err)
	} else {

		if frameInfo != nil {
			t.Log(frameInfo)
		}

		if audioInfo != nil {
			t.Log(audioInfo)
		}

	}

}
