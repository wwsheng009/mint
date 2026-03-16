# Tabs Default Initialization Design

## Summary

`ui/components/tabs` should be able to render with readable default styles even when a host program does not initialize the global theme system. The component should keep following host-provided theme and explicit style configuration when available, but it must stop depending on them for baseline usability.

This design adds component-local fallback behavior for tab header rendering only. It does not change tab switching, focus, intent, or global theme initialization behavior.

## Problem

Recent fixes made `tabs` work correctly in the example app after global theme initialization was corrected, but that still leaves an integration trap:

- Moving `tabs` into a different program can silently break the selected-state appearance if that program never initializes the theme system.
- Selected header readability currently depends on theme data being present and well-formed.
- Callers can still accidentally override selected foreground color in a way that reduces contrast.

The result is a component that is functionally usable but not self-sufficient. The failure mode is subtle enough that a user can end up debugging the host app instead of the component.

## Goals

- Make `tabs` visually usable without requiring `InitTheme()`.
- Keep the fallback aligned with terminal defaults instead of introducing a fixed palette.
- Preserve existing theme-driven styling when the host app provides it.
- Preserve explicit caller styling where it does not break selected-state readability.
- Limit the change to the tab header rendering path.

## Non-Goals

- Do not initialize the global theme manager from inside `tabs`.
- Do not change the builder API or add a required initialization call.
- Do not create a component-local theme system with named palettes.
- Do not change tab switching, focus handling, intent emission, or layout behavior.
- Do not guarantee pixel-identical appearance across terminals.

## Current Context

The relevant behavior is concentrated in:

- `ui/components/tabs/instance.go`
- `ui/components/tabs/vnode.go`
- `ui/components/tabs/builder.go`

Current active-header rendering already attempts to compose:

1. built-in active emphasis
2. base tab style
3. theme-derived selected style
4. explicit `ActiveTabStyle`
5. a foreground protection step to preserve readability

That is the correct integration point. The gap is that theme-derived styling is still treated as the normal path, while the component-local fallback is too implicit and not clearly defined as the component's own responsibility.

## Design

### Responsibility Boundary

`tabs` owns its own visual fallback for header rendering. It should not assume that any theme provider exists.

The component will:

- read component styles and semantic colors from the theme when available
- synthesize a minimal selected style from semantic theme colors when component-level tab styles are missing
- fall back to terminal-friendly local defaults when theme access yields no usable style information

The component will not:

- call global theme initialization
- mutate global theme state
- expose a new initialization API

### Exact Theme Inputs

This change reads only the following theme inputs:

- component style: `style.GetStyle("tabs", "select")`
- semantic selected background: `fwtheme.Select()`
- semantic contrasting foreground: `fwtheme.BG()`
- semantic disabled foreground: `fwtheme.DisabledFG()`

No other component-level theme keys are introduced in this design.

For this spec:

- a component theme style is "usable" when `style.GetStyle("tabs", "select")` is not empty
- a semantic color is "usable" when it is not `style.NoColor`
- "host did not initialize theme" means the style bridge was never registered, so `style.GetStyle(...)` returns empty; semantic colors may still be usable because `framework/theme` carries its own default manager

### Fallback Trigger Rules

Fallback is evaluated at render time for each tab header.

Active tab resolution is deterministic:

1. If `style.GetStyle("tabs", "select")` is non-empty, use that as the active-state theme style source.
2. Otherwise, if both `fwtheme.Select()` and `fwtheme.BG()` are usable, synthesize the active-state theme style as:
   - `BG = fwtheme.Select()`
   - `FG = fwtheme.BG()`
   - `Bold = true`
3. Otherwise, use the component-local active fallback:
   - `Reverse = true`
   - `Bold = true`

Disabled tab resolution is deterministic:

1. If `fwtheme.DisabledFG()` is usable, use it as the disabled baseline foreground.
2. Otherwise, use the component-local disabled fallback:
   - `Italic = true`
   - no foreground color

Normal tab resolution is deterministic:

- no theme lookup is required
- foreground and background remain unset at the baseline layer

This keeps the component usable in three host states:

- fully themed host with registered style bridge
- host without registered style bridge but with semantic colors available
- host where both theme bridge and semantic colors are unavailable

### Local Default Style Model

The local fallback must follow terminal defaults as much as possible.

#### Normal Tab

- `FG = NoColor`
- `BG = NoColor`
- no style flags are set at the baseline layer
- spacing, divider, and label logic remain unchanged

#### Active Tab

- preferred active style source order is:
  1. `tabs/select` component style
  2. semantic synthesis from `fwtheme.Select()` and `fwtheme.BG()`
  3. local fallback `Reverse(true).Bold(true)`
- local fallback never introduces fixed color values
- the local fallback is the final safety net only when both theme style and semantic colors are unavailable

#### Disabled Tab

- baseline style uses `fwtheme.DisabledFG()` when available
- otherwise baseline style is `Italic(true)` with unset colors
- this avoids introducing a fixed disabled color in the component-local fallback

#### Divider And Hotkeys

- divider remains neutral and does not participate in selected background filling
- hotkey display remains text-only in the current component and does not add a separate style layer

### State Precedence

Header state precedence is:

1. hidden: item is not rendered
2. disabled: disabled style path wins over active style path
3. active: active style path applies only to visible, enabled items whose index equals `activeTab`
4. normal: all other visible items

`normalizeActiveTab()` already attempts to avoid active disabled tabs. If the render path still sees an active+disabled overlap, disabled wins so the header does not present a non-interactive tab as selected.

### Style Composition Order

State baselines are:

- normal: empty style
- disabled: resolved disabled baseline
- active: resolved active baseline from component theme style, semantic synthesis, or local reverse fallback

The final style order is state-specific:

#### Normal

1. empty baseline
2. shared explicit `TabStyle`

#### Disabled

1. disabled baseline
2. shared explicit `TabStyle`
3. explicit `DisabledTabStyle`

#### Active

1. shared explicit `TabStyle`
2. active baseline
3. explicit `ActiveTabStyle`
4. selected-state readability protection

For active tabs, this means the selected baseline replaces any shared `TabStyle.FG` or `TabStyle.BG` before `ActiveTabStyle` is merged.

The active-tab readability protection rule is exact:

1. If the active baseline came from `tabs/select` and that baseline specifies `FG`, final `FG` is forced back to that baseline `FG` after all merges, regardless of whether the conflicting foreground came from `TabStyle` or `ActiveTabStyle`.
2. If the active baseline came from semantic synthesis, final `FG` is forced to the synthesized `FG` after all merges.
3. If the active baseline came from the local reverse fallback, final color overrides from `ActiveTabStyle` are ignored:
   - `FG` remains `NoColor`
   - `BG` remains `NoColor`
   - `Reverse = true`
   - shared `TabStyle` colors are also cleared
   - non-color flags from `TabStyle` and `ActiveTabStyle` may still merge
4. `ActiveTabStyle.BG` is allowed to override the theme-derived selected background only when the active baseline came from component theme style or semantic synthesis. In that case the component guarantees foreground preservation only; it does not attempt runtime contrast computation for arbitrary caller-provided backgrounds.

This preserves caller customization while preventing the specific failure mode that made selected labels hard to read.

### Implementation Shape

The change should stay inside the tabs rendering helpers rather than spreading theme checks across unrelated code paths.

Expected units:

- `resolveActiveBaseline() (style.Style, activeBaselineSource)`
  - inputs: `style.GetStyle("tabs", "select")`, `fwtheme.Select()`, `fwtheme.BG()`
  - outputs: the active baseline style and its source enum: `componentTheme`, `semanticTheme`, or `localFallback`
- `resolveDisabledBaseline() style.Style`
  - inputs: `fwtheme.DisabledFG()`
  - output: disabled baseline style
- `resolveTabStyle(index int) style.Style`
  - inputs:
    - tab state
    - shared explicit `TabStyle`
    - explicit `ActiveTabStyle` for active state
    - explicit `DisabledTabStyle` for disabled state
    - resolved baselines
  - output: final tab header style for one visible tab

No changes are required in:

- event handling
- layout computation
- intent types
- app-level theme bootstrap

## Compatibility

The behavior is intentionally conservative.

- Existing apps with explicit tab styling should continue to render almost identically.
- Hosts with working theme initialization should keep using theme-provided tab styles.
- Hosts with missing or incomplete theme initialization will now get a readable selected state instead of a confusing near-default result.

The main visible change is in previously broken integrations, not in already-correct ones.

## Error Handling

Theme lookup failure is not treated as an error condition. Missing theme data is expected input and should transparently fall through to the next fallback layer.

No warnings, logs, or panics are introduced for this path.

## Testing

Add or adjust tests in `ui/components/tabs/tabs_test.go` to cover exact invariants:

1. no theme initialization
   - simulate missing style bridge so `style.GetStyle("tabs", "select")` returns empty
   - verify active tab resolves to `BG = fwtheme.Select()`, `FG = fwtheme.BG()`, `Bold = true`
   - verify disabled tab resolves to `FG = fwtheme.DisabledFG()` when available
   - verify normal tab keeps `FG = NoColor` and `BG = NoColor` at the baseline layer

2. partial theme availability
   - semantic colors available but no `tabs/select` style
   - verify the component uses semantic synthesis, not local reverse fallback
   - assert `IsReverse() == false` for this case
   - verify semantic synthesis is skipped when either `fwtheme.Select()` or `fwtheme.BG()` is unavailable

3. local fallback path
   - simulate both missing style bridge and missing semantic colors
   - verify active tab resolves to `Reverse = true`, `Bold = true`, `FG = NoColor`, `BG = NoColor`
   - verify `TabStyle` and `ActiveTabStyle` color overrides do not replace the reverse fallback colors

4. explicit caller styles
   - caller-supplied `TabStyle` and `ActiveTabStyle` still participate in composition
   - when the active baseline comes from `tabs/select`, final `FG` equals the theme-selected `FG` even if `TabStyle.FG` or `ActiveTabStyle.FG` is set
   - when the active baseline comes from semantic synthesis, final `FG` equals `fwtheme.BG()` even if `TabStyle.FG` or `ActiveTabStyle.FG` is set

5. state precedence and neutrality of non-selected elements
   - disabled wins over active if a broken test fixture creates that overlap
   - divider and unselected tabs do not inherit selected background behavior
   - hotkey rendering does not add a separate style layer

## Verification

Minimum verification after implementation:

- `go test ./ui/components/tabs ./ui ./examples/tabs ./framework`

Manual spot-check after implementation:

- run the tabs example in an environment with normal app theme initialization
- run a focused test or small harness where `tabs` is created without theme bootstrap
- confirm selected labels remain readable in both cases

## Open Questions Resolved

- Responsibility lives in the `tabs` component, not in app bootstrap.
- The component must cover both runtime dependency fallback and visual readability fallback.
- When no theme exists, fallback should follow terminal defaults and use the minimum highlighting needed for clarity.
