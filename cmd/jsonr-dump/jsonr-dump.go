// jsonr-dump tool
// Write out JSONR in a line-oriented key-value notation
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

var usage = `Simple tool to dump a JSON obect as flat list of line-oriented key path and value pairs.

  jsonr-dump something.jsonr
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	useExpr := flag.Bool("use-expr", false, "Use expression notation.")
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
		var out string
		if *useExpr {
			out = ast.FmtKeyValue(root, ast.OptionKeyFormatter(ast.FmtKeyAsExpression))
		} else {
			out = ast.FmtKeyValue(root)
		}
		_, err = os.Stdout.Write([]byte(out))
		if err != nil {
			log.Fatal(err)
		}
	}
}
