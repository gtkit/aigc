package aigc

import (
	"bytes"
	"testing"

	"github.com/gtkit/json/v2"
)

func TestIdentifierJSON(t *testing.T) {
	t.Parallel()

	id := Identifier{
		Label:           LabelMaybe,
		ContentProducer: "test-producer-code",
		ProduceID:       "produce-001",
	}
	got, err := id.JSON()
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	var back Identifier
	if err := json.Unmarshal([]byte(got), &back); err != nil {
		t.Fatalf("unmarshal AIGC json failed: %v", err)
	}
	if back != id {
		t.Fatalf("round-trip mismatch: got %+v want %+v", back, id)
	}
}

func TestIdentifierValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		id      Identifier
		wantErr error
	}{
		{"label 非法", Identifier{Label: "9", ContentProducer: "x"}, ErrInvalidLabel},
		{"缺生成者", Identifier{Label: LabelIs}, ErrProducerRequired},
		{"合法", Identifier{Label: LabelIs, ContentProducer: "x"}, nil},
	}
	for _, tc := range cases {
		if err := tc.id.Validate(); err != tc.wantErr {
			t.Errorf("%s: Validate() = %v want %v", tc.name, err, tc.wantErr)
		}
	}
}

// TestWriteMP3 验证：写入后字节以 ID3v2.4 标签开头，且能按 synchsafe/帧结构解析出
// description=AIGC 的 TXXX 值，反序列化与原标识一致；标签之后紧跟原 mp3 裸帧。
func TestWriteMP3(t *testing.T) {
	t.Parallel()

	raw := []byte{0xff, 0xfb, 0x90, 0x00, 0x11, 0x22} // 伪 mp3 帧
	id := Identifier{Label: LabelMaybe, ContentProducer: "producer", ProduceID: "p1"}

	out, err := WriteMP3(raw, id)
	if err != nil {
		t.Fatalf("WriteMP3 error = %v", err)
	}

	// 标签头
	if string(out[:3]) != "ID3" || out[3] != 0x04 || out[4] != 0x00 {
		t.Fatalf("缺少 ID3v2.4 标签头: % x", out[:5])
	}
	tagSize := readSynchsafe(out[6:10])
	tagEnd := 10 + tagSize

	// 尾部应为原始 mp3
	if !bytes.Equal(out[tagEnd:], raw) {
		t.Fatalf("标签后未原样跟随 mp3: % x", out[tagEnd:])
	}

	// 解析 TXXX 帧
	frame := out[10:tagEnd]
	if string(frame[:4]) != "TXXX" {
		t.Fatalf("首帧不是 TXXX: %q", frame[:4])
	}
	frameSize := readSynchsafe(frame[4:8])
	body := frame[10 : 10+frameSize]
	if body[0] != 0x03 {
		t.Fatalf("encoding 应为 UTF-8(0x03), got 0x%02x", body[0])
	}
	nul := bytes.IndexByte(body[1:], 0x00)
	if nul < 0 {
		t.Fatal("description 缺少结束符")
	}
	desc := string(body[1 : 1+nul])
	if desc != "AIGC" {
		t.Fatalf("description = %q want AIGC", desc)
	}
	value := body[1+nul+1:]

	var back Identifier
	if err := json.Unmarshal(value, &back); err != nil {
		t.Fatalf("AIGC 值无法反序列化: %v (%q)", err, value)
	}
	if back != id {
		t.Fatalf("AIGC 值不一致: got %+v want %+v", back, id)
	}
}

func readSynchsafe(b []byte) int {
	return int(b[0])<<21 | int(b[1])<<14 | int(b[2])<<7 | int(b[3])
}
