// Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
// Use of this document is governed by a license found in the LICENSE document.

/*

reactGen is a go generate generator that helps to automate the process of
writing GopherJS React web applications.

For more information see https://github.com/myitcv/x/blob/master/react/_doc/README.md

*/
package main

import (
	"flag"
	"log"
)

const (
	reactGenCmd = "myitcv.io/react/cmd/reactGen"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix(reactGenCmd + ": ")

	flag.Usage = usage
	flag.Parse()

	doinit(*fInit.val)
}
