package tests

import (
	"strings"
	"testing"

	"webtyp.com/erd"
	"webtyp.com/model"
)

func has(fs []erd.Finding, want erd.Finding) bool {
	for _, f := range fs {
		if f == want {
			return true
		}
	}
	return false
}

func TestReadApp_CleanSchemaHasNoFindings(t *testing.T) {
	d := erd.ReadApp(erd.Group{Name: "auth", Models: []model.Model{tenant(), user()}})
	if fs := d.Findings(); len(fs) != 0 {
		t.Fatalf("findings = %+v, want none", fs)
	}
	if err := d.Check(); err != nil {
		t.Fatalf("Check() = %v, want nil", err)
	}
}

func TestReadApp_ReferenceToATableNoModuleCreatesIsAnError(t *testing.T) {
	d := erd.ReadApp(erd.Group{Name: "auth", Models: []model.Model{user()}})
	want := erd.Finding{Rule: erd.RuleMissingTable, Table: "user", Column: "tenant_id", Target: "tenant"}
	if !has(d.Findings(), want) {
		t.Fatalf("findings = %+v, want %+v", d.Findings(), want)
	}
	err := d.Check()
	if err == nil || !strings.Contains(err.Error(), "user.tenant_id") {
		t.Fatalf("Check() = %v, want an error naming user.tenant_id", err)
	}
}

func TestReadModule_TableOfAnotherModuleIsExternalNotAnError(t *testing.T) {
	d := erd.ReadModule("auth", []model.Model{user()})
	if fs := d.Findings(); len(fs) != 0 {
		t.Fatalf("findings = %+v, want none: tenant belongs to another module", fs)
	}
	svg := d.SVG()
	if !strings.Contains(svg, `id="`+erd.IDPrefix+`tenant"`) || !strings.Contains(svg, "erd-external") {
		t.Fatal("the referenced table must be drawn, marked external")
	}
}

func TestReadApp_ForeignKeyTypeMustMatchThePrimaryKey(t *testing.T) {
	numericTenant := newTable("tenant", pk("id", model.Int()))
	d := erd.ReadApp(erd.Group{Name: "auth", Models: []model.Model{numericTenant, user()}})
	want := erd.Finding{Rule: erd.RuleTypeMismatch, Table: "user", Column: "tenant_id", Target: "tenant"}
	if !has(d.Findings(), want) {
		t.Fatalf("findings = %+v, want %+v", d.Findings(), want)
	}
}

func TestReadApp_TableWithoutPrimaryKeyIsAnError(t *testing.T) {
	d := erd.ReadApp(erd.Group{Name: "log", Models: []model.Model{newTable("event", col("name", model.Text()))}})
	if !has(d.Findings(), erd.Finding{Rule: erd.RuleNoPrimaryKey, Table: "event"}) {
		t.Fatalf("findings = %+v, want no_primary_key on event", d.Findings())
	}
}

func TestReadApp_ForeignKeyCycleIsAnError(t *testing.T) {
	a := newTable("a", pk("id", model.Text()), col("b_id", model.Text())).withFK("b_id", "b")
	b := newTable("b", pk("id", model.Text()), col("a_id", model.Text())).withFK("a_id", "a")
	d := erd.ReadApp(erd.Group{Name: "m", Models: []model.Model{a, b}})
	for _, name := range []string{"a", "b"} {
		if !has(d.Findings(), erd.Finding{Rule: erd.RuleCycle, Table: name}) {
			t.Errorf("findings = %+v, want cycle on %s", d.Findings(), name)
		}
	}
}

func TestReadApp_TableCreatedByTwoModulesIsAnError(t *testing.T) {
	d := erd.ReadApp(
		erd.Group{Name: "auth", Models: []model.Model{tenant()}},
		erd.Group{Name: "billing", Models: []model.Model{tenant()}},
	)
	if !has(d.Findings(), erd.Finding{Rule: erd.RuleDuplicateTable, Table: "tenant", Target: "auth"}) {
		t.Fatalf("findings = %+v, want duplicate_table naming the first module", d.Findings())
	}
}

func TestReadApp_IDColumnWithoutForeignKeyIsAWarning(t *testing.T) {
	join := newTable("role_permission", pk("role_id", model.Text()), pk("permission_id", model.Text()))
	single := newTable("session", pk("id", model.Text()), col("user_id", model.Text()))
	d := erd.ReadApp(erd.Group{Name: "rbac", Models: []model.Model{join, single}})
	for _, want := range []erd.Finding{
		{Rule: erd.RuleUndeclaredReference, Table: "role_permission", Column: "role_id"},
		{Rule: erd.RuleUndeclaredReference, Table: "role_permission", Column: "permission_id"},
		{Rule: erd.RuleUndeclaredReference, Table: "session", Column: "user_id"},
	} {
		if !has(d.Findings(), want) {
			t.Errorf("findings = %+v, want %+v", d.Findings(), want)
		}
	}
	if err := d.Check(); err != nil {
		t.Fatalf("warnings must not fail Check(): %v", err)
	}
}

func TestReadApp_TableWithoutRelationsIsAWarningOnlyAmongOthers(t *testing.T) {
	alone := erd.ReadApp(erd.Group{Name: "m", Models: []model.Model{tenant()}})
	if len(alone.Findings()) != 0 {
		t.Fatalf("a single table has nothing to relate to: %+v", alone.Findings())
	}
	d := erd.ReadApp(erd.Group{Name: "m", Models: []model.Model{tenant(), user(), newTable("setting", pk("id", model.Text()))}})
	if !has(d.Findings(), erd.Finding{Rule: erd.RuleIsolatedTable, Table: "setting"}) {
		t.Fatalf("findings = %+v, want isolated_table on setting", d.Findings())
	}
}

func TestReadApp_BrokenReferenceIsNotAlsoIsolated(t *testing.T) {
	// user references tenant, which no group creates: one error, and no isolated_table on user on top of it.
	d := erd.ReadApp(erd.Group{Name: "auth", Models: []model.Model{user(), newTable("setting", pk("id", model.Text()))}})
	if has(d.Findings(), erd.Finding{Rule: erd.RuleIsolatedTable, Table: "user"}) {
		t.Fatalf("findings = %+v: a table with a foreign key is not isolated, even when its target is missing", d.Findings())
	}
	if !has(d.Findings(), erd.Finding{Rule: erd.RuleIsolatedTable, Table: "setting"}) {
		t.Fatalf("findings = %+v, want isolated_table on setting", d.Findings())
	}
}

func TestFindings_ErrorsComeFirst(t *testing.T) {
	d := erd.ReadApp(erd.Group{Name: "m", Models: []model.Model{
		newTable("audit", pk("id", model.Text()), col("actor_id", model.Text())),
		newTable("event", col("name", model.Text())),
	}})
	fs := d.Findings()
	if len(fs) == 0 || fs[0].Rule.Level() != erd.LevelError {
		t.Fatalf("findings = %+v, want the error first", fs)
	}
}

func TestSVG_SameSchemaSameBytes(t *testing.T) {
	read := func() string {
		return erd.ReadApp(
			erd.Group{Name: "auth", Models: []model.Model{tenant(), user()}},
			erd.Group{Name: "log", Models: []model.Model{newTable("event", col("name", model.Text()))}},
		).SVG()
	}
	first := read()
	if first != read() {
		t.Fatal("two reads of the same schema drew different SVG")
	}
	for _, want := range []string{`<svg xmlns="http://www.w3.org/2000/svg"`, `id="` + erd.IDPrefix + `user"`, "auth", "log", "</svg>"} {
		if !strings.Contains(first, want) {
			t.Errorf("SVG lacks %q", want)
		}
	}
}

func TestSVG_NamesAreEscaped(t *testing.T) {
	svg := erd.ReadModule(`a<b>&"c`, []model.Model{tenant()}).SVG()
	if strings.Contains(svg, `a<b>`) || !strings.Contains(svg, "a&lt;b&gt;&amp;&quot;c") {
		t.Fatal("a band name must be escaped in the SVG")
	}
}
