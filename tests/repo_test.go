package tests

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"webtyp.com/erd"
	"webtyp.com/erd/repo"
	"webtyp.com/ormc"
)

func TestWrite_DrawsOnceAndSkipsWhenUnchanged(t *testing.T) {
	out := filepath.Join(t.TempDir(), "docs", "diagrams", "database.svg")
	written, err := repo.Write("testdata/shop", out)
	if err != nil || !written {
		t.Fatalf("first Write = %v, %v; want true, nil", written, err)
	}
	svg, _ := os.ReadFile(out)
	for _, want := range []string{`id="` + erd.IDPrefix + `customer"`, `id="` + erd.IDPrefix + `invoice"`, ">shop<"} {
		if !strings.Contains(string(svg), want) {
			t.Errorf("diagram lacks %q", want)
		}
	}
	written, err = repo.Write("testdata/shop", out)
	if err != nil || written {
		t.Fatalf("second Write = %v, %v; want false, nil (nothing changed)", written, err)
	}
}

func TestWrite_NoModelsWritesNothing(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "database.svg")
	if _, err := repo.Write(dir, out); !errors.Is(err, ormc.ErrNoModelsFound) {
		t.Fatalf("err = %v, want ormc.ErrNoModelsFound", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("no file may be written when there are no models")
	}
}
