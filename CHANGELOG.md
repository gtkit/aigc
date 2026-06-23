# Changelog

本项目遵循 [Keep a Changelog 1.1.0](https://keepachangelog.com/zh-CN/1.1.0/) 与 [语义化版本 2.0.0](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added

- `WriteMetadata(audio, format, id)`：纯 Go 写 AIGC 隐式标识字段块——wav 写 RIFF 的 `AIGC` chunk（同时写 `LIST/INFO` 降险），mp3 写 ID3v2.4 TXXX。
- `PrependCue(audio, cue, format, pos)`：显式标识——把提示音/摩斯码拼到正文起始或末尾；wav 采样层无损拼接，mp3 同参数帧拼接。
- `Format`（`WAV`/`MP3`）、`Position`（`AtStart`/`AtEnd`）。
- `RhythmWAV(sampleRate)` 合成「短长短短」摩斯节奏提示音 WAV 素材。
- `Identifier` 七要素 + `Validate`（必填 + 附录 E j) 字符集校验）。
- 错误值 `ErrUnsupportedFormat` / `ErrInvalidWAV` / `ErrWAVParamMismatch` / `ErrAlreadyLabeled` / `ErrInvalidLabel` / `ErrProducerRequired` / `ErrInvalidCharset`。

### Changed

- 实现改为**纯 Go**：内存 `[]byte` 操作，零外部进程依赖，不再依赖 ffmpeg、不落临时文件。

### Removed

- 移除基于 ffmpeg 的实现（`Label`/`Source`/`WrapPCM`/`ConcatCue` 等）：wav 自定义元数据 ffmpeg 写入不可靠，且本场景无需格式转换，纯 Go 更贴合。
