//go:build !wasm

package erd

import "webtyp.com/css"

// Standalone is the diagram as a file opened outside any app (docs/diagrams/database.svg on
// GitHub): SVG() plus one <style> generated from the theme's tokens and RenderCSS, the same
// sheet an app serves. Nothing in it is written by hand.
//
//	os.WriteFile(repo.File, []byte(d.Standalone()), 0o644)
func (d Diagram) Standalone() string {
	sheet := css.Theme(css.DefaultLight()).String() + d.RenderCSS().String()
	return d.svg("<style>" + sheet + "</style>")
}
