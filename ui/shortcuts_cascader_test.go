package ui

import (
	"testing"

	"github.com/wwsheng009/mint/ui/components/cascader"
)

func TestNewCascaderBuilder(t *testing.T) {
	vnode := NewCascaderBuilder().
		Options([]cascader.Option{
			cascader.Node("zj", "Zhejiang", cascader.Leaf("hz", "Hangzhou")),
		}).
		Build()
	if vnode == nil {
		t.Fatal("NewCascaderBuilder().Build() returned nil")
	}
	if vnode.Tag() != "cascader" {
		t.Fatalf("Tag = %q, want cascader", vnode.Tag())
	}
}
