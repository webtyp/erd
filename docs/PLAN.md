---
PLAN: "feat: erd.ReadSteps — the app diagram from its migration steps, for server and browser"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 11908226592401344753
---

> Read AGENTS.md at the repo root before any change.

# `erd.ReadSteps`

## Context

The «Database» screen draws the diagram an app serves at `SchemaPath`. The server builds it in `appkit` with
`ddl.Capture(steps)`, then a private `groupsOf` (one group per migration step) and `erd.ReadApp(groups...)`
(`appkit/server.go:58-60`, `appkit/schema.go`). A module's demo runs in the browser with no server and must draw the
same diagram from its own migration steps. So the step-to-diagram path moves here, as `erd.ReadSteps(steps...)`; a
later plan makes `appkit` call it and deletes `groupsOf`.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go get webtyp.com/ddl@v0.0.36
go mod tidy
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`.

## Ecosystem search

Searched `groupsOf`, `ReadSteps`, `ddl.Capture` over `webtyp/`: `ddl.Capture` (reuse) and `appkit.groupsOf` (private,
moves here). **Extend** `erd`; `groupsOf` is deleted from `appkit` in the next plan.

## Design gate

1. **Prior art.** SchemaSpy and Django's `graph_models` draw the schema from the migrations/models the app declares,
   not from a live server (unverified: pages not read in this session).
2. **Novice name.** `erd.ReadSteps(steps...)` next to `erd.ReadApp(groups...)` and `erd.ReadModule(name, models)`.
3. **Complexity ledger.** Concepts +0 · ways to go from steps to a diagram: 1 (after the `appkit` plan).
4. **Where it belongs.** `erd`, owner of the diagram; `ddl` stays the owner of `Capture`.
5. **Deletes.** Nothing here; `appkit/schema.go` in the next plan.

## APIs used

| Symbol | Exact signature | Tag |
|---|---|---|
| `ddl.Step` | `struct { Name string; Run func(conn Execer, compiler Compiler) error }` | `webtyp.com/ddl` v0.0.36 |
| `ddl.Capture` | `func Capture(steps []Step) ([]Table, error)` — runs each step against a recorder; on failure `fmt.Err("ddl: capture: step", name+":", err.Error())` | v0.0.36 |
| `ddl.Table` | `struct { Step string; Model model.Model }` | v0.0.36 |
| `ddl.New(conn, c).CreateTable(m)` | how a test step creates a table | v0.0.36 |
| `erd.ReadApp`, `erd.Group`, `(Diagram).Report`, `Report.SVG` | as in `read.go`, `erd.go` | this repo |

If a signature differs, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — tests first: `tests/read_steps_test.go` (new, `package tests`)

Use the test models already defined in `tests/` (pick two or three with no foreign key between them, or with one).
A step is built as `ddl.Step{Name: "catalog", Run: func(c ddl.Execer, k ddl.Compiler) error { return ddl.New(c, k).CreateTable(&B{}) }}`.
Red after this stage.

| Test | Setup | Expected |
|---|---|---|
| `TestReadSteps_SameAsReadApp` | steps `auth` (creates A) and `catalog` (creates B, then C) | `d, err := erd.ReadSteps(...)`: `err == nil`; `d.Report().SVG == erd.ReadApp(erd.Group{Name: "auth", Models: [A]}, erd.Group{Name: "catalog", Models: [B, C]}).Report().SVG` |
| `TestReadSteps_StepFails` | a step `broken` whose `Run` returns `errors.New("boom")` | `err != nil`, `err.Error()` contains `"broken"` and `"boom"` |
| `TestReadSteps_NoSteps` | no steps | `err == nil`; `d.Report().SVG == erd.ReadApp().Report().SVG` |

Commit after this stage.

## Stage 2 — `steps.go` (new)

```go
package erd

import (
	"webtyp.com/ddl"
	"webtyp.com/model"
)

// ReadSteps reads the app's diagram from its migration steps, one group per
// step in step order: what an app serves at SchemaPath, and what a page with
// no server (a module's demo) draws from its own migrations.
//
// Example: d, err := erd.ReadSteps(steps...)
func ReadSteps(steps ...ddl.Step) (Diagram, error) {
	tables, err := ddl.Capture(steps)
	if err != nil {
		return Diagram{}, err
	}
	var groups []Group
	for _, t := range tables {
		if n := len(groups); n > 0 && groups[n-1].Name == t.Step {
			groups[n-1].Models = append(groups[n-1].Models, t.Model)
			continue
		}
		groups = append(groups, Group{Name: t.Step, Models: []model.Model{t.Model}})
	}
	return ReadApp(groups...), nil
}
```

Run `gotest`: green. Commit after this stage.

## Stage 3 — docs

`README.md`: add `ReadSteps` next to `ReadApp`/`ReadModule` with one sentence and the example. Commit.

## Delivery check

1. `gotest` passes. A PR that does not compile, or whose tests could not run, is a failed delivery.
2. `git diff --stat main` lists only `tests/read_steps_test.go`, `steps.go`, `README.md`, `go.mod`, `go.sum`.

## Delivery rules

- Never pause to ask permission or report progress: when the delivery check passes, commit, push and open the PR. Stuck
  for 30 minutes: open the PR anyway, listing what fails.
- Commit after every stage. Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

## Stages

| # | Files | Result |
|---|---|---|
| 1 | `tests/read_steps_test.go` | Same as ReadApp, failing step, no steps; red |
| 2 | `steps.go` | `ReadSteps`; green |
| 3 | `README.md` | Documented |
