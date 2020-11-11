// json-fmt tool
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

var usage = `Simple tool to canonically format JSONR.

  jsonr-fmt < something.jsonr > formatted.jsonr
  jsonr-fmt something.jsonr
  jsonr-fmt -w something.jsonr

`

func main() {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	overwrite := flag.Bool("w", false, "write result to source file instead of stdout")
	sortKeys := flag.Bool("s", false, "sort object keys")
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

		_ = f.Close()

		root, err := ast.Parse(string(in))
		if err != nil {
			log.Fatal(err)
		}
		opts := []ast.Option{}
		if *sortKeys {
			opts = append(opts, ast.SortKeys)
		}
		out := ast.JsonrFmt(root, opts...)
		outFile := os.Stdout
		if *overwrite {
			outFile, err = os.OpenFile(p, os.O_TRUNC|os.O_WRONLY, 0664)
			if err != nil {
				log.Fatal(err)
			}
		}
		_, err = outFile.Write([]byte(out))
		if err != nil {
			log.Fatal(err)
		}
	}
}
