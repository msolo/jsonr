package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	root, err := ast.Parse(string(in))
	if err != nil {
		log.Fatal(err)
	}
	out := ast.JsonFmt(root)
	_, err = os.Stdout.Write([]byte(out))
	if err != nil {
		log.Fatal(err)
	}
	// var x interface{}
	// err = jsonr.Unmarshal(in, &x)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// out, err := json.MarshalIndent(x, "", "  ")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = os.Stdout.Write(out)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = os.Stdout.WriteString("\n")
}
