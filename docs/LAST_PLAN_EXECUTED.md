---
PLAN: "feat!: erd styled from the theme (css.go), draws only what the model declares, soft links dashed, storage types"
EXECUTOR: local
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# Plan — `webtyp.com/erd`: no hand-written CSS; the model decides what is drawn

## Why

1. `draw.go` embedded a `<style>` with hex colours and its own dark `@media`: inside an app set to light the
   diagram stayed dark (**widget-styling**: no CSS text outside `css.go`).
2. `undeclared_reference` guessed relations from column names: 13 of its 22 findings in mjosefa-cms were
   false (`tenant_id`, `project_id`, `identity.provider_id`). The diagram draws what `model.Definition`
   declares (DB_DIAGRAM P12).
3. A select filled from another table (`Options: &ZoneModel`) is a relation without a foreign key: drawn dashed,
   never a warning (D15).
4. Columns showed the form widget (`checkbox`, `select`); they show the storage type, and the unit for an
   instant (`unix_time`, `nano_time`) (D16).

## Ecosystem search

`widget/style` now has `Paint`, `Ink`, `Line`, `Dashed` (widget plan) — reuse. `css.Theme(css.DefaultLight())`
generates the token definitions — reuse for the file opened outside the app. `model.UnixTime().Name()` and
`model.NanoTime().Name()` name the instant kinds — reuse, nothing new in `model`. `widget.Name` gives typed
part classes — reuse.

## Design gate

1. **Prior art:** dbdiagram.io, DBeaver, SchemaSpy draw declared foreign keys solid and «virtual»/«implied»
   relations dashed; Mermaid `erDiagram` shows the column's data type, never the UI control.
2. **Names:** `(*Diagram).RenderCSS` (the producer every
   css.go declares), `(Diagram).Standalone()` — «the diagram as a standalone file, with its styles».
3. **Ledger:** concepts −1 (`RuleUndeclaredReference`) +1 (`Standalone`); ways to colour the
   diagram 2 (hand CSS + theme) → 1 (theme); ways to draw a relation 1 → 2 (constraint, link) — different facts.
4. **Belongs:** the drawing's look in its own `css.go`; colours in `widget/style`; tokens in `css`.
5. **Deletes:** the `style` constant, `RuleUndeclaredReference`, its text and lang key, `suffixID`, the warning
   colouring in the drawing (warnings stay in `Findings`).

## Stages

1. `erd.go`: unexported `nameDiagram` and part constants; delete `RuleUndeclaredReference`.
2. `css.go` (`!wasm`): `(*Diagram) WidgetName/WidgetKind/RenderCSS` with `Paint`/`Ink`/`Line`/`Dashed`, `FontSize`,
   `FontWeight`. Errors use `Danger`; external tables and soft links `Dashed()`.
3. `draw.go`: no `<style>`; classes from `nameDiagram.Class(part)`; only errors are marked in the drawing.
4. `file.go` (`!wasm`): `Standalone()` = `SVG()` with one `<style>`: `css.Theme(css.DefaultLight()).String()` +
   `RenderCSS().String()`. `repo.Write` writes `Standalone()`.
5. `read.go`: `Options` → soft link (dashed, counts as related, not in cycles); delete the name guess; the column
   type is the storage type or the instant kind's name.
6. Tests: soft link related and not a finding; `tenant_id` without `Ref` is not a finding; instant shows its unit;
   `Standalone()` carries the theme and no hex written here; `SVG()` carries no `<style>`.
7. Docs: README rules table, ARCHITECTURE, `lang.json`.

## Additions (review with the user, 2026-10-28)

- No background rect: the canvas panel shows through, so the edges are even (a Page-coloured rect inside a
  Panel canvas left an uneven strip).
- Contrast: table heads `Paint(Primary)` + names `Ink(Primary)`; relations `Line(Primary)`.
- Each band reads «Module <name>», title `Glyph(Primary)`. `ModuleLabel` is English in the SVG: the server has
  no dictionary (sitec injects it into the page), so the screen translates it.
- Even edges: bands go edge to edge of the drawing (x=0, full width) with `margin` as inner padding on every
  side; the SVG is `width="100%"` (the viewBox keeps proportions), so it fills its container with no strip.
- Each row: name, then the type (`column-type`, `Ink(Subtle)`) and the constraint (`column-key`, Primary, bold)
  in their own spans.
- README shows the Database screen (`docs/img/schema-screen.png`).
