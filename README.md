# aigc

`github.com/gtkit/aigc` —— AI 生成合成内容的**文件元数据隐式标识**构造与写入。

依据《人工智能生成合成内容标识办法》及强制性国标 **GB 45438-2025**，把 AIGC 标识写入文件元数据。当前仅实现 mp3 音频：将标识 JSON 写入 ID3v2.4 的 `TXXX` 帧（`description=AIGC`），标签前置于音频帧。

**合规依据**：字段名与值结构已对齐 GB 45438-2025 附录 E（规范性）——隐式标识扩展字段名含 `AIGC`，其值为字符串 `{"AIGC":{"Label","ContentProducer","ProduceID","ReservedCode1","ContentPropagator","PropagateID","ReservedCode2"}}`。

> ⚠ **仍待确认**：附录 F.2「音频文件元数据隐式标识示例」以图示给出，未明文规定 mp3 的承载方式；本包采用 ID3v2.4 TXXX（mp3 元数据事实标准），满足附录 E a) 对字段命名的要求。附录 E j) 对各要素值的字符集约束（限 GB18030 指定的可见 ASCII 码位）本包未强制校验，由调用方保证。

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
| `WriteMP3(mp3, id)` | 在裸 mp3 帧前置写入承载标识的 ID3v2.4 标签；已含标识则拒绝 |
| `ReadMP3(mp3)` | 读回本包写入的 AIGC 标识，返回 `(Identifier, found, error)` |
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
| `ErrAlreadyLabeled` | `WriteMP3` 入参已含本包写入的 AIGC 标识 |
| `ErrCorruptTag` | mp3 以 ID3 开头但标签结构损坏（长度越界） |

均为包级 `error` 值，可用 `errors.Is` 判定。

## 读取与重复打标防护

```go
id, found, err := aigc.ReadMP3(data)
// found=true 表示已含本包写入的标识，可据此校验或避免重复打标
```

`WriteMP3` 已内建防护：入参若已含本包标识，直接返回 `ErrAlreadyLabeled`，不会叠加第二层标签。`ReadMP3` 仅识别本包 `WriteMP3` 写入的格式（ID3v2.4、标志位为 0、`description=AIGC` 的 TXXX 帧），其他情形返回 `found=false`。

## 输入约定

- `WriteMP3` 的入参 `mp3` 应为**不含 ID3 标签的裸 mp3 帧流**（百度流式 aue=3 的输出即是裸帧）。
- ID3v2.4 帧长用 synchsafe 编码，单帧承载上限 `2^28-1` 字节，标识 JSON 远小于此。

## License

随仓库 LICENSE。
