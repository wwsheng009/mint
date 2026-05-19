# 语法高亮设计文档

**版本**: v1.0
**日期**: 2026-01-31
**来源**: demo5_ide.md
**状态**: 🟢 低优先级

---

## 一、概述

### 1.1 设计目标

实现增量词法分析器，为代码编辑器提供语法高亮功能。

### 1.2 核心挑战

| 问题 | 错误方案 | 正确方案 |
|------|---------|---------|
| 性能 | 每帧全量解析 | 增量解析 + 缓存 |
| 状态 | 单行独立解析 | 跨行状态传播 |
| 触发 | 定时轮询 | 脏行驱动 |

### 1.3 为什么不能简单正则？

假设：2000 行文件，60 FPS

```go
// ❌ 错误：每帧正则扫描全文件
for each frame:
    for each line:
        regex.Match(line)  // 2000 × 60 = 120,000 次/秒
```

```go
// ✅ 正确：只解析修改过的行
onLineChanged(lineNumber):
    tokens[lineNumber] = lexLine(lineNumber)
```

---

## 二、Token 模型

### 2.1 TokenType 定义

```go
// framework/editor/token.go

package editor

// TokenType Token 类型
type TokenType int

const (
    TokenEOF TokenType = iota

    // 关键字
    TokenKeyword
    TokenControl   // if/else/for/return

    // 字面量
    TokenString
    TokenChar
    TokenNumber
    TokenBool

    // 标识符
    TokenIdent
    TokenBuiltin   // 内置函数

    // 注释
    TokenLineComment
    TokenBlockComment

    // 符号
    TokenOperator
    TokenPunctuation
    TokenBracket

    // 其他
    TokenWhitespace
    TokenNewline
    TokenUnknown
)
```

### 2.2 Token 结构

```go
// framework/editor/token.go

// Token 单个词法单元
type Token struct {
    Type  TokenType
    Start int      // 行内起始位置
    End   int      // 行内结束位置
    Value string   // 原始文本（可选，用于调试）
}

// IsValid 检查 Token 是否有效
func (t Token) IsValid() bool {
    return t.Type != TokenEOF && t.Start < t.End
}

// Length 返回 Token 长度
func (t Token) Length() int {
    return t.End - t.Start
}
```

### 2.3 Token 样式映射

```go
// framework/editor/token_style.go

package editor

// TokenStyle Token 到样式的映射
type TokenStyle struct {
    Fg       *color.Color
    Bg       *color.Color
    Bold     bool
    Italic   bool
    Underline bool
}

// DefaultTokenStyles 默认样式表
var DefaultTokenStyles = map[TokenType]TokenStyle{
    TokenKeyword:    {Fg: color.Magenta, Bold: true},
    TokenControl:    {Fg: color.Magenta, Bold: true},
    TokenString:     {Fg: color.Green},
    TokenChar:       {Fg: color.Green},
    TokenNumber:     {Fg: color.Yellow},
    TokenBool:       {Fg: color.Yellow, Bold: true},
    TokenIdent:      {Fg: color.White},
    TokenBuiltin:    {Fg: color.Cyan},
    TokenLineComment: {Fg: color.Gray},
    TokenBlockComment: {Fg: color.Gray, Italic: true},
    TokenOperator:   {Fg: color.White},
    TokenPunctuation: {Fg: color.White},
    TokenBracket:    {Fg: color.White},
}

// GetStyle 获取 Token 样式
func GetStyle(tokenType TokenType) TokenStyle {
    if style, ok := DefaultTokenStyles[tokenType]; ok {
        return style
    }
    return TokenStyle{Fg: color.White}
}
```

---

## 三、行状态管理

### 3.1 LineState 定义

```go
// framework/editor/linestate.go

package editor

// LineState 行状态（用于跨行解析）
type LineState struct {
    // 在块注释中
    InBlockComment bool
    // 在字符串中
    InString bool
    // 字符串引号类型
    StringQuote rune
    // 在多行字符串中（如 Go 的 `）
    InRawString bool
}

// Equal 比较两个状态是否相等
func (s LineState) Equal(other LineState) bool {
    return s.InBlockComment == other.InBlockComment &&
           s.InString == other.InString &&
           s.StringQuote == other.StringQuote &&
           s.InRawString == other.InRawString
}
```

### 3.2 状态传播

```go
// framework/editor/state_propagation.go

package editor

// LineStateMap 行状态映射
type LineStateMap struct {
    states map[int]LineState
    dirty  map[int]bool // 脏行标记
}

// NewLineStateMap 创建行状态映射
func NewLineStateMap() *LineStateMap {
    return &LineStateMap{
        states: make(map[int]LineState),
        dirty:  make(map[int]bool),
    }
}

// Get 获取行状态
func (m *LineStateMap) Get(line int) LineState {
    if state, ok := m.states[line]; ok {
        return state
    }
    return LineState{} // 默认状态
}

// Set 设置行状态
func (m *LineStateMap) Set(line int, state LineState) {
    old := m.Get(line)
    if !old.Equal(state) {
        m.states[line] = state
        // 标记后续行为脏
        m.MarkDirtyFrom(line + 1)
    }
}

// MarkDirtyFrom 标记从某行开始的所有行为脏
func (m *LineStateMap) MarkDirtyFrom(line int) {
    // 状态变化会影响后续行
    m.dirty[line] = true
}

// ClearDirty 清除脏标记
func (m *LineStateMap) ClearDirty(line int) {
    delete(m.dirty, line)
}

// IsDirty 检查行是否脏
func (m *LineStateMap) IsDirty(line int) bool {
    return m.dirty[line]
}
```

---

## 四、增量 Lexer

### 4.1 Lexer 接口

```go
// framework/editor/lexer.go

package editor

// Lexer 词法分析器接口
type Lexer interface {
    // TokenizeLine 对单行进行词法分析
    TokenizeLine(lineNum int, line []rune, state LineState) ([]Token, LineState)
}

// GoLexer Go 语言词法分析器
type GoLexer struct{}

// NewGoLexer 创建 Go Lexer
func NewGoLexer() *GoLexer {
    return &GoLexer{}
}
```

### 4.2 核心词法分析

```go
// framework/editor/go_lexer.go

package editor

// TokenizeLine 实现 Go 词法分析
func (l *GoLexer) TokenizeLine(lineNum int, line []rune, state LineState) ([]Token, LineState) {
    tokens := []Token{}
    pos := 0
    len := len(line)
    newState := state

    for pos < len {
        // 跳过空白
        if unicode.IsSpace(line[pos]) {
            start := pos
            for pos < len && unicode.IsSpace(line[pos]) {
                pos++
            }
            tokens = append(tokens, Token{
                Type:  TokenWhitespace,
                Start: start,
                End:   pos,
            })
            continue
        }

        // 块注释开始
        if !state.InBlockComment && pos+1 < len && line[pos] == '/' && line[pos+1] == '*' {
            newState.InBlockComment = true
            pos += 2
            tokens = append(tokens, Token{
                Type:  TokenBlockComment,
                Start: pos - 2,
                End:   pos,
            })
            continue
        }

        // 块注释结束
        if state.InBlockComment && pos+1 < len && line[pos] == '*' && line[pos+1] == '/' {
            newState.InBlockComment = false
            pos += 2
            tokens = append(tokens, Token{
                Type:  TokenBlockComment,
                Start: pos - 2,
                End:   pos,
            })
            continue
        }

        // 在块注释中
        if state.InBlockComment {
            start := pos
            for pos < len && !(pos+1 < len && line[pos] == '*' && line[pos+1] == '/') {
                pos++
            }
            tokens = append(tokens, Token{
                Type:  TokenBlockComment,
                Start: start,
                End:   pos,
            })
            continue
        }

        // 行注释
        if pos < len && line[pos] == '/' {
            tokens = append(tokens, Token{
                Type:  TokenLineComment,
                Start: pos,
                End:   len,
            })
            break // 行注释后不再处理
        }

        // 字符串字面量
        if line[pos] == '"' || line[pos] == '\'' || line[pos] == '`' {
            quote := line[pos]
            start := pos
            pos++

            for pos < len {
                if line[pos] == '\\' && pos+1 < len {
                    pos += 2 // 转义字符
                    continue
                }
                if line[pos] == quote {
                    pos++
                    break
                }
                if line[pos] == '\n' && quote != '`' {
                    break
                }
                pos++
            }

            var tokenType TokenType
            if quote == '`' {
                tokenType = TokenString
            } else if quote == '"' {
                tokenType = TokenString
            } else {
                tokenType = TokenChar
            }

            tokens = append(tokens, Token{
                Type:  tokenType,
                Start: start,
                End:   pos,
            })
            continue
        }

        // 数字
        if unicode.IsDigit(line[pos]) {
            start := pos
            for pos < len && (unicode.IsDigit(line[pos]) || line[pos] == '.') {
                pos++
            }
            tokens = append(tokens, Token{
                Type:  TokenNumber,
                Start: start,
                End:   pos,
            })
            continue
        }

        // 标识符或关键字
        if unicode.IsLetter(line[pos]) || line[pos] == '_' {
            start := pos
            for pos < len && (unicode.IsLetter(line[pos]) || unicode.IsDigit(line[pos]) || line[pos] == '_') {
                pos++
            }

            ident := string(line[start:pos])
            tokenType := l.classifyIdentifier(ident)

            tokens = append(tokens, Token{
                Type:  tokenType,
                Start: start,
                End:   pos,
            })
            continue
        }

        // 符号
        start = pos
        for pos < len && isOperator(line[pos]) {
            pos++
        }
        tokens = append(tokens, Token{
            Type:  TokenOperator,
            Start: start,
            End:   pos,
        })
    }

    return tokens, newState
}

// classifyIdentifier 分类标识符
func (l *GoLexer) classifyIdentifier(ident string) TokenType {
    // 关键字
    keywords := map[string]TokenType{
        "func":   TokenKeyword,
        "var":    TokenKeyword,
        "const":  TokenKeyword,
        "type":   TokenKeyword,
        "struct": TokenKeyword,
        "interface": TokenKeyword,
        "if":     TokenControl,
        "else":   TokenControl,
        "for":    TokenControl,
        "return": TokenControl,
        "break":  TokenControl,
        "continue": TokenControl,
        "go":     TokenKeyword,
        "defer":  TokenKeyword,
        "select": TokenKeyword,
        "case":   TokenControl,
        "default": TokenControl,
        "switch": TokenControl,
    }

    if tok, ok := keywords[ident]; ok {
        return tok
    }

    // 预定义标识符
    builtins := map[string]bool{
        "append": true, "copy": true, "delete": true,
        "len": true, "cap": true, "make": true, "new": true,
        "print": true, "println": true,
        "true": true, "false": true, "iota": true, "nil": true,
    }

    if builtins[ident] {
        return TokenBuiltin
    }

    return TokenIdent
}

func isOperator(r rune) bool {
    operators := "+-*/%&|^<>=!:.()[]{};,"
    return strings.ContainsRune(operators, r)
}
```

### 4.3 增量解析管理器

```go
// framework/editor/incremental_lexer.go

package editor

// IncrementalLexer 增量词法分析器
type IncrementalLexer struct {
    lexer    Lexer
    tokens   map[int][]Token      // 行号 → Token 列表
    states   *LineStateMap         // 行状态
    buffer   *TextBuffer           // 文本缓冲
}

// NewIncrementalLexer 创建增量词法分析器
func NewIncrementalLexer(buffer *TextBuffer) *IncrementalLexer {
    return &IncrementalLexer{
        lexer:  NewGoLexer(),
        tokens: make(map[int][]Token),
        states: NewLineStateMap(),
        buffer: buffer,
    }
}

// RetokenizeLine 重新标记某行
func (l *IncrementalLexer) RetokenizeLine(lineNum int) {
    line := l.buffer.GetLine(lineNum)
    state := l.states.Get(lineNum)

    tokens, newState := l.lexer.TokenizeLine(lineNum, line, state)

    l.tokens[lineNum] = tokens
    l.states.Set(lineNum, newState)
    l.states.ClearDirty(lineNum)

    // 如果状态变化，传播到后续行
    if !state.Equal(newState) {
        l.propagateState(lineNum + 1)
    }
}

// propagateState 传播状态变化
func (l *IncrementalLexer) propagateState(startLine int) {
    prevState := l.states.Get(startLine - 1)
    for i := startLine; i < l.buffer.LineCount(); i++ {
        line := l.buffer.GetLine(i)
        tokens, newState := l.lexer.TokenizeLine(i, line, prevState)

        // 状态稳定，停止传播
        if prevState.Equal(newState) && l.states.Get(i).Equal(newState) {
            break
        }

        l.tokens[i] = tokens
        l.states.Set(i, newState)
        prevState = newState
    }
}

// GetTokens 获取某行的 Token
func (l *IncrementalLexer) GetTokens(lineNum int) []Token {
    if tokens, ok := l.tokens[lineNum]; ok {
        return tokens
    }
    return []Token{}
}

// MarkDirty 标记某行为脏
func (l *IncrementalLexer) MarkDirty(lineNum int) {
    l.states.MarkDirtyFrom(lineNum)
}

// OnLineChange 行变化回调
func (l *IncrementalLexer) OnLineChange(lineNum int, changeType ChangeType) {
    switch changeType {
    case ChangeModified, ChangeInserted:
        l.RetokenizeLine(lineNum)
    case ChangeDeleted:
        // 删除行需要重新处理后续行的状态
        l.propagateState(lineNum)
    }
}
```

---

## 五、渲染集成

### 5.1 高亮渲染

```go
// framework/editor/highlight_render.go

package editor

// PaintLineWithHighlight 渲染带高亮的行
func PaintLineWithHighlight(buffer *paint.Buffer, lineNum int, x, y int, lexer *IncrementalLexer) {
    line := lexer.buffer.GetLine(lineNum)
    tokens := lexer.GetTokens(lineNum)

    for _, token := range tokens {
        style := GetStyle(token.Type)

        // 提取 Token 文本
        text := string(line[token.Start:token.End])

        // 绘制
        buffer.DrawString(x+token.Start, y, text, style)
    }
}

// PaintLineRange 渲染多行
func PaintLineRange(buffer *paint.Buffer, startLine, endLine int, x, y int, lexer *IncrementalLexer) {
    currentY := y
    for line := startLine; line < endLine; line++ {
        if lexer.states.IsDirty(line) {
            lexer.RetokenizeLine(line)
        }
        PaintLineWithHighlight(buffer, line, x, currentY, lexer)
        currentY++
    }
}
```

### 5.2 虚拟列表集成

```go
// framework/editor/vlist_highlight.go

package editor

// HighlightedVirtualList 支持高亮的虚拟列表
type HighlightedVirtualList struct {
    *VirtualList
    lexer *IncrementalLexer
}

// NewHighlightedVirtualList 创建支持高亮的虚拟列表
func NewHighlightedVirtualList(buffer *TextBuffer) *HighlightedVirtualList {
    return &HighlightedVirtualList{
        VirtualList: NewVirtualList(buffer),
        lexer:       NewIncrementalLexer(buffer),
    }
}

// Render 渲染可见行
func (v *HighlightedVirtualList) Render(buffer *paint.Buffer) {
    visible := v.GetVisibleRange()

    for i := visible.Start; i < visible.End; i++ {
        if v.lexer.states.IsDirty(i) {
            v.lexer.RetokenizeLine(i)
        }
    }

    v.VirtualList.Render(buffer)
}
```

---

## 六、性能优化

### 6.1 延迟解析

```go
// framework/editor/lazy_lexer.go

package editor

// LazyLexer 延迟词法分析器
type LazyLexer struct {
    *IncrementalLexer
    pending map[int]bool // 待解析行
}

// MarkDirty 标记脏行（不立即解析）
func (l *LazyLexer) MarkDirty(lineNum int) {
    l.pending[lineNum] = true
}

// ProcessPending 处理待解析行（在空闲时）
func (l *LazyLexer) ProcessPending(budget time.Duration) int {
    deadline := time.Now().Add(budget)
    processed := 0

    for line := range l.pending {
        if time.Now().After(deadline) {
            break
        }
        l.RetokenizeLine(line)
        delete(l.pending, line)
        processed++
    }

    return processed
}
```

### 6.2 Token 缓存

```go
// framework/editor/token_cache.go

package editor

// TokenCache Token 缓存
type TokenCache struct {
    cache map[string][]Token
    hits  int
    misses int
}

// ComputeHash 计算行内容的哈希
func (c *TokenCache) ComputeHash(line []rune) string {
    h := fnv.New32()
    for _, r := range line {
        h.Write([]byte(string(r)))
    }
    return fmt.Sprintf("%x", h.Sum32())
}

// Get 获取缓存的 Token
func (c *TokenCache) Get(line []rune) ([]Token, bool) {
    key := c.ComputeHash(line)
    if tokens, ok := c.cache[key]; ok {
        c.hits++
        return tokens, true
    }
    c.misses++
    return nil, false
}

// Set 设置缓存
func (c *TokenCache) Set(line []rune, tokens []Token) {
    key := c.ComputeHash(line)
    c.cache[key] = tokens
}
```

---

## 七、多语言支持

### 7.1 Lexer 工厂

```go
// framework/editor/lexer_factory.go

package editor

// LanguageType 语言类型
type LanguageType string

const (
    LangGo         LanguageType = "go"
    LangJavaScript LanguageType = "javascript"
    LangPython     LanguageType = "python"
    LangJSON       LanguageType = "json"
    LangMarkdown   LanguageType = "markdown"
)

// LexerFactory 词法分析器工厂
type LexerFactory struct{}

// NewLexerFactory 创建工厂
func NewLexerFactory() *LexerFactory {
    return &LexerFactory{}
}

// Create 创建指定语言的 Lexer
func (f *LexerFactory) Create(lang LanguageType) Lexer {
    switch lang {
    case LangGo:
        return NewGoLexer()
    case LangJavaScript:
        return NewJavaScriptLexer()
    case LangPython:
        return NewPythonLexer()
    case LangJSON:
        return NewJSONLexer()
    case LangMarkdown:
        return NewMarkdownLexer()
    default:
        return NewPlainTextLexer()
    }
}

// DetectFromFilename 从文件名检测语言
func (f *LexerFactory) DetectFromFilename(filename string) LanguageType {
    ext := filepath.Ext(filename)
    switch ext {
    case ".go":
        return LangGo
    case ".js", ".ts":
        return LangJavaScript
    case ".py":
        return LangPython
    case ".json":
        return LangJSON
    case ".md":
        return LangMarkdown
    default:
        return "plaintext"
    }
}
```

---

## 八、使用示例

### 8.1 基础使用

```go
// 示例：创建带高亮的编辑器
buffer := input.NewTextBuffer()
lexer := editor.NewIncrementalLexer(buffer)

// 文本变化时
buffer.OnChange(func(line int) {
    lexer.MarkDirty(line)
})

// 渲染时
func RenderEditor(buffer *paint.Buffer) {
    for line := viewport.Start; line < viewport.End; line++ {
        editor.PaintLineWithHighlight(buffer, line, 0, line-viewport.Start, lexer)
    }
}
```

### 8.2 多语言编辑器

```go
// 示例：多语言支持
factory := editor.NewLexerFactory()

func OpenFile(filename string) {
    lang := factory.DetectFromFilename(filename)
    lexer := factory.Create(lang)

    // 使用 lexer...
}
```

---

## 九、实施计划

### 阶段 1: 核心实现

- [ ] 实现 Token/TokenType
- [ ] 实现 TokenStyle 映射
- [ ] 实现 LineState
- [ ] 实现 GoLexer

### 阶段 2: 增量解析

- [ ] 实现 IncrementalLexer
- [ ] 实现状态传播
- [ ] 实现脏行标记
- [ ] 实现延迟解析

### 阶段 3: 渲染集成

- [ ] 实现高亮渲染
- [ ] 集成到 Editor 组件
- [ ] 性能优化

### 阶段 4: 多语言支持

- [ ] 实现 JavaScript Lexer
- [ ] 实现 Python Lexer
- [ ] 实现 LexerFactory

---

## 十、测试策略

```go
// framework/editor/lexer_test.go

func TestGoLexerKeywords(t *testing.T) {
    lexer := NewGoLexer()
    line := []rune("func main() {")
    state := LineState{}

    tokens, _ := lexer.TokenizeLine(0, line, state)

    assert.Equal(t, TokenKeyword, tokens[0].Type)  // func
    assert.Equal(t, TokenIdent, tokens[1].Type)   // main
    assert.Equal(t, TokenBracket, tokens[2].Type) // (
}

func TestStatePropagation(t *testing.T) {
    lexer := NewIncrementalLexer(buffer)

    // 插入多行注释开始
    buffer.Insert(0, 0, []rune("/* comment"))

    lexer.RetokenizeLine(0)

    // 验证状态传播
    state := lexer.states.Get(1)
    assert.True(t, state.InBlockComment)
}
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-31
