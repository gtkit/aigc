package aigc

import "strings"

// LabelingConfig 是 AIGC 标识的运行时配置。
// MarkerMP3 由上层（runtime 接线层）按配置路径预加载，空表示不注入显式标识。
type LabelingConfig struct {
	Enabled           bool
	Label             string // 1=是 / 2=可能是 / 3=疑似；空则默认按 LabelIs 处理
	ContentProducer   string
	ContentPropagator string
	ReserveCode1      string
	ReserveCode2      string
	ArchiveDir        string // 隐式标识归档目录；空则不落盘
	MarkerMP3         []byte // 显式标识素材（mp3 裸帧），空则不前置注入
}

// Active 判断隐式标识是否应启用：开关打开且生成服务提供者非空。
func (c LabelingConfig) Active() bool {
	return c.Enabled && strings.TrimSpace(c.ContentProducer) != ""
}

// HasMarker 判断是否有可注入的显式标识素材。
func (c LabelingConfig) HasMarker() bool {
	return c.Enabled && len(c.MarkerMP3) > 0
}

// NewIdentifier 用配置构造一次合成的 AIGC 标识；produceID 为本次内容生成编号。
func (c LabelingConfig) NewIdentifier(produceID string) Identifier {
	label := strings.TrimSpace(c.Label)
	if label == "" {
		label = LabelIs
	}

	return Identifier{
		Label:             label,
		ContentProducer:   c.ContentProducer,
		ProduceID:         produceID,
		ContentPropagator: c.ContentPropagator,
		ReserveCode1:      c.ReserveCode1,
		ReserveCode2:      c.ReserveCode2,
	}
}
