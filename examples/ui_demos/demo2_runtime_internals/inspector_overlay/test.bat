@echo off
REM Demo2 Inspector 快速测试脚本 (Windows)

cd /d "%~dp0"

echo ===================================
echo Demo2 Inspector 测试脚本
echo ===================================
echo.

REM 检查编译
if not exist "demo2_inspector.exe" (
    echo [!] demo2_inspector.exe 未找到，开始编译...
    go build -o demo2_inspector.exe main.go
    if errorlevel 1 (
        echo [X] 编译失败
        pause
        exit /b 1
    )
    echo [√] 编译成功
)

echo.
echo 测试选项：
echo 1. 基础运行（无调试）
echo 2. 基础调试（TUI_DEBUG）
echo 3. Inspector 详细模式（TUI_INSPECTOR_VERBOSE）
echo 4. Layer 系统调试（TUI_LAYER_DEBUG）
echo 5. 完整诊断（所有调试）
echo 6. 输出到文件（保存调试日志）
echo.

set /p choice="请选择 (1-6): "

if "%choice%"=="1" (
    echo.
    echo [√] 启动 demo2（无调试）...
    demo2_inspector.exe
) else if "%choice%"=="2" (
    echo.
    echo [*] 启动 demo2（基础调试）...
    set TUI_DEBUG=true
    set TUI_DEBUG_UI=true
    demo2_inspector.exe
) else if "%choice%"=="3" (
    echo.
    echo [?] 启动 demo2（Inspector 详细）...
    set TUI_DEBUG=true
    set TUI_INSPECTOR_VERBOSE=true
    demo2_inspector.exe
) else if "%choice%"=="4" (
    echo.
    echo [£] 启动 demo2（Layer 调试）...
    set TUI_LAYER_DEBUG=true
    set TUI_DEBUG_RENDERING=true
    demo2_inspector.exe
) else if "%choice%"=="5" (
    echo.
    echo [§] 启动 demo2（完整诊断）...
    set TUI_DEBUG=true
    set TUI_DEBUG_UI=true
    set TUI_INSPECTOR_VERBOSE=true
    set TUI_LAYER_DEBUG=true
    set TUI_DEBUG_RENDERING=true
    demo2_inspector.exe 2>&1 | tee demo2_debug_%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%.log
) else if "%choice%"=="6" (
    set log_file=demo2_debug_%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%.log
    echo.
    echo [^] 启动 demo2（保存到 %log_file%）...
    set TUI_DEBUG=true
    set TUI_DEBUG_UI=true
    set TUI_INSPECTOR_VERBOSE=true
    set TUI_LAYER_DEBUG=true
    set TUI_DEBUG_RENDERING=true
    demo2_inspector.exe 2>&1 | tee %log_file%
    echo.
    echo [√] 调试日志已保存到: %log_file%
) else (
    echo.
    echo [X] 无效选择
    pause
    exit /b 1
)

pause
