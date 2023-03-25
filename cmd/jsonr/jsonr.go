package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/msolo/jsonr/ast"
)

var usage = `Simple tool to convert from JSONR to plain-old JSON.

  jsonr < something.jsonr > something.json
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			os.Exit(1) // Nothing to do and probably an error.
		} else {
			paths = []string{"/dev/stdin"}
		}
	}

	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			log.Fatal(err)
		}
		in, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		root, err := ast.Parse(string(in))
		if err != nil {
			log.Fatal(err)
		}
		out := ast.FmtJson(root)
		_, err = os.Stdout.Write([]byte(out))
		if err != nil {
			log.Fatal(err)
		}
	}
}
