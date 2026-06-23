# Changelog

本项目遵循 [Keep a Changelog 1.1.0](https://keepachangelog.com/zh-CN/1.1.0/) 与 [语义化版本 2.0.0](https://semver.org/lang/zh-CN/)。

## [Unreleased]

## [1.1.0] - 2026-06-25

### Added

- `Label(audio, cue, format, pos, id)`：端到端组合入口，一步完成显式标识（cue 非空时）与隐式标识；cue 为 nil 时只做隐式标识。

### Changed

- wav 隐式标识新增重复打标检测：已含 `AIGC` chunk 时返回 `ErrAlreadyLabeled`（与 mp3 对称）。
- `ErrAlreadyLabeled` 文案改为通用「音频已含 AIGC 标识，拒绝重复打标」，wav/mp3 共用（导出名不变，`errors.Is` 兼容）。

## [1.0.0] - 2026-06-25

### Added

- `WriteMetadata(audio, format, id)`：写 AIGC 隐式标识字段块——wav 写 RIFF 的 `AIGC` chunk（同时写 `LIST/INFO` 降险），mp3 写 ID3v2.4 TXXX。
- `PrependCue(audio, cue, format, pos)`：显式标识——把提示音/摩斯码拼到正文起始或末尾；wav 采样层无损拼接，mp3 同参数帧拼接。
- `RhythmWAV(sampleRate)`：合成「短长短短」摩斯节奏提示音（单声道 16-bit PCM WAV）素材。
- `Identifier` 七要素 + `Validate`（必填校验 + GB 45438-2025 附录 E j) 字符集校验）。
- `Format`（`WAV`/`MP3`）、`Position`（`AtStart`/`AtEnd`）。
- 错误值 `ErrUnsupportedFormat` / `ErrInvalidWAV` / `ErrWAVParamMismatch` / `ErrAlreadyLabeled` / `ErrInvalidLabel` / `ErrProducerRequired` / `ErrInvalidCharset`。

### 说明

- 纯 Go 实现：内存 `[]byte` 操作、零外部进程依赖（不依赖 ffmpeg、不落临时文件）。
- 依据《人工智能生成合成内容标识办法》及强制性国标 GB 45438-2025。
