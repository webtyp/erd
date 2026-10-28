package erd

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// Group is the tables one module creates, drawn together in one band.
type Group struct {
	Name   string
	Models []model.Model
}

// Diagram is a read schema: its tables, its foreign keys and what is wrong with them.
// Build it with ReadApp or ReadModule; it never changes afterwards.
type Diagram struct {
	tables   []*table
	edges    []edge
	bands    []string
	findings []Finding
}

type column struct {
	name    string
	kind    string
	storage model.FieldType
	pk      bool
	ref     string
	mark    Level
}

type table struct {
	name     string
	band     int
	columns  []column
	external bool
	cycle    bool
	depth    int
	mark     Level
	x, y     int
	w, h     int
}

type edge struct {
	from, to int // table indexes: from holds the foreign key, to is referenced
	column   int // index of the foreign-key column in from
}

// scope says whether the tables are the whole app or one module seen alone.
type scope int

const (
	scopeApp scope = iota
	scopeModule
)

const suffixID = "_id"

// ReadApp reads every table an app creates, one Group per module, in migration order.
// A reference to a table no group creates is an error (RuleMissingTable).
//
//	d := erd.ReadApp(erd.Group{Name: "auth", Models: authModels}, erd.Group{Name: "rbac", Models: rbacModels})
func ReadApp(groups ...Group) Diagram {
	return read(scopeApp, groups)
}

// ReadModule reads the tables of one module seen alone: a reference to a table of another
// module (tenant, user) is drawn as a dashed external table, not reported.
//
//	d := erd.ReadModule("alert_manager", models)
func ReadModule(name string, models []model.Model) Diagram {
	return read(scopeModule, []Group{{Name: name, Models: models}})
}

// Findings lists every problem, errors first, then by table and column.
func (d Diagram) Findings() []Finding {
	out := make([]Finding, len(d.findings))
	copy(out, d.findings)
	return out
}

const errSchema = "erd: the schema has errors:"

// Check returns nil when no finding is an error; otherwise one error listing them, so a test
// can fail on a broken schema: if err := erd.ReadApp(groups...).Check(); err != nil { t.Fatal(err) }.
func (d Diagram) Check() error {
	var b fmt.Builder
	n := 0
	for _, f := range d.findings {
		if f.Rule.Level() != LevelError {
			continue
		}
		n++
		b.WriteString("\n  ").WriteString(f.Table)
		if f.Column != "" {
			b.WriteString(".").WriteString(f.Column)
		}
		b.WriteString(": ").WriteString(string(f.Rule.Text()))
		if f.Target != "" {
			b.WriteString(" (").WriteString(f.Target).WriteString(")")
		}
	}
	if n == 0 {
		return nil
	}
	return fmt.Err(errSchema + b.String())
}

// Report is what Mount serves: the drawing and the findings.
func (d Diagram) Report() Report {
	return Report{SVG: d.SVG(), Findings: d.Findings()}
}

func read(sc scope, groups []Group) Diagram {
	var d Diagram
	for gi, g := range groups {
		d.bands = append(d.bands, g.Name)
		for _, m := range g.Models {
			d.addTable(gi, m)
		}
	}
	d.resolveReferences(sc)
	d.checkTables()
	d.markCycles()
	d.layer()
	d.sortFindings()
	d.layout()
	return d
}

func (d *Diagram) find(name string) int {
	for i, t := range d.tables {
		if t.name == name {
			return i
		}
	}
	return -1
}

func (d *Diagram) addTable(band int, m model.Model) {
	t := &table{name: m.ModelName(), band: band}
	if prev := d.find(t.name); prev >= 0 {
		d.add(Finding{Rule: RuleDuplicateTable, Table: t.name, Target: d.bands[d.tables[prev].band]})
	}
	for _, f := range m.Schema() {
		c := column{name: f.Name}
		if f.Type != nil {
			c.kind = f.Type.Name()
			c.storage = f.Type.Storage()
		}
		c.pk = f.DB != nil && f.DB.PK
		t.columns = append(t.columns, c)
	}
	if ext, ok := m.(interface{ SchemaExt() []model.FieldExt }); ok {
		for _, fe := range ext.SchemaExt() {
			for ci := range t.columns {
				if t.columns[ci].name == fe.Name {
					t.columns[ci].ref = fe.Ref
				}
			}
		}
	}
	d.tables = append(d.tables, t)
}

func (d *Diagram) add(f Finding) {
	d.findings = append(d.findings, f)
	i := d.find(f.Table)
	if i < 0 {
		return
	}
	t := d.tables[i]
	t.mark = worse(t.mark, f.Rule.Level())
	for ci := range t.columns {
		if t.columns[ci].name == f.Column {
			t.columns[ci].mark = worse(t.columns[ci].mark, f.Rule.Level())
		}
	}
}

func worse(a, b Level) Level {
	if a == LevelError || b == LevelError {
		return LevelError
	}
	if a == LevelWarning || b == LevelWarning {
		return LevelWarning
	}
	return ""
}

// resolveReferences turns each foreign key into an edge. A target nobody creates is an
// error in an app, and an external table in a module.
func (d *Diagram) resolveReferences(sc scope) {
	n := len(d.tables)
	for ti := 0; ti < n; ti++ {
		t := d.tables[ti]
		for ci, c := range t.columns {
			if c.ref == "" {
				continue
			}
			to := d.find(c.ref)
			if to < 0 && sc == scopeApp {
				d.add(Finding{Rule: RuleMissingTable, Table: t.name, Column: c.name, Target: c.ref})
				continue
			}
			if to < 0 {
				d.tables = append(d.tables, &table{name: c.ref, band: t.band, external: true})
				to = len(d.tables) - 1
			}
			d.edges = append(d.edges, edge{from: ti, to: to, column: ci})
			target := d.tables[to]
			if target.external {
				continue
			}
			for _, pc := range target.columns {
				if pc.pk && pc.storage != c.storage {
					d.add(Finding{Rule: RuleTypeMismatch, Table: t.name, Column: c.name, Target: c.ref})
					break
				}
			}
		}
	}
}

func (d *Diagram) checkTables() {
	real := 0
	for _, t := range d.tables {
		if !t.external {
			real++
		}
	}
	for ti, t := range d.tables {
		if t.external {
			continue
		}
		pks := 0
		for _, c := range t.columns {
			if c.pk {
				pks++
			}
		}
		for _, c := range t.columns {
			// A column named *_id without a foreign key, unless it is the table's only key
			// (a part of a composite key, as in a join table, still references something).
			if c.ref == "" && !(c.pk && pks == 1) && len(c.name) > len(suffixID) && c.name[len(c.name)-len(suffixID):] == suffixID {
				d.add(Finding{Rule: RuleUndeclaredReference, Table: t.name, Column: c.name})
			}
		}
		if pks == 0 {
			d.add(Finding{Rule: RuleNoPrimaryKey, Table: t.name})
		}
		if real > 1 && !d.related(ti) {
			d.add(Finding{Rule: RuleIsolatedTable, Table: t.name})
		}
	}
}

func (d *Diagram) related(ti int) bool {
	for _, e := range d.edges {
		if e.from == ti || e.to == ti {
			return true
		}
	}
	return false
}

// markCycles runs Kahn's sort over every edge (a self-reference counts, as it does for
// ddl.TopologicalSort, which refuses to migrate it): whatever cannot be sorted is in a cycle.
func (d *Diagram) markCycles() {
	in := make([]int, len(d.tables))
	for _, e := range d.edges {
		in[e.from]++
	}
	queue := make([]int, 0, len(d.tables))
	for i := range d.tables {
		if in[i] == 0 {
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		to := queue[0]
		queue = queue[1:]
		for _, e := range d.edges {
			if e.to == to {
				in[e.from]--
				if in[e.from] == 0 {
					queue = append(queue, e.from)
				}
			}
		}
	}
	for i, t := range d.tables {
		if in[i] > 0 {
			t.cycle = true
			d.add(Finding{Rule: RuleCycle, Table: t.name})
		}
	}
}

// layer gives each table its column inside its band: 0 for a table that references nothing
// in its band, otherwise one more than the deepest table it references there. Tables in a
// cycle go one column after the deepest of their band.
func (d *Diagram) layer() {
	for changed := true; changed; {
		changed = false
		for _, e := range d.edges {
			from, to := d.tables[e.from], d.tables[e.to]
			if from.cycle || to.cycle || from.band != to.band || e.from == e.to {
				continue
			}
			if from.depth < to.depth+1 {
				from.depth = to.depth + 1
				changed = true
			}
		}
	}
	for _, t := range d.tables {
		if !t.cycle {
			continue
		}
		deepest := 0
		for _, o := range d.tables {
			if o.band == t.band && !o.cycle && o.depth > deepest {
				deepest = o.depth
			}
		}
		t.depth = deepest + 1
	}
}

// sortFindings orders errors first, then by table, column and rule, so the same schema
// always yields the same list (insertion sort: the list is short and no map is used).
func (d *Diagram) sortFindings() {
	less := func(a, b Finding) bool {
		if a.Rule.Level() != b.Rule.Level() {
			return a.Rule.Level() == LevelError
		}
		if a.Table != b.Table {
			return a.Table < b.Table
		}
		if a.Column != b.Column {
			return a.Column < b.Column
		}
		return a.Rule < b.Rule
	}
	f := d.findings
	for i := 1; i < len(f); i++ {
		for j := i; j > 0 && less(f[j], f[j-1]); j-- {
			f[j], f[j-1] = f[j-1], f[j]
		}
	}
}
