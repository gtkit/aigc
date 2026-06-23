package aigc

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// ErrInvalidWAV 表示数据不是合法的 WAV(RIFF/WAVE) 容器。
var ErrInvalidWAV = errors.New("aigc: 不是合法的 WAV(RIFF/WAVE) 数据")

// ErrWAVParamMismatch 表示拼接的两段 WAV 的 fmt 参数（采样率/声道/位深）不一致。
var ErrWAVParamMismatch = errors.New("aigc: 拼接的两段 WAV 格式参数不一致")

// wavMetadataChunkID 是承载 AIGC 隐式标识的自定义 chunk 标识。
const wavMetadataChunkID = "AIGC"

// wavWriteMetadata 在 WAV 的 RIFF 容器末尾追加 AIGC 隐式标识。
// 降险写两处：自定义 "AIGC" chunk + LIST/INFO 的 AIGC 子块，兼容不同检测工具。
func wavWriteMetadata(audio, value []byte) ([]byte, error) {
	if !isRIFFWAVE(audio) {
		return nil, ErrInvalidWAV
	}
	extra := buildChunk(wavMetadataChunkID, value)
	extra = append(extra, buildListInfo(wavMetadataChunkID, value)...)

	out := make([]byte, len(audio)+len(extra))
	copy(out, audio)
	copy(out[len(audio):], extra)
	riffSize := binary.LittleEndian.Uint32(out[4:8])
	binary.LittleEndian.PutUint32(out[4:8], riffSize+uint32(len(extra)))

	return out, nil
}

// wavPrependCue 把 cue 的采样数据拼接到 audio 的 data 起始/末尾，要求两段 fmt 一致。
func wavPrependCue(audio, cue []byte, pos Position) ([]byte, error) {
	if !isRIFFWAVE(audio) || !isRIFFWAVE(cue) {
		return nil, ErrInvalidWAV
	}
	aFmt, aData, err := wavFmtData(audio)
	if err != nil {
		return nil, err
	}
	cFmt, cData, err := wavFmtData(cue)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(aFmt, cFmt) {
		return nil, ErrWAVParamMismatch
	}

	merged := make([]byte, 0, len(aData)+len(cData))
	if pos == AtEnd {
		merged = append(append(merged, aData...), cData...)
	} else {
		merged = append(append(merged, cData...), aData...)
	}

	return buildWAV(aFmt, merged), nil
}

// isRIFFWAVE 报告 b 是否以 RIFF/WAVE 头开头。
func isRIFFWAVE(b []byte) bool {
	return len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WAVE"
}

// wavFmtData 遍历 RIFF chunks，返回 fmt 块内容与 data 块采样数据。
func wavFmtData(b []byte) (fmtBody, dataBody []byte, err error) {
	off := 12
	for off+8 <= len(b) {
		id := string(b[off : off+4])
		size := int(binary.LittleEndian.Uint32(b[off+4 : off+8]))
		start := off + 8
		if size < 0 || start+size > len(b) {
			return nil, nil, ErrInvalidWAV
		}
		switch id {
		case "fmt ":
			fmtBody = b[start : start+size]
		case "data":
			dataBody = b[start : start+size]
		}
		off = start + size
		if size%2 == 1 {
			off++ // chunk 按偶数字节对齐
		}
	}
	if fmtBody == nil || dataBody == nil {
		return nil, nil, ErrInvalidWAV
	}

	return fmtBody, dataBody, nil
}

// buildChunk 构造一个 RIFF chunk：id(4) + 小端长度(4) + data + 奇数长度补齐字节。
func buildChunk(id string, data []byte) []byte {
	out := make([]byte, 0, 8+len(data)+1)
	out = append(out, id...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(data)))
	out = append(out, data...)
	if len(data)%2 == 1 {
		out = append(out, 0x00)
	}

	return out
}

// buildListInfo 构造 LIST/INFO chunk，内含一个 subID 子块承载 value。
func buildListInfo(subID string, value []byte) []byte {
	sub := buildChunk(subID, value)
	body := make([]byte, 0, 4+len(sub))
	body = append(body, "INFO"...)
	body = append(body, sub...)

	return buildChunk("LIST", body)
}

// buildWAV 用 fmt 块与采样数据重建一个最小 WAV。
func buildWAV(fmtBody, dataBody []byte) []byte {
	body := make([]byte, 0, 4+8+len(fmtBody)+8+len(dataBody)+2)
	body = append(body, "WAVE"...)
	body = append(body, buildChunk("fmt ", fmtBody)...)
	body = append(body, buildChunk("data", dataBody)...)

	out := make([]byte, 0, 8+len(body))
	out = append(out, "RIFF"...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(body)))
	out = append(out, body...)

	return out
}
