# Tabs Default Initialization Implementation Plan

## Scope

Implement the approved tabs default-initialization design in the header rendering path only.

## Work Items

1. Add regression coverage in `ui/components/tabs/tabs_test.go`
   - no style bridge: active state falls back to semantic colors
   - semantic colors missing: active state falls back to reverse + bold
   - disabled semantic color missing: disabled state falls back to italic
   - explicit `TabStyle` and `ActiveTabStyle` cannot break active contrast
   - component `tabs/select` branches: FG-only, BG-only, reverse-only, non-protectable-but-non-empty

2. Refactor tabs style resolution in `ui/components/tabs/instance.go`
   - introduce active baseline source tracking
   - split active and disabled baseline resolution helpers
   - apply state-specific merge order and readability protection rules

3. Verify behavior
   - targeted package tests for `ui/components/tabs`
   - broader regression run for `./ui`, `./examples/tabs`, and `./framework`

## Risks

- Theme state is global, so tests must restore style getter and theme manager state.
- Reverse-only handling must not accidentally preserve caller colors.

## Exit Criteria

- All new tabs-style branches are covered by deterministic tests.
- `go test ./ui/components/tabs ./ui ./examples/tabs ./framework` passes.
