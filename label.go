package aigc

import (
	"errors"

	"github.com/gtkit/json/v2"
)

// Format 标识音频容器格式。
type Format int

const (
	// WAV RIFF/WAVE 容器（百度流式默认输出）。
	WAV Format = iota
	// MP3 mp3 文件。
	MP3
)

// Position 指定显式标识提示音拼接到正文的位置。
type Position int

const (
	// AtStart 拼接到正文起始位置。
	AtStart Position = iota
	// AtEnd 拼接到正文末尾。
	AtEnd
)

// ErrUnsupportedFormat 表示传入了本包不支持的音频格式。
var ErrUnsupportedFormat = errors.New("aigc: 不支持的音频格式")

// jsonValue 返回写入元数据的 AIGC 值：七要素紧凑 JSON（对齐 GB 45438-2025 附录 E）。
func (id Identifier) jsonValue() ([]byte, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}

	return json.Marshal(id)
}

// WriteMetadata 给完整音频写入 AIGC 隐式标识（文件元数据字段块），返回带标识的新字节。
// 纯 Go 实现，内存操作，不依赖 ffmpeg、不落临时文件。入参须为完整音频（流式须先拼接完整）。
func WriteMetadata(audio []byte, format Format, id Identifier) ([]byte, error) {
	value, err := id.jsonValue()
	if err != nil {
		return nil, err
	}
	switch format {
	case WAV:
		return wavWriteMetadata(audio, value)
	case MP3:
		return mp3WriteMetadata(audio, value)
	default:
		return nil, ErrUnsupportedFormat
	}
}

// PrependCue 给完整音频拼接显式标识提示音 cue（摩斯码或预录语音），返回拼接后的新字节。
// pos 指定拼到起始或末尾。cue 与 audio 须为同一格式、同编码参数。
func PrependCue(audio, cue []byte, format Format, pos Position) ([]byte, error) {
	switch format {
	case WAV:
		return wavPrependCue(audio, cue, pos)
	case MP3:
		return mp3PrependCue(audio, cue, pos)
	default:
		return nil, ErrUnsupportedFormat
	}
}
