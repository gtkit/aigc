package aigc_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/gtkit/aigc"
)

// findWAVChunk 在 RIFF 里查找指定 id 的 chunk 内容（测试辅助）。
func findWAVChunk(b []byte, id string) []byte {
	if len(b) < 12 {
		return nil
	}
	off := 12
	for off+8 <= len(b) {
		cid := string(b[off : off+4])
		size := int(binary.LittleEndian.Uint32(b[off+4 : off+8]))
		start := off + 8
		if start+size > len(b) {
			break
		}
		if cid == id {
			return b[start : start+size]
		}
		off = start + size
		if size%2 == 1 {
			off++
		}
	}
	return nil
}

func TestWAVWriteMetadata(t *testing.T) {
	wav := aigc.RhythmWAV(16000)
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "PRODUCER-001", ProduceID: "X-1"}

	out, err := aigc.WriteMetadata(wav, aigc.WAV, id)
	if err != nil {
		t.Fatalf("WriteMetadata: %v", err)
	}
	if len(out) <= len(wav) {
		t.Fatal("输出未增长，未写入标识")
	}
	chunk := findWAVChunk(out, "AIGC")
	if chunk == nil {
		t.Fatal("未找到 AIGC chunk")
	}
	if !bytes.Contains(chunk, []byte(`"Label":"1"`)) ||
		!bytes.Contains(chunk, []byte(`"ContentProducer":"PRODUCER-001"`)) {
		t.Fatalf("AIGC chunk 内容不符: %s", chunk)
	}
	if findWAVChunk(out, "data") == nil || findWAVChunk(out, "fmt ") == nil {
		t.Fatal("写标识后 fmt/data 块丢失")
	}
	if findWAVChunk(out, "LIST") == nil {
		t.Fatal("未写入 LIST/INFO 降险副本")
	}
}

func TestWAVWriteMetadataAlreadyLabeled(t *testing.T) {
	wav := aigc.RhythmWAV(16000)
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "P"}
	once, err := aigc.WriteMetadata(wav, aigc.WAV, id)
	if err != nil {
		t.Fatalf("首次 WriteMetadata: %v", err)
	}
	if _, err := aigc.WriteMetadata(once, aigc.WAV, id); !errors.Is(err, aigc.ErrAlreadyLabeled) {
		t.Fatalf("重复打标 err = %v, 期望 ErrAlreadyLabeled", err)
	}
}

func TestWAVPrependCue(t *testing.T) {
	body := aigc.RhythmWAV(16000)
	cue := aigc.RhythmWAV(16000)
	bodyData := findWAVChunk(body, "data")
	cueData := findWAVChunk(cue, "data")

	out, err := aigc.PrependCue(body, cue, aigc.WAV, aigc.AtStart)
	if err != nil {
		t.Fatalf("PrependCue: %v", err)
	}
	outData := findWAVChunk(out, "data")
	if len(outData) != len(bodyData)+len(cueData) {
		t.Fatalf("拼接后 data 长度 = %d, 期望 %d", len(outData), len(bodyData)+len(cueData))
	}
	if !bytes.HasPrefix(outData, cueData) {
		t.Fatal("AtStart 时 cue 应在 data 起始")
	}
}

func TestWAVPrependCueMismatch(t *testing.T) {
	body := aigc.RhythmWAV(16000)
	cue := aigc.RhythmWAV(44100)
	if _, err := aigc.PrependCue(body, cue, aigc.WAV, aigc.AtStart); !errors.Is(err, aigc.ErrWAVParamMismatch) {
		t.Fatalf("err = %v, 期望 ErrWAVParamMismatch", err)
	}
}

func TestWAVInvalid(t *testing.T) {
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "P"}
	if _, err := aigc.WriteMetadata([]byte("not a wav at all"), aigc.WAV, id); !errors.Is(err, aigc.ErrInvalidWAV) {
		t.Fatalf("err = %v, 期望 ErrInvalidWAV", err)
	}
}

func TestWAVPrependCueCorruptSize(t *testing.T) {
	corrupt := []byte("RIFF")
	corrupt = append(corrupt, 0x00, 0x00, 0x00, 0x00)
	corrupt = append(corrupt, "WAVE"...)
	corrupt = append(corrupt, "fmt "...)
	corrupt = append(corrupt, 0xff, 0xff, 0xff, 0x7f) // chunk size = 0x7fffffff，远超实际
	cue := aigc.RhythmWAV(16000)

	if _, err := aigc.PrependCue(corrupt, cue, aigc.WAV, aigc.AtStart); err == nil {
		t.Fatal("越界 chunk size 应返回错误，而非 panic 或成功")
	}
}
