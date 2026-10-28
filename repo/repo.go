// Package repo writes a repo's database diagram from the models its model.go/models.go declare.
package repo

import (
	"bytes"
	"os"
	"path/filepath"

	"webtyp.com/erd"
	"webtyp.com/ormc"
)

// File is where a repo keeps its database diagram, relative to its root. It is generated: never edit it by hand.
const File = "docs/diagrams/database.svg"

// Write draws the diagram of the models under dir (one module: tables of other modules are drawn as
// external) into out, and writes only when the bytes changed. It reports whether it wrote. When dir declares
// no database model it returns ormc.ErrNoModelsFound and writes nothing.
//
//	written, err := repo.Write(".", repo.File)
func Write(dir, out string) (bool, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	models, err := ormc.New().Models(abs)
	if err != nil {
		return false, err
	}
	svg := []byte(erd.ReadModule(filepath.Base(abs), models).SVG())
	if old, err := os.ReadFile(out); err == nil && bytes.Equal(old, svg) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(out, svg, 0o644); err != nil {
		return false, err
	}
	return true, nil
}
