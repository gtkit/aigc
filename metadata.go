// Package aigc 实现《人工智能生成合成内容标识办法》及强制性国标 GB 45438-2025
// 要求的「文件元数据隐式标识」构造与写入。
//
// 当前仅实现 mp3 音频：把 AIGC 标识 JSON 写入 ID3v2.4 的 TXXX 帧（description=AIGC）。
//
// 重要合规说明（落地前必须以官方 PDF 核对）：
//   - 字段键名以 GB 45438-2025 附录 E 为准。本文件采用
//     Label/ContentProducer/ProduceID/ContentPropagator/PropagateID/ReserveCode1/ReserveCode2，
//     但 vivo 线上实现里出现过 Propagator/PropatorID 等不同写法，二者需以标准原文校正。
//   - mp3 用 ID3v2 TXXX 承载 AIGC 字段，是否完全符合附录 E 对「音频文件元数据」落点的要求，
//     需以官方《文件元数据隐式标识 音频文件》实践指南核对后确认。
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

// Identifier 是文件元数据隐式标识的五要素（含两个预留码）。
// 字段顺序与 JSON 键名以标准附录 E 为准；这里用普通结构体便于核对后调整。
type Identifier struct {
	// Label 生成合成标识：1=是 / 2=可能是 / 3=疑似。
	Label string `json:"Label"`
	// ContentProducer 生成服务提供者名称或编码。
	ContentProducer string `json:"ContentProducer"`
	// ProduceID 内容生成编号（同一服务提供者内唯一）。
	ProduceID string `json:"ProduceID,omitempty"`
	// ContentPropagator 内容传播服务提供者名称或编码。
	ContentPropagator string `json:"ContentPropagator,omitempty"`
	// PropagateID 内容传播编号。
	PropagateID string `json:"PropagateID,omitempty"`
	// ReserveCode1 预留字段 1。
	ReserveCode1 string `json:"ReserveCode1,omitempty"`
	// ReserveCode2 预留字段 2。
	ReserveCode2 string `json:"ReserveCode2,omitempty"`
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

// JSON 返回写入元数据的 AIGC 值（紧凑 JSON 字符串）。
func (id Identifier) JSON() (string, error) {
	if err := id.Validate(); err != nil {
		return "", err
	}
	b, err := json.Marshal(id)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
