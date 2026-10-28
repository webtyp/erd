// Command erd writes the database diagram of a repo's models:
//
//	erd              # docs/diagrams/database.svg for the models under .
//	erd ../billing   # ../billing/docs/diagrams/database.svg
//	erd -o out.svg   # another file
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"webtyp.com/erd/repo"
	"webtyp.com/ormc"
)

func main() {
	out := flag.String("o", "", "file to write (default <dir>/"+repo.File+")")
	flag.Parse()
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	if *out == "" {
		*out = filepath.Join(dir, repo.File)
	}
	written, err := repo.Write(dir, *out)
	switch {
	case errors.Is(err, ormc.ErrNoModelsFound):
		fmt.Fprintln(os.Stderr, "erd: no models under", dir, "— nothing written")
	case err != nil:
		fmt.Fprintln(os.Stderr, "erd:", err)
		os.Exit(1)
	case written:
		fmt.Fprintln(os.Stderr, "erd: wrote", *out)
	default:
		fmt.Fprintln(os.Stderr, "erd:", *out, "unchanged")
	}
}
