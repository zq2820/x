package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

const (
	templateMinimal = "minimal"
)

// TODO this needs a significant overhaul to support multiple templates
// extensibility etc but works for now
func doinit(tmplName string) {
	switch tmplName {
	case templateMinimal:
		break
	default:
		panic(fmt.Errorf("unknown template %q", tmplName))
	}

	_, err := git.PlainClone("./test", false, &git.CloneOptions{
		URL:      "https://github.com/zq2820/helloworld.git",
		Progress: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}
