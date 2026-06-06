# PageViewport

`PageViewport` is an interactive clipped viewport for dense TUI pages. Unlike
the text-oriented `ScrollView`, it keeps the child VNode tree intact so child
controls continue to participate in layout, painting, hit testing, focus, and
event routing.

Supported scope:

- controlled vertical `ScrollOffset`
- uncontrolled vertical scrolling when no `ScrollOffset` prop is provided
- wheel scrolling and PageUp/PageDown/Home/End navigation actions
- visible vertical scroll indicator when content overflows the viewport
- stable viewport `Width` and `Height`
- layout offset for in-flow children
- paint and hitmap clipping to the viewport rectangle
- child control event routing through the preserved Fiber tree

Use `PageViewport` for complex page content with buttons, inputs, tables, tabs,
and other interactive children. Use the older `ScrollView` for text-only blocks
where extracting and painting text lines is sufficient.

```go
viewport := pageviewport.NewBuilder().
	Child(content).
	Width(100).
	Height(18).
	Build()
```

Use `ScrollOffset(...)` only when the application wants to fully control the
offset. Without it, the instance owns the offset and handles scroll actions.

The scroll indicator is enabled by default. It is post-painted over the right
edge after children have rendered, so it does not participate in layout, focus,
hit testing, or event routing. Disable it for custom chrome with
`ShowIndicator(false)`.
