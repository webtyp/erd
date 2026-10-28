package tests

import (
	"testing"

	"webtyp.com/erd"
	"webtyp.com/json"
	"webtyp.com/model"
	"webtyp.com/router/mock"
)

// mounted guards the diagram the way an app does: only "admin" may read erd.ResourceSchema.
func mounted(d erd.Diagram) *mock.Router {
	r := &mock.Router{}
	r.Configure(mock.Config{Authorize: func(userID string, res model.Resource, act model.Action) bool {
		return userID == "admin" && res == erd.ResourceSchema && act == model.Read
	}})
	erd.Mount(r, erd.SchemaPath, d).Requires(erd.ResourceSchema, model.Read)
	return r
}

func TestMount_WithoutThePermissionIsForbidden(t *testing.T) {
	r := mounted(erd.ReadApp(erd.Group{Name: "auth", Models: []model.Model{tenant()}}))
	ctx := &mock.Context{InMethod: "GET", InPath: erd.SchemaPath}
	ctx.SetUserID("clerk")
	r.Invoke("GET", erd.SchemaPath, ctx)
	if ctx.Status != 403 {
		t.Fatalf("status = %d, want 403", ctx.Status)
	}
}

func TestMount_ServesTheReport(t *testing.T) {
	d := erd.ReadApp(erd.Group{Name: "auth", Models: []model.Model{user()}})
	r := mounted(d)

	ctx := &mock.Context{InMethod: "GET", InPath: erd.SchemaPath}
	ctx.SetUserID("admin")
	r.Invoke("GET", erd.SchemaPath, ctx)
	if ctx.Status != 200 {
		t.Fatalf("status = %d, want 200", ctx.Status)
	}
	var rep erd.Report
	if err := json.Decode(ctx.ResponseBody(), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.SVG != d.SVG() {
		t.Error("the served SVG differs from Diagram.SVG()")
	}
	want := erd.Finding{Rule: erd.RuleMissingTable, Table: "user", Column: "tenant_id", Target: "tenant"}
	if len(rep.Findings) != 1 || rep.Findings[0] != want {
		t.Fatalf("findings = %+v, want [%+v]", rep.Findings, want)
	}
}
