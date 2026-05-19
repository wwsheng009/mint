# Cleanup Archive 2026-05-19

This archive contains documentation and examples moved out of the active SDK surface during the 2026-05-19 cleanup.

## Why These Files Were Archived

The active repository had current SDK-facing docs mixed with:

- historical fix reports
- one-off implementation notes
- design alternatives
- investigation logs
- temporary debug probes
- very small examples duplicated by broader maintained demos

Those materials remain useful for archaeology, but they should not be part of the default SDK learning path.

## Contents

| Path | Description |
|---|---|
| `docs/` | Archived documents moved from `docs/` |
| `_examples/` | Archived example directories moved from `examples/`; the leading underscore keeps Go tooling from treating them as active packages |
| `_framework_docs/` | Archived framework design drafts, implementation plans, review notes, and historical reports moved from `framework/docs/`; the leading underscore keeps Go tooling from treating embedded `.go` snippets as active packages |
| `moved-docs.txt` | Relative source paths moved from `docs/` |
| `moved-examples.txt` | Relative source paths moved from `examples/` |
| `moved-framework-docs.txt` | Relative source paths moved from `framework/docs/` |
| `moved-root-docs.txt` | Root-level historical markdown files moved into `root/` |
| `root/` | Archived root-level fix/status reports |

## Restoration

If a file is needed again, restore it deliberately and update the current docs index. Do not move archived materials back into `docs/` or `examples/` without checking whether they are still accurate against the current source.
