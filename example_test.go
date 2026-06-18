package aigc_test

import (
	"bytes"
	"fmt"

	"github.com/gtkit/aigc"
)

// ExampleIdentifier_JSON 演示把 AIGC 标识序列化为写入文件元数据的紧凑 JSON。
func ExampleIdentifier_JSON() {
	id := aigc.Identifier{
		Label:           aigc.LabelIs,
		ContentProducer: "PRODUCER-001",
		ProduceID:       "20260618-0001",
	}

	s, err := id.JSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	// Output:
	// {"AIGC":{"Label":"1","ContentProducer":"PRODUCER-001","ProduceID":"20260618-0001"}}
}

// ExampleLabelingConfig_NewIdentifier 演示用运行时配置构造一次合成的标识，
// Label 留空时默认按「是 AI 生成合成」处理。
func ExampleLabelingConfig_NewIdentifier() {
	cfg := aigc.LabelingConfig{
		Enabled:         true,
		ContentProducer: "PRODUCER-001",
	}

	fmt.Println("隐式标识启用:", cfg.Active())

	id := cfg.NewIdentifier("20260618-0001")
	fmt.Println("默认 Label:", id.Label)
	// Output:
	// 隐式标识启用: true
	// 默认 Label: 1
}

// ExampleWriteMP3 演示把隐式标识前置写入裸 mp3 帧：输出以 ID3v2.4 标签开头，
// 原始音频帧原样保留在标签之后。
func ExampleWriteMP3() {
	raw := []byte{0xff, 0xfb, 0x90, 0x00} // 裸 mp3 帧（示意）
	id := aigc.Identifier{Label: aigc.LabelMaybe, ContentProducer: "PRODUCER-001"}

	out, err := aigc.WriteMP3(raw, id)
	if err != nil {
		panic(err)
	}

	fmt.Println("以 ID3 开头:", string(out[:3]) == "ID3")
	fmt.Println("尾部保留原帧:", bytes.Equal(out[len(out)-len(raw):], raw))
	// Output:
	// 以 ID3 开头: true
	// 尾部保留原帧: true
}

// ExampleReadMP3 演示读回隐式标识做校验，以及裸帧不含标识；
// found 返回值即可用于写入前的重复打标防护。
func ExampleReadMP3() {
	id := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "PRODUCER-001"}
	labeled, _ := aigc.WriteMP3([]byte{0xff, 0xfb, 0x90, 0x00}, id)

	got, found, err := aigc.ReadMP3(labeled)
	if err != nil {
		panic(err)
	}
	fmt.Println("已标识:", found)
	fmt.Println("生成者:", got.ContentProducer)

	_, bare, _ := aigc.ReadMP3([]byte{0xff, 0xfb, 0x90, 0x00})
	fmt.Println("裸帧已标识:", bare)
	// Output:
	// 已标识: true
	// 生成者: PRODUCER-001
	// 裸帧已标识: false
}
