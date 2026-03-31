package main

import (
	gallerydemo "github.com/wwsheng009/mint/examples/charts_gallery_demo/gallery"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(ChartsGalleryDemo,
		ui.WithWidth(80),
		ui.WithHeight(24),
		ui.WithTitle("Charts Gallery Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func ChartsGalleryDemo() ui.VNode {
	return gallerydemo.Build()
}
