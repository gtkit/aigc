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

func TestLabel(t *testing.T) {
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "P-001"}
	wav := aigc.RhythmWAV(16000)
	cue := aigc.RhythmWAV(16000)
	baseData := findWAVChunk(wav, "data")

	// 显式 + 隐式：拼 cue 且写 AIGC
	out, err := aigc.Label(wav, cue, aigc.WAV, aigc.AtStart, id)
	if err != nil {
		t.Fatalf("Label: %v", err)
	}
	if findWAVChunk(out, "AIGC") == nil {
		t.Fatal("Label 后缺 AIGC 隐式标识")
	}
	if len(findWAVChunk(out, "data")) <= len(baseData) {
		t.Fatal("Label 未拼接显式标识（data 未增长）")
	}

	// cue 为空：只隐式，不拼接
	out2, err := aigc.Label(wav, nil, aigc.WAV, aigc.AtStart, id)
	if err != nil {
		t.Fatalf("Label(nil cue): %v", err)
	}
	if findWAVChunk(out2, "AIGC") == nil {
		t.Fatal("Label(nil cue) 后缺 AIGC 隐式标识")
	}
	if len(findWAVChunk(out2, "data")) != len(baseData) {
		t.Fatal("Label(nil cue) 不应改动 data")
	}
}
