package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
)

const (
	templateMinimal = "minimal"
)

// TODO this needs a significant overhaul to support multiple templates
// extensibility etc but works for now
func doinit(tmplName string, projectName string) {
	switch tmplName {
	case templateMinimal:
		break
	default:
		panic(fmt.Errorf("unknown template %q", tmplName))
	}

	output := fmt.Sprintf("./%s", projectName)
	_, err := git.PlainClone(output, false, &git.CloneOptions{
		URL:      "https://github.com/zq2820/helloworld.git",
		Progress: os.Stdout,
	})
	exec.Command("bash", "-c", fmt.Sprintf("cd %s;rm -rf .git;git init", projectName))

	replaceModule(output, projectName)
	if err != nil {
		panic(err)
	}
}

func replaceModule(dir, projectName string) {
	if items, err := os.ReadDir(dir); err == nil {
		for _, item := range items {
			if item.IsDir() && item.Name() != ".git" {
				replaceModule(path.Join(dir, item.Name()), projectName)
			} else if strings.HasSuffix(item.Name(), ".go") || strings.HasSuffix(item.Name(), ".mod") {
				file := path.Join(dir, item.Name())
				if buffer, err := os.ReadFile(file); err == nil {
					content := string(buffer)
					content = strings.ReplaceAll(content, "github.com/zq2820/helloworld", projectName)
					os.WriteFile(file, []byte(content), 0644)
				}
			}
		}
	}
}
