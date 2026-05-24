package image

import (
	stdimage "image"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey         = "key"
	propID          = "id"
	propAlt         = "alt"
	propDataURI     = "dataURI"
	propSourceImage = "sourceImage"
	propRGBA        = "rgba"
	propPixelWidth  = "pixelWidth"
	propPixelHeight = "pixelHeight"
	propWidthCells  = "widthCells"
	propHeightCells = "heightCells"
	propStyle       = "style"
)

// VNode is the immutable description of an Image component.
type VNode struct {
	*rtui.ElementVNode

	key         string
	id          string
	alt         string
	dataURI     string
	sourceImage stdimage.Image
	rgba        []byte
	pixelWidth  int
	pixelHeight int
	widthCells  int
	heightCells int
	imageStyle  style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Image VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("image"),
		alt:          "image",
	}
}

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	v.Props()[propKey] = key
	return v
}

func (v *VNode) SetID(id string) rtui.VNode {
	v.id = id
	v.Props()[propID] = id
	return v
}

func (v *VNode) SetAlt(alt string) *VNode {
	v.alt = alt
	v.Props()[propAlt] = alt
	return v
}

func (v *VNode) SetDataURI(dataURI string) *VNode {
	v.dataURI = dataURI
	v.Props()[propDataURI] = dataURI
	return v
}

func (v *VNode) SetSourceImage(img stdimage.Image) *VNode {
	v.sourceImage = img
	v.Props()[propSourceImage] = img
	return v
}

func (v *VNode) SetRGBA(rgba []byte, width, height int) *VNode {
	v.rgba = append([]byte(nil), rgba...)
	v.pixelWidth = width
	v.pixelHeight = height
	v.Props()[propRGBA] = append([]byte(nil), rgba...)
	v.Props()[propPixelWidth] = width
	v.Props()[propPixelHeight] = height
	return v
}

func (v *VNode) SetSize(widthCells, heightCells int) *VNode {
	v.widthCells = widthCells
	v.heightCells = heightCells
	v.Props()[propWidthCells] = widthCells
	v.Props()[propHeightCells] = heightCells
	return v
}

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.imageStyle = s
	v.Props()[propStyle] = s
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}
