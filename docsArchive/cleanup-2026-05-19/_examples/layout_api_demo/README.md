# Layout API Demo

Demonstrates the full box-model layout API: padding, margin, flex, gap, and text alignment.

Consolidated from `box_model_demo`, `elegant_api_demo`, and `final_box_demo`.

## Run

```bash
go run ./examples/layout_api_demo
```

## Sections

1. **Padding** — `PaddingAll`, `PaddingH`, `PaddingAll` on Text, asymmetric `PaddingH`
2. **Margin** — `MarginV`, `MarginH`, `Margin` (4-sided), `MarginAll`
3. **Flex + TextAlign** — `Flex(1)` distribution with `AlignStart / AlignCenter / AlignEnd`
4. **Gap** — `VStackBuilder.Gap(1)` vs per-button `MarginV`
