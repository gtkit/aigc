// Package aigc 实现《人工智能生成合成内容标识办法》及强制性国标 GB 45438-2025
// 要求的 AIGC 标识：对百度流式合成的音频（mp3 会员 / pcm 免费）打显式标识与隐式标识。
//
// 通过 Label 统一入口封装 ffmpeg：可选把显式标识提示音（预录提示语音或 RhythmWAV 生成的
// 摩斯码）拼接到正文起始/末尾，统一编码输出 mp3，并把 Identifier 七要素写入文件元数据的
// AIGC 键（隐式标识）。
//
// 字段名与值结构依据 GB 45438-2025 附录 E（规范性）；各要素值的字符集按附录 E j) 强制校验。
package aigc

import (
	"errors"
	"strings"
)

// Label 取值：生成合成标识属性。依据 GB 45438-2025。
const (
	LabelIs        = "1" // 是 AI 生成合成
	LabelMaybe     = "2" // 可能是
	LabelSuspected = "3" // 疑似
)

// Identifier 是文件元数据隐式标识的七要素，字段名与顺序对齐 GB 45438-2025 附录 E（规范性）。
// 写入元数据时以 json tag 的键名序列化为紧凑 JSON，作为文件元数据 AIGC 键的值。
type Identifier struct {
	// Label 生成合成标识：1=是 / 2=可能是 / 3=疑似。
	Label string `json:"Label"`
	// ContentProducer 生成合成服务提供者名称或编码。
	ContentProducer string `json:"ContentProducer"`
	// ProduceID 内容制作编号（同一生成合成服务提供者内唯一）。
	ProduceID string `json:"ProduceID,omitempty"`
	// ReservedCode1 预留字段 1，可存储生成合成服务提供者的安全防护信息。
	ReservedCode1 string `json:"ReservedCode1,omitempty"`
	// ContentPropagator 内容传播服务提供者名称或编码。
	ContentPropagator string `json:"ContentPropagator,omitempty"`
	// PropagateID 内容传播编号。
	PropagateID string `json:"PropagateID,omitempty"`
	// ReservedCode2 预留字段 2，可存储内容传播服务提供者的安全防护信息。
	ReservedCode2 string `json:"ReservedCode2,omitempty"`
}

// ErrInvalidLabel 表示 Label 取值不在 1/2/3 之内。
var ErrInvalidLabel = errors.New("aigc: Label 必须为 1/2/3")

// ErrProducerRequired 表示缺少生成服务提供者标识。
var ErrProducerRequired = errors.New("aigc: ContentProducer 不能为空")

// ErrInvalidCharset 表示要素值含 GB 45438-2025 附录 E j) 允许范围
// （GB18030 指定的可见 ASCII 码位 0x20–0x7E）之外的字符。
var ErrInvalidCharset = errors.New("aigc: 要素值含 0x20-0x7E 之外的字符")

// Validate 做必填校验与附录 E j) 字符集校验。
// 生成服务提供者与 Label 是隐式标识的核心要素；各要素值仅允许 GB18030 指定的可见 ASCII（0x20–0x7E）。
func (id Identifier) Validate() error {
	switch id.Label {
	case LabelIs, LabelMaybe, LabelSuspected:
	default:
		return ErrInvalidLabel
	}
	if strings.TrimSpace(id.ContentProducer) == "" {
		return ErrProducerRequired
	}
	for _, v := range []string{
		id.ContentProducer, id.ProduceID, id.ReservedCode1,
		id.ContentPropagator, id.PropagateID, id.ReservedCode2,
	} {
		if !isVisibleASCII(v) {
			return ErrInvalidCharset
		}
	}

	return nil
}

// isVisibleASCII 报告 s 是否仅含 GB18030 指定的可见 ASCII 码位（0x20–0x7E）。
// 按字节校验：多字节 UTF-8（如中文）必有字节 > 0x7E，会被判为非法。
func isVisibleASCII(s string) bool {
	for i := range len(s) {
		if s[i] < 0x20 || s[i] > 0x7e {
			return false
		}
	}

	return true
}
