# aigc

`github.com/gtkit/aigc` —— 给 AI 生成合成音频打 **AIGC 标识**的**纯 Go**工具（零外部进程依赖）：对百度流式合成的 **wav / mp3** 文件，加**显式标识**（起始/末尾拼提示音或摩斯码）与**隐式标识**（写文件元数据字段块）。

依据《人工智能生成合成内容标识办法》及强制性国标 **GB 45438-2025**。

## 特点

- **纯 Go、不依赖 ffmpeg**：内存 `[]byte` 进出，不落临时文件，部署无外部依赖。
- **wav + mp3 双格式**：
  - 隐式标识：wav 写 RIFF 的 `AIGC` chunk（同时写 `LIST/INFO` 降险）；mp3 写 ID3v2.4 TXXX。
  - 显式标识：wav 采样层无损拼接；mp3 同参数帧拼接。
- 字段结构对齐 GB 45438-2025 附录 E，要素值按附录 E j) 强制字符集校验。
- 重复打标保护：已打标的 wav/mp3 再次写入返回 `ErrAlreadyLabeled`。

> 不做格式转换（pcm→mp3 那种需要编码器，本包不涉及）。百度流式裸流请先按序拼接成完整 wav/mp3 再调用。

## 安装

需 Go 1.26+：

```bash
go get github.com/gtkit/aigc@v1.1.0
```

## 快速上手

```go
import "github.com/gtkit/aigc"

id := aigc.Identifier{
    Label:           aigc.LabelIs, // 1=是 / 2=可能是 / 3=疑似
    ContentProducer: "PRODUCER-001",
    ProduceID:       "20260625-0001",
}

// 端到端一步完成：显式标识(拼提示音) + 隐式标识(写字段块)
cue := aigc.RhythmWAV(16000) // 摩斯码提示音；须与正文同采样率/声道
final, _ := aigc.Label(wavAudio, cue, aigc.WAV, aigc.AtStart, id)

// 也可分两步（更灵活）：
withCue, _ := aigc.PrependCue(wavAudio, cue, aigc.WAV, aigc.AtStart) // 仅显式
onlyMeta, _ := aigc.WriteMetadata(wavAudio, aigc.WAV, id)            // 仅隐式
_ = withCue
_ = onlyMeta
```

`Label` 的 `cue` 传 `nil` 即只做隐式标识。

## 端到端示例（百度流式对接）

```go
// 已从百度流式接口按到达顺序拼接得到完整音频字节 audio（wav 或 mp3）。
// mp3Cue 为预置的、与正文同编码参数的 mp3 提示音素材；wav 用 RhythmWAV 现合成。
func labelBaiduAudio(audio, mp3Cue []byte, format aigc.Format) ([]byte, error) {
    id := aigc.Identifier{
        Label:           aigc.LabelIs,
        ContentProducer: "你的机构编码", // 对齐 GB 45438-2025 附录 E
        ProduceID:       "20260625-0001",
    }

    var cue []byte
    switch format {
    case aigc.WAV:
        cue = aigc.RhythmWAV(16000) // 须与正文同采样率(16k)、单声道
    case aigc.MP3:
        cue = mp3Cue
    default:
        return nil, aigc.ErrUnsupportedFormat
    }

    // 一步完成：起始拼摩斯提示音(显式) + 写 AIGC 字段块(隐式)
    return aigc.Label(audio, cue, format, aigc.AtStart, id)
}
```

## 工作流程

```
百度流式分块 → 内存按序拼成完整 wav/mp3 []byte
  → Label（或 PrependCue + WriteMetadata）
      · 显式标识：拼提示音/摩斯码到起始或末尾（可选）
      · 隐式标识：写 AIGC 字段块
  → 带标识的 []byte（落盘或回传）
```

> 隐式标识须基于**完整音频 + 容器**写入，流式须先拼完整；但纯 Go 在内存完成，无需临时文件。

## API 概览

| 符号 | 说明 |
|------|------|
| `Label(audio, cue, format, pos, id)` | 端到端：显式(cue 非空时)+ 隐式，一步完成 |
| `WriteMetadata(audio, format, id)` | 隐式标识：写 AIGC 字段块（wav→RIFF chunk / mp3→ID3 TXXX） |
| `PrependCue(audio, cue, format, pos)` | 显式标识：把提示音拼到起始/末尾 |
| `RhythmWAV(sampleRate)` | 合成「短长短短」摩斯节奏提示音（单声道 16-bit PCM WAV 素材） |
| `Identifier` | 隐式标识七要素（对齐 GB 45438-2025 附录 E） |
| `Identifier.Validate()` | 必填校验 + 附录 E j) 字符集校验 |
| `Format`：`WAV` / `MP3` | 音频容器格式 |
| `Position`：`AtStart` / `AtEnd` | 显式标识拼接位置 |

常量：`LabelIs`(="1") / `LabelMaybe`(="2") / `LabelSuspected`(="3")。

## 源码结构

| 文件 | 职责 |
|------|------|
| `label.go` | 公共入口：`Label`/`WriteMetadata`/`PrependCue` + `Format`/`Position` + 通用错误（`ErrUnsupportedFormat`/`ErrAlreadyLabeled`） |
| `metadata.go` | `Identifier` 七要素 + `Validate`（必填 + 附录 E j) 字符集校验）+ `Label*` 常量 |
| `wav.go` | WAV(RIFF) 实现：隐式（`AIGC` chunk + `LIST/INFO`）、显式（采样层拼接）、重复打标检测、整数溢出加固 |
| `mp3.go` | MP3 实现：隐式（ID3v2.4 TXXX）、显式（同参数帧拼接） |
| `rhythm.go` | `RhythmWAV` 合成「短长短短」摩斯节奏提示音（WAV 素材） |
| `version.go` | 包版本常量 `Version` |

## 错误处理

| 错误 | 触发条件 |
|------|----------|
| `ErrInvalidLabel` | `Label` 不在 1/2/3 之内 |
| `ErrProducerRequired` | `ContentProducer` 为空 |
| `ErrInvalidCharset` | 要素值含 0x20–0x7E 之外的字符（违反附录 E j)） |
| `ErrUnsupportedFormat` | 传入了不支持的 `Format` |
| `ErrInvalidWAV` | 数据不是合法的 WAV(RIFF/WAVE) |
| `ErrWAVParamMismatch` | 拼接的两段 WAV 的 fmt 参数不一致 |
| `ErrAlreadyLabeled` | 音频已含 AIGC 标识（wav 已有 `AIGC` chunk / mp3 已有 ID3），拒绝重复打标 |

均为包级 `error` 值，可用 `errors.Is` 判定。

## 说明与约束

- **wav 隐式标识位置**：默认同时写自定义 `AIGC` chunk 与 `LIST/INFO` 子块（兼容不同检测工具）。最终以审核方检测工具读取的位置为准，确认后可收敛到一处。
- **mp3 显式标识**：需 cue 与正文为**同编码参数**的裸 mp3 帧；库的 `RhythmWAV`（wav）不能直接拼进 mp3，mp3 档的提示音请预置 mp3 素材。拼接点可能有极轻微瑕疵（bit reservoir），提示音场景基本无感。
- **不支持格式转换**：本包只在原格式上加标识，不做 pcm↔mp3↔wav 编码转换。

## License

随仓库 LICENSE。
