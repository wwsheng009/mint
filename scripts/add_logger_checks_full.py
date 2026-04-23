#!/usr/bin/env python3
"""
批量添加 logger.IfEnabled() 检查的脚本（增强版）

该脚本会递归查找所有 .go 文件，为所有 logger 调用添加 IfEnabled() 检查。
会自动跳过：
1. 已用 IfEnabled() 包装的调用
2. if Enabled() { } 块内的调用
3. 测试文件（_test.go）
4. 已在处理列表中的文件
"""

import os
import re
import glob
from pathlib import Path
from typing import Set, Tuple

# 需要跳过的目录
SKIP_DIRS = {
    ".git", ".vscode", ".idea", "node_modules", "vendor",
    "bin", ".claude", ".crush", "logs", "sandbox", "trash",
    "__debug_bin", "dist", "build",
}

# Logger 类型列表
LOGGER_TYPES = [
    "RenderLogger", "PaintLogger", "FiberLogger", "LayoutLogger",
    "FocusLogger", "HitMapLogger", "KeyLogger", "EngineLogger",
    "UILogger", "PipelineLogger", "EventLogger", "InputLogger",
    "PlatFormLogger", "ActionLogger", "IntentLogger", "ButtonLogger",
    "BorderLogger", "WrapLogger", "PumpLogger", "FormLogger",
    "CursorLogger", "InspectorLogger", "LayerLogger", "TempLogger",
    "WinLogger", "LinuxLogger", "ValidationLogger", "RenderingLogger",
    "MessageLogger",
]

# 正则表达式匹配 logger 调用
LOG_PATTERN = re.compile(
    r'\blog\.(' + '|'.join(LOGGER_TYPES) + r')\.(Debug|Info|Warn|Error)\(',
    re.MULTILINE
)


def extract_logger_call(line: str, pos: int) -> Tuple[str, int]:
    """提取完整的 logger 调用，返回 (调用字符串, 下一位置)"""
    start = pos
    depth = 0
    in_string = False
    escape = False
    string_char = None

    # 从 '(' 开始计算
    open_paren_pos = line.find('(', pos)
    if open_paren_pos == -1:
        return None, pos

    depth = 1
    pos = open_paren_pos + 1

    while pos < len(line):
        char = line[pos]

        # 处理字符串
        if escape:
            escape = False
        elif char == '\\':
            escape = True
        elif char in ('"', '`', "'"):
            if not in_string:
                in_string = True
                string_char = char
            elif char == string_char:
                in_string = False

        # 不在字符串中时计算括号
        if not in_string and not escape:
            if char == '(':
                depth += 1
            elif char == ')':
                depth -= 1
                if depth == 0:
                    return line[start:pos+1], pos + 1

        pos += 1

    return None, pos


def should_wrap_line(line: str, logger_name: str, log_method: str) -> bool:
    """判断是否需要包装这行"""
    # 检查是否已经用 IfEnabled() 包装
    if f"log.{logger_name}.IfEnabled().{log_method}(" in line:
        return False

    # 检查是否在 if 块中（简化判断）
    if f"if log.{logger_name}.Enabled()" in line:
        return False

    return True


def process_file(filepath: str, processed_files: Set[str]) -> int:
    """处理单个文件，返回修改数量"""
    filepath_norm = os.path.normpath(filepath)

    # 跳过已处理的文件
    if filepath_norm in processed_files:
        return 0

    # 跳过测试文件
    if filepath.endswith('_test.go'):
        return 0

    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()

        # 跳过不包含日志调用的文件
        if 'log.' not in content:
            return 0

        lines = content.split('\n')
        modified_lines = []
        changes = []
        i = 0

        while i < len(lines):
            line = lines[i]

            # 检查是否有 logger 调用
            match = LOG_PATTERN.search(line)
            if match:
                logger_name = match.group(1)
                log_method = match.group(2)
                call_start = match.start()

                # 检查是否需要包装
                if should_wrap_line(line, logger_name, log_method):
                    # 提取完整的调用
                    full_call, new_pos = extract_logger_call(line, call_start)
                    if full_call:
                        # 替换为：log.RenderLogger.IfEnabled().Debug(...)
                        # 注意：full_call 包含 "log.RenderLogger.Debug(...)"
                        # 我们需要将其改为 "log.RenderLogger.IfEnabled().Debug(...)"
                        prefix = f"log.{logger_name}.{log_method}"
                        wrapped_call = f"log.{logger_name}.IfEnabled().{log_method}{full_call[len(prefix):]}"
                        new_line = line[:call_start] + wrapped_call + line[new_pos:]
                        modified_lines.append(new_line)
                        changes.append(f"  Line {i+1}: {logger_name}.{log_method}()")
                        i += 1
                        continue

            modified_lines.append(line)
            i += 1

        if modified_lines != lines:
            # 写回文件
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write('\n'.join(modified_lines))

            print(f"[OK] Modified: {filepath}")
            for change in changes:
                print(change)

            # 标记为已处理
            processed_files.add(filepath_norm)
            return len(changes)

        return 0

    except UnicodeDecodeError:
        print(f"[SKIP] {filepath} (encoding error)")
        return 0
    except Exception as e:
        print(f"[ERROR] {filepath}: {e}")
        return 0


def find_go_files(root_dir: str) -> list:
    """递归查找所有 .go 文件"""
    go_files = []
    root_path = Path(root_dir)

    for filepath in root_path.rglob("*.go"):
        rel_path = filepath.relative_to(root_path)

        # 跳过特定目录
        skip = False
        for skip_dir in SKIP_DIRS:
            if skip_dir in rel_path.parts:
                skip = True
                break

        if not skip:
            go_files.append(str(filepath))

    return go_files


def main():
    """主函数"""
    import sys

    # 获取脚本所在目录
    script_dir = Path(__file__).parent.parent  # scripts/ -> project root
    root_dir = script_dir

    print("=" * 70)
    print("批量添加 logger.IfEnabled() 检查（全项目模式）")
    print("=" * 70)

    # 如果指定了参数，处理指定文件
    if len(sys.argv) > 1:
        processed_files = set()
        total_changes = 0
        for arg in sys.argv[1:]:
            if os.path.exists(arg):
                changes = process_file(arg, processed_files)
                total_changes += changes

        print("=" * 70)
        print(f"总修改数: {total_changes}")
        print("=" * 70)
        return

    # 批量处理整个项目
    print(f"扫描目录: {root_dir}")
    print()

    # 查找所有 .go 文件
    go_files = find_go_files(str(root_dir))
    print(f"找到 .go 文件: {len(go_files)}")
    print()

    # 优先处理核心路径
    priority_files = []
    for pattern in [
        "runtime/paint",
        "internal/render",
        "internal/reconciler",
        "runtime/event",
        "internal/state",
        "runtime/bridge",
        "runtime/action",
        "runtime/platform",
        "framework/event",
    ]:
        for filepath in go_files:
            if pattern in filepath:
                priority_files.append(filepath)

    # 去重并保持顺序
    seen = set()
    unique_priority = []
    for f in priority_files:
        if f not in seen:
            seen.add(f)
            unique_priority.append(f)

    # 剩余文件
    remaining_files = [f for f in go_files if f not in seen]

    # 合并：优先文件在前
    all_files = unique_priority + remaining_files

    print(f"处理顺序: {len(unique_priority)} 个核心文件 + {len(remaining_files)} 个其他文件")
    print()

    # 分批处理
    processed_files = set()
    total_files = 0
    total_changes = 0
    modified_files = 0

    for filepath in all_files:
        total_files += 1
        changes = process_file(filepath, processed_files)
        if changes > 0:
            total_changes += changes
            modified_files += 1
            print()

    print("=" * 70)
    print(f"扫描文件数: {total_files}")
    print(f"修改文件数: {modified_files}")
    print(f"总修改数: {total_changes}")
    print("=" * 70)
    print()
    print("提示：使用 `git diff` 查看具体改动")


if __name__ == "__main__":
    main()
