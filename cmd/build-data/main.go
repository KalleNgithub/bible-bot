package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: build-data download|embed")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "download":
		if err = download("web", "data/web.tsv"); err == nil {
			err = download("pyharaamattu1933", "data/fi.tsv")
		}
	case "embed":
		err = embedCmd(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
