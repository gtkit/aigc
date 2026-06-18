// Package aigc 实现《人工智能生成合成内容标识办法》及强制性国标 GB 45438-2025
// 要求的「文件元数据隐式标识」构造与写入。
//
// 当前仅实现 mp3 音频：把 AIGC 标识 JSON 写入 ID3v2.4 的 TXXX 帧（description=AIGC）。
//
// 字段名与值结构依据 GB 45438-2025 附录 E（规范性）：隐式标识扩展字段的名称或关键词中
// 应含 "AIGC"，其值为字符串 {"AIGC":{"Label",...,"ReservedCode2"}}。本包以 ID3v2.4 的
// TXXX 帧（description=AIGC）承载该值，满足附录 E a) 对字段命名的要求。
//
// 注：附录 F.2 的「音频文件元数据隐式标识示例」以图示给出，未明文规定 mp3 必须使用 ID3；
// ID3v2 TXXX 是 mp3 元数据的事实标准，本包据此实现。
package aigc

import (
	"errors"
	"strings"

	"github.com/gtkit/json/v2"
)

// Label 取值：生成合成标识属性。依据 GB 45438-2025（待 PDF 核对）。
const (
	LabelIs        = "1" // 是 AI 生成合成
	LabelMaybe     = "2" // 可能是
	LabelSuspected = "3" // 疑似
)

// metadataKey 是写入文件元数据时使用的扩展字段键名（ID3 TXXX 的 description）。
const metadataKey = "AIGC"

// Identifier 是文件元数据隐式标识的七要素，字段名与顺序对齐 GB 45438-2025 附录 E（规范性）。
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

// Validate 做最小必填校验。生成服务提供者与 Label 是隐式标识的核心要素。
func (id Identifier) Validate() error {
	switch id.Label {
	case LabelIs, LabelMaybe, LabelSuspected:
	default:
		return ErrInvalidLabel
	}
	if strings.TrimSpace(id.ContentProducer) == "" {
		return ErrProducerRequired
	}

	return nil
}

// envelope 是写入元数据的外层结构。附录 E b) 要求扩展字段的值为 {"AIGC":{...}} 字符串。
type envelope struct {
	AIGC Identifier `json:"AIGC"`
}

// JSON 返回写入元数据的标识值（紧凑 JSON 字符串），格式为 {"AIGC":{...}}，对齐附录 E b)。
func (id Identifier) JSON() (string, error) {
	if err := id.Validate(); err != nil {
		return "", err
	}
	b, err := json.Marshal(envelope{AIGC: id})
	if err != nil {
		return "", err
	}

	return string(b), nil
}
