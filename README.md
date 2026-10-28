# erd
<img src="docs/img/badges.svg">

Database diagram (ERD) of webtyp models: tables, foreign keys and model errors, drawn as one SVG.

```go
// a module repo, from the models ormc scans
d := erd.ReadModule("alert_manager", models)
os.WriteFile("docs/diagrams/database.svg", []byte(d.SVG()), 0o644)

// an app, one group per migration step: a reference to a table nobody creates is an error
d := erd.ReadApp(erd.Group{Name: "auth", Models: authModels}, erd.Group{Name: "rbac", Models: rbacModels})
if err := d.Check(); err != nil { t.Fatal(err) } // in a test: fail on schema errors
erd.Mount(r, erd.SchemaPath, d).Requires(erd.ResourceSchema, model.Read) // GET /_schema → Report
```

| Symbol | What |
|---|---|
| `ReadApp(groups ...Group) Diagram` | every table of an app, one band per module |
| `ReadModule(name, models) Diagram` | one module alone; tables of other modules are dashed «external» boxes |
| `Diagram.SVG() string` | standalone SVG (light/dark), same schema → same bytes; each table is `<g id="erd-table-<name>">` |
| `Diagram.Findings() []Finding` / `Check() error` | problems, errors first / an error only when one is an error |
| `Mount(r, path, d) router.Route` | read-only GET answering `Report{SVG, Findings}`; the caller declares access, never `Public()` |
| `ResourceSchema`, `SchemaPath`, `Report`, `Finding`, `Rule` | what a browser screen needs to fetch and list the findings |

## Findings

| Rule | Level | When |
|---|---|---|
| `missing_table` | error | a foreign key names a table no group creates (`ReadApp` only) |
| `type_mismatch` | error | the foreign key's storage type differs from the referenced primary key |
| `no_primary_key` | error | a table has no primary key |
| `cycle` | error | tables reference each other in a loop (a self-reference too: `ddl.TopologicalSort` refuses it) |
| `duplicate_table` | error | two groups create the same table |
| `undeclared_reference` | warning | a `*_id` column without foreign key (a part of a composite key too) |
| `isolated_table` | warning | a table with no foreign key in or out, among several tables |

Design: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).
