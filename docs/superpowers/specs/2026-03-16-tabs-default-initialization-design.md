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

### Fallback Trigger Rules

Fallback is evaluated at render time for each tab header.

1. If the caller supplied an explicit tab style, use it as part of the composed result.
2. If the theme exposes `tabs/select` or related component-level styles, prefer those.
3. If the theme exposes semantic colors but not `tabs/select`, synthesize a minimal selected style from semantic colors.
4. If no usable theme data exists, use component-local fallback defaults.

This keeps the component usable in three host states:

- fully themed host
- partially themed host
- host with no theme initialization

### Local Default Style Model

The local fallback must follow terminal defaults as much as possible.

#### Normal Tab

- leave background unset
- leave foreground unset unless needed to preserve existing behavior
- keep current spacing, divider, and label logic unchanged

#### Active Tab

- prefer `Reverse(true)` as the primary signal
- add `Bold(true)` when needed to keep the active tab visually distinct
- avoid fixed color values
- if theme semantic colors are available, a synthesized selected style may use them, but the local fallback remains the final safety net

#### Disabled Tab

- use a light weakening effect such as `Faint(true)` when available
- do not hard-code gray or other fixed color values

#### Divider And Hotkeys

- divider remains neutral and does not participate in selected background filling
- hotkey emphasis may remain, but it must not reduce active-tab contrast

### Style Composition Order

The final style order remains explicit and stable:

1. component-local baseline emphasis for the current state
2. shared explicit tab style from the caller
3. theme-derived component style
4. explicit per-state style from the caller
5. selected-state readability protection

The readability protection rule is narrow:

- if a tab is active and the composed foreground color would undermine contrast against the selected background or reverse state, preserve the high-contrast selected foreground instead of blindly applying the caller's foreground override

This preserves caller customization while preventing the specific failure mode that made selected labels hard to read.

### Implementation Shape

The change should stay inside the tabs rendering helpers rather than spreading theme checks across unrelated code paths.

Expected units:

- a helper that resolves theme-derived selected styling
- a helper that resolves component-local fallback styles when theme data is unavailable
- a single composition path for normal, active, and disabled tab header styles

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

Add or adjust tests in `ui/components/tabs/tabs_test.go` to cover:

1. no theme initialization
   - render a normal tab, active tab, and disabled tab
   - verify active styling includes a stable readable fallback without depending on `InitTheme()`

2. partial theme availability
   - semantic colors available but no `tabs/select` style
   - verify the component synthesizes selected styling and stays readable

3. explicit caller styles
   - caller-supplied `TabStyle` and `ActiveTabStyle` still participate in composition
   - selected foreground protection prevents unreadable active text

4. neutrality of non-selected elements
   - divider and unselected tabs do not inherit selected background behavior

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
