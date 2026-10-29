# ARCHITECTURE — erd

## Why it exists

A webtyp app declares its schema as Go models spread over many module repos. Nothing showed it whole,
and the hand-written Mermaid ERDs in each module (`docs/diagrams/database.md`) drifted from the code.
`erd` draws the schema from the models themselves and marks what is wrong in them.

## Ecosystem search (before creating the repo)

- `webtyp/ddlc`: DDL export only; its root was dead code (no importer). Not the place: drawing and
  checking a schema is another concern, and it needs no SQL dialect.
- `webtyp/svg`: icons (`sprite`, `sanitize`); no drawing primitives. The drawing helpers stay private
  here until a second consumer needs them, then they move to a `svg` subpackage.
- `webtyp/chart`: draws on a PDF canvas, not SVG.
- `webtyp/ddl.TopologicalSort`: orders tables for migration but reports only «cycle», not which tables
  nor their depth; `erd` needs both, so it layers the graph itself.
- `router.MountIntrospection`: the pattern `Mount` copies (returns the Route, the caller declares access).

## Prior art

- Mermaid `erDiagram`: text, rendered by GitHub; needs a JS runtime in the page, has no grouping by
  module and no way to mark errors.
- Graphviz / `dot`: excellent layout, but a C toolchain or a large WASM build.
- dbdiagram.io, DBeaver, pgAdmin: read the live database; they lose which module owns a table and
  cannot see a model that was never migrated.

`erd` reads the models (what the code says), groups by module, and keeps the layout simple and
deterministic: layers by foreign-key depth, parents left.

## Layout

Bands top to bottom, one per group, in the given order. Inside a band, a table's layer is 0 when it
references nothing in its band, else one more than the deepest table it references there; a layer with
more than six tables wraps into the next column. Tables in a cycle go after the deepest layer. Curves
go from the foreign-key row of the child to the name row of the parent, drawn under the tables.

## Colours

No CSS is written here. The SVG carries the classes of the widget `erd` (`erd__box`, `erd__edge-soft`, …) and
geometry; `css.go` colours them with the `widget/style` SVG recipes (`Paint`, `Ink`, `Line`, `Dashed`) from the
surface roles, so the diagram follows the app's theme and its light/dark switch. A file opened outside any app
(`Standalone`, what `repo.Write` writes) embeds the same sheet generated: `css.Theme(css.DefaultLight())` plus
`RenderCSS`. Text inherits the app's font; a standalone file uses the viewer's default font.

## What is drawn

What `model.Definition` declares and nothing else: `Ref` (foreign key, solid line) and `Options` (the table
that fills a select, dashed line, no constraint, never part of a cycle). Column names are never read for
meaning: a guessed «`*_id` without foreign key» rule flagged scope columns (`tenant_id`, `project_id`) and
provider ids as errors in 13 of 22 cases, so it was removed.
