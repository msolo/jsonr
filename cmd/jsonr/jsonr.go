package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/msolo/jsonr"
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
	var x interface{}
	err := jsonr.Unmarshal(&x, in)
	if err != nil {
		log.Fatal(err)
	}
	out, err := json.Marshal(x)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stdout.Write(out)
	if err != nil {
		log.Fatal(err)
	}
}
