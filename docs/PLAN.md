---
PLAN: "feat: erd CLI and repo.Write — a repo's diagram from its models.go"
TAG: v0.0.3
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 17258430540539880632
---

> Read AGENTS.md at the repo root before any change.

<!-- Dispatch after webtyp.com/ormc v0.1.27 (ormc.Models) is published. -->

# Plan — `webtyp.com/erd`: `repo.Write` and `cmd/erd`

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go get webtyp.com/ormc@v0.1.27
```

## Why

A module repo keeps its database diagram in `docs/diagrams/database.svg`, generated from its `models.go`, never
drawn by hand. Two callers need the same steps (scan → draw → write only if changed): the `erd` CLI (any agent
or CI) and the `webtyp/app` dev console (on save). They live once, in a subpackage `repo` — not in the `erd`
root, which a browser screen imports and must not pull `ormc` (and `go/parser`) into a WASM build.

## Ecosystem search

`ormc.Models` (v0.1.27) scans a repo's `model.go`/`models.go`; `erd.ReadModule` draws. Nothing else writes
diagrams. New: subpackage `repo` + `cmd/erd`, both thin wiring.

## Design gate

1. **Prior art:** `sqlc generate`, `buf generate`, `go generate`: a generator CLI that writes a checked-in file
   and leaves it untouched when nothing changed. Same here.
2. **Names:** `repo.Write(dir, out)` — «write the diagram of the repo at dir»; `repo.File` — where it lives;
   command `erd`.
3. **Ledger:** concepts +2 (`repo.Write`, `erd` command) / −1 (the hand-written Mermaid ERD); ways to make a
   diagram +0 (the hand-written one is deleted module by module).
4. **Belongs:** the scan stays in `ormc`, the drawing in `erd`; `repo` only wires them to a file.
5. **Deletes:** nothing in this repo; it lets 13 modules delete `docs/diagrams/database.md`.

## APIs used

| Symbol | Signature | Tag | File |
|---|---|---|---|
| `ormc.New` | `func New() *Generator` | ormc v0.1.27 | `generator.go` |
| `(*ormc.Generator).Models` | `func (g *Generator) Models(root string) ([]model.Model, error)` | ormc v0.1.27 | `generate.go` |
| `ormc.ErrNoModelsFound` | `var ErrNoModelsFound error` (returned when root has no database model) | ormc v0.1.27 | `errors.go` |
| `erd.ReadModule` | `func ReadModule(name string, models []model.Model) Diagram` | this repo | `read.go` |
| `(erd.Diagram).SVG` | `func (d Diagram) SVG() string` | this repo | `draw.go` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by
guessing.

## Stage 1 — `repo/repo.go` (new)

This is backend tooling: stdlib `os`, `path/filepath`, `bytes` are legitimate here; do NOT replace them.

```go
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
```

## Stage 2 — `cmd/erd/main.go` (new)

```go
// Command erd writes the database diagram of a repo's models:
//
//	erd              # docs/diagrams/database.svg for the models under .
//	erd ../billing   # ../billing/docs/diagrams/database.svg
//	erd -o out.svg   # another file
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"webtyp.com/erd/repo"
	"webtyp.com/ormc"
)

func main() {
	out := flag.String("o", "", "file to write (default <dir>/"+repo.File+")")
	flag.Parse()
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	if *out == "" {
		*out = filepath.Join(dir, repo.File)
	}
	written, err := repo.Write(dir, *out)
	switch {
	case errors.Is(err, ormc.ErrNoModelsFound):
		fmt.Fprintln(os.Stderr, "erd: no models under", dir, "— nothing written")
	case err != nil:
		fmt.Fprintln(os.Stderr, "erd:", err)
		os.Exit(1)
	case written:
		fmt.Fprintln(os.Stderr, "erd: wrote", *out)
	default:
		fmt.Fprintln(os.Stderr, "erd:", *out, "unchanged")
	}
}
```

## Stage 3 — tests

`tests/testdata/shop/models.go` (fixture; `testdata` is ignored by `go build` and `gotest`):

```go
package shop

import "webtyp.com/model"

var CustomerModel = model.Definition{
	Name: "customer",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: model.Text()},
	},
}

var InvoiceModel = model.Definition{
	Name: "invoice",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "customer_id", Type: model.Text(), Ref: &CustomerModel},
	},
}
```

`tests/repo_test.go`:

```go
package tests

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"webtyp.com/erd"
	"webtyp.com/erd/repo"
	"webtyp.com/ormc"
)

func TestWrite_DrawsOnceAndSkipsWhenUnchanged(t *testing.T) {
	out := filepath.Join(t.TempDir(), "docs", "diagrams", "database.svg")
	written, err := repo.Write("testdata/shop", out)
	if err != nil || !written {
		t.Fatalf("first Write = %v, %v; want true, nil", written, err)
	}
	svg, _ := os.ReadFile(out)
	for _, want := range []string{`id="` + erd.IDPrefix + `customer"`, `id="` + erd.IDPrefix + `invoice"`, ">shop<"} {
		if !strings.Contains(string(svg), want) {
			t.Errorf("diagram lacks %q", want)
		}
	}
	written, err = repo.Write("testdata/shop", out)
	if err != nil || written {
		t.Fatalf("second Write = %v, %v; want false, nil (nothing changed)", written, err)
	}
}

func TestWrite_NoModelsWritesNothing(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "database.svg")
	if _, err := repo.Write(dir, out); !errors.Is(err, ormc.ErrNoModelsFound) {
		t.Fatalf("err = %v, want ormc.ErrNoModelsFound", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("no file may be written when there are no models")
	}
}
```

If `ormc` cannot parse the fixture, stop and paste its error in the PR; do not change `repo.go` to work around it.

## Stage 4 — docs

- `README.md`: after the first code block add:

```markdown
## In a repo

`go install webtyp.com/erd/cmd/erd@latest`, then `erd` at the repo root writes `docs/diagrams/database.svg`
(only when it changed). Never edit that file by hand; link it from `docs/ARCHITECTURE.md` as
`![Database](diagrams/database.svg)`.
```

- `AGENTS.md`, `## This repo`, first bullet: replace «It never opens a database, never scans files (`ormc` does
  that) and never runs migrations (`ddl.Capture` does that).» with «It never opens a database and never runs
  migrations (`ddl.Capture` does). The root never imports `ormc`: only `repo` and `cmd/erd` do, so a browser screen
  that imports the root never compiles `go/parser`.»

## Delivery

- `gotest` green. Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`. A PR that does
  not compile, or whose tests could not run, is a failed delivery.
- `grep -rn '"webtyp.com/ormc"' --include=*.go . | grep -v "^./repo/\|^./cmd/\|^./tests/"` → empty.
- `git diff --stat main` lists exactly: `go.mod`, `go.sum`, `repo/repo.go`, `cmd/erd/main.go`,
  `tests/testdata/shop/models.go`, `tests/repo_test.go`, `README.md`, `AGENTS.md`.
- Never pause to ask permission or report progress: when the check passes, commit, push and open the PR. Stuck 30
  minutes → open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
  `git checkout -- <file>` or `git stash` on your own work.

| Stage | File | Action |
|---|---|---|
| 1 | `repo/repo.go` | `File`, `Write` |
| 2 | `cmd/erd/main.go` | the CLI |
| 3 | `tests/testdata/shop/models.go`, `tests/repo_test.go` | writes once, skips unchanged, nothing without models |
| 4 | `README.md`, `AGENTS.md` | how a repo uses it |
