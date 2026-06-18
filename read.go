package aigc

import (
	"bytes"
	"errors"

	"github.com/gtkit/json/v2"
)

// ErrCorruptTag 表示 mp3 以 ID3 标签开头，但标签结构损坏（声明长度越界）。
var ErrCorruptTag = errors.New("aigc: ID3 标签结构损坏")

// readSynchsafe 解码 4 字节 synchsafe 整数（synchsafe 的逆运算）。
func readSynchsafe(b []byte) int {
	return int(b[0])<<21 | int(b[1])<<14 | int(b[2])<<7 | int(b[3])
}

// ReadMP3 从 mp3 字节读取本包 WriteMP3 写入的 AIGC 隐式标识。
//
// 仅识别本包写入的格式：ID3v2.4、标志位为 0、承载 description=AIGC 的 TXXX 帧。
// 返回 found=false 且 err=nil 表示「不含本包标识」——包括不以 ID3 开头（如裸帧）、
// 非 2.4 版本、含本包不写入的标志位、标签内无 AIGC 帧等情形。
// 仅当标签结构损坏（声明长度越界）或 AIGC 值无法反序列化时返回 err。
func ReadMP3(mp3 []byte) (id Identifier, found bool, err error) {
	const headerLen = 10
	if len(mp3) < headerLen || string(mp3[:3]) != "ID3" {
		return Identifier{}, false, nil
	}
	if mp3[3] != 0x04 || mp3[5] != 0x00 { // 仅 2.4 版本、标志位为 0
		return Identifier{}, false, nil
	}

	tagSize := readSynchsafe(mp3[6:headerLen])
	if tagSize < 0 || headerLen+tagSize > len(mp3) {
		return Identifier{}, false, ErrCorruptTag
	}
	body := mp3[headerLen : headerLen+tagSize]

	for len(body) >= headerLen {
		if body[0] == 0x00 { // 进入 padding 区，结束遍历
			break
		}
		frameID := string(body[:4])
		frameSize := readSynchsafe(body[4:8])
		if frameSize < 0 || headerLen+frameSize > len(body) {
			return Identifier{}, false, ErrCorruptTag
		}
		frameBody := body[headerLen : headerLen+frameSize]

		if frameID == "TXXX" {
			if value, ok := aigcValue(frameBody); ok {
				var env envelope
				if e := json.Unmarshal(value, &env); e != nil {
					return Identifier{}, false, e
				}

				return env.AIGC, true, nil
			}
		}
		body = body[headerLen+frameSize:]
	}

	return Identifier{}, false, nil
}

// aigcValue 从 TXXX 帧体中提取 description=AIGC 时的 value 字节。
// 帧体结构：encoding(1) + description + 0x00 + value，仅支持 UTF-8(0x03)。
func aigcValue(frameBody []byte) ([]byte, bool) {
	if len(frameBody) < 2 || frameBody[0] != 0x03 {
		return nil, false
	}
	rest := frameBody[1:]
	nul := bytes.IndexByte(rest, 0x00)
	if nul < 0 || string(rest[:nul]) != metadataKey {
		return nil, false
	}

	return rest[nul+1:], true
}
