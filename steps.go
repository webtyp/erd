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
