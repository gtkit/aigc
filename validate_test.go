package aigc_test

import (
	"errors"
	"testing"

	"github.com/gtkit/aigc"
)

func TestValidateCharset(t *testing.T) {
	base := aigc.Identifier{Label: aigc.LabelIs, ContentProducer: "PRODUCER-001"}
	tests := []struct {
		name    string                 // 用例名
		mutate  func(*aigc.Identifier) // 在基础标识上做修改
		wantErr error                  // 期望错误
	}{
		{"合法仅必填", func(*aigc.Identifier) {}, nil},
		{"空可选值不触发", func(id *aigc.Identifier) { id.ProduceID = ""; id.ReservedCode1 = "" }, nil},
		{"ContentProducer含中文", func(id *aigc.Identifier) { id.ContentProducer = "提供者甲" }, aigc.ErrInvalidCharset},
		{"ReservedCode1含中文", func(id *aigc.Identifier) { id.ReservedCode1 = "预留" }, aigc.ErrInvalidCharset},
		{"含控制字符制表符", func(id *aigc.Identifier) { id.ProduceID = "a\tb" }, aigc.ErrInvalidCharset},
		{"Label非法", func(id *aigc.Identifier) { id.Label = "9" }, aigc.ErrInvalidLabel},
		{"ContentProducer空", func(id *aigc.Identifier) { id.ContentProducer = "  " }, aigc.ErrProducerRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := base
			tt.mutate(&id)
			if err := id.Validate(); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, 期望 %v", err, tt.wantErr)
			}
		})
	}
}
