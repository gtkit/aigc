package aigc

import (
	"testing"
)

// TestReadMP3RoundTrip 验证 WriteMP3 写入的标识能被 ReadMP3 原样读回。
func TestReadMP3RoundTrip(t *testing.T) {
	t.Parallel()

	raw := []byte{0xff, 0xfb, 0x90, 0x00, 0x11, 0x22}
	id := Identifier{
		Label:             LabelMaybe,
		ContentProducer:   "producer",
		ProduceID:         "p1",
		ContentPropagator: "propagator",
		ReservedCode1:     "r1",
	}

	out, err := WriteMP3(raw, id)
	if err != nil {
		t.Fatalf("WriteMP3 error = %v", err)
	}

	got, found, err := ReadMP3(out)
	if err != nil {
		t.Fatalf("ReadMP3 error = %v", err)
	}
	if !found {
		t.Fatal("ReadMP3 未找到 AIGC 标识")
	}
	if got != id {
		t.Fatalf("round-trip 不一致: got %+v want %+v", got, id)
	}
}

// TestReadMP3NotLabeled 验证不含本包标识的输入返回 found=false 且无错误。
func TestReadMP3NotLabeled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   []byte
	}{
		{"裸 mp3 帧", []byte{0xff, 0xfb, 0x90, 0x00}},
		{"空输入", nil},
		{"过短", []byte{'I', 'D'}},
		{"非 2.4 版本", []byte{'I', 'D', '3', 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"标志位非 0", []byte{'I', 'D', '3', 0x04, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00}},
	}
	for _, tc := range cases {
		got, found, err := ReadMP3(tc.in)
		if err != nil {
			t.Errorf("%s: 期望 err=nil, got %v", tc.name, err)
		}
		if found {
			t.Errorf("%s: 期望 found=false, got %+v", tc.name, got)
		}
	}
}

// TestReadMP3Corrupt 验证标签声明长度越界时返回 ErrCorruptTag。
func TestReadMP3Corrupt(t *testing.T) {
	t.Parallel()

	// 合法 2.4 头，但 tagSize 声明 127 字节，实际无后续数据。
	corrupt := []byte{'I', 'D', '3', 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7f}
	if _, _, err := ReadMP3(corrupt); err != ErrCorruptTag {
		t.Fatalf("期望 ErrCorruptTag, got %v", err)
	}
}

// TestWriteMP3RejectRelabel 验证已含标识的 mp3 再次写入被拒。
func TestWriteMP3RejectRelabel(t *testing.T) {
	t.Parallel()

	id := Identifier{Label: LabelIs, ContentProducer: "producer"}
	out, err := WriteMP3([]byte{0xff, 0xfb}, id)
	if err != nil {
		t.Fatalf("首次 WriteMP3 error = %v", err)
	}
	if _, err := WriteMP3(out, id); err != ErrAlreadyLabeled {
		t.Fatalf("重复打标期望 ErrAlreadyLabeled, got %v", err)
	}
}
