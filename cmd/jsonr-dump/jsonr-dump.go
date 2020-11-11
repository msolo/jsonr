// json-lines tool
// Write out JSON in a line-oriented notation
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

var usage = `Simple tool to dump a JSON obect as flat list of key paths.

  jsonr-dump something.jsonr
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
		out := ast.JsonLineFmt(root)
		_, err = os.Stdout.Write([]byte(out))
		if err != nil {
			log.Fatal(err)
		}
	}
}
