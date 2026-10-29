//go:build !wasm

package erd

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func (d *Diagram) WidgetName() widget.Name { return nameDiagram }
func (d *Diagram) WidgetKind() widget.Kind { return widget.Region }

// RenderCSS colours the drawing from the app's theme: tables as panels with a Primary head,
// relations in Primary (dashed when soft), errors in Danger, outside tables dashed. Standalone embeds the same sheet.
func (d *Diagram) RenderCSS() *css.Stylesheet {
	return style.For(d).
		Root(style.Typeface(), style.FontSize(style.TextXs)).
		Part(partBand, style.Paint(style.Inset)).
		Part(partBandTitle, style.Glyph(style.Primary), style.FontSize(style.TextSm), style.FontWeight(style.WeightBold)).
		Part(partBox, style.Paint(style.Panel)).
		Part(partBoxExternal, style.Paint(style.Page), style.Dashed()).
		Part(partBoxError, style.Paint(style.DangerWash)).
		Part(partHead, style.Paint(style.Primary)).
		Part(partName, style.Ink(style.Primary), style.FontWeight(style.WeightBold)).
		Part(partColumn, style.Ink(style.Panel)).
		Part(partColumnError, style.Glyph(style.Danger)).
		Part(partColumnType, style.Ink(style.Subtle)).
		Part(partColumnKey, style.Glyph(style.Primary), style.FontWeight(style.WeightBold)).
		Part(partEdge, style.Line(style.Primary)).
		Part(partEdgeSoft, style.Line(style.Primary), style.Dashed()).
		Part(partEdgeError, style.Line(style.Danger)).
		Stylesheet()
}
