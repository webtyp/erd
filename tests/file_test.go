//go:build !wasm

package tests

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/erd"
	"webtyp.com/model"
)

// The drawing carries no CSS of its own: inside an app the theme colours it. A file
// opened outside the app embeds the theme's generated tokens.
func TestStandalone_EmbedsTheThemeNotHandWrittenCSS(t *testing.T) {
	d := erd.ReadApp(erd.Group{Name: "auth", Models: []model.Model{tenant(), user()}})
	if strings.Contains(d.SVG(), "<style") {
		t.Fatal("SVG() must carry no <style>: the app's stylesheet colours it")
	}
	file := d.Standalone()
	if !strings.Contains(file, "<style>") || !strings.Contains(file, css.ColorSurface.Name+":") {
		t.Fatal("Standalone() must embed the theme's token definitions")
	}
	if !strings.Contains(file, `id="`+erd.IDPrefix+`user"`) {
		t.Fatal("Standalone() must draw the same tables as SVG()")
	}
}
