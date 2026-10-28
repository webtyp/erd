// Package erd draws the database diagram of webtyp models — tables, foreign keys and the
// model errors that would break or mislead the schema — as one SVG.
//
// A screen that shows the diagram imports only ResourceSchema, SchemaPath and Report; the
// linker leaves the reading and drawing out of a binary that never calls them.
package erd

import (
	"webtyp.com/lang"
	"webtyp.com/model"
)

// ResourceSchema is the RBAC resource that guards the diagram at SchemaPath and every
// screen that shows it. Mount it as
//
//	erd.Mount(r, erd.SchemaPath, diagram).Requires(erd.ResourceSchema, model.Read)
const ResourceSchema model.Resource = "db_schema"

// SchemaPath is where an app serves its database diagram.
const SchemaPath = "/_schema"

// Level says how bad a Finding is.
type Level string

const (
	LevelError   Level = "error"   // the schema is wrong: a migration fails or data cannot be trusted
	LevelWarning Level = "warning" // probably a mistake: worth a look
)

// Rule names what a Finding found. Its Level and its text follow from it.
type Rule string

const (
	RuleMissingTable        Rule = "missing_table"
	RuleTypeMismatch        Rule = "type_mismatch"
	RuleNoPrimaryKey        Rule = "no_primary_key"
	RuleCycle               Rule = "cycle"
	RuleDuplicateTable      Rule = "duplicate_table"
	RuleUndeclaredReference Rule = "undeclared_reference"
	RuleIsolatedTable       Rule = "isolated_table"
)

// Level is LevelWarning for the two rules that may be intended, LevelError for the rest.
func (r Rule) Level() Level {
	if r == RuleUndeclaredReference || r == RuleIsolatedTable {
		return LevelWarning
	}
	return LevelError
}

// Text is the English translation key that explains the rule:
// lang.Translate(f.Rule.Text()).
func (r Rule) Text() lang.Text {
	switch r {
	case RuleMissingTable:
		return "references a table no module creates"
	case RuleTypeMismatch:
		return "type differs from the referenced key"
	case RuleNoPrimaryKey:
		return "table without primary key"
	case RuleCycle:
		return "circular reference"
	case RuleDuplicateTable:
		return "table created by more than one module"
	case RuleUndeclaredReference:
		return "looks like a reference but has no foreign key"
	default:
		return "not related to any other table"
	}
}

// Finding is one problem in the schema. Column is empty for a table-level rule; Target is
// the referenced table (RuleMissingTable, RuleTypeMismatch) or the other module
// (RuleDuplicateTable), empty otherwise.
type Finding struct {
	Rule   Rule
	Table  string
	Column string
	Target string
}

// Report is what GET SchemaPath answers: the drawing and its findings.
type Report struct {
	SVG      string
	Findings []Finding
}

const (
	keySVG      = "svg"
	keyFindings = "findings"
	keyRule     = "rule"
	keyTable    = "table"
	keyColumn   = "column"
	keyTarget   = "target"
)

func (r *Report) IsNil() bool { return r == nil }

func (r *Report) EncodeFields(w model.FieldWriter) {
	w.String(keySVG, r.SVG)
	arr := w.Array(keyFindings, len(r.Findings))
	for i := range r.Findings {
		arr.Object(&r.Findings[i])
	}
	arr.Close()
}

func (r *Report) DecodeFields(rd model.FieldReader) {
	r.SVG, _ = rd.String(keySVG)
	if arr, ok := rd.Array(keyFindings); ok {
		r.Findings = make([]Finding, arr.Len())
		for i := 0; i < arr.Len(); i++ {
			arr.Object(i, &r.Findings[i])
		}
	}
}

func (f *Finding) IsNil() bool { return f == nil }

func (f *Finding) EncodeFields(w model.FieldWriter) {
	w.String(keyRule, string(f.Rule))
	w.String(keyTable, f.Table)
	w.String(keyColumn, f.Column)
	w.String(keyTarget, f.Target)
}

func (f *Finding) DecodeFields(rd model.FieldReader) {
	rule, _ := rd.String(keyRule)
	f.Rule = Rule(rule)
	f.Table, _ = rd.String(keyTable)
	f.Column, _ = rd.String(keyColumn)
	f.Target, _ = rd.String(keyTarget)
}
