# aigc

`github.com/gtkit/aigc` —— AI 生成合成内容的**文件元数据隐式标识**构造与写入。

依据《人工智能生成合成内容标识办法》及强制性国标 **GB 45438-2025**，把 AIGC 标识写入文件元数据。当前仅实现 mp3 音频：将标识 JSON 写入 ID3v2.4 的 `TXXX` 帧（`description=AIGC`），标签前置于音频帧。

> ⚠ **合规未决项**：标识字段键名与音频落点以 GB 45438-2025 附录 E 为准，本包当前实现**尚未与官方 PDF 逐字核对**（详见 `metadata.go` 顶部注释）。在用于正式合规场景前，请先核对字段名（`Label`/`ContentProducer`/`ProduceID`/`ContentPropagator`/`PropagateID`/`ReserveCode1`/`ReserveCode2`）与落点是否符合标准原文。字段若调整属破坏性变更。

## 安装

```bash
go get github.com/gtkit/aigc
```

## 快速上手

```go
import "github.com/gtkit/aigc"

// 1. 构造标识
id := aigc.Identifier{
    Label:           aigc.LabelIs, // 1=是 / 2=可能是 / 3=疑似
    ContentProducer: "PRODUCER-001",
    ProduceID:       "20260618-0001",
}

// 2. 写入裸 mp3 帧（百度流式 aue=3 的输出即是裸帧）
labeled, err := aigc.WriteMP3(rawMP3, id)
if err != nil {
    return err
}
// labeled 以 ID3v2.4 标签开头，原始音频帧原样跟随其后
```

也可用运行时配置批量构造标识：

```go
cfg := aigc.LabelingConfig{
    Enabled:         true,
    ContentProducer: "PRODUCER-001",
}
if cfg.Active() {
    id := cfg.NewIdentifier(produceID) // Label 留空时默认按 LabelIs
    labeled, _ := aigc.WriteMP3(rawMP3, id)
    _ = labeled
}
```

更多可运行示例见 `example_test.go`。

## API 概览

| 符号 | 说明 |
|------|------|
| `Identifier` | 文件元数据隐式标识五要素（含两个预留码） |
| `Identifier.Validate()` | 最小必填校验：`Label` 须为 1/2/3，`ContentProducer` 非空 |
| `Identifier.JSON()` | 返回写入元数据的紧凑 JSON 字符串 |
| `WriteMP3(mp3, id)` | 在裸 mp3 帧前置写入承载标识的 ID3v2.4 标签 |
| `LabelingConfig` | 标识的运行时配置 |
| `LabelingConfig.Active()` | 隐式标识是否应启用（开关开 + 生成者非空） |
| `LabelingConfig.HasMarker()` | 是否有可注入的显式标识素材 |
| `LabelingConfig.NewIdentifier(produceID)` | 用配置构造一次标识 |

常量：`LabelIs`(="1") / `LabelMaybe`(="2") / `LabelSuspected`(="3")。

## 错误处理

| 错误 | 触发条件 |
|------|----------|
| `ErrInvalidLabel` | `Label` 取值不在 1/2/3 之内 |
| `ErrProducerRequired` | `ContentProducer` 为空 |
| `ErrEmptyValue` | 待写入的 AIGC 值为空 |

均为包级 `error` 值，可用 `errors.Is` 判定。

## 输入约定

- `WriteMP3` 的入参 `mp3` 应为**不含 ID3 标签的裸 mp3 帧流**；若已带 ID3 标签会被前置叠加，需调用方自行保证。
- ID3v2.4 帧长用 synchsafe 编码，单帧承载上限 `2^28-1` 字节，标识 JSON 远小于此。

## License

随仓库 LICENSE。
