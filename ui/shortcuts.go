// Package ui provides convenient shortcuts for common UI components
//
// This package re-exports:
// 1. Quick shortcut functions for common use cases (e.g., Text(), Button(), Input())
// 2. Builder factory functions (e.g., NewButtonBuilder(), NewInputBuilder())
//
// Usage examples:
//
// Quick shortcuts:
//
//	ui.Text("Hello")
//	ui.Button("Click Me")
//	ui.Input("Placeholder")
//
// Full Builder patterns:
//
//	ui.NewTextBuilder("Hello").Bold().FgColor("red").Build()
//	ui.NewButtonBuilder("Click").Primary().Large().OnPress(intent).Build()
//	ui.NewInputBuilder().Placeholder("...").Value("x").OnChange(intent).Build()
package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"

	"github.com/wwsheng009/mint/ui/components/absolute"
	"github.com/wwsheng009/mint/ui/components/alert"
	anchorcomp "github.com/wwsheng009/mint/ui/components/anchor"
	"github.com/wwsheng009/mint/ui/components/badge"
	"github.com/wwsheng009/mint/ui/components/breadcrumb"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/cascader"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/clock"
	"github.com/wwsheng009/mint/ui/components/collapse"
	"github.com/wwsheng009/mint/ui/components/datepicker"
	"github.com/wwsheng009/mint/ui/components/descriptions"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/drawer"
	"github.com/wwsheng009/mint/ui/components/empty"
	formcomp "github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/grid"
	imagecomp "github.com/wwsheng009/mint/ui/components/image"
	"github.com/wwsheng009/mint/ui/components/input"
	layoutcomp "github.com/wwsheng009/mint/ui/components/layout"
	"github.com/wwsheng009/mint/ui/components/list"
	"github.com/wwsheng009/mint/ui/components/modal"
	"github.com/wwsheng009/mint/ui/components/notification"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
	"github.com/wwsheng009/mint/ui/components/pagination"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/popconfirm"
	"github.com/wwsheng009/mint/ui/components/popover"
	"github.com/wwsheng009/mint/ui/components/progress"
	"github.com/wwsheng009/mint/ui/components/radio"
	"github.com/wwsheng009/mint/ui/components/rate"
	"github.com/wwsheng009/mint/ui/components/result"
	rowcolcomp "github.com/wwsheng009/mint/ui/components/rowcol"
	"github.com/wwsheng009/mint/ui/components/scrollview"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
	"github.com/wwsheng009/mint/ui/components/skeleton"
	"github.com/wwsheng009/mint/ui/components/slider"
	spacecomp "github.com/wwsheng009/mint/ui/components/space"
	"github.com/wwsheng009/mint/ui/components/spin"
	"github.com/wwsheng009/mint/ui/components/statistic"
	"github.com/wwsheng009/mint/ui/components/statusbar"
	"github.com/wwsheng009/mint/ui/components/steps"
	switchcomp "github.com/wwsheng009/mint/ui/components/switch"
	"github.com/wwsheng009/mint/ui/components/table"
	"github.com/wwsheng009/mint/ui/components/tabs"
	"github.com/wwsheng009/mint/ui/components/tag"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/textarea"
	"github.com/wwsheng009/mint/ui/components/timeline"
	"github.com/wwsheng009/mint/ui/components/timepicker"
	"github.com/wwsheng009/mint/ui/components/toast"
	"github.com/wwsheng009/mint/ui/components/tooltip"
	"github.com/wwsheng009/mint/ui/components/transfer"
	"github.com/wwsheng009/mint/ui/components/treeview"
	"github.com/wwsheng009/mint/ui/components/virtuallist"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

// =============================================================================
// Builder Factory Functions - Re-exported from component packages
// =============================================================================

// Form Components
func NewInputBuilder() *input.Builder {
	return input.NewBuilder()
}

func NewSearchInputBuilder() *input.Builder {
	return input.NewBuilder().Search()
}

func NewTextareaBuilder() *textarea.Builder {
	return textarea.NewBuilder()
}

func NewDatePickerBuilder() *datepicker.Builder {
	return datepicker.NewBuilder()
}

func NewFormBuilder() formcomp.Builder {
	return formcomp.NewBuilder()
}

func NewTimePickerBuilder() *timepicker.Builder {
	return timepicker.NewBuilder()
}

func NewAnchorBuilder() *anchorcomp.Builder {
	return anchorcomp.NewBuilder()
}

func NewCascaderBuilder() *cascader.Builder {
	return cascader.NewBuilder()
}

func NewTransferBuilder() *transfer.Builder {
	return transfer.NewBuilder()
}

func NewSpaceBuilder() *spacecomp.Builder {
	return spacecomp.NewBuilder()
}

func NewLayoutBuilder() *layoutcomp.Builder {
	return layoutcomp.NewBuilder()
}

func NewRowBuilder() *rowcolcomp.RowBuilder {
	return rowcolcomp.NewRowBuilder()
}

func NewColBuilder() *rowcolcomp.ColBuilder {
	return rowcolcomp.NewColBuilder()
}

func NewCheckboxBuilder() *checkbox.Builder {
	return checkbox.NewBuilder()
}

func NewCheckboxGroupBuilder(options []checkbox.Option) *checkbox.GroupBuilder {
	return checkbox.NewGroupBuilder(options)
}

func NewRadioBuilder() *radio.Builder {
	return radio.NewBuilder()
}

func NewSwitchBuilder() *switchcomp.Builder {
	return switchcomp.NewBuilder()
}

func NewRadioGroupBuilder(options []radio.Option) *radio.GroupBuilder {
	return radio.NewGroupBuilder(options)
}

// Button Components
func NewButtonBuilder(label string) *button.Builder {
	return button.NewBuilder(label)
}

// Text Components
func NewTextBuilder(content string) *text.Builder {
	return text.NewBuilder(content)
}

func NewWrapBuilder() *wrap.Builder {
	return wrap.NewBuilder()
}

func NewGridBuilder() *grid.Builder {
	return grid.NewBuilder()
}

func NewAbsoluteBuilder(child rtui.VNode) *absolute.Builder {
	return absolute.NewBuilder(child)
}

// Container Components

// NewBorderBuilder creates a border builder.
func NewPanelBuilder() *panel.Builder {
	return panel.NewBuilder()
}

func NewPaginationBuilder() *pagination.Builder {
	return pagination.NewBuilder()
}

func NewScrollViewBuilder() *scrollview.Builder {
	return scrollview.NewBuilder()
}

func NewEmptyBuilder() *empty.Builder {
	return empty.NewBuilder()
}

func NewImageBuilder() *imagecomp.Builder {
	return imagecomp.NewBuilder()
}

func NewAlertBuilder(message string) *alert.Builder {
	return alert.NewBuilder(message)
}

func NewStatusBarBuilder() *statusbar.Builder {
	return statusbar.NewBuilder()
}

func StatusBar(left, center, right []statusbar.Section) rtui.VNode {
	return statusbar.NewBuilder().
		LeftSections(left...).
		CenterSections(center...).
		RightSections(right...).
		Build()
}

func StatusBarWithHelp(theme statusbar.Theme, helpFallback string, left, center, right []statusbar.Section) rtui.VNode {
	return StatusBarWithHelpMode(theme, helpFallback, statusbar.HelpDisplayInline, left, center, right)
}

func StatusBarWithHelpMode(theme statusbar.Theme, helpFallback string, mode statusbar.HelpDisplayMode, left, center, right []statusbar.Section) rtui.VNode {
	return statusbar.NewBuilder().
		Theme(theme).
		HelpFallback(helpFallback).
		HelpDisplayMode(mode).
		LeftSections(left...).
		CenterSections(center...).
		RightSections(right...).
		BuildWithHelp()
}

// Data Display Components
func NewListBuilder() *list.Builder {
	return list.NewBuilder()
}

func NewBadgeBuilder(label string) *badge.Builder {
	return badge.NewBuilder(label)
}

func NewTagBuilder(text string) *tag.Builder {
	return tag.NewBuilder(text)
}

func NewCollapseBuilder() *collapse.Builder {
	return collapse.NewBuilder()
}

func NewDescriptionsBuilder() *descriptions.Builder {
	return descriptions.NewBuilder()
}

func NewPopoverBuilder(content rtui.VNode) *popover.Builder {
	return popover.NewBuilder(content)
}

func NewPopconfirmBuilder(content rtui.VNode) *popconfirm.Builder {
	return popconfirm.NewBuilder(content)
}

func NewStatisticBuilder() *statistic.Builder {
	return statistic.NewBuilder()
}

func NewSkeletonBuilder() *skeleton.Builder {
	return skeleton.NewBuilder()
}

func NewTableBuilder() *table.Builder {
	return table.NewBuilder()
}

func NewTreeViewBuilder() *treeview.Builder {
	return treeview.NewBuilder()
}

func NewVirtualListBuilder() *virtuallist.Builder {
	return virtuallist.NewBuilder()
}

func NewSliderBuilder() *slider.Builder {
	return slider.NewBuilder()
}

func NewRateBuilder() *rate.Builder {
	return rate.NewBuilder()
}

func NewOptionGroupBuilder(options []optiongroup.Option) *optiongroup.Builder {
	return optiongroup.NewBuilder(options)
}

func NewTabsBuilder() *tabs.Builder {
	return tabs.NewBuilder()
}

func NewBreadcrumbBuilder() *breadcrumb.Builder {
	return breadcrumb.NewBuilder()
}

func NewStepsBuilder() *steps.Builder {
	return steps.NewBuilder()
}

func NewTimelineBuilder() *timeline.Builder {
	return timeline.NewBuilder()
}

func NewSelectBuilder() *selectcomp.Builder {
	return selectcomp.NewBuilder()
}

// Navigation Components
// Note: NewTabsBuilder() is declared above

// Divider Components
func NewDividerBuilder() *divider.Builder {
	return divider.NewBuilder()
}

// Progress Components
func NewProgressBuilder() *progress.Builder {
	return progress.NewBuilder()
}

func NewClockBuilder() *clock.Builder {
	return clock.NewBuilder()
}

func NewSpinBuilder() *spin.Builder {
	return spin.NewBuilder()
}

func NewResultBuilder() *result.Builder {
	return result.NewBuilder()
}

func NewNotificationBuilder(message string) *notification.Builder {
	return notification.NewBuilder(message)
}

// Modal Components
func NewModalBuilder() *modal.Builder {
	return modal.NewBuilder()
}

func NewDrawerBuilder() *drawer.Builder {
	return drawer.NewBuilder()
}

// Tooltip Components
func NewTooltipBuilder(content rtui.VNode, tooltipText string) *tooltip.Builder {
	return tooltip.NewBuilder(content, tooltipText)
}

func NewToastBuilder(message string) *toast.ToastBuilder {
	return toast.NewToastBuilder(message)
}

// =============================================================================
// Common Type Re-exports
// =============================================================================

// Button Types
type (
	ButtonVariant    = button.Variant
	ButtonSize       = button.Size
	ButtonFocusStyle = button.FocusStyle
)

const (
	ButtonVariantDefault   = button.VariantDefault
	ButtonVariantPrimary   = button.VariantPrimary
	ButtonVariantSecondary = button.VariantSecondary
	ButtonVariantDanger    = button.VariantDanger
	ButtonVariantSuccess   = button.VariantSuccess

	ButtonSmall  = button.SizeSmall
	ButtonMedium = button.SizeMedium
	ButtonLarge  = button.SizeLarge

	FocusStyleReverse   = button.FocusStyleReverse
	FocusStyleUnderline = button.FocusStyleUnderline
	FocusStyleBracket   = button.FocusStyleBracket
	FocusStyleBold      = button.FocusStyleBold
)

// Divider Types
type (
	DividerStyle       = divider.Style
	DividerOrientation = divider.Orientation
)

const (
	DividerSolid  = divider.StyleSolid
	DividerDashed = divider.StyleDashed
	DividerDotted = divider.StyleDotted
	DividerDouble = divider.StyleDouble

	HorizontalDivider = divider.Horizontal
	VerticalDivider   = divider.Vertical
)

// StatusBar Types
type (
	StatusBarSection           = statusbar.Section
	StatusBarTheme             = statusbar.Theme
	StatusBarOverflow          = statusbar.OverflowMode
	StatusBarHelpDisplay       = statusbar.HelpDisplayMode
	StatusBarTooltipPlacement  = statusbar.TooltipPlacement
	StatusBarTooltipArrowStyle = statusbar.TooltipArrowStyle
)

const (
	StatusBarOverflowEllipsis = statusbar.OverflowEllipsis
	StatusBarOverflowClip     = statusbar.OverflowClip

	StatusBarHelpInline  = statusbar.HelpDisplayInline
	StatusBarHelpOverlay = statusbar.HelpDisplayOverlay
	StatusBarHelpBoth    = statusbar.HelpDisplayBoth

	StatusBarTooltipAuto   = statusbar.TooltipPlacementAuto
	StatusBarTooltipTop    = statusbar.TooltipPlacementTop
	StatusBarTooltipBottom = statusbar.TooltipPlacementBottom

	StatusBarTooltipArrowDefault = statusbar.TooltipArrowStyleDefault
	StatusBarTooltipArrowSharp   = statusbar.TooltipArrowStyleSharp
	StatusBarTooltipArrowRounded = statusbar.TooltipArrowStyleRounded
)

// Grid Types
type (
	GridDimension = grid.Dimension
	GridFixed     = grid.Fixed
	GridFlex      = grid.Flex
	GridAuto      = grid.Auto
)

// Grid dimension helper functions - avoid conflict with ui.Flex function
func FixedDim(size int) grid.Fixed { return grid.Fixed(size) }
func FlexDim(factor int) grid.Flex { return grid.Flex{Factor: factor} }
func AutoDim() grid.Auto           { return grid.Auto{} }

// Absolute Position Types
type PositionValue = absolute.PositionValue
type Anchor = absolute.Anchor

// Tab Types
type TabPosition = tabs.TabPosition
type TabItem = tabs.TabItem
type FormLayout = formcomp.FormLayout
type InputType = input.Type
type AlertType = alert.AlertType
type BreadcrumbItem = breadcrumb.Item
type BadgeStatus = badge.Status
type DrawerPlacement = drawer.Placement
type NotificationType = notification.NotificationType
type NotificationPlacement = notification.Placement
type CollapseItem = collapse.Item
type DescriptionsItem = descriptions.Item
type DescriptionsLayout = descriptions.Layout
type TagColor = tag.TagColor
type TimelineItem = timeline.Item
type TimelineStatus = timeline.Status
type ResultStatus = result.Status
type SkeletonShape = skeleton.Shape
type SpinSize = spin.Size
type PopoverPlacement = popover.Placement
type PopoverTriggerMode = popover.TriggerMode
type PopconfirmPlacement = popconfirm.Placement
type PopconfirmTriggerMode = popconfirm.TriggerMode
type StatisticTrend = statistic.Trend
type StepsDirection = steps.Direction
type StepsStatus = steps.Status
type StepsItem = steps.Item

const (
	FormVertical   = formcomp.LayoutVertical
	FormHorizontal = formcomp.LayoutHorizontal
	FormInline     = formcomp.LayoutInline

	InputText     = input.TypeText
	InputPassword = input.TypePassword
	InputNumber   = input.TypeNumber
	InputEmail    = input.TypeEmail

	AlertInfo    = alert.AlertInfo
	AlertSuccess = alert.AlertSuccess
	AlertWarning = alert.AlertWarning
	AlertError   = alert.AlertError

	TabPositionTop    = tabs.TabPositionTop
	TabPositionBottom = tabs.TabPositionBottom
	TabPositionLeft   = tabs.TabPositionLeft
	TabPositionRight  = tabs.TabPositionRight

	DrawerRight  = drawer.PlacementRight
	DrawerLeft   = drawer.PlacementLeft
	DrawerTop    = drawer.PlacementTop
	DrawerBottom = drawer.PlacementBottom

	NotificationInfo        = notification.NotificationInfo
	NotificationSuccess     = notification.NotificationSuccess
	NotificationWarning     = notification.NotificationWarning
	NotificationError       = notification.NotificationError
	NotificationTopRight    = notification.PlacementTopRight
	NotificationTopLeft     = notification.PlacementTopLeft
	NotificationBottomRight = notification.PlacementBottomRight
	NotificationBottomLeft  = notification.PlacementBottomLeft

	BadgeStatusDefault    = badge.StatusDefault
	BadgeStatusPrimary    = badge.StatusPrimary
	BadgeStatusSuccess    = badge.StatusSuccess
	BadgeStatusWarning    = badge.StatusWarning
	BadgeStatusError      = badge.StatusError
	BadgeStatusProcessing = badge.StatusProcessing

	TagColorDefault    = tag.ColorDefault
	TagColorPrimary    = tag.ColorPrimary
	TagColorSuccess    = tag.ColorSuccess
	TagColorWarning    = tag.ColorWarning
	TagColorError      = tag.ColorError
	TagColorProcessing = tag.ColorProcessing

	DescriptionsHorizontal = descriptions.LayoutHorizontal
	DescriptionsVertical   = descriptions.LayoutVertical

	TimelineStatusDefault = timeline.StatusDefault
	TimelineStatusSuccess = timeline.StatusSuccess
	TimelineStatusWarning = timeline.StatusWarning
	TimelineStatusError   = timeline.StatusError
	TimelineStatusPending = timeline.StatusPending

	ResultStatusInfo    = result.StatusInfo
	ResultStatusSuccess = result.StatusSuccess
	ResultStatusWarning = result.StatusWarning
	ResultStatusError   = result.StatusError
	ResultStatus403     = result.Status403
	ResultStatus404     = result.Status404
	ResultStatus500     = result.Status500

	SkeletonShapeSquare = skeleton.ShapeSquare
	SkeletonShapeRound  = skeleton.ShapeRound

	SpinSizeSmall   = spin.SizeSmall
	SpinSizeDefault = spin.SizeDefault
	SpinSizeLarge   = spin.SizeLarge

	PopoverPlacementAuto        = popover.PlacementAuto
	PopoverPlacementTop         = popover.PlacementTop
	PopoverPlacementTopLeft     = popover.PlacementTopLeft
	PopoverPlacementTopRight    = popover.PlacementTopRight
	PopoverPlacementBottom      = popover.PlacementBottom
	PopoverPlacementBottomLeft  = popover.PlacementBottomLeft
	PopoverPlacementBottomRight = popover.PlacementBottomRight

	PopoverTriggerClick  = popover.TriggerClick
	PopoverTriggerHover  = popover.TriggerHover
	PopoverTriggerManual = popover.TriggerManual

	PopconfirmPlacementAuto        = popconfirm.PlacementAuto
	PopconfirmPlacementTop         = popconfirm.PlacementTop
	PopconfirmPlacementTopLeft     = popconfirm.PlacementTopLeft
	PopconfirmPlacementTopRight    = popconfirm.PlacementTopRight
	PopconfirmPlacementBottom      = popconfirm.PlacementBottom
	PopconfirmPlacementBottomLeft  = popconfirm.PlacementBottomLeft
	PopconfirmPlacementBottomRight = popconfirm.PlacementBottomRight

	PopconfirmTriggerClick  = popconfirm.TriggerClick
	PopconfirmTriggerHover  = popconfirm.TriggerHover
	PopconfirmTriggerManual = popconfirm.TriggerManual

	StatisticTrendNone = statistic.TrendNone
	StatisticTrendUp   = statistic.TrendUp
	StatisticTrendDown = statistic.TrendDown

	StepsHorizontal = steps.DirectionHorizontal
	StepsVertical   = steps.DirectionVertical

	StepsStatusAuto    = steps.StatusAuto
	StepsStatusWait    = steps.StatusWait
	StepsStatusProcess = steps.StatusProcess
	StepsStatusFinish  = steps.StatusFinish
	StepsStatusError   = steps.StatusError
)

// Table Types
type TableColumn = table.TableColumn

// Tree Types
type TreeNode = treeview.TreeNode
type TreeSelectionMode = treeview.SelectionMode

const (
	TreeSelectionNone     = treeview.SelectionNone
	TreeSelectionSingle   = treeview.SelectionSingle
	TreeSelectionMultiple = treeview.SelectionMultiple
)

// Toast Types
type ToastType = toast.ToastType

const (
	ToastTypeInfo    = toast.ToastInfo
	ToastTypeSuccess = toast.ToastSuccess
	ToastTypeWarning = toast.ToastWarning
	ToastTypeError   = toast.ToastError
)

// Select Types
type SelectOption = selectcomp.Option
type OptionGroupOption = optiongroup.Option
type OptionGroupMode = optiongroup.SelectMode
type OptionGroupOrientation = optiongroup.Orientation
type AnchorItem = anchorcomp.Item
type CascaderOption = cascader.Option
type TransferItem = transfer.Item
type SpaceDirection = spacecomp.Direction

const (
	OptionGroupSingle     = optiongroup.ModeSingle
	OptionGroupMultiple   = optiongroup.ModeMultiple
	OptionGroupVertical   = optiongroup.OrientationVertical
	OptionGroupHorizontal = optiongroup.OrientationHorizontal

	SpaceHorizontal = spacecomp.DirectionHorizontal
	SpaceVertical   = spacecomp.DirectionVertical

	SpaceSizeSmall  = spacecomp.SizeSmall
	SpaceSizeMiddle = spacecomp.SizeMiddle
	SpaceSizeLarge  = spacecomp.SizeLarge
)

// Radio Types
type (
	CheckboxOption      = checkbox.Option
	CheckboxOrientation = checkbox.Orientation
	RadioOption         = radio.Option
	RadioOrientation    = radio.Orientation
)

const (
	CheckboxVertical   = checkbox.OrientationVertical
	CheckboxHorizontal = checkbox.OrientationHorizontal

	RadioVertical   = radio.OrientationVertical
	RadioHorizontal = radio.OrientationHorizontal
)

// NewSelectOption creates a new select option with value and label
func NewSelectOption(value, label string) selectcomp.Option {
	return selectcomp.Option{Value: value, Label: label}
}

// NewOptionGroupOption creates a new option group option with value and label.
func NewOptionGroupOption(value, label string) optiongroup.Option {
	return optiongroup.Option{Value: value, Label: label}
}

// NewAnchorItem creates an anchor item with optional child items.
func NewAnchorItem(key, title string, children ...anchorcomp.Item) anchorcomp.Item {
	return anchorcomp.NewItem(key, title, children...)
}

// NewCascaderOption creates a cascader option with optional child options.
func NewCascaderOption(value, label string, children ...cascader.Option) cascader.Option {
	return cascader.Node(value, label, children...)
}

// NewTransferItem creates a transfer item with key and title.
func NewTransferItem(key, title string) transfer.Item {
	return transfer.NewItem(key, title)
}

// NewBreadcrumbItem creates a breadcrumb item with the given label.
func NewBreadcrumbItem(label string) breadcrumb.Item {
	return breadcrumb.Crumb(label)
}

// NewCollapseItem creates a collapse item with the given header and content.
func NewCollapseItem(header string, content rtui.VNode) collapse.Item {
	return collapse.Section(header, content)
}

// NewDescriptionsItem creates a descriptions item with the given label and content.
func NewDescriptionsItem(label string, content rtui.VNode) descriptions.Item {
	return descriptions.Entry(label, content)
}

// NewDescriptionsField creates a descriptions item from text.
func NewDescriptionsField(label, value string) descriptions.Item {
	return descriptions.Field(label, value)
}

// NewDescriptionsValue creates a descriptions item from an arbitrary value.
func NewDescriptionsValue(label string, value interface{}) descriptions.Item {
	return descriptions.Value(label, value)
}

// NewSensitiveDescriptionsItem creates a masked descriptions item.
func NewSensitiveDescriptionsItem(label string, value interface{}) descriptions.Item {
	return descriptions.SensitiveField(label, value)
}

// NewTimelineItem creates a timeline item with content text.
func NewTimelineItem(content string) timeline.Item {
	return timeline.Event(content)
}

// NewResult creates a result builder with a status.
func NewResult(status result.Status) *result.VNode {
	return result.New().SetStatus(status)
}

// NewPopover creates a popover builder with anchor content.
func NewPopover(content rtui.VNode) *popover.VNode {
	return popover.New(content)
}

// NewPopconfirm creates a popconfirm builder around an anchor content node.
func NewPopconfirm(content rtui.VNode) *popconfirm.VNode {
	return popconfirm.New(content)
}

// NewStatistic creates a statistic builder with title and value.
func NewStatistic(title string, value interface{}) *statistic.VNode {
	return statistic.New().SetTitle(title).SetValue(value)
}

// NewBadge creates a badge builder seed with the given label.
func NewBadge(label string) *badge.VNode {
	return badge.New(label)
}

// NewEmpty creates an empty-state builder seed.
func NewEmpty() *empty.VNode {
	return empty.New()
}

// NewForm creates a form builder seed with the given key.
func NewForm(key string) *formcomp.VNode {
	return formcomp.NewForm(key)
}

// NewFormItem creates a FormItem wrapper around a field component.
func NewFormItem(field string, child rtui.VNode) *formcomp.ItemBuilder {
	return formcomp.NewItem(field, child)
}

// NewSlider creates a slider builder seed.
func NewSlider() *slider.VNode {
	return slider.NewSlider()
}

// NewRate creates a rate builder seed.
func NewRate() *rate.VNode {
	return rate.NewRate()
}

// NewOptionGroup creates an option group builder seed.
func NewOptionGroup(options []optiongroup.Option) *optiongroup.VNode {
	return optiongroup.New(options)
}

// NewAlert creates an alert builder seed with the given message.
func NewAlert(message string) *alert.VNode {
	return alert.New().SetMessage(message)
}

// NewNotification creates a notification builder seed with the given message.
func NewNotification(message string) *notification.VNode {
	return notification.New().SetMessage(message)
}

// NewTag creates a tag builder seed with the given text.
func NewTag(text string) *tag.VNode {
	return tag.New(text)
}

// NewSpin creates a spin builder seed in spinning state.
func NewSpin() *spin.VNode {
	return spin.New()
}

// NewToast creates a toast builder seed with the given message.
func NewToast(message string) *toast.ToastVNode {
	return toast.NewToast(message)
}

// NewStepsItem creates a steps item with the given title.
func NewStepsItem(title string) steps.Item {
	return steps.Step(title)
}

// NewRadioOption creates a new radio option with value and label.
func NewRadioOption(value, label string) radio.Option {
	return radio.Option{Value: value, Label: label}
}

// NewCheckboxOption creates a new checkbox option with value and label.
func NewCheckboxOption(value, label string) checkbox.Option {
	return checkbox.Option{Value: value, Label: label}
}

// =============================================================================
// NOTE: BorderStyle constants are in ui/layout.go (re-exported from rtui)
// =============================================================================
// NOTE: HStackBuilder, VStackBuilder are in ui/layout.go (runtime/ui.LayoutBuilder)
// NOTE: ModalBuilder, TooltipBuilder are in ui/layer.go (ui-specific implementations)

// =============================================================================
// Quick Shortcut Functions (for common use cases)
// =============================================================================

// Form Components shortcuts

// Form creates a form containing the provided children.
func Form(children ...rtui.VNode) rtui.VNode {
	return formcomp.NewBuilder().AddChildren(children...)
}

// Input creates an input field with placeholder
func Input(placeholder string) rtui.VNode {
	return input.NewBuilder().Placeholder(placeholder).Build()
}

// SearchInput creates a search input field with placeholder.
func SearchInput(placeholder string) rtui.VNode {
	return input.NewBuilder().Search().Placeholder(placeholder).Build()
}

// InputWithValue creates an input field with initial value and placeholder
func InputWithValue(placeholder, value string) rtui.VNode {
	return input.NewBuilder().Placeholder(placeholder).Value(value).Build()
}

// Textarea creates a textarea field with placeholder
func Textarea(placeholder string) rtui.VNode {
	return textarea.NewBuilder().Placeholder(placeholder).Build()
}

// TextareaWithValue creates a textarea with initial value
func TextareaWithValue(placeholder, value string) rtui.VNode {
	return textarea.NewBuilder().Placeholder(placeholder).Value(value).Build()
}

// Checkbox creates a checkbox
func Checkbox(label string, checked bool) rtui.VNode {
	return checkbox.NewBuilder().Label(label).Checked(checked).Build()
}

// CheckboxGroup creates a checkbox group builder.
func CheckboxGroup(options []checkbox.Option) *checkbox.GroupBuilder {
	return checkbox.NewGroupBuilder(options)
}

// Radio creates a radio button.
func Radio(label string, checked bool) rtui.VNode {
	return radio.NewBuilder().Label(label).Checked(checked).Build()
}

// Switch creates a switch.
func Switch(label string, checked bool) rtui.VNode {
	return switchcomp.NewBuilder().Label(label).Checked(checked).Build()
}

// RadioGroup creates a radio group builder.
func RadioGroup(options []radio.Option) *radio.GroupBuilder {
	return radio.NewGroupBuilder(options)
}

// Select creates a select dropdown with options
func Select(options []map[string]interface{}) rtui.VNode {
	// Convert options to selectcomp.Option format
	opts := make([]selectcomp.Option, len(options))
	for i, opt := range options {
		value, _ := opt["value"].(string)
		label, _ := opt["label"].(string)
		opts[i] = selectcomp.Option{Value: value, Label: label}
	}
	return selectcomp.NewBuilder().Options(opts).Build()
}

// Slider creates a slider builder.
func Slider() *slider.Builder {
	return slider.NewBuilder()
}

// Rate creates a rate builder.
func Rate() *rate.Builder {
	return rate.NewBuilder()
}

// OptionGroup creates an option group builder with the provided options.
func OptionGroup(options []optiongroup.Option) *optiongroup.Builder {
	return optiongroup.NewBuilder(options)
}

// Cascader creates a cascader with hierarchical options.
func Cascader(options []cascader.Option) rtui.VNode {
	return cascader.NewBuilder().Options(options).Build()
}

// AnchorNav creates an anchor navigation component from the provided items.
func AnchorNav(items []anchorcomp.Item) rtui.VNode {
	return anchorcomp.Of(items)
}

// Transfer creates a transfer component from the provided items.
func Transfer(items []transfer.Item) rtui.VNode {
	return transfer.Of(items)
}

// Space creates a horizontal spacing layout from the provided children.
func Space(children ...rtui.VNode) rtui.VNode {
	return spacecomp.NewBuilder().Children(children...).Build()
}

// Layout creates a layout shell with a single content section.
func Layout(content rtui.VNode) rtui.VNode {
	return layoutcomp.NewBuilder().Content(content).Build()
}

// Button shortcuts

// Button creates a button with label (no click handler)
// Note: This version does not support onClick. Use ButtonWithIntent for actions.
func Button(label string) rtui.VNode {
	return button.NewBuilder(label).Build()
}

// ButtonWithIntent creates a button with Intent (Fiber-first pattern)
func ButtonWithIntent(label string, pressIntent intent.Intent) rtui.VNode {
	return button.NewBuilder(label).OnPress(pressIntent).Build()
}

// Display Components shortcuts

// Text shortcuts

// Text creates a simple text VNode
func Text(content string) rtui.VNode {
	return text.T(content)
}

// Textf creates a formatted text VNode
// Note: actual formatting should be done by caller with fmt.Sprintf
func Textf(format string, args ...interface{}) rtui.VNode {
	return text.NewBuilder(format).Build()
}

// TextWithStyle creates a styled text VNode
func TextWithStyle(content string, s style.Style) rtui.VNode {
	return text.Styled(content, s)
}

// TextBold creates a bold text VNode
func TextBold(content string) rtui.VNode {
	return text.Bold(content)
}

// TextColored creates a colored text VNode
func TextColored(content string, fg style.Color) rtui.VNode {
	return text.Colored(content, fg)
}

// StatusBarText creates a plain status bar section.
func StatusBarText(content string) statusbar.Section {
	return statusbar.Text(content)
}

// StatusBarActionText creates a clickable plain status bar section.
func StatusBarActionText(content string, pressIntent intent.Intent) statusbar.Section {
	return statusbar.ActionText(content, pressIntent)
}

func StatusBarSections(sections ...statusbar.Section) []statusbar.Section {
	return statusbar.Sections(sections...)
}

// StatusBarBadge creates a highlighted status bar section.
func StatusBarBadge(content, fgColor, bgColor string) statusbar.Section {
	return statusbar.Badge(content, fgColor, bgColor)
}

// StatusBarActionBadge creates a clickable highlighted status bar section.
func StatusBarActionBadge(content, fgColor, bgColor string, pressIntent intent.Intent) statusbar.Section {
	return statusbar.ActionBadge(content, fgColor, bgColor, pressIntent)
}

// StatusBarHelp creates a help/tooltip text for a section.
func StatusBarHelp(section statusbar.Section, helpText string) statusbar.Section {
	return section.WithHelp(helpText)
}

// StatusBarThemeDefault returns the default status bar theme.
func StatusBarThemeDefault() statusbar.Theme {
	return statusbar.DefaultTheme()
}

// StatusBarThemeMuted returns the muted status bar theme.
func StatusBarThemeMuted() statusbar.Theme {
	return statusbar.MutedTheme()
}

// StatusBarThemeContrast returns the contrast status bar theme.
func StatusBarThemeContrast() statusbar.Theme {
	return statusbar.ContrastTheme()
}

func StatusBarWithTheme(theme statusbar.Theme, left, center, right []statusbar.Section) rtui.VNode {
	return statusbar.NewBuilder().
		Theme(theme).
		LeftSections(left...).
		CenterSections(center...).
		RightSections(right...).
		Build()
}

// TextAlign creates a text VNode with horizontal alignment
func TextAlign(content string, align string) rtui.VNode {
	var a rtui.Align
	switch align {
	case "center":
		a = rtui.AlignCenter
	case "right":
		a = rtui.AlignEnd
	default:
		a = rtui.AlignStart
	}
	return text.NewBuilder(content).TextAlign(a).Build()
}

// TextCenter creates a centered text VNode
func TextCenter(content string) rtui.VNode {
	return TextAlign(content, "center")
}

// TextRight creates a right-aligned text VNode
func TextRight(content string) rtui.VNode {
	return TextAlign(content, "right")
}

// Badge creates an inline badge with a numeric count.
func Badge(label string, count int) rtui.VNode {
	return badge.NewBuilder(label).Count(count).Build()
}

// Collapse creates a collapse component from items.
func Collapse(items []collapse.Item) rtui.VNode {
	return collapse.Of(items)
}

// Descriptions creates a descriptions component from items.
func Descriptions(items []descriptions.Item) rtui.VNode {
	return descriptions.Of(items)
}

// Timeline creates a timeline component from items.
func Timeline(items []timeline.Item) rtui.VNode {
	return timeline.Of(items)
}

// Result creates a result component with status, title, and subtitle.
func Result(status result.Status, title, subtitle string) rtui.VNode {
	return result.New().SetStatus(status).SetTitle(title).SetSubtitle(subtitle)
}

// Popover creates a popover component with anchor, title, and body.
func Popover(anchor rtui.VNode, title, body string) rtui.VNode {
	return popover.New(anchor).SetTitle(title).SetBody(body)
}

// Popconfirm creates a popconfirm component with anchor, title, and description.
func Popconfirm(anchor rtui.VNode, title, description string) rtui.VNode {
	return popconfirm.New(anchor).SetTitle(title).SetDescription(description)
}

// Statistic creates a statistic component with title and value.
func Statistic(title string, value interface{}) rtui.VNode {
	return statistic.New().SetTitle(title).SetValue(value)
}

// Alert creates an inline alert with a message.
func Alert(message string) rtui.VNode {
	return alert.NewBuilder(message).Build()
}

// Notification creates a notification with a message.
func Notification(message string) rtui.VNode {
	return notification.NewBuilder(message).Build()
}

// Progress shortcuts

// Progress creates a progress bar
func Progress(value, max int) rtui.VNode {
	return progress.NewBuilder().Value(value).Max(max).Build()
}

// ProgressPercent creates a progress bar with percentage
func ProgressPercent(percent int) rtui.VNode {
	return progress.NewBuilder().Value(percent).Max(100).Build()
}

// ProgressIndeterminate creates an animated progress bar for work without a known total.
func ProgressIndeterminate(label string) rtui.VNode {
	return progress.NewBuilder().Indeterminate().Label(label).Build()
}

// Clock creates a realtime clock with the requested radius.
func Clock(radius int) rtui.VNode {
	return clock.NewBuilder().Radius(radius).Build()
}

// Spin creates a spinner with an optional tip.
func Spin(tip string) rtui.VNode {
	return spin.NewBuilder().Tip(tip).Build()
}

// Steps creates a steps component from items.
func Steps(items []steps.Item) rtui.VNode {
	return steps.NewBuilder().Items(items).Build()
}

// Tag creates a tag with text.
func Tag(text string) rtui.VNode {
	return tag.NewBuilder(text).Build()
}

// Wrap shortcuts

// WrapWithWidth creates a wrap layout with specified width
func WrapWithWidth(width int, children ...rtui.VNode) rtui.VNode {
	w := wrap.W()
	w.Width(width)
	return w.Children(children...).Build()
}

// WrapWithGap creates a wrap layout with specified gap
func WrapWithGap(gap int, children ...rtui.VNode) rtui.VNode {
	return wrap.W().Gap(gap).Children(children...).Build()
}

// Grid shortcuts

// Grid creates a simple grid with specified number of columns
func Grid(numCols int, children ...rtui.VNode) rtui.VNode {
	return grid.SimpleGrid(numCols, children...)
}

// TwoColumnGrid creates a two-column grid layout
func TwoColumnGrid(children ...rtui.VNode) rtui.VNode {
	return grid.TwoColumnGrid(children...)
}

// ThreeColumnGrid creates a three-column grid layout
func ThreeColumnGrid(children ...rtui.VNode) rtui.VNode {
	return grid.ThreeColumnGrid(children...)
}

// Absolute shortcuts

// At places a child at absolute coordinates (x, y)
func At(child rtui.VNode, x, y int) rtui.VNode {
	return absolute.At(child, x, y)
}

// TopLeft places a child at top-left corner
func TopLeft(child rtui.VNode) rtui.VNode {
	return absolute.TopLeft(child)
}

// TopRight places a child at top-right corner
func TopRight(child rtui.VNode) rtui.VNode {
	return absolute.TopRight(child)
}

// BottomLeft places a child at bottom-left corner
func BottomLeft(child rtui.VNode) rtui.VNode {
	return absolute.BottomLeft(child)
}

// BottomRight places a child at bottom-right corner
func BottomRight(child rtui.VNode) rtui.VNode {
	return absolute.BottomRight(child)
}

// CenterAbs places a child at center of container
func CenterAbs(child rtui.VNode) rtui.VNode {
	return absolute.Center(child)
}

// =============================================================================
// Container Components shortcuts
// =============================================================================

// Panel shortcuts

// Panel creates a panel with content
func Panel(content rtui.VNode) rtui.VNode {
	return panel.Of(content)
}

// PanelOfSize creates a panel with specified size
func PanelOfSize(content rtui.VNode, width, height int) rtui.VNode {
	return panel.OfSize(content, width, height)
}

// PanelTitled creates a panel with title
func PanelTitled(title string, content rtui.VNode) rtui.VNode {
	return panel.Titled(title, content)
}

// PanelBordered creates a panel with border and specified size
func PanelBordered(content rtui.VNode, width, height int) rtui.VNode {
	return panel.Bordered(content, width, height)
}

// ScrollView shortcuts

// ScrollView creates a scrollable view
func ScrollView(child rtui.VNode) rtui.VNode {
	return scrollview.Scroll(child)
}

// Scroll creates a scrollable view (alias for ScrollView)
func Scroll(child rtui.VNode) rtui.VNode {
	return scrollview.Scroll(child)
}

// ScrollSize creates a scrollable view with specified size
func ScrollSize(child rtui.VNode, width, height int) rtui.VNode {
	return scrollview.ScrollSize(child, width, height)
}

// ScrollBordered creates a bordered scrollable view with specified size
func ScrollBordered(child rtui.VNode, width, height int) rtui.VNode {
	return scrollview.Bordered(child, width, height)
}

// =============================================================================
// Data Display Components shortcuts
// =============================================================================

// List shortcuts

// List creates a list component
func List() *list.Builder {
	return list.NewBuilder()
}

// ListOf creates a list with given rows
func ListOf(rows []string) rtui.VNode {
	return list.Of(rows)
}

// ListWithHeader creates a list with header and rows
func ListWithHeader(header string, rows []string) rtui.VNode {
	return list.WithHeader(header).Rows(rows).Build()
}

// Table shortcuts

// TableOf creates a table with columns and rows
func TableOf(columns []string, rows [][]string) rtui.VNode {
	// Convert columns to TableColumn type
	cols := make([]table.TableColumn, len(columns))
	for i, col := range columns {
		cols[i] = table.TableColumn{Title: col}
	}
	return table.Of(cols, rows)
}

// Pagination shortcuts

// Pagination creates a pagination component builder.
func Pagination() *pagination.Builder {
	return pagination.NewBuilder()
}

// PaginationOf creates a pagination component with the provided state.
func PaginationOf(total, pageSize, currentPage int) rtui.VNode {
	return pagination.NewBuilder().
		Total(total).
		PageSize(pageSize).
		CurrentPage(currentPage).
		Build()
}

// TreeView shortcuts

// TreeView creates a tree view
func TreeView() *treeview.Builder {
	return treeview.NewBuilder()
}

// TreeViewOf creates a tree view with nodes
func TreeViewOf(nodes []treeview.TreeNode) rtui.VNode {
	return treeview.Of(nodes)
}

// VirtualList shortcuts

// VirtualList creates a virtual list
func VirtualList() *virtuallist.Builder {
	return virtuallist.NewBuilder()
}

// VirtualListOfSize creates a virtual list with items and explicit dimensions.
func VirtualListOfSize(items []string, width, height int) rtui.VNode {
	return virtuallist.OfSize(items, width, height)
}

// =============================================================================
// Navigation Components shortcuts
// =============================================================================

// Tabs creates a tabs component with tab items
func Tabs(tabItems []tabs.TabItem) rtui.VNode {
	return tabs.Of(tabItems)
}

// Breadcrumb creates a breadcrumb component from the provided items.
func Breadcrumb(items []breadcrumb.Item) rtui.VNode {
	return breadcrumb.Of(items)
}

// =============================================================================
// Divider shortcuts
// =============================================================================

// Divider creates a simple horizontal divider
func Divider() rtui.VNode {
	return divider.D()
}

// Empty creates an empty-state component with an optional description override.
func Empty(description string) rtui.VNode {
	builder := empty.NewBuilder()
	if description != "" {
		builder.Description(description)
	}
	return builder.Build()
}

// DividerWithLabel creates a horizontal divider with label
func DividerWithLabel(label string) rtui.VNode {
	return divider.H(label)
}

// VDivider creates a vertical divider
func VDivider() rtui.VNode {
	return divider.V()
}

// HDivider creates a horizontal divider (alias for Divider)
func HDivider() rtui.VNode {
	return divider.D()
}

// DividerSection creates a section divider with title
func DividerSection(title string) rtui.VNode {
	return divider.Section(title)
}

// =============================================================================
// Modal shortcuts
// =============================================================================

// Note: Modal() function exists in ui/layer.go (returns *ModalBuilder)
// The following are shortcuts for ui/components/modal.Of():

// ModalOfSize creates a modal with specified size
func ModalOfSize(content rtui.VNode, width, height int) rtui.VNode {
	return modal.OfSize(content, width, height)
}

// Drawer creates a drawer with the provided content.
func Drawer(content rtui.VNode) rtui.VNode {
	return drawer.Of(content)
}

// DrawerTitled creates a titled drawer.
func DrawerTitled(title string, content rtui.VNode) rtui.VNode {
	return drawer.Titled(title, content)
}

// ModalTitled creates a modal with title
func ModalTitled(title string, content rtui.VNode) rtui.VNode {
	return modal.Titled(title, content)
}

// ModalAlert creates an alert modal dialog
func ModalAlert(title, message string) rtui.VNode {
	return modal.Alert(title, message).Build()
}

// ModalConfirm creates a confirm modal dialog
func ModalConfirm(title, message string) rtui.VNode {
	return modal.Confirm(title, message).Build()
}

// ModalInfo creates an informational modal dialog.
func ModalInfo(message string) rtui.VNode {
	return modal.Info(message).Build()
}

// ModalSuccess creates a success modal dialog.
func ModalSuccess(message string) rtui.VNode {
	return modal.Success(message).Build()
}

// ModalWarning creates a warning modal dialog.
func ModalWarning(message string) rtui.VNode {
	return modal.Warning(message).Build()
}

// ModalError creates an error modal dialog.
func ModalError(message string) rtui.VNode {
	return modal.Error(message).Build()
}

// =============================================================================
// Style Helpers - VNode manipulation helpers
// =============================================================================

// Styled applies a style to a VNode (only works for ElementVNode)
func Styled(vnode rtui.VNode, s style.Style) rtui.VNode {
	if elem, ok := vnode.(*rtui.ElementVNode); ok {
		elem.SetStyle(s)
		return elem
	}
	return vnode
}

// WithStyle is a fluent wrapper for Styled
func WithStyle(s style.Style) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return Styled(vnode, s)
	}
}

// =============================================================================
// Key Helpers
// =============================================================================

// WithKey adds a key to a VNode for reconciliation
func WithKey(key string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		if elem, ok := vnode.(*rtui.ElementVNode); ok {
			elem.SetKey(key)
			return elem
		}
		if comp, ok := vnode.(*rtui.ComponentVNode); ok {
			comp.SetKey(key)
			return comp
		}
		return vnode
	}
}

// =============================================================================
// ID Helpers
// =============================================================================

// WithID adds an ID to a VNode for Portal anchoring
func WithID(id string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		vnode.SetID(id)
		return vnode
	}
}

// =============================================================================
// Props Helpers
// =============================================================================

// WithProp adds a single property to a VNode
func WithProp(key string, value interface{}) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		if elem, ok := vnode.(*rtui.ElementVNode); ok {
			elem.Props().Set(key, value)
			return elem
		}
		return vnode
	}
}

// WithProps adds multiple properties to a VNode
func WithProps(props map[string]interface{}) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		if elem, ok := vnode.(*rtui.ElementVNode); ok {
			for k, v := range props {
				elem.Props().Set(k, v)
			}
			return elem
		}
		return vnode
	}
}

// =============================================================================
// Portal Helpers - Functional-style helpers for Portal configuration
// =============================================================================

// WithPortalRoot is a functional helper that sets the portalRoot property
func WithPortalRoot(portalRootID string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalRoot(portalRootID)
	}
}

// WithAnchorTo is a functional helper that sets anchorId and anchor properties
func WithAnchorTo(anchorID string, anchor Anchor) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetAnchorTo(anchorID, anchor)
	}
}

// WithPortalPosition is a functional helper that sets the position property
func WithPortalPosition(position types.PositionType) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalPosition(position)
	}
}

// WithPortalPriority is a functional helper that sets the priority property
func WithPortalPriority(priority int) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalPriority(priority)
	}
}

// WithPortalRootId is a functional helper that sets the portalRootId property
func WithPortalRootId(portalRootId string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalRootId(portalRootId)
	}
}

// =============================================================================
// Error Boundary Shortcuts - Runtime feature
// =============================================================================

// ErrorBoundary creates a new error boundary wrapper
func ErrorBoundary(name string, component rtui.ComponentFunc, fallback rtui.VNode) rtui.VNode {
	return rtui.ErrorBoundary(name, component, fallback)
}

// FallbackText creates a simple text fallback
func FallbackText(text string) rtui.VNode {
	return rtui.FallbackText(text)
}

// FallbackError creates an error message fallback with details
func FallbackError(prefix string) rtui.VNode {
	return rtui.FallbackError(prefix)
}

// FallbackBox creates a boxed error message
func FallbackBox(title, message string) rtui.VNode {
	return rtui.FallbackBox(title, message)
}

// =============================================================================
// Memo Shortcuts - Runtime optimization
// =============================================================================

// Memo wraps a component to memoize its output
func Memo(component rtui.VNode) rtui.VNode {
	return rtui.NewMemo(component)
}

// MemoWithCompare wraps a component with a custom comparison function
func MemoWithCompare(component rtui.VNode, compare rtui.PropsEqual) rtui.VNode {
	return rtui.NewMemoWithCompare(component, compare)
}

// MemoComponent wraps a component function with memoization
func MemoizedComponent(name string, fn rtui.ComponentFunc) rtui.VNode {
	return rtui.MemoComponent(name, fn)
}

// ShallowPropsEqual performs shallow comparison of two Props objects
func ShallowPropsEqual(oldProps, newProps rtui.Props) bool {
	return rtui.ShallowPropsEqual(oldProps, newProps)
}

// PropsEqualExcept creates a comparison function that ignores specific keys
func PropsEqualExcept(exceptKeys ...string) rtui.PropsEqual {
	return rtui.PropsEqualExcept(exceptKeys...)
}

// PropsEqualOnly creates a comparison function that only checks specific keys
func PropsEqualOnly(keys ...string) rtui.PropsEqual {
	return rtui.PropsEqualOnly(keys...)
}

// PureComponent creates a memoized component that only re-renders when props change
func PureComponent(name string, fn rtui.ComponentFunc) rtui.VNode {
	return rtui.MemoComponent(name, fn)
}

// PureComponentWithProps creates a memoized component with props
func PureComponentWithProps(name string, fn rtui.ComponentFuncWithProps) rtui.VNode {
	component := rtui.NewComponentWithProps(name, fn)
	return rtui.NewMemo(component)
}

// =============================================================================
// Toast Notifications shortcuts
// =============================================================================

// Toast creates a default info toast notification.
func Toast(message string) rtui.VNode {
	return toast.Toast(message)
}

// ToastInfo creates an info toast notification
func ToastInfo(message string) rtui.VNode {
	return toast.Info(message)
}

// ToastSuccess creates a success toast notification
func ToastSuccess(message string) rtui.VNode {
	return toast.Success(message)
}

// ToastWarning creates a warning toast notification
func ToastWarning(message string) rtui.VNode {
	return toast.Warning(message)
}

// ToastError creates an error toast notification
func ToastError(message string) rtui.VNode {
	return toast.Error(message)
}

// Tooltip shortcut

// TooltipFor creates a tooltip for a content element
func TooltipFor(content rtui.VNode, tooltipText string) rtui.VNode {
	return tooltip.Tooltip(content, tooltipText)
}
