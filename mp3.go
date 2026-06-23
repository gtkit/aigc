package aigc

// mp3MetadataDesc 是 ID3 TXXX 帧承载 AIGC 标识时的 description。
const mp3MetadataDesc = "AIGC"

// mp3WriteMetadata 在 mp3 前置 ID3v2.4 标签，用 TXXX 帧（description=AIGC）承载隐式标识。
// 入参应为不含 ID3 标签的裸 mp3（百度流式输出即裸帧）；已含标签则返回 ErrAlreadyLabeled。
func mp3WriteMetadata(audio, value []byte) ([]byte, error) {
	if len(audio) >= 3 && string(audio[0:3]) == "ID3" {
		return nil, ErrAlreadyLabeled
	}
	tag := id3v24Tag(mp3MetadataDesc, value)
	out := make([]byte, len(tag)+len(audio))
	copy(out, tag)
	copy(out[len(tag):], audio)

	return out, nil
}

// mp3PrependCue 把 cue 的 mp3 帧字节拼接到 audio 起始/末尾。
// 要求 cue 与 audio 为同编码参数的裸 mp3 帧；拼接点可能有极轻微瑕疵（bit reservoir）。
func mp3PrependCue(audio, cue []byte, pos Position) ([]byte, error) {
	out := make([]byte, 0, len(audio)+len(cue))
	if pos == AtEnd {
		out = append(out, audio...)
		out = append(out, cue...)
	} else {
		out = append(out, cue...)
		out = append(out, audio...)
	}

	return out, nil
}

// synchsafe 把 ≤ 2^28-1 的整数编码为 4 字节 synchsafe 整数（每字节仅用低 7 位）。
func synchsafe(n int) [4]byte {
	return [4]byte{
		byte((n >> 21) & 0x7f),
		byte((n >> 14) & 0x7f),
		byte((n >> 7) & 0x7f),
		byte(n & 0x7f),
	}
}

// id3v24Tag 用单个 TXXX 帧构造完整 ID3v2.4 标签字节。
// 帧体：encoding(0x03=UTF-8) + description + 0x00 + value。
func id3v24Tag(desc string, value []byte) []byte {
	body := make([]byte, 0, 1+len(desc)+1+len(value))
	body = append(body, 0x03)
	body = append(body, desc...)
	body = append(body, 0x00)
	body = append(body, value...)

	fsize := synchsafe(len(body))
	frame := make([]byte, 0, 10+len(body))
	frame = append(frame, 'T', 'X', 'X', 'X')
	frame = append(frame, fsize[0], fsize[1], fsize[2], fsize[3])
	frame = append(frame, 0x00, 0x00)
	frame = append(frame, body...)

	tsize := synchsafe(len(frame))
	tag := make([]byte, 0, 10+len(frame))
	tag = append(tag, 'I', 'D', '3')
	tag = append(tag, 0x04, 0x00)
	tag = append(tag, 0x00)
	tag = append(tag, tsize[0], tsize[1], tsize[2], tsize[3])
	tag = append(tag, frame...)

	return tag
}
