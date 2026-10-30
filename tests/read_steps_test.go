package tests

import (
	"errors"
	"strings"
	"testing"

	"webtyp.com/ddl"
	"webtyp.com/erd"
	"webtyp.com/model"
)

func TestReadSteps_SameAsReadApp(t *testing.T) {
	a := tenant()
	b := user()
	c := newTable("setting", pk("id", model.Text()))

	steps := []ddl.Step{
		{
			Name: "auth",
			Run: func(conn ddl.Execer, compiler ddl.Compiler) error {
				return ddl.New(conn, compiler).CreateTable(a)
			},
		},
		{
			Name: "catalog",
			Run: func(conn ddl.Execer, compiler ddl.Compiler) error {
				d := ddl.New(conn, compiler)
				if err := d.CreateTable(b); err != nil {
					return err
				}
				return d.CreateTable(c)
			},
		},
	}

	d, err := erd.ReadSteps(steps...)
	if err != nil {
		t.Fatalf("ReadSteps() = %v, want nil", err)
	}

	want := erd.ReadApp(
		erd.Group{Name: "auth", Models: []model.Model{a}},
		erd.Group{Name: "catalog", Models: []model.Model{b, c}},
	).Report().SVG

	if got := d.Report().SVG; got != want {
		t.Errorf("ReadSteps() SVG = %v\nwant %v", got, want)
	}
}

func TestReadSteps_StepFails(t *testing.T) {
	steps := []ddl.Step{
		{
			Name: "broken",
			Run: func(conn ddl.Execer, compiler ddl.Compiler) error {
				return errors.New("boom")
			},
		},
	}

	_, err := erd.ReadSteps(steps...)
	if err == nil {
		t.Fatal("ReadSteps() = nil, want error")
	}

	if !strings.Contains(err.Error(), "broken") || !strings.Contains(err.Error(), "boom") {
		t.Errorf("ReadSteps() error = %v, want it to contain 'broken' and 'boom'", err)
	}
}

func TestReadSteps_NoSteps(t *testing.T) {
	d, err := erd.ReadSteps()
	if err != nil {
		t.Fatalf("ReadSteps() = %v, want nil", err)
	}

	want := erd.ReadApp().Report().SVG
	if got := d.Report().SVG; got != want {
		t.Errorf("ReadSteps() SVG = %v\nwant %v", got, want)
	}
}
