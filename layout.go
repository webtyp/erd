package erd

// Geometry, in SVG user units (px at 100%). Integers only: the same schema must always
// produce the same bytes.
const (
	charW     = 7  // width of one monospace character at fontSize
	fontSize  = 12 // px
	rowH      = 18 // one column row
	headH     = 24 // table name row
	padX      = 10 // text inset inside a table
	gapX      = 64 // between layer columns (room for the curves)
	gapY      = 16 // between tables of one layer column
	bandHead  = 30 // band title row
	bandGap   = 24 // between bands
	margin    = 16 // around everything
	perColumn = 6  // tables per drawn column: a layer with more wraps into the next column
)

const (
	markPK = "PK"
	markFK = "FK"
	sep    = "  "
)

// rowText is what one column row shows: "tenant_id  text  FK" — its width sizes the table.
func rowText(c column) string {
	s := c.name + sep + c.typ
	if k := rowKeys(c); k != "" {
		s += sep + k
	}
	return s
}

// rowKeys is the constraint part of a row: "PK", "FK" or "PK FK".
func rowKeys(c column) string {
	var k string
	if c.pk {
		k = markPK
	}
	if c.ref != "" {
		if k != "" {
			k += " "
		}
		k += markFK
	}
	return k
}

func (t *table) size() {
	chars := len(t.name)
	for _, c := range t.columns {
		if n := len(rowText(c)); n > chars {
			chars = n
		}
	}
	t.w = padX*2 + chars*charW
	t.h = headH + len(t.columns)*rowH
}

// layout places every table: bands stacked top to bottom in group order; inside a band,
// layer columns left to right (referenced tables left), and in each layer column the
// tables top to bottom by name. It fills x, y, w, h and returns nothing.
func (d *Diagram) layout() {
	for _, t := range d.tables {
		t.size()
	}
	y := 0 // a band touches the drawing's edges; margin is its inner padding on every side
	for b := range d.bands {
		y += bandHead
		x := margin
		bandBottom := y
		for depth := 0; depth <= d.maxDepth(b); depth++ {
			layer := d.layerTables(b, depth)
			for start := 0; start < len(layer); start += perColumn {
				end := start + perColumn
				if end > len(layer) {
					end = len(layer)
				}
				colW, cy := 0, y
				for _, t := range layer[start:end] {
					t.x, t.y = x, cy
					cy += t.h + gapY
					if t.w > colW {
						colW = t.w
					}
				}
				if cy-gapY > bandBottom {
					bandBottom = cy - gapY
				}
				x += colW + gapX
			}
		}
		y = bandBottom + bandGap
	}
}

func (d *Diagram) maxDepth(band int) int {
	m := 0
	for _, t := range d.tables {
		if t.band == band && t.depth > m {
			m = t.depth
		}
	}
	return m
}

// layerTables returns the tables of one band and layer, sorted by name (insertion sort).
func (d *Diagram) layerTables(band, depth int) []*table {
	var out []*table
	for _, t := range d.tables {
		if t.band == band && t.depth == depth {
			out = append(out, t)
		}
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].name < out[j-1].name; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// bounds is the size of the whole drawing.
func (d *Diagram) bounds() (w, h int) {
	for _, t := range d.tables {
		if r := t.x + t.w + margin; r > w {
			w = r
		}
		if b := t.y + t.h + margin; b > h {
			h = b
		}
	}
	return w, h
}
