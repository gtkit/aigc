package aigc_test

import (
	"encoding/binary"
	"testing"

	"github.com/gtkit/aigc"
)

func TestRhythmWAV(t *testing.T) {
	tests := []struct {
		name       string // 用例名
		sampleRate int    // 入参采样率
		wantRate   uint32 // WAV 头部期望采样率
	}{
		{"44100", 44100, 44100},
		{"16000", 16000, 16000},
		{"非法回退默认", 0, 16000},
		{"负值回退默认", -1, 16000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wav := aigc.RhythmWAV(tt.sampleRate)
			if len(wav) <= 44 {
				t.Fatalf("WAV 长度 %d，应含头部与数据", len(wav))
			}
			if string(wav[:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
				t.Fatalf("缺少 RIFF/WAVE 头: %q %q", wav[:4], wav[8:12])
			}
			if ch := binary.LittleEndian.Uint16(wav[22:24]); ch != 1 {
				t.Fatalf("声道数 = %d, 期望 1", ch)
			}
			if bits := binary.LittleEndian.Uint16(wav[34:36]); bits != 16 {
				t.Fatalf("位深 = %d, 期望 16", bits)
			}
			if rate := binary.LittleEndian.Uint32(wav[24:28]); rate != tt.wantRate {
				t.Fatalf("采样率 = %d, 期望 %d", rate, tt.wantRate)
			}
		})
	}
}

func BenchmarkRhythmWAV(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = aigc.RhythmWAV(44100)
	}
}

func ExampleRhythmWAV() {
	wav := aigc.RhythmWAV(16000) // “短长短短”提示音，单声道 16-bit PCM WAV
	_ = wav                      // 由上层编码为 mp3 后经 PrependMP3 注入正文起始位置
}
