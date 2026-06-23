package aigc_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gtkit/aigc"
)

// fakeMP3 是测试用的裸 mp3 帧（非真实可播放音频，仅用于字节级断言）。
var fakeMP3 = []byte{0xff, 0xfb, 0x90, 0x00, 0x01, 0x02, 0x03, 0x04}

func TestMP3WriteMetadata(t *testing.T) {
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "PRODUCER-001"}
	out, err := aigc.WriteMetadata(fakeMP3, aigc.MP3, id)
	if err != nil {
		t.Fatalf("WriteMetadata: %v", err)
	}
	if string(out[0:3]) != "ID3" {
		t.Fatal("输出未以 ID3 开头")
	}
	if !bytes.Contains(out, []byte("TXXX")) ||
		!bytes.Contains(out, []byte("AIGC")) ||
		!bytes.Contains(out, []byte(`"ContentProducer":"PRODUCER-001"`)) {
		t.Fatal("ID3 TXXX 内容不符")
	}
	if !bytes.HasSuffix(out, fakeMP3) {
		t.Fatal("原始 mp3 帧未原样保留在末尾")
	}
}

func TestMP3WriteMetadataAlreadyLabeled(t *testing.T) {
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "P"}
	withTag := []byte("ID3\x04\x00\x00\x00\x00\x00\x00")
	if _, err := aigc.WriteMetadata(withTag, aigc.MP3, id); !errors.Is(err, aigc.ErrAlreadyLabeled) {
		t.Fatalf("err = %v, 期望 ErrAlreadyLabeled", err)
	}
}

func TestMP3PrependCue(t *testing.T) {
	audio := []byte{0xff, 0xfb, 0x01, 0x02}
	cue := []byte{0xff, 0xfb, 0x03, 0x04}

	start, _ := aigc.PrependCue(audio, cue, aigc.MP3, aigc.AtStart)
	if !bytes.Equal(start, append(append([]byte{}, cue...), audio...)) {
		t.Fatalf("AtStart 应为 cue+audio: %v", start)
	}
	end, _ := aigc.PrependCue(audio, cue, aigc.MP3, aigc.AtEnd)
	if !bytes.Equal(end, append(append([]byte{}, audio...), cue...)) {
		t.Fatalf("AtEnd 应为 audio+cue: %v", end)
	}
}
