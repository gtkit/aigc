package aigc

import (
	"encoding/binary"
	"math"
)

// 本文件零依赖合成“短长短短”摩斯节奏提示音（AI 摩斯码），输出单声道 16-bit PCM WAV。
// 不引入 mp3 编码器：WAV 由上层音频管线编码为 mp3 后，经 PrependMP3 注入正文起始位置。

// 节奏与音色参数。“短长短短” = ·−·· ，dot 为短拍、dash 为长拍，符号间以等长静音分隔。
const (
	rhythmDefaultSampleRate = 16000 // 默认采样率，与百度 aue=3 mp3 常用采样率一致
	rhythmFreq              = 800.0 // 提示音频率（Hz）
	rhythmDotMillis         = 120   // 短拍时长（毫秒）
	rhythmDashMillis        = 360   // 长拍时长（毫秒）
	rhythmGapMillis         = 120   // 符号间静音时长（毫秒）
	rhythmAmplitude         = 0.6   // 振幅（占满幅比例），留余量避免削顶
)

// rhythmPattern 为“短长短短”的拍长序列（true=长拍 dash，false=短拍 dot）。
var rhythmPattern = []bool{false, true, false, false}

// RhythmWAV 合成“短长短短”摩斯节奏提示音，返回单声道 16-bit PCM 的 WAV 字节。
// sampleRate 应与正文音频一致以便注入；sampleRate <= 0 时回退到包内默认采样率。
func RhythmWAV(sampleRate int) []byte {
	if sampleRate <= 0 {
		sampleRate = rhythmDefaultSampleRate
	}

	pcm := synthRhythmPCM(sampleRate)

	return wrapWAV(pcm, sampleRate)
}

// synthRhythmPCM 按节奏序列生成单声道 16-bit PCM 小端样本。
func synthRhythmPCM(sampleRate int) []byte {
	gap := sampleRate * rhythmGapMillis / 1000
	var samples []int16
	for i, dash := range rhythmPattern {
		ms := rhythmDotMillis
		if dash {
			ms = rhythmDashMillis
		}
		samples = appendTone(samples, sampleRate, ms)
		if i < len(rhythmPattern)-1 {
			samples = append(samples, make([]int16, gap)...)
		}
	}

	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
	}

	return buf
}

// appendTone 追加一段 millis 毫秒的正弦蜂鸣样本，首尾各做一小段线性淡入淡出消除爆音。
func appendTone(dst []int16, sampleRate, millis int) []int16 {
	n := sampleRate * millis / 1000
	if n == 0 {
		return dst
	}
	fade := min(n/8, sampleRate/200+1)
	for i := range n {
		v := math.Sin(2 * math.Pi * rhythmFreq * float64(i) / float64(sampleRate))
		gain := rhythmAmplitude
		switch {
		case i < fade:
			gain *= float64(i) / float64(fade)
		case i >= n-fade:
			gain *= float64(n-1-i) / float64(fade)
		}
		dst = append(dst, int16(v*gain*math.MaxInt16))
	}

	return dst
}

// wrapWAV 把单声道 16-bit PCM 字节封装为标准 WAV（RIFF/WAVE）。
func wrapWAV(pcm []byte, sampleRate int) []byte {
	const (
		numChannels   = 1
		bitsPerSample = 16
	)
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8

	buf := make([]byte, 0, 44+len(pcm))
	buf = append(buf, "RIFF"...)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(36+len(pcm)))
	buf = append(buf, "WAVE"...)
	buf = append(buf, "fmt "...)
	buf = binary.LittleEndian.AppendUint32(buf, 16) // fmt chunk 长度
	buf = binary.LittleEndian.AppendUint16(buf, 1)  // 音频格式：1=PCM
	buf = binary.LittleEndian.AppendUint16(buf, numChannels)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(sampleRate))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(byteRate))
	buf = binary.LittleEndian.AppendUint16(buf, uint16(blockAlign))
	buf = binary.LittleEndian.AppendUint16(buf, bitsPerSample)
	buf = append(buf, "data"...)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(pcm)))
	buf = append(buf, pcm...)

	return buf
}
