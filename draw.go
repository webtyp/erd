package erd

import "webtyp.com/fmt"

// IDPrefix starts the id of every table group in the SVG: "erd-table-" + table name. A
// screen finds a table with it (to highlight or scroll to it).
const IDPrefix = "erd-table-"

const curve = 40 // horizontal pull of a foreign-key curve

// style colours the drawing with plain colours (every SVG renderer reads them), light by
// default and dark under prefers-color-scheme. A page that shows the SVG in line re-colours
// it from its theme with more specific selectors (.schemaview .erd-box { fill: … }).
const style = `.erd-bg{fill:#ffffff}.erd-band{fill:#f5f7fa}.erd-box{fill:#ffffff;stroke:#9aa5b1}.erd-head{fill:#e8edf3}` +
	`.erd text{fill:#1f2933;font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:12px}` +
	`.erd-band-title,.erd-name{font-weight:bold}.erd-external .erd-box{stroke-dasharray:4 3}.erd .erd-external text{fill:#616e7c}` +
	`.erd-edge{fill:none;stroke:#9aa5b1}` +
	`.erd-warning .erd-box{stroke:#c99a06;stroke-width:2}.erd-error .erd-box{stroke:#d64545;stroke-width:2}` +
	`.erd text.erd-warning{fill:#c99a06}.erd text.erd-error{fill:#d64545}.erd-edge.erd-warning{stroke:#c99a06}.erd-edge.erd-error{stroke:#d64545}` +
	`@media (prefers-color-scheme:dark){.erd-bg{fill:#111827}.erd-band{fill:#1a2230}.erd-box{fill:#1f2937;stroke:#616e7c}.erd-head{fill:#2b3646}` +
	`.erd text{fill:#e5e7eb}.erd .erd-external text{fill:#9aa5b1}.erd-edge{stroke:#7b8794}}`

// SVG draws the diagram as one standalone SVG document: bands, foreign-key curves under
// the tables, then the tables. Same schema, same bytes.
func (d Diagram) SVG() string {
	w, h := d.bounds()
	var b fmt.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" class="erd" viewBox="0 0 `).
		WriteString(itoa(w)).WriteString(" ").WriteString(itoa(h)).
		WriteString(`" width="`).WriteString(itoa(w)).WriteString(`" height="`).WriteString(itoa(h)).WriteString(`">`)
	b.WriteString("<style>").WriteString(style).WriteString("</style>")
	b.WriteString(`<rect class="erd-bg" width="100%" height="100%"/>`)
	for bi, name := range d.bands {
		d.drawBand(&b, bi, name, w)
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
	b.WriteString(`<rect class="erd-band" x="`).WriteString(itoa(margin / 2)).
		WriteString(`" y="`).WriteString(itoa(y)).
		WriteString(`" width="`).WriteString(itoa(width - margin)).
		WriteString(`" height="`).WriteString(itoa(bottom - y + margin/2)).WriteString(`" rx="6"/>`)
	b.WriteString(`<text class="erd-band-title" x="`).WriteString(itoa(margin)).
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
	cls := "erd-edge"
	if m := from.columns[e.column].mark; m != "" {
		cls += " erd-" + string(m)
	}
	b.WriteString(`<path class="`).WriteString(cls).WriteString(`" d="M`).
		WriteString(itoa(x1)).WriteString(" ").WriteString(itoa(y1)).
		WriteString(" C").WriteString(itoa(c1)).WriteString(" ").WriteString(itoa(y1)).
		WriteString(" ").WriteString(itoa(c2)).WriteString(" ").WriteString(itoa(y2)).
		WriteString(" ").WriteString(itoa(x2)).WriteString(" ").WriteString(itoa(y2)).WriteString(`"/>`)
}

func drawTable(b *fmt.Builder, t *table) {
	cls := "erd-table"
	if t.external {
		cls += " erd-external"
	}
	if t.mark != "" {
		cls += " erd-" + string(t.mark)
	}
	b.WriteString(`<g class="`).WriteString(cls).WriteString(`" id="`).WriteString(IDPrefix).WriteString(escape(t.name)).WriteString(`">`)
	b.WriteString("<title>").WriteString(escape(t.name)).WriteString("</title>")
	box(b, "erd-box", t.x, t.y, t.w, t.h)
	box(b, "erd-head", t.x+1, t.y+1, t.w-2, headH-1)
	text(b, "erd-name", t.x+padX, t.y+headH-8, t.name)
	for i, c := range t.columns {
		ccls := "erd-col"
		if c.mark != "" {
			ccls += " erd-" + string(c.mark)
		}
		text(b, ccls, t.x+padX, t.y+headH+i*rowH+rowH-5, rowText(c))
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
