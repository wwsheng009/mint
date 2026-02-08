#!/bin/bash
# Demo2 Inspector 快速测试脚本

cd "$(dirname "$0")"

echo "==================================="
echo "Demo2 Inspector 测试脚本"
echo "==================================="
echo ""

# 检查编译
if [ ! -f "./demo2_inspector.exe" ]; then
    echo "⚠️  demo2_inspector.exe 未找到，开始编译..."
    go build -o demo2_inspector.exe main.go
    if [ $? -ne 0 ]; then
        echo "❌ 编译失败"
        exit 1
    fi
    echo "✅ 编译成功"
fi

echo ""
echo "测试选项："
echo "1. 基础运行（无调试）"
echo "2. 基础调试（TUI_DEBUG）"
echo "3. Inspector 详细模式（TUI_INSPECTOR_VERBOSE）"
echo "4. Layer 系统调试（TUI_LAYER_DEBUG）"
echo "5. 完整诊断（所有调试）"
echo "6. 输出到文件（保存调试日志）"
echo ""

read -p "请选择 (1-6): " choice

case $choice in
    1)
        echo "🚀 启动 demo2（无调试）..."
        ./demo2_inspector.exe
        ;;
    2)
        echo "🐛 启动 demo2（基础调试）..."
        TUI_DEBUG=true TUI_DEBUG_UI=true ./demo2_inspector.exe
        ;;
    3)
        echo "🔍 启动 demo2（Inspector 详细）..."
        TUI_DEBUG=true TUI_INSPECTOR_VERBOSE=true ./demo2_inspector.exe
        ;;
    4)
        echo "📊 启动 demo2（Layer 调试）..."
        TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true ./demo2_inspector.exe
        ;;
    5)
        echo "🔬 启动 demo2（完整诊断）..."
        TUI_DEBUG=true TUI_DEBUG_UI=true TUI_INSPECTOR_VERBOSE=true \
        TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true \
        ./demo2_inspector.exe 2>&1 | tee demo2_debug_$(date +%Y%m%d_%H%M%S).log
        ;;
    6)
        log_file="demo2_debug_$(date +%Y%m%d_%H%M%S).log"
        echo "📝 启动 demo2（保存到 $log_file）..."
        TUI_DEBUG=true TUI_DEBUG_UI=true TUI_INSPECTOR_VERBOSE=true \
        TUI_LAYER_DEBUG=true TUI_DEBUG_RENDERING=true \
        ./demo2_inspector.exe 2>&1 | tee $log_file
        echo ""
        echo "✅ 调试日志已保存到: $log_file"
        ;;
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac
