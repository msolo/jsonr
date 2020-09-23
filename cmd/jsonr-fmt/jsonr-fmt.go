// json-fmt tool
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/msolo/jsonr/ast"
)

var usage = `Simple tool to canonically format JSONR.

  jsonr-fmt < something.jsonr > something.json
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	// overwrite := flag.Bool("-w", false, "overwrite files with formatting modifications")
	flag.Parse()

	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	root, err := ast.Parse(string(in))
	if err != nil {
		log.Fatal(err)
	}
	out := ast.JsonrFmt(root)
	_, err = os.Stdout.Write([]byte(out))
	if err != nil {
		log.Fatal(err)
	}

}
