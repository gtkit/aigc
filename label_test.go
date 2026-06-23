package aigc_test

import (
	"errors"
	"testing"

	"github.com/gtkit/aigc"
)

func TestWriteMetadataUnsupportedFormat(t *testing.T) {
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "P"}
	if _, err := aigc.WriteMetadata([]byte("x"), aigc.Format(9), id); !errors.Is(err, aigc.ErrUnsupportedFormat) {
		t.Fatalf("err = %v, 期望 ErrUnsupportedFormat", err)
	}
}

func TestWriteMetadataInvalidID(t *testing.T) {
	if _, err := aigc.WriteMetadata(aigc.RhythmWAV(16000), aigc.WAV, aigc.Identifier{Label: "9"}); !errors.Is(err, aigc.ErrInvalidLabel) {
		t.Fatalf("err = %v, 期望 ErrInvalidLabel", err)
	}
}

func TestPrependCueUnsupportedFormat(t *testing.T) {
	if _, err := aigc.PrependCue([]byte("a"), []byte("b"), aigc.Format(9), aigc.AtStart); !errors.Is(err, aigc.ErrUnsupportedFormat) {
		t.Fatalf("err = %v, 期望 ErrUnsupportedFormat", err)
	}
}
