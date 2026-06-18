package aigc

import "errors"

// 本文件实现最小化的 ID3v2.4 标签写入：仅写一个 TXXX 帧承载 AIGC 标识。
// 选 v2.4 是因为它原生支持 UTF-8（encoding=0x03），帧长用 synchsafe 编码。
// 标签写在 mp3 文件最前面，符合 ID3v2 规范（标签前置于音频帧）。

// ErrEmptyValue 表示要写入的 AIGC 值为空。
var ErrEmptyValue = errors.New("aigc: 写入的 AIGC 值为空")

// ErrAlreadyLabeled 表示入参 mp3 已含本包写入的 AIGC 标识，拒绝重复打标。
var ErrAlreadyLabeled = errors.New("aigc: mp3 已含 AIGC 标识")

// synchsafe 把一个 ≤ 2^28-1 的整数编码为 4 字节 synchsafe 整数（每字节高位恒为 0，只用低 7 位）。
func synchsafe(n int) [4]byte {
	return [4]byte{
		byte((n >> 21) & 0x7f),
		byte((n >> 14) & 0x7f),
		byte((n >> 7) & 0x7f),
		byte(n & 0x7f),
	}
}

// buildTXXX 构造一个 ID3v2.4 的 TXXX 帧（不含标签头）。
// 帧内容：encoding(0x03=UTF-8) + description + 0x00 + value。
func buildTXXX(description, value string) []byte {
	body := make([]byte, 0, 1+len(description)+1+len(value))
	body = append(body, 0x03) // UTF-8
	body = append(body, description...)
	body = append(body, 0x00) // description 结束符
	body = append(body, value...)

	size := synchsafe(len(body))
	frame := make([]byte, 0, 10+len(body))
	frame = append(frame, 'T', 'X', 'X', 'X') // 帧 ID
	frame = append(frame, size[0], size[1], size[2], size[3])
	frame = append(frame, 0x00, 0x00) // 帧标志
	frame = append(frame, body...)

	return frame
}

// id3v24Tag 用单个 TXXX 帧构造完整 ID3v2.4 标签字节。
func id3v24Tag(description, value string) []byte {
	frame := buildTXXX(description, value)

	size := synchsafe(len(frame))
	tag := make([]byte, 0, 10+len(frame))
	tag = append(tag, 'I', 'D', '3') // 标识
	tag = append(tag, 0x04, 0x00)    // 版本 2.4.0
	tag = append(tag, 0x00)          // 标志
	tag = append(tag, size[0], size[1], size[2], size[3])
	tag = append(tag, frame...)

	return tag
}

// WriteMP3 在 mp3 音频字节前置写入承载 AIGC 标识的 ID3v2.4 标签，返回带隐式标识的完整 mp3。
// 入参 mp3 应为不含 ID3 标签的裸 mp3 帧流（百度流式 aue=3 的输出即是裸帧）。
func WriteMP3(mp3 []byte, id Identifier) ([]byte, error) {
	if _, found, rerr := ReadMP3(mp3); rerr != nil {
		return nil, rerr
	} else if found {
		return nil, ErrAlreadyLabeled
	}

	value, err := id.JSON()
	if err != nil {
		return nil, err
	}
	if value == "" {
		return nil, ErrEmptyValue
	}

	tag := id3v24Tag(metadataKey, value)
	out := make([]byte, 0, len(tag)+len(mp3))
	out = append(out, tag...)
	out = append(out, mp3...)

	return out, nil
}
