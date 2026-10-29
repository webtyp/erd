package erd

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// IDPrefix starts the id of every table group in the SVG: "erd-table-" + table name. A
// screen finds a table with it (to highlight or scroll to it).
const IDPrefix = "erd-table-"

const curve = 40 // horizontal pull of a foreign-key curve

// SVG draws the diagram for a page that already carries the app's stylesheet (RenderCSS
// colours it): bands, foreign-key curves under the tables, then the tables. Classes and
// geometry only; each band titled ModuleLabel + " " + its name. Same schema, same bytes.
func (d Diagram) SVG() string { return d.svg("") }

// svg writes the drawing; head goes first inside <svg> (Standalone's generated <style>).
// No background: the page or the viewer shows through.
func (d Diagram) svg(head string) string {
	w, h := d.bounds()
	var b fmt.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" class="`).WriteString(nameDiagram.Root().String()).WriteString(`" viewBox="0 0 `).
		WriteString(itoa(w)).WriteString(" ").WriteString(itoa(h)).
		WriteString(`" width="100%">`) // fills its container; the viewBox keeps the proportions
	b.WriteString(head)
	for bi, name := range d.bands {
		d.drawBand(&b, bi, string(ModuleLabel)+" "+name, w)
	}
	for _, e := range d.edges {
		d.drawEdge(&b, e)
	}
	for _, t := range d.tables {
		drawTable(&b, t)
	}
	b.WriteString("</svg>")
	return b.String()
}

func (d *Diagram) drawBand(b *fmt.Builder, band int, name string, width int) {
	top, bottom := -1, 0
	for _, t := range d.tables {
		if t.band != band {
			continue
		}
		if top < 0 || t.y < top {
			top = t.y
		}
		if t.y+t.h > bottom {
			bottom = t.y + t.h
		}
	}
	if top < 0 {
		return
	}
	y := top - bandHead
	// Edge to edge, the same inner padding (margin) on every side: tables start margin in from
	// the left and the drawing ends margin after the widest one (bounds).
	b.WriteString(`<rect class="`).WriteString(class(partBand)).WriteString(`" x="0" y="`).WriteString(itoa(y)).
		WriteString(`" width="`).WriteString(itoa(width)).
		WriteString(`" height="`).WriteString(itoa(bottom - y + margin)).WriteString(`" rx="6"/>`)
	b.WriteString(`<text class="`).WriteString(class(partBandTitle)).WriteString(`" x="`).WriteString(itoa(margin)).
		WriteString(`" y="`).WriteString(itoa(y + 20)).WriteString(`">`).WriteString(escape(name)).WriteString("</text>")
}

// drawEdge curves from the foreign-key row of the child to the name row of the parent.
func (d *Diagram) drawEdge(b *fmt.Builder, e edge) {
	from, to := d.tables[e.from], d.tables[e.to]
	y1 := from.y + headH + e.column*rowH + rowH/2
	y2 := to.y + headH/2
	// Parent to the left: leave left, enter its right side. Parent to the right: the
	// mirror. Same column (a reference across bands): leave and enter on the left.
	x1, c1 := from.x, from.x-curve
	x2, c2 := to.x+to.w, to.x+to.w+curve
	switch {
	case to.x >= from.x+from.w:
		x1, c1 = from.x+from.w, from.x+from.w+curve
		x2, c2 = to.x, to.x-curve
	case to.x+to.w > from.x:
		x2, c2 = to.x, to.x-curve
	}
	cls := class(partEdge)
	switch {
	case from.columns[e.column].mark == LevelError:
		cls += " " + class(partEdgeError)
	case e.soft:
		cls += " " + class(partEdgeSoft)
	}
	b.WriteString(`<path class="`).WriteString(cls).WriteString(`" d="M`).
		WriteString(itoa(x1)).WriteString(" ").WriteString(itoa(y1)).
		WriteString(" C").WriteString(itoa(c1)).WriteString(" ").WriteString(itoa(y1)).
		WriteString(" ").WriteString(itoa(c2)).WriteString(" ").WriteString(itoa(y2)).
		WriteString(" ").WriteString(itoa(x2)).WriteString(" ").WriteString(itoa(y2)).WriteString(`"/>`)
}

func drawTable(b *fmt.Builder, t *table) {
	cls := class(partBox)
	switch {
	case t.mark == LevelError:
		cls += " " + class(partBoxError)
	case t.external:
		cls += " " + class(partBoxExternal)
	}
	b.WriteString(`<g id="`).WriteString(IDPrefix).WriteString(escape(t.name)).WriteString(`">`)
	b.WriteString("<title>").WriteString(escape(t.name)).WriteString("</title>")
	box(b, cls, t.x, t.y, t.w, t.h)
	box(b, class(partHead), t.x+1, t.y+1, t.w-2, headH-1)
	text(b, class(partName), t.x+padX, t.y+headH-8, t.name)
	for i, c := range t.columns {
		ccls := class(partColumn)
		if c.mark == LevelError {
			ccls += " " + class(partColumnError)
		}
		row(b, ccls, t.x+padX, t.y+headH+i*rowH+rowH-5, c)
	}
	b.WriteString("</g>")
}

func box(b *fmt.Builder, cls string, x, y, w, h int) {
	b.WriteString(`<rect class="`).WriteString(cls).WriteString(`" x="`).WriteString(itoa(x)).
		WriteString(`" y="`).WriteString(itoa(y)).WriteString(`" width="`).WriteString(itoa(w)).
		WriteString(`" height="`).WriteString(itoa(h)).WriteString(`" rx="4"/>`)
}

func text(b *fmt.Builder, cls string, x, y int, s string) {
	b.WriteString(`<text class="`).WriteString(cls).WriteString(`" x="`).WriteString(itoa(x)).
		WriteString(`" y="`).WriteString(itoa(y)).WriteString(`">`).WriteString(escape(s)).WriteString("</text>")
}

// row writes one column: its name, then the type and the constraint in their own spans, so
// the eye tells them apart («tenant_id  text  FK»).
func row(b *fmt.Builder, cls string, x, y int, c column) {
	b.WriteString(`<text class="`).WriteString(cls).WriteString(`" x="`).WriteString(itoa(x)).
		WriteString(`" y="`).WriteString(itoa(y)).WriteString(`">`).WriteString(escape(c.name)).
		WriteString(sep).WriteString(`<tspan class="`).WriteString(class(partColumnType)).WriteString(`">`).
		WriteString(escape(c.typ)).WriteString("</tspan>")
	if k := rowKeys(c); k != "" {
		b.WriteString(sep).WriteString(`<tspan class="`).WriteString(class(partColumnKey)).WriteString(`">`).
			WriteString(k).WriteString("</tspan>")
	}
	b.WriteString("</text>")
}

// class is the SVG class of one part of the drawing.
func class(p widget.Part) string { return nameDiagram.Class(p).String() }

func itoa(n int) string { return fmt.Convert(n).String() }

// escape makes a name safe inside SVG text and attributes.
func escape(s string) string {
	var b fmt.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteString(s[i : i+1])
		}
	}
	return b.String()
}
