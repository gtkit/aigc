# Changelog

本项目遵循 [Keep a Changelog 1.1.0](https://keepachangelog.com/zh-CN/1.1.0/) 与 [SemVer 2.0.0](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added

- 从 `sleep_client` 迁入独立模块，提供 AI 生成合成内容的文件元数据隐式标识能力（GB 45438-2025）。
- `Identifier` 标识七要素，字段名与顺序对齐附录 E（规范性）；含 `Validate` 必填校验与 `JSON` 序列化，值结构为 `{"AIGC":{...}}`（附录 E b)）。
- `WriteMP3`：在裸 mp3 帧前置写入承载标识的 ID3v2.4 `TXXX` 帧；入参已含本包标识时返回 `ErrAlreadyLabeled`，防止重复打标。
- `ReadMP3`：读回本包写入的 AIGC 标识，返回 `(Identifier, found, error)`，可用于校验与写入前防重。
- `LabelingConfig` 运行时配置，含 `Active` / `HasMarker` / `NewIdentifier`。
- `LabelIs` / `LabelMaybe` / `LabelSuspected` 标识取值常量。
- `ErrInvalidLabel` / `ErrProducerRequired` / `ErrEmptyValue` / `ErrAlreadyLabeled` / `ErrCorruptTag` 错误值。

### Security

- JSON 序列化改用 `github.com/gtkit/json/v2`，对齐 gtkit 生态统一规范。
